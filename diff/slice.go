package diff

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	myersdiff "github.com/mb0/diff"
	"github.com/yazgazan/jaydiff/jpath"
)

type slice struct {
	diffs   []Differ
	indices []int
	lhs     interface{}
	rhs     interface{}
}

type sliceMissing struct {
	value interface{}
}

type sliceExcess struct {
	value interface{}
}

type diffData struct {
	lhs       reflect.Value
	rhs       reflect.Value
	visited   *visited
	lastError error
	c         config
}

func (d *diffData) Equal(i, j int) bool {
	diff, err := diff(d.c, d.lhs.Index(i).Interface(), d.rhs.Index(j).Interface(), d.visited)
	if err != nil {
		d.lastError = err
		return false
	}

	return diff.Diff() == Identical
}

func myersToDiff(conf config, lhs, rhs reflect.Value, changes []myersdiff.Change) ([]Differ, []int) {
	res := []Differ{}
	indices := []int{}

	lhsIdx := 0
	rhsIdx := 0
	for _, c := range changes {
		for i := 0; lhsIdx+i < c.A; i++ {
			diff, _ := diff(conf, lhs.Index(lhsIdx+i).Interface(), rhs.Index(rhsIdx+i).Interface(), &visited{})
			res = append(res, diff)
			indices = append(indices, lhsIdx+i)
		}
		lhsIdx = c.A
		rhsIdx = c.B
		for d := 0; d < c.Del; d++ {
			res = append(res, sliceMissing{lhs.Index(lhsIdx + d).Interface()})
			indices = append(indices, lhsIdx+d)
		}
		for i := 0; i < c.Ins; i++ {
			res = append(res, sliceExcess{rhs.Index(rhsIdx + i).Interface()})
			indices = append(indices, lhsIdx+i)
		}
		lhsIdx += c.Del
		rhsIdx += c.Ins
	}

	for lhsIdx < lhs.Len() && rhsIdx < rhs.Len() {
		diff, _ := diff(conf, lhs.Index(lhsIdx).Interface(), rhs.Index(rhsIdx).Interface(), &visited{})
		res = append(res, diff)
		indices = append(indices, lhsIdx)
		lhsIdx++
		rhsIdx++
	}
	return res, indices
}

func newMyersSlice(c config, lhs, rhs interface{}, visited *visited) (Differ, error) {
	var diffs []Differ
	var indices []int

	lhsVal := reflect.ValueOf(lhs)
	rhsVal := reflect.ValueOf(rhs)

	if typesDiffer, err := sliceTypesDiffer(lhs, rhs); err != nil {
		return slice{
			lhs: lhs,
			rhs: rhs,
		}, err
	} else if !typesDiffer {
		dData := diffData{
			lhs:     lhsVal,
			rhs:     rhsVal,
			visited: visited,
			c:       c,
		}
		myers := myersdiff.Diff(lhsVal.Len(), rhsVal.Len(), &dData)

		diffs, indices = myersToDiff(c, lhsVal, rhsVal, myers)
		if dData.lastError != nil {
			return slice{
				lhs:     lhs,
				rhs:     rhs,
				diffs:   diffs,
				indices: indices,
			}, dData.lastError
		}
	}

	return slice{
		lhs:     lhs,
		rhs:     rhs,
		diffs:   diffs,
		indices: indices,
	}, nil
}

func newSlice(c config, lhs, rhs interface{}, visited *visited) (Differ, error) {
	var (
		diffs       []Differ
		indices     []int
		err         error
		typesDiffer bool
	)

	lhsVal := reflect.ValueOf(lhs)
	rhsVal := reflect.ValueOf(rhs)

	if typesDiffer, err = sliceTypesDiffer(lhs, rhs); err != nil {
		return slice{
			lhs: lhs,
			rhs: rhs,
		}, err
	} else if !typesDiffer {
		indices, diffs, err = sliceNewSameTypes(c, lhsVal, rhsVal, visited)
	}

	return slice{
		lhs:     lhs,
		rhs:     rhs,
		diffs:   diffs,
		indices: indices,
	}, err
}

func sliceNewSameTypes(c config, lhsVal, rhsVal reflect.Value, visited *visited) (indices []int, diffs []Differ, err error) {
	nElems := lhsVal.Len()
	if rhsVal.Len() > nElems {
		nElems = rhsVal.Len()
	}

	for i := 0; i < nElems; i++ {
		indices = append(indices, i)
		if i < lhsVal.Len() && i < rhsVal.Len() {
			diff, err := diff(c, lhsVal.Index(i).Interface(), rhsVal.Index(i).Interface(), visited)
			diffs = append(diffs, diff)

			if err != nil {
				return indices, diffs, err
			}
			continue
		}
		if i >= rhsVal.Len() {
			diffs = append(diffs, sliceMissing{lhsVal.Index(i).Interface()})
			continue
		}
		diffs = append(diffs, sliceExcess{rhsVal.Index(i).Interface()})
	}

	return indices, diffs, nil
}

func sliceTypesDiffer(lhs, rhs interface{}) (bool, error) {
	if lhs == nil {
		return true, errInvalidType{Value: lhs, For: "slice"}
	}
	if rhs == nil {
		return true, errInvalidType{Value: rhs, For: "slice"}
	}

	lhsVal := reflect.ValueOf(lhs)
	lhsElType := lhsVal.Type().Elem()
	rhsVal := reflect.ValueOf(rhs)
	rhsElType := rhsVal.Type().Elem()

	return lhsElType.Kind() != rhsElType.Kind(), nil
}

func (s slice) Diff() Type {
	if ok, err := sliceTypesDiffer(s.lhs, s.rhs); err != nil {
		return Invalid
	} else if ok {
		return TypesDiffer
	}

	for _, d := range s.diffs {
		if d.Diff() != Identical {
			return ContentDiffer
		}
	}

	return Identical
}

func (s slice) Strings() []string {
	switch s.Diff() {
	case Identical:
		return []string{fmt.Sprintf("  %T %v", s.lhs, s.lhs)}
	case TypesDiffer:
		return []string{
			fmt.Sprintf("- %T %v", s.lhs, s.lhs),
			fmt.Sprintf("+ %T %v", s.rhs, s.rhs),
		}
	case ContentDiffer:
		var ss = []string{"["}

		for _, d := range s.diffs {
			ss = append(ss, d.Strings()...)
		}

		return append(ss, "]")
	}

	return []string{}
}

func (s slice) StringIndent(key, prefix string, conf Output) string {
	switch s.Diff() {
	case Identical:
		return " " + prefix + key + conf.white(s.lhs)
	case TypesDiffer:
		return "-" + prefix + key + conf.red(s.lhs) + newLineSeparatorString(conf) +
			"+" + prefix + key + conf.green(s.rhs)
	case ContentDiffer:
		var ss = []string{}

		for _, d := range s.diffs {
			s := d.StringIndent("", prefix+conf.Indent, conf)
			if s != "" {
				ss = append(ss, s)
			}
		}

		return strings.Join(
			[]string{
				s.openString(key, prefix, conf),
				strings.Join(ss, newLineSeparatorString(conf)),
				" " + prefix + "]",
			}, "\n",
		)
	}

	return ""
}

func (s *slice) Add(path jpath.Path, i interface{}) error {
	if len(path) == 0 {
		return errors.New("cannot add value to empty path")
	}
	idx, ok := path[0].(jpath.PathIndex)
	if !ok {
		return fmt.Errorf("cannot add value to %T path", path[0])
	}

	return s.add(int(idx), path[1:], i)
}

func (s *slice) add(idx int, path jpath.Path, i interface{}) error {
	if len(s.diffs) <= idx {
		s.addIgnoreUntil(idx)
	}
	d := s.diffs[idx]
	if len(path) > 0 && IsIgnore(d) {
		d = emptyContainer(path[0])
		s.diffs[idx] = d
	}
	if len(path) > 0 {
		diffBuilder, ok := d.(DiffBuilder)
		if !ok {
			return fmt.Errorf("cannot add value to %T(%+v)", d, d)
		}
		return diffBuilder.Add(path, i)
	}

	if len(path) == 0 && IsIgnore(d) {
		s.diffs[idx] = sliceExcess{
			value: i,
		}

		return nil
	}

	if t, ok := d.(sliceMissing); ok {
		s.diffs[idx] = valueDiffers{
			lhs: t.value,
			rhs: i,
		}

		return nil
	}

	return fmt.Errorf("cannot add value to %T(%+v)", d, d)
}

func (s *slice) Delete(path jpath.Path, i interface{}) error {
	if len(path) == 0 {
		return errors.New("cannot delete value from empty path")
	}
	idx, ok := path[0].(jpath.PathIndex)
	if !ok {
		return fmt.Errorf("cannot delete value from %T path", path[0])
	}

	return s.delete(int(idx), path[1:], i)
}

func (s *slice) delete(idx int, path jpath.Path, i interface{}) error {
	if len(s.diffs) <= idx {
		s.addIgnoreUntil(idx)
	}
	d := s.diffs[idx]
	if len(path) > 0 && IsIgnore(d) {
		d = emptyContainer(path[0])
		s.diffs[idx] = d
	}
	if len(path) > 0 {
		diffBuilder, ok := d.(DiffBuilder)
		if !ok {
			return fmt.Errorf("cannot delete value from %T(%+v)", d, d)
		}
		return diffBuilder.Delete(path, i)
	}

	if len(path) == 0 && IsIgnore(d) {
		s.diffs[idx] = sliceMissing{
			value: i,
		}

		return nil
	}

	return fmt.Errorf("cannot delete value to %T(%+v)", d, d)
}

func (s *slice) addIgnoreUntil(idx int) {
	for len(s.diffs) <= idx {
		s.diffs = append(s.diffs, ignore{})
		s.indices = append(s.indices, len(s.diffs)-1)
	}
}

func (s slice) openString(key, prefix string, conf Output) string {
	if conf.JSON {
		return " " + prefix + key + "["
	}

	return " " + prefix + key + conf.typ(s.lhs) + "["
}

func (s slice) Walk(path string, fn WalkFn) error {
	for i, diff := range s.diffs {
		d, err := walk(s, diff, path+"["+strconv.Itoa(s.lhsIndex(i))+"]", fn)
		if err != nil {
			return err
		}
		if d != nil {
			s.diffs[i] = d
		}
	}

	return nil
}

func (s slice) lhsIndex(i int) int {
	return s.indices[i]
}

func (s slice) LHS() interface{} {
	return s.lhs
}

func (s slice) RHS() interface{} {
	return s.rhs
}

func (m sliceMissing) Diff() Type {
	return ContentDiffer
}

func (m sliceMissing) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", m.value, m.value),
	}
}

func (m sliceMissing) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.red(m.value)
}

func (m sliceMissing) LHS() interface{} {
	return m.value
}

func (e sliceExcess) Diff() Type {
	return ContentDiffer
}

func (e sliceExcess) Strings() []string {
	return []string{
		fmt.Sprintf("+ %T %v", e.value, e.value),
	}
}

func (e sliceExcess) StringIndent(key, prefix string, conf Output) string {
	return "+" + prefix + key + conf.green(e.value)
}

func (e sliceExcess) RHS() interface{} {
	return e.value
}

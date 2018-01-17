package diff

import (
	"fmt"
	"reflect"
	"strings"

	myersdiff "github.com/mb0/diff"
)

type slice struct {
	diffs []Differ
	lhs   interface{}
	rhs   interface{}
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

func myersToDiff(conf config, lhs, rhs reflect.Value, changes []myersdiff.Change) []Differ {
	res := []Differ{}

	lhsIdx := 0
	rhsIdx := 0
	for _, c := range changes {
		for i := 0; lhsIdx+i < c.A; i++ {
			diff, _ := diff(conf, lhs.Index(lhsIdx+i).Interface(), rhs.Index(rhsIdx+i).Interface(), &visited{})
			res = append(res, diff)
		}
		lhsIdx = c.A
		rhsIdx = c.B
		for d := 0; d < c.Del; d++ {
			res = append(res, sliceMissing{lhs.Index(lhsIdx + d).Interface()})
		}
		lhsIdx += c.Del
		for i := 0; i < c.Ins; i++ {
			res = append(res, sliceExcess{rhs.Index(rhsIdx + i).Interface()})
		}
		rhsIdx += c.Ins
	}

	for lhsIdx < lhs.Len() && rhsIdx < rhs.Len() {
		diff, _ := diff(conf, lhs.Index(lhsIdx).Interface(), rhs.Index(rhsIdx).Interface(), &visited{})
		res = append(res, diff)
		lhsIdx++
		rhsIdx++
	}
	return res
}

func newMyersSlice(c config, lhs, rhs interface{}, visited *visited) (Differ, error) {
	var diffs []Differ

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

		diffs = myersToDiff(c, lhsVal, rhsVal, myers)
		if dData.lastError != nil {
			return slice{
				lhs:   lhs,
				rhs:   rhs,
				diffs: diffs,
			}, dData.lastError
		}
	}

	return slice{
		lhs:   lhs,
		rhs:   rhs,
		diffs: diffs,
	}, nil
}

func newSlice(c config, lhs, rhs interface{}, visited *visited) (Differ, error) {
	var diffs []Differ

	lhsVal := reflect.ValueOf(lhs)
	rhsVal := reflect.ValueOf(rhs)

	if typesDiffer, err := sliceTypesDiffer(lhs, rhs); err != nil {
		return slice{
			lhs: lhs,
			rhs: rhs,
		}, err
	} else if !typesDiffer {
		nElems := lhsVal.Len()
		if rhsVal.Len() > nElems {
			nElems = rhsVal.Len()
		}

		for i := 0; i < nElems; i++ {
			if i < lhsVal.Len() && i < rhsVal.Len() {
				diff, err := diff(c, lhsVal.Index(i).Interface(), rhsVal.Index(i).Interface(), visited)
				diffs = append(diffs, diff)

				if err != nil {
					return slice{
						lhs:   lhs,
						rhs:   rhs,
						diffs: diffs,
					}, err
				}
				continue
			}
			if i >= rhsVal.Len() {
				diffs = append(diffs, sliceMissing{lhsVal.Index(i).Interface()})
				continue
			}
			diffs = append(diffs, sliceExcess{rhsVal.Index(i).Interface()})
		}
	}

	return slice{
		lhs:   lhs,
		rhs:   rhs,
		diffs: diffs,
	}, nil
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
		return "-" + prefix + key + conf.red(s.lhs) + "\n" +
			"+" + prefix + key + conf.green(s.rhs)
	case ContentDiffer:
		var ss = []string{" " + prefix + key + conf.typ(s.lhs) + "["}

		for _, d := range s.diffs {
			s := d.StringIndent("", prefix+conf.Indent, conf)
			if s != "" {
				ss = append(ss, s)
			}
		}

		return strings.Join(append(ss, " "+prefix+"]"), "\n")
	}

	return ""
}

func (s slice) Walk(path string, fn WalkFn) error {
	for i, diff := range s.diffs {
		d, err := walk(s, diff, path+"[]", fn)
		if err != nil {
			return err
		}
		if d != nil {
			s.diffs[i] = d
		}
	}

	return nil
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

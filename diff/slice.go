package diff

import (
	"fmt"
	"reflect"
	"strings"
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

func newSlice(lhs, rhs interface{}) (Differ, error) {
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
				diff, err := Diff(lhsVal.Index(i).Interface(), rhsVal.Index(i).Interface())
				if diff.Diff() != Identical {
				}
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
		return true, ErrInvalidType{Value: lhs, For: "slice"}
	}
	if rhs == nil {
		return true, ErrInvalidType{Value: rhs, For: "slice"}
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

func (m sliceMissing) Diff() Type {
	return ContentDiffer
}

func (m sliceMissing) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", m.value, m.value),
	}
}

func (m sliceMissing) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.red(m.value) +
		"\n+" + prefix + key
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
	return "-" + prefix + key +
		"\n+" + prefix + key + conf.green(e.value)
}

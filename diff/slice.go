package diff

import (
	"fmt"
	"reflect"
	"strings"
)

type Slice struct {
	Diffs []Differ
	LHS   interface{}
	RHS   interface{}
}

type SliceMissing struct {
	Value interface{}
}

type SliceExcess struct {
	Value interface{}
}

func NewSlice(lhs, rhs interface{}) (*Slice, error) {
	var diffs []Differ

	lhsVal := reflect.ValueOf(lhs)
	rhsVal := reflect.ValueOf(rhs)

	if typesDiffer, err := sliceTypesDiffer(lhs, rhs); err != nil {
		return &Slice{
			LHS: lhs,
			RHS: rhs,
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
					return &Slice{
						LHS:   lhs,
						RHS:   rhs,
						Diffs: diffs,
					}, err
				}
				continue
			}
			if i >= rhsVal.Len() {
				missing := &SliceMissing{lhsVal.Index(i).Interface()}
				diffs = append(diffs, missing)
				continue
			}
			excess := &SliceExcess{rhsVal.Index(i).Interface()}
			diffs = append(diffs, excess)
		}
	}

	return &Slice{
		LHS:   lhs,
		RHS:   rhs,
		Diffs: diffs,
	}, nil
}

func sliceTypesDiffer(lhs, rhs interface{}) (bool, error) {
	if lhs == nil {
		return true, InvalidType{Value: lhs, For: "slice"}
	}
	if rhs == nil {
		return true, InvalidType{Value: rhs, For: "slice"}
	}

	lhsVal := reflect.ValueOf(lhs)
	lhsElType := lhsVal.Type().Elem()
	rhsVal := reflect.ValueOf(rhs)
	rhsElType := rhsVal.Type().Elem()

	return lhsElType.Kind() != rhsElType.Kind(), nil
}

func (s Slice) Diff() Type {
	if ok, err := sliceTypesDiffer(s.LHS, s.RHS); err != nil {
		return Invalid
	} else if ok {
		return TypesDiffer
	}

	for _, d := range s.Diffs {
		if d.Diff() != Identical {
			return ContentDiffer
		}
	}

	return Identical
}

func (s Slice) Strings() []string {
	switch s.Diff() {
	case Identical:
		return []string{fmt.Sprintf("  %T %v", s.LHS, s.LHS)}
	case TypesDiffer:
		return []string{
			fmt.Sprintf("- %T %v", s.LHS, s.LHS),
			fmt.Sprintf("+ %T %v", s.RHS, s.RHS),
		}
	case ContentDiffer:
		var ss = []string{"["}

		for _, d := range s.Diffs {
			ss = append(ss, d.Strings()...)
		}

		return append(ss, "]")
	}

	return []string{}
}

func (s Slice) StringIndent(key, prefix string, conf Output) string {
	switch s.Diff() {
	case Identical:
		return " " + prefix + key + conf.White(s.LHS)
	case TypesDiffer:
		return "-" + prefix + key + conf.Red(s.LHS) + "\n" +
			"+" + prefix + key + conf.Green(s.RHS)
	case ContentDiffer:
		var ss = []string{" " + prefix + key + conf.Type(s.LHS) + "["}

		for _, d := range s.Diffs {
			s := d.StringIndent("", prefix+conf.Indent, conf)
			if s != "" {
				ss = append(ss, s)
			}
		}

		return strings.Join(append(ss, " "+prefix+"]"), "\n")
	}

	return ""
}

func (s Slice) Walk(path string, fn WalkFn) error {
	for _, diff := range s.Diffs {
		err := walk(s, diff, path+"[]", fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m SliceMissing) Diff() Type {
	return ContentDiffer
}

func (m SliceMissing) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", m.Value, m.Value),
	}
}

func (m SliceMissing) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.Red(m.Value) +
		"\n+" + prefix + key
}

func (e SliceExcess) Diff() Type {
	return ContentDiffer
}

func (e SliceExcess) Strings() []string {
	return []string{
		fmt.Sprintf("+ %T %v", e.Value, e.Value),
	}
}

func (e SliceExcess) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key +
		"\n+" + prefix + key + conf.Green(e.Value)
}

package diff

import (
	"fmt"
	"reflect"
	"strings"
)

type Slice struct {
	Type
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
	var Type = Identical
	var diffs []Differ

	lhsVal := reflect.ValueOf(lhs)
	lhsElType := lhsVal.Type().Elem()
	rhsVal := reflect.ValueOf(rhs)
	rhsElType := rhsVal.Type().Elem()

	if lhsElType.Kind() != rhsElType.Kind() {
		Type = TypesDiffer
	} else {
		nElems := lhsVal.Len()
		if rhsVal.Len() > nElems {
			nElems = rhsVal.Len()
		}

		for i := 0; i < nElems; i++ {
			if i < lhsVal.Len() && i < rhsVal.Len() {
				diff, err := Diff(lhsVal.Index(i).Interface(), rhsVal.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				if diff.Diff() != Identical {
					Type = ContentDiffer
				}
				diffs = append(diffs, diff)
				continue
			}
			Type = ContentDiffer
			if i >= rhsVal.Len() {
				diffs = append(diffs, &SliceMissing{lhsVal.Index(i).Interface()})
				continue
			}
			diffs = append(diffs, &SliceExcess{rhsVal.Index(i).Interface()})
		}
	}

	return &Slice{
		Type:  Type,
		LHS:   lhs,
		RHS:   rhs,
		Diffs: diffs,
	}, nil
}

func (s Slice) Diff() Type {
	return s.Type
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
		return prefix + key + conf.White(s.LHS)
	case TypesDiffer:
		return "-" + prefix + key + conf.Red(s.LHS) + "\n" +
			"+" + prefix + key + conf.Green(s.RHS)
	case ContentDiffer:
		var ss = []string{prefix + key + conf.Type(s.LHS) + "["}

		for _, d := range s.Diffs {
			ss = append(ss, d.StringIndent("", prefix+conf.Indent, conf))
		}

		return strings.Join(append(ss, prefix+"]"), "\n")
	}

	return ""
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

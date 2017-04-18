package diff

import (
	"fmt"
	"reflect"
)

type Scalar struct {
	LHS interface{}
	RHS interface{}
}

func (s Scalar) Diff() Type {
	lhsVal := reflect.ValueOf(s.LHS)
	rhsVal := reflect.ValueOf(s.RHS)

	if lhsVal.Kind() != rhsVal.Kind() {
		return TypesDiffer
	}
	if s.LHS != s.RHS {
		return ContentDiffer
	}

	return Identical
}

func (s Scalar) Strings() []string {
	if s.Diff() == Identical {
		return []string{
			fmt.Sprintf("  %T %v", s.LHS, s.LHS),
		}
	}

	return []string{
		fmt.Sprintf("- %T %v", s.LHS, s.LHS),
		fmt.Sprintf("+ %T %v", s.RHS, s.RHS),
	}
}

func (s Scalar) StringIndent(key, prefix string, conf Output) string {
	if s.Diff() == Identical {
		return " " + prefix + key + conf.White(s.LHS)
	}

	return "-" + prefix + key + conf.Red(s.LHS) + "\n" +
		"+" + prefix + key + conf.Green(s.RHS)
}

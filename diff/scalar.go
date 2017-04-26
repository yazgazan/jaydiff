package diff

import (
	"fmt"
	"reflect"
)

type scalar struct {
	lhs interface{}
	rhs interface{}
}

func IsScalar(d Differ) bool {
	_, ok := d.(scalar)

	return ok
}

func (s scalar) Diff() Type {
	lhsVal := reflect.ValueOf(s.lhs)
	rhsVal := reflect.ValueOf(s.rhs)

	if lhsVal.Kind() != rhsVal.Kind() {
		return TypesDiffer
	}
	if s.lhs != s.rhs {
		return ContentDiffer
	}

	return Identical
}

func (s scalar) Strings() []string {
	if s.Diff() == Identical {
		return []string{
			fmt.Sprintf("  %T %v", s.lhs, s.lhs),
		}
	}

	return []string{
		fmt.Sprintf("- %T %v", s.lhs, s.lhs),
		fmt.Sprintf("+ %T %v", s.rhs, s.rhs),
	}
}

func (s scalar) StringIndent(key, prefix string, conf Output) string {
	if s.Diff() == Identical {
		return " " + prefix + key + conf.white(s.lhs)
	}

	return "-" + prefix + key + conf.red(s.lhs) + "\n" +
		"+" + prefix + key + conf.green(s.rhs)
}

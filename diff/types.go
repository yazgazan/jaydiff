package diff

import (
	"fmt"
)

type types struct {
	lhs interface{}
	rhs interface{}
}

func (t types) Diff() Type {
	return TypesDiffer
}

func (t types) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", t.lhs, t.lhs),
		fmt.Sprintf("+ %T %v", t.rhs, t.rhs),
	}
}

func (t types) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.red(t.lhs) + newLineSeparatorString(conf) +
		"+" + prefix + key + conf.green(t.rhs)
}

func (t types) LHS() interface{} {
	return t.lhs
}

func (t types) RHS() interface{} {
	return t.rhs
}

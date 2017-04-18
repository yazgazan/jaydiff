package diff

import (
	"fmt"
)

type Types struct {
	LHS interface{}
	RHS interface{}
}

func (t Types) Diff() Type {
	return TypesDiffer
}

func (t Types) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", t.LHS, t.LHS),
		fmt.Sprintf("+ %T %v", t.RHS, t.RHS),
	}
}

func (t Types) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.Red(t.LHS) + "\n" +
		"+" + prefix + key + conf.Green(t.RHS)
}

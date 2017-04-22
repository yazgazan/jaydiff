package diff

import (
	"reflect"
)

type Unsupported struct {
	LHS reflect.Type
	RHS reflect.Type
}

func (e Unsupported) Error() string {
	return "unsupported types: " + e.LHS.String() + ", " + e.RHS.String()
}

package diff

import (
	"fmt"
	"reflect"
)

type Unsupported struct {
	LHS reflect.Type
	RHS reflect.Type
}

func (e Unsupported) Error() string {
	return "unsupported types: " + e.LHS.String() + ", " + e.RHS.String()
}

type InvalidType struct {
	Value interface{}
	For   string
}

func (e InvalidType) Error() string {
	return fmt.Sprintf("%T is not a valid type for %s", e.Value, e.For)
}

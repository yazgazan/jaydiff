package diff

import (
	"fmt"
	"reflect"
)

type ErrUnsupported struct {
	LHS reflect.Type
	RHS reflect.Type
}

func (e ErrUnsupported) Error() string {
	return "unsupported types: " + e.LHS.String() + ", " + e.RHS.String()
}

type ErrInvalidType struct {
	Value interface{}
	For   string
}

func (e ErrInvalidType) Error() string {
	return fmt.Sprintf("%T is not a valid type for %s", e.Value, e.For)
}

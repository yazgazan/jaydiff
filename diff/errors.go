package diff

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrUnsupported is returned when an unsupported type is encountered (func, struct ...).
type ErrUnsupported struct {
	LHS reflect.Type
	RHS reflect.Type
}

func (e ErrUnsupported) Error() string {
	return "unsupported types: " + e.LHS.String() + ", " + e.RHS.String()
}

type errInvalidType struct {
	Value interface{}
	For   string
}

func (e errInvalidType) Error() string {
	return fmt.Sprintf("%T is not a valid type for %s", e.Value, e.For)
}

// ErrCyclic is returned when one of the compared values contain circular references
var ErrCyclic = errors.New("circular references not supported")

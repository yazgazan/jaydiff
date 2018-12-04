package diff

import (
	"errors"
	"reflect"
	"unsafe"
)

// ErrCyclic is returned when one of the compared values contain circular references
var ErrCyclic = errors.New("circular references not supported")

// visited is used to detect cyclic structures.
// It is not safe for concurent use.
type visited struct {
	lhs []unsafe.Pointer
	rhs []unsafe.Pointer
}

// add will try to add the value's pointers to the list. It will return an error
// if the value is already in the list.
// visited.remove should be called whether an error occured or not.
func (v *visited) add(lhs, rhs reflect.Value) error {
	if canAddr(lhs) && !isEmptyMapOrSlice(lhs) {
		if inPointers(v.lhs, unsafe.Pointer(lhs.Pointer())) {
			return ErrCyclic
		}
		v.lhs = append(v.lhs, unsafe.Pointer(lhs.Pointer()))
	}
	if canAddr(rhs) && !isEmptyMapOrSlice(rhs) {
		if inPointers(v.rhs, unsafe.Pointer(rhs.Pointer())) {
			return ErrCyclic
		}
		v.rhs = append(v.rhs, unsafe.Pointer(rhs.Pointer()))
	}

	return nil
}

func (v *visited) remove(lhs, rhs reflect.Value) {
	if canAddr(lhs) && isLastPointer(v.lhs, unsafe.Pointer(lhs.Pointer())) {
		v.lhs = v.lhs[:len(v.lhs)-1]
	}

	if canAddr(rhs) && isLastPointer(v.rhs, unsafe.Pointer(rhs.Pointer())) {
		v.rhs = v.rhs[:len(v.rhs)-1]
	}
}

func isLastPointer(pointers []unsafe.Pointer, val unsafe.Pointer) bool {
	if len(pointers) == 0 {
		return false
	}

	return pointers[len(pointers)-1] == val
}

func isEmptyMapOrSlice(v reflect.Value) bool {
	// we don't want to include empty slices and maps in our cyclic check, since these are not problematic
	return (v.Kind() == reflect.Slice || v.Kind() == reflect.Map) && v.Len() == 0
}

func inPointers(pointers []unsafe.Pointer, val unsafe.Pointer) bool {
	for _, lhs := range pointers {
		if lhs == val {
			return true
		}
	}

	return false
}

func canAddr(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map:
		fallthrough
	case reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return true
	}

	return false
}

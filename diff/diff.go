// Package diff provides utilities to generate deep, walkable diffs of maps and slices
package diff

import (
	"reflect"
)

// Type is used to specify the nature of the difference
type Type int

const (
	// TypesDiffer is used when two values cannot be compared due to types differences
	// (for example: comparing a slice to an int)
	TypesDiffer Type = iota
	// ContentDiffer is used when the types matches but the content differs
	ContentDiffer
	// Identical is used when both the type and the content match.
	Identical
	// Invalid is used when calling Diff() on an inproperly constructed node
	Invalid
)

// Differ is implemented by all nodes in a diff-tree.
type Differ interface {
	Diff() Type
	Strings() []string
	StringIndent(key, prefix string, conf Output) string
}

type diffFn func(c config, lhs, rhs interface{}, visited *visited) (Differ, error)

// Diff generates a tree representing differences and similarities between two objects.
//
// Diff supports maps, slices and scalars (comparables types such as int, string, etc ...).
// When an unsupported type is encountered, an ErrUnsupported error is returned.
func Diff(lhs, rhs interface{}, opts ...ConfigOpt) (Differ, error) {
	c := defaultConfig()
	for _, opt := range opts {
		c = opt(c)
	}

	return diff(c, lhs, rhs, &visited{})
}

func diff(c config, lhs, rhs interface{}, visited *visited) (Differ, error) {
	lhsVal := reflect.ValueOf(lhs)
	rhsVal := reflect.ValueOf(rhs)

	if d, ok := nilCheck(lhs, rhs); ok {
		return d, nil
	}
	if err := visited.add(lhsVal, rhsVal); err != nil {
		return types{lhs, rhs}, ErrCyclic
	}
	if lhsVal.Type().Comparable() && rhsVal.Type().Comparable() {
		return scalar{lhs, rhs}, nil
	}
	if lhsVal.Kind() != rhsVal.Kind() {
		return types{lhs, rhs}, nil
	}

	if lhsVal.Kind() == reflect.Slice {
		return c.sliceFn(c, lhs, rhs, visited)
	}
	if lhsVal.Kind() == reflect.Map {
		return newMap(c, lhs, rhs, visited)
	}

	return types{lhs, rhs}, &ErrUnsupported{lhsVal.Type(), rhsVal.Type()}
}

func nilCheck(lhs, rhs interface{}) (Differ, bool) {
	if lhs == nil && rhs == nil {
		return scalar{lhs, rhs}, true
	}
	if lhs == nil || rhs == nil {
		return types{lhs, rhs}, true
	}

	return nil, false
}

func (t Type) String() string {
	switch t {
	case Identical:
		return "identical"
	case ContentDiffer:
		return "content differ"
	case TypesDiffer:
		return "types differ"
	}

	return "invalid type"
}

// IsExcess returns true if d represent value missing from the LHS (in a map or an array)
func IsExcess(d Differ) bool {
	switch d.(type) {
	default:
		return false
	case mapExcess:
		return true
	case sliceExcess:
		return true
	}
}

// IsMissing returns true if d represent value missing from the RHS (in a map or an array)
func IsMissing(d Differ) bool {
	switch d.(type) {
	default:
		return false
	case mapMissing:
		return true
	case sliceMissing:
		return true
	}
}

// IsScalar returns true of d is a diff between two values that can be compared (int, float64, string, ...)
func IsScalar(d Differ) bool {
	_, ok := d.(scalar)

	return ok
}

// IsTypes returns true if d is a diff between two values of different types that cannot be compared
func IsTypes(d Differ) bool {
	_, ok := d.(types)

	return ok
}

// IsIgnore returns true if d is a diff created by NewIgnore
func IsIgnore(d Differ) bool {
	_, ok := d.(ignore)

	return ok
}

// IsMap returns true if d is a diff between towo maps
func IsMap(d Differ) bool {
	_, ok := d.(mapDiff)

	return ok
}

// IsSlice returns true if d is a diff between towo slices
func IsSlice(d Differ) bool {
	_, ok := d.(slice)

	return ok
}

type lhsGetter interface {
	LHS() interface{}
}

type rhsGetter interface {
	RHS() interface{}
}

// LHS returns the lhs value associated with the Differ.
func LHS(d Differ) (interface{}, error) {
	if lhs, ok := d.(lhsGetter); ok {
		return lhs.LHS(), nil
	}

	return nil, ErrLHSNotSupported{Diff: d}
}

// RHS returns the rhs value associated with the Differ.
func RHS(d Differ) (interface{}, error) {
	if rhs, ok := d.(rhsGetter); ok {
		return rhs.RHS(), nil
	}

	return nil, ErrRHSNotSupported{Diff: d}
}

type visited struct {
	lhs []uintptr
	rhs []uintptr
}

func (v *visited) add(lhs, rhs reflect.Value) error {
	if canAddr(lhs) && !isEmptyMapOrSlice(lhs) {
		if inPointers(v.lhs, lhs) {
			return ErrCyclic
		}
		v.lhs = append(v.lhs, lhs.Pointer())
	}
	if canAddr(rhs) && !isEmptyMapOrSlice(rhs) {
		if inPointers(v.rhs, rhs) {
			return ErrCyclic
		}
		v.rhs = append(v.rhs, rhs.Pointer())
	}

	return nil
}

func isEmptyMapOrSlice(v reflect.Value) bool {
	// we don't want to include empty slices and maps in our cyclic check, since these are not problematic
	return (v.Kind() == reflect.Slice || v.Kind() == reflect.Map) && v.Len() == 0
}

func inPointers(pointers []uintptr, val reflect.Value) bool {
	for _, lhs := range pointers {
		if lhs == val.Pointer() {
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

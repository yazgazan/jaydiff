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
// Diff supports maps, slices, Stream and scalars (comparables types such as int, string, etc ...).
// When an unsupported type is encountered, an ErrUnsupported error is returned.
func Diff(lhs, rhs interface{}, opts ...ConfigOpt) (Differ, error) {
	c := defaultConfig()
	for _, opt := range opts {
		c = opt(c)
	}

	return diff(c, lhs, rhs, &visited{})
}

func diff(c config, lhs, rhs interface{}, visited *visited) (Differ, error) {
	lhsVal, lhs := indirectValueOf(lhs)
	rhsVal, rhs := indirectValueOf(rhs)

	if d, ok := nilCheck(lhsVal, rhsVal, lhs, rhs); ok {
		return d, nil
	}

	err := visited.push(lhsVal, rhsVal)
	defer visited.pop(lhsVal, rhsVal)
	if err != nil {
		return types{lhs, rhs}, ErrCyclic
	}

	return diffValues(c, lhsVal, rhsVal, visited)
}

func diffValues(c config, lhs, rhs reflect.Value, visited *visited) (Differ, error) {
	if valueIsStream(lhs) && valueIsStream(rhs) {
		return newStream(c, lhs.Interface(), rhs.Interface(), visited)
	}

	if valueIsScalar(lhs) && valueIsScalar(rhs) {
		return scalar{lhs.Interface(), rhs.Interface()}, nil
	}
	if lhs.Kind() != rhs.Kind() {
		return types{lhs.Interface(), rhs.Interface()}, nil
	}

	switch lhs.Kind() {
	case reflect.Slice, reflect.Array:
		return c.sliceFn(c, lhs.Interface(), rhs.Interface(), visited)
	case reflect.Map:
		return newMap(c, lhs.Interface(), rhs.Interface(), visited)
	case reflect.Struct:
		return newStruct(c, lhs.Interface(), rhs.Interface(), visited)
	}

	return types{lhs.Interface(), rhs.Interface()}, &ErrUnsupported{lhs.Type(), rhs.Type()}
}

func indirectValueOf(i interface{}) (reflect.Value, interface{}) {
	if _, ok := i.(Stream); ok {
		return reflect.ValueOf(i), i
	}

	v := reflect.Indirect(reflect.ValueOf(i))
	if !v.IsValid() || !v.CanInterface() {
		return reflect.ValueOf(i), i
	}

	return v, v.Interface()
}

func valueIsStream(v reflect.Value) bool {
	if !v.IsValid() || !v.CanInterface() {
		return false
	}

	_, ok := v.Interface().(Stream)

	return ok
}

func valueIsScalar(v reflect.Value) bool {
	switch v.Kind() {
	default:
		return v.Type().Comparable()
	case reflect.Struct, reflect.Array, reflect.Ptr, reflect.Chan:
		return false
	}
}

func nilCheck(lhsVal, rhsVal reflect.Value, lhs, rhs interface{}) (Differ, bool) {
	lhsNil := lhs == nil || (lhsVal.Kind() == reflect.Ptr && lhsVal.IsNil())
	rhsNil := rhs == nil || (rhsVal.Kind() == reflect.Ptr && rhsVal.IsNil())

	if lhsNil && rhsNil {
		return scalar{lhsVal, rhsVal}, true
	}
	if lhsNil || rhsNil {
		return types{lhsVal, rhsVal}, true
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
	case mapExcess, sliceExcess, streamExcess:
		return true
	}
}

// IsMissing returns true if d represent value missing from the RHS (in a map or an array)
func IsMissing(d Differ) bool {
	switch d.(type) {
	default:
		return false
	case mapMissing, sliceMissing, streamMissing:
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

// IsStream returns true if d is a diff between towo slices
func IsStream(d Differ) bool {
	_, ok := d.(stream)

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

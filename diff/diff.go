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

// Diff generates a tree representing differences and similarities between two objects.
//
// Diff supports maps, slices and scalars (comparables types such as int, string, etc ...).
// When an unsupported type is encountered, an ErrUnsupported error is returned.
//
// BUG(yazgazan): An infinite recursion is possible if the lhs and/or rhs objects are cyclic
func Diff(lhs, rhs interface{}) (Differ, error) {
	lhsVal := reflect.ValueOf(lhs)
	rhsVal := reflect.ValueOf(rhs)

	if lhs == nil && rhs == nil {
		return scalar{lhs, rhs}, nil
	}
	if lhs == nil || rhs == nil {
		return types{lhs, rhs}, nil
	}
	if lhsVal.Type().Comparable() && rhsVal.Type().Comparable() {
		return scalar{lhs, rhs}, nil
	}
	if lhsVal.Kind() != rhsVal.Kind() {
		return types{lhs, rhs}, nil
	}
	if lhsVal.Kind() == reflect.Slice {
		return newSlice(lhs, rhs)
	}
	if lhsVal.Kind() == reflect.Map {
		return newMap(lhs, rhs)
	}

	return types{lhs, rhs}, &ErrUnsupported{lhsVal.Type(), rhsVal.Type()}
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

// IsExcess can be used in a WalkFn to find values missing from the LHS
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

// IsMissing can be used in a WalkFn to find values missing from the RHS
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

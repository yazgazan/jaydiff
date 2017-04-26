package diff

import (
	"reflect"
)

type Type int

const (
	TypesDiffer Type = iota
	ContentDiffer
	Identical
	Invalid
)

type Differ interface {
	Diff() Type
	Strings() []string
	StringIndent(key, prefix string, conf Output) string
}

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

// IsExcess returns true if the provided Differ refers to a value missing in LHS
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

// IsMissing returns true if the provided Differ refers to a value missing in LHS
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

// Code generated by "stringer -type PathKind"; DO NOT EDIT.

package jpath

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PathKindIndex-0]
	_ = x[PathKindKey-1]
}

const _PathKind_name = "PathKindIndexPathKindKey"

var _PathKind_index = [...]uint8{0, 13, 24}

func (i PathKind) String() string {
	if i < 0 || i >= PathKind(len(_PathKind_index)-1) {
		return "PathKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _PathKind_name[_PathKind_index[i]:_PathKind_index[i+1]]
}

package jpath

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrInvalidPath      = errors.New("invalid path")
	ErrNotSlice         = errors.New("not a slice")
	ErrNotMap           = errors.New("not a map")
	ErrOutOfBounds      = errors.New("index out of bounds")
	ErrInvalidInterface = errors.New("cannot get interface of value")
	ErrNil              = errors.New("cannot get index or key of nil")
	ErrKeyType          = errors.New("cannot handle this type of key")
)

// StripIndices removes the characters in between brackets in a json path
func StripIndices(path string) string {
	r := make([]byte, 0, len(path))

	i := 0
	start := -1
	escaped := false
	for i < len(path) {
		if escaped && isUnescapedQuote(path[i-1:i+1]) {
			escaped = false
		} else if path[i] == '"' {
			escaped = true
		}
		if escaped {
			r = append(r, path[i])
			i++
			continue
		}
		if start != -1 && path[i] == ']' {
			start = -1
		}
		if start == -1 {
			r = append(r, path[i])
		}
		if path[i] == '[' {
			start = i
		}
		i++
	}

	return string(r)
}

func isUnescapedQuote(s string) bool {
	return s[0] != '\\' && s[1] == '"'
}

// HasSuffix tests whether the string s ends with suffix, ignoring indices in brackets.
func HasSuffix(s, suffix string) bool {
	stripped := StripIndices(s)

	return strings.HasSuffix(stripped, suffix)
}

// EscapeKey produces a string from a key (int, string, etc), enclosing it in quotes when necessary.
// This should be used when generating json paths to avoid ambiguity from keys such as `foo.bar`.
func EscapeKey(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	if s != "" && !strings.ContainsAny(s, "[].\"") {
		return s
	}
	return fmt.Sprintf("%q", s)
}

func Split(path string) (head, tail string) {
	if path == "" {
		return "", ""
	}

	if path[0] == '[' {
		for i := 1; i < len(path); i++ {
			if path[i] == ']' {
				return path[0 : i+1], path[i+1:]
			}
		}
	}
	// Skipping first character as we espect the path to start with a dot.
	for i := 1; i < len(path); i++ {
		if path[i] == '.' || path[i] == '[' {
			return path[0:i], path[i:]
		}
	}

	// tail not found, returning the path as head
	return path, ""
}

func getKey(s string, kind reflect.Kind) (reflect.Value, error) {
	switch kind {
	default:
		return reflect.Value{}, ErrKeyType
	case reflect.Int:
		i, err := strconv.Atoi(s)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i), nil
	case reflect.String:
		return reflect.ValueOf(s), nil
	}
}

func ExecutePath(path string, i interface{}) (interface{}, error) {
	// TODO(yazgazan): better errors
	head, tail := Split(path)
	if head == "" {
		return i, nil
	}

	v := reflect.ValueOf(i)

	switch head[0] {
	default:
		return nil, ErrInvalidPath
	case '[':
		return executeSlice(head, tail, v)
	case '.':
		return executeMap(head, tail, v)
	}
}

func executeSlice(head, tail string, v reflect.Value) (interface{}, error) {
	if v.Kind() != reflect.Slice {
		return nil, ErrNotSlice
	}
	if v.IsNil() {
		return nil, ErrNil
	}
	if head[len(head)-1] != ']' {
		return nil, ErrInvalidPath
	}
	index, err := strconv.Atoi(head[1 : len(head)-1])
	if err != nil {
		return nil, err
	}
	if index >= v.Len() {
		return nil, ErrOutOfBounds
	}
	val := v.Index(index)
	if !val.CanInterface() {
		return nil, ErrInvalidInterface
	}
	return ExecutePath(tail, val.Interface())
}

func executeMap(head, tail string, v reflect.Value) (interface{}, error) {
	if v.Kind() != reflect.Map {
		return nil, ErrNotMap
	}
	keyStr := head[1:]
	if keyStr == "" {
		return nil, ErrInvalidPath
	}
	key, err := getKey(keyStr, v.Type().Key().Kind())
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, ErrNil
	}
	val := v.MapIndex(key)
	if !val.CanInterface() {
		return nil, ErrInvalidInterface
	}
	return ExecutePath(tail, val.Interface())
}

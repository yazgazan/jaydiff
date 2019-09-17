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

const keySpecials = "[].\":"

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
		s = fmt.Sprintf("%v", v)
	}
	if s != "" && !strings.ContainsAny(s, keySpecials) {
		return s
	}
	return strconv.Quote(s)
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
	pp, _, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	return executePath(pp, i)
}

func executePath(path []pathPart, i interface{}) (interface{}, error) {
	if len(path) == 0 {
		return i, nil
	}
	head, tail := path[0], path[1:]

	v := reflect.ValueOf(i)

	switch head.Kind() {
	default:
		return nil, ErrInvalidPath
	case pathKindIndex:
		return executeSlice(head.(pathIndex), tail, v)
	case pathKindKey:
		return executeMap(head.(pathKey), tail, v)
	}
}

func executeSlice(idx pathIndex, tail []pathPart, v reflect.Value) (interface{}, error) {
	if v.Kind() != reflect.Slice {
		return nil, ErrNotSlice
	}
	if v.IsNil() {
		return nil, ErrNil
	}

	if int(idx) >= v.Len() {
		return nil, ErrOutOfBounds
	}
	val := v.Index(int(idx))
	if !val.CanInterface() {
		return nil, ErrInvalidInterface
	}
	return executePath(tail, val.Interface())
}

func executeMap(keyStr pathKey, tail []pathPart, v reflect.Value) (interface{}, error) {
	if v.Kind() != reflect.Map {
		return nil, ErrNotMap
	}

	key, err := getKey(string(keyStr), v.Type().Key().Kind())
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
	return executePath(tail, val.Interface())
}

type pathKind int

const (
	pathKindUnknown pathKind = iota
	pathKindIndex
	pathKindKey
)

type pathPart interface {
	Kind() pathKind
	String() string
}

type pathIndex int

func (i pathIndex) Kind() pathKind {
	return pathKindIndex
}

func (i pathIndex) String() string {
	return "[" + strconv.Itoa(int(i)) + "]"
}

type pathKey string

func (k pathKey) Kind() pathKind {
	return pathKindKey
}

func (k pathKey) String() string {
	return "." + EscapeKey(string(k))
}

func parsePath(path string) ([]pathPart, int, error) {
	var (
		part pathPart
		i    int
		err  error
	)

	if path == "" {
		return nil, 0, nil
	}

	switch path[0] {
	default:
		return nil, 0, nil
	case '.':
		part, i, err = parseKey(path)
	case '[':
		part, i, err = parseIndex(path)
	}
	if err != nil {
		return nil, 0, err
	}
	path = path[i:]

	parts, j, err := parsePath(path)
	if err != nil {
		return []pathPart{part}, i + j, err
	}

	return append([]pathPart{part}, parts...), i + j, nil
}

func parseKey(path string) (pathKey, int, error) {
	i := 1
	if len(path) <= 1 {
		return "", i, errors.New("expected key after '.'")
	}

	if path[i] == '"' {
		return parseQuotedKey(path)
	}

	for ; i < len(path) && !strings.ContainsAny(keySpecials, path[i:i+1]); i++ {
	}

	return pathKey(path[1:i]), i, nil
}

func parseQuotedKey(path string) (pathKey, int, error) {
	i := 1

	i++
	escaping := false
	for ; i < len(path); i++ {
		if escaping {
			escaping = false
			continue
		}
		if path[i] == '\\' {
			escaping = true
			continue
		}
		if path[i] == '"' {
			break
		}
	}
	if escaping || i >= len(path) {
		return "", i, errors.New("malformed key")
	}
	s, err := strconv.Unquote(path[1 : i+1])
	if err != nil {
		fmt.Println(path[1 : i+1])
		return "", i, err
	}

	return pathKey(s), i + 1, nil
}

func parseIndex(path string) (pathIndex, int, error) {
	i := 1
	if len(path) < 3 {
		return 0, i, errors.New("expected index to be of the form [number]")
	}

	for ; i < len(path) && path[i] != ']'; i++ {
	}

	if i == len(path) {
		return 0, i, errors.New("expected index to be of the form [number]")
	}

	n, err := strconv.ParseInt(path[1:i], 10, strconv.IntSize)
	if err != nil {
		return 0, i, err
	}

	return pathIndex(n), i + 1, nil
}

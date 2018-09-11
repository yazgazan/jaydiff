package jpath

import (
	"fmt"
	"strings"
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

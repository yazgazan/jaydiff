package jpath

import (
	"reflect"
	"testing"
)

func TestStripIndices(t *testing.T) {
	for _, test := range []struct {
		In       string
		Expected string
	}{
		{".foo.bar", ".foo.bar"},
		{"", ""},
		{".foo[].bar", ".foo[].bar"},
		{"[].bar", "[].bar"},
		{".foo[]", ".foo[]"},
		{".foo[341].bar", ".foo[].bar"},
		{"[1].bar", "[].bar"},
		{".foo[22]", ".foo[]"},
		{`."f[oo]"[22]`, `."f[oo]"[]`},
		{`."f[00]"[22]`, `."f[00]"[]`},
		{`."f[0\"]"[22]`, `."f[0\"]"[]`},
		{`."foo[42]`, `."foo[42]`},
	} {
		got := StripIndices(test.In)
		if got != test.Expected {
			t.Errorf("StripIndices(%q) = %q, expected %q", test.In, got, test.Expected)
		}
	}
}

func TestEscapeKey(t *testing.T) {
	type CustomString string

	for _, test := range []struct {
		In       interface{}
		Expected string
	}{
		{"", `""`},
		{"foo", `foo`},
		{"42", `42`},
		{`"foo`, `"\"foo"`},
		{"[foo]", `"[foo]"`},
		{"foo:bar", `"foo:bar"`},
		{CustomString("foo:bar"), `"foo:bar"`},
		{42, "42"},
	} {
		got := EscapeKey(test.In)
		if got != test.Expected {
			t.Errorf("EscapeKey(%q) = %q, expected %q", test.In, got, test.Expected)
		}
	}
}

func TestHasSuffix(t *testing.T) {
	for _, test := range []struct {
		In       string
		Suffix   string
		Expected bool
	}{
		{".foo.bar", ".bar", true},
		{".foo.bar", ".foo", false},
		{".foo[].bar", ".bar", true},
		{".foo.bar[]", ".bar[]", true},
		{".foo.bar[24]", ".bar[]", true},
		{".foo.bar[24].fizz", ".bar[]", false},
		{".foo.bar[24].fizz", ".bar[].fizz", true},
	} {
		got := HasSuffix(test.In, test.Suffix)
		if got != test.Expected {
			t.Errorf("HasSuffix(%q, %q) = %v, expected %v", test.In, test.Suffix, got, test.Expected)
		}
	}
}

func TestSplit(t *testing.T) {
	for _, test := range []struct {
		In   string
		Head string
		Tail string
	}{
		{"", "", ""},
		{".", ".", ""},
		{"foo", "foo", ""},
		{".foo.", ".foo", "."},
		{".foo", ".foo", ""},
		{".foo.bar", ".foo", ".bar"},
		{".foo[2].bar.fizz", ".foo", "[2].bar.fizz"},
		{"[2].bar.fizz", "[2]", ".bar.fizz"},
		{"[2]", "[2]", ""},
	} {
		head, tail := Split(test.In)
		if head != test.Head || tail != test.Tail {
			t.Errorf("Split(%q) = (%q, %q), expected (%q, %q)", test.In, head, tail, test.Head, test.Tail)
		}
	}
}

func TestExecutePath(t *testing.T) {
	for _, test := range []struct {
		I        interface{}
		Path     string
		Expected interface{}
	}{
		{
			map[string]int{"foo": 42},
			".foo",
			42,
		},
		{
			map[string][]int{
				"foo": []int{1, 2, 3},
				"bar": []int{4, 5, 6},
			},
			".foo[1]",
			2,
		},
		{
			map[string][]map[int]string{
				"foo": []map[int]string{
					map[int]string{
						23: "ha",
						44: "bar",
					},
				},
			},
			".foo[0].23",
			"ha",
		},
		{
			map[string]interface{}{
				"foo": []interface{}{
					42,
					23,
				},
			},
			`."foo"[1]`,
			23,
		},
		{
			map[string]interface{}{
				"foo[0]": []interface{}{
					42,
					23,
				},
			},
			`."foo[0]"[1]`,
			23,
		},
		{
			[]interface{}{
				map[string]interface{}{
					"foo": 23,
					"bar": 42,
				},
			},
			"[0].foo",
			23,
		},
	} {
		got, err := ExecutePath(test.Path, test.I)
		if err != nil {
			t.Errorf("ExecutePath(%q, %+v): unexpected error: %s", test.Path, test.I, err)
			continue
		}
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("ExecutePath(%q, %+v) = %+v, expected %+v", test.Path, test.I, got, test.Expected)
		}
	}

	malformedPath := `."foo[4]`
	_, err := ExecutePath(malformedPath, map[string][]string{
		`"foo`: []string{"a", "b", "c", "d"},
	})
	if err == nil {
		t.Errorf("ExecutePath(%q): expected error, got nil", malformedPath)
	}
}

func TestParseKey(t *testing.T) {
	for _, test := range []struct {
		In          string
		Expected    pathKey
		I           int
		ExpectError bool
	}{
		{
			In:       ".foo[42]",
			Expected: "foo",
			I:        4,
		},
		{
			In:       ".foo",
			Expected: "foo",
			I:        4,
		},
		{
			In:       ".foo: ",
			Expected: "foo",
			I:        4,
		},
		{
			In:       `."foo".bar`,
			Expected: "foo",
			I:        6,
		},
		{
			In:       `."foo"`,
			Expected: "foo",
			I:        6,
		},
		{
			In:       `."foo": `,
			Expected: "foo",
			I:        6,
		},
		{
			In:       `."foo\"bar"`,
			Expected: `foo"bar`,
			I:        11,
		},
		{
			In:          ".",
			ExpectError: true,
		},
		{
			In:          `."foo`,
			ExpectError: true,
		},
		{
			In:          `."foo\"bar`,
			ExpectError: true,
		},
		{
			In:          `."foo\`,
			ExpectError: true,
		},
	} {
		got, i, err := parseKey(test.In)
		if test.ExpectError && err == nil {
			t.Errorf("parseKey(%q): expected error, got nil", test.In)
			continue
		}
		if !test.ExpectError && err != nil {
			t.Errorf("parseKey(%q): unexpected error: %v", test.In, err)
		}
		if err != nil {
			continue
		}

		if got != test.Expected {
			t.Errorf("parseKey(%q) = %q, expected %q", test.In, got, test.Expected)
		}
		if i != test.I {
			t.Errorf("parseKey(%q) = [%d], expected [%d]", test.In, i, test.I)
		}
	}
}

func TestParseIndex(t *testing.T) {
	for _, test := range []struct {
		In          string
		Expected    pathIndex
		I           int
		ExpectError bool
	}{
		{
			In:       "[42].foo",
			Expected: 42,
			I:        4,
		},
		{
			In:       "[3].foo",
			Expected: 3,
			I:        3,
		},
		{
			In:       "[3]",
			Expected: 3,
			I:        3,
		},
		{
			In:       "[3]:",
			Expected: 3,
			I:        3,
		},
		{
			In:          "[a]",
			ExpectError: true,
		},
		{
			In:          "[]",
			ExpectError: true,
		},
		{
			In:          "[",
			ExpectError: true,
		},
		{
			In:          "[42",
			ExpectError: true,
		},
	} {
		got, i, err := parseIndex(test.In)
		if test.ExpectError && err == nil {
			t.Errorf("parseIndex(%q): expected error, got nil", test.In)
			continue
		}
		if !test.ExpectError && err != nil {
			t.Errorf("parseIndex(%q): unexpected error: %v", test.In, err)
		}
		if err != nil {
			continue
		}

		if got != test.Expected {
			t.Errorf("parseIndex(%q) = %d, expected %d", test.In, got, test.Expected)
		}
		if i != test.I {
			t.Errorf("parseIndex(%q) = [%d], expected [%d]", test.In, i, test.I)
		}
	}
}

func TestParsePath(t *testing.T) {
	for _, test := range []struct {
		In          string
		Expected    []pathPart
		I           int
		ExpectError bool
	}{
		{
			In: ".foo.bar",
			Expected: []pathPart{
				pathKey("foo"),
				pathKey("bar"),
			},
			I: 8,
		},
		{
			In: "[2]",
			Expected: []pathPart{
				pathIndex(2),
			},
			I: 3,
		},
		{
			In: "[2].foo: bar",
			Expected: []pathPart{
				pathIndex(2),
				pathKey("foo"),
			},
			I: 7,
		},
		{
			In: `[2]."": bar`,
			Expected: []pathPart{
				pathIndex(2),
				pathKey(""),
			},
			I: 6,
		},
		{
			In: ".foo.bar: ",
			Expected: []pathPart{
				pathKey("foo"),
				pathKey("bar"),
			},
			I: 8,
		},
		{
			In: `.foo.bar[42]."hello world!": `,
			Expected: []pathPart{
				pathKey("foo"),
				pathKey("bar"),
				pathIndex(42),
				pathKey("hello world!"),
			},
			I: 27,
		},
		{
			In:          `.foo.bar[42]."hello world!: `,
			ExpectError: true,
		},
	} {
		got, i, err := parsePath(test.In)
		if test.ExpectError && err == nil {
			t.Errorf("parsePath(%q): expected error, got nil", test.In)
			continue
		}
		if !test.ExpectError && err != nil {
			t.Errorf("parsePath(%q): unexpected error: %v", test.In, err)
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("parsePath(%q) = %v, expected %v", test.In, got, test.Expected)
		}
		if i != test.I {
			t.Errorf("parsePath(%q) = [%d], expected [%d]", test.In, i, test.I)
		}
	}
}

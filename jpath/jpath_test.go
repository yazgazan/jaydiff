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
	} {
		got := StripIndices(test.In)
		if got != test.Expected {
			t.Errorf("StripIndices(%q) = %q, expected %q", test.In, got, test.Expected)
		}
	}
}

func TestEscapeKey(t *testing.T) {
	for _, test := range []struct {
		In       interface{}
		Expected string
	}{
		{"", `""`},
		{"foo", `foo`},
		{"42", `42`},
		{`"foo`, `"\"foo"`},
		{"[foo]", `"[foo]"`},
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
}

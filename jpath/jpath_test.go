package jpath

import "testing"

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

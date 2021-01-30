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

func TestGetKey(t *testing.T) {
	for _, test := range []struct {
		S           string
		Kind        reflect.Kind
		Expected    interface{}
		ExpectError bool
	}{
		{
			S:        "foo",
			Kind:     reflect.String,
			Expected: "foo",
		},
		{
			S:        "42",
			Kind:     reflect.Int,
			Expected: 42,
		},
		{
			S:           "foo",
			Kind:        reflect.Struct,
			ExpectError: true,
		},
		{
			S:           "foo",
			Kind:        reflect.Int,
			ExpectError: true,
		},
	} {
		got, err := getKey(test.S, test.Kind)
		if test.ExpectError && err == nil {
			t.Errorf("getKey(%q, %v): expected error, got nil", test.S, test.Kind)
		}
		if !test.ExpectError && err != nil {
			t.Errorf("getKey(%q, %v): unexpected error: %v", test.S, test.Kind, err)
		}
		if err != nil || test.ExpectError {
			continue
		}

		i := got.Interface()
		if !reflect.DeepEqual(i, test.Expected) {
			t.Errorf("getKey(%q, %v) = %v, expected %v", test.S, test.Kind, got, test.Expected)
		}
	}
}

func TestExecutePath(t *testing.T) {
	for _, test := range []struct {
		I           interface{}
		Path        string
		Expected    interface{}
		ExpectError bool
	}{
		{
			I:        map[string]int{"foo": 42},
			Path:     ".foo",
			Expected: 42,
		},
		{
			I: map[string][]int{
				"foo": []int{1, 2, 3},
				"bar": []int{4, 5, 6},
			},
			Path:     ".foo[1]",
			Expected: 2,
		},
		{
			I: map[string][]map[int]string{
				"foo": []map[int]string{
					map[int]string{
						23: "ha",
						44: "bar",
					},
				},
			},
			Path:     ".foo[0].23",
			Expected: "ha",
		},
		{
			I: map[string]interface{}{
				"foo": []interface{}{
					42,
					23,
				},
			},
			Path:     `."foo"[1]`,
			Expected: 23,
		},
		{
			I: map[string]interface{}{
				"foo[0]": []interface{}{
					42,
					23,
				},
			},
			Path:     `."foo[0]"[1]`,
			Expected: 23,
		},
		{
			I: []interface{}{
				map[string]interface{}{
					"foo": 23,
					"bar": 42,
				},
			},
			Path:     "[0].foo",
			Expected: 23,
		},
		{
			I: []interface{}{
				42,
			},
			Path:        ".foo",
			ExpectError: true,
		},
		{
			I: map[string]interface{}{
				"foo": 42,
			},
			Path:        "[2]",
			ExpectError: true,
		},
		{
			I:           []interface{}(nil),
			Path:        "[2]",
			ExpectError: true,
		},
		{
			I:           map[string]interface{}(nil),
			Path:        ".foo",
			ExpectError: true,
		},
		{
			I:           map[struct{}]interface{}{},
			Path:        ".foo",
			ExpectError: true,
		},
		{
			I:           []interface{}{},
			Path:        "[2]",
			ExpectError: true,
		},
	} {
		got, err := ExecutePath(test.Path, test.I)
		if !test.ExpectError && err != nil {
			t.Errorf("ExecutePath(%q, %+v): unexpected error: %v", test.Path, test.I, err)
		}
		if test.ExpectError && err == nil {
			t.Errorf("ExecutePath(%q, %+v): expected error: got nil", test.Path, test.I)
		}
		if test.ExpectError || err != nil {
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

type invalidPathPart struct{}

func (p invalidPathPart) Kind() PathKind {
	return -1
}

func (p invalidPathPart) String() string {
	return "<invalid>"
}

func TestExecutePath2(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		_, err := executePath(Path{invalidPathPart{}}, 42)

		if err == nil {
			t.Error("executePath(Path{invalidPathPart{}}): expected error, got nil")
		}
	})
}

func TestExecuteSlice(t *testing.T) {
	t.Run("accessing private field", func(t *testing.T) {
		type Foo struct {
			private []int
		}

		f := Foo{
			private: []int{1, 2, 3},
		}
		v := reflect.ValueOf(f).FieldByName("private")

		_, err := executeSlice(PathIndex(0), Path{}, v)
		if err == nil {
			t.Error("executeSlice(0, [], fromPrivateField): expected error, got nil")
		}
	})
}

func TestExecuteMap(t *testing.T) {
	t.Run("accessing private field", func(t *testing.T) {
		type Foo struct {
			private map[string]int
		}

		f := Foo{
			private: map[string]int{
				"foo": 42,
			},
		}
		v := reflect.ValueOf(f).FieldByName("private")

		_, err := executeMap(PathKey("foo"), Path{}, v)
		if err == nil {
			t.Error(`executeMap("foo", [], fromPrivateField): expected error, got nil`)
		}
	})
}

func TestParseKey(t *testing.T) {
	for _, test := range []struct {
		In          string
		Expected    PathKey
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
		Expected    PathIndex
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
		Expected    Path
		I           int
		ExpectError bool
	}{
		{
			In: ".foo.bar",
			Expected: Path{
				PathKey("foo"),
				PathKey("bar"),
			},
			I: 8,
		},
		{
			In: "[2]",
			Expected: Path{
				PathIndex(2),
			},
			I: 3,
		},
		{
			In: "[2].foo: bar",
			Expected: Path{
				PathIndex(2),
				PathKey("foo"),
			},
			I: 7,
		},
		{
			In: `[2]."": bar`,
			Expected: Path{
				PathIndex(2),
				PathKey(""),
			},
			I: 6,
		},
		{
			In: ".foo.bar: ",
			Expected: Path{
				PathKey("foo"),
				PathKey("bar"),
			},
			I: 8,
		},
		{
			In: `.foo.bar[42]."hello world!": `,
			Expected: Path{
				PathKey("foo"),
				PathKey("bar"),
				PathIndex(42),
				PathKey("hello world!"),
			},
			I: 27,
		},
		{
			In:       ": ",
			Expected: Path{},
			I:        0,
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

func TestParse(t *testing.T) {
	for _, test := range []struct {
		In          string
		Expected    Path
		ExpectError bool
	}{
		{
			In: ".foo.bar",
			Expected: Path{
				PathKey("foo"),
				PathKey("bar"),
			},
		},
		{
			In: "[2]",
			Expected: Path{
				PathIndex(2),
			},
		},
		{
			In:          "[2].foo: bar",
			ExpectError: true,
		},
		{
			In: `[2].""`,
			Expected: Path{
				PathIndex(2),
				PathKey(""),
			},
		},
		{
			In: ".foo.bar",
			Expected: Path{
				PathKey("foo"),
				PathKey("bar"),
			},
		},
		{
			In: `.foo.bar[42]."hello world!"`,
			Expected: Path{
				PathKey("foo"),
				PathKey("bar"),
				PathIndex(42),
				PathKey("hello world!"),
			},
		},
		{
			In:       "",
			Expected: Path{},
		},
		{
			In:          `.foo."ba\xzar"`,
			ExpectError: true,
		},
		{
			In:          `.foo.bar[42]."hello world!: `,
			ExpectError: true,
		},
	} {
		got, err := Parse(test.In)
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
	}
}

func TestPathIndex(t *testing.T) {
	t.Run("Kind()", func(t *testing.T) {
		k := PathIndex(2).Kind()

		if k != PathKindIndex {
			t.Errorf("PathIndex().Kind() = %v, expected %v", k, PathKindIndex)
		}
	})

	t.Run("String()", func(t *testing.T) {
		s := PathIndex(2).String()

		if s != "[2]" {
			t.Errorf("PathIndex(2).String() = %q, expected %q", s, "[2]")
		}
	})
}

func TestPathKey(t *testing.T) {
	t.Run("Kind()", func(t *testing.T) {
		k := PathKey("foo").Kind()

		if k != PathKindKey {
			t.Errorf("PathKey().Kind() = %v, expected %v", k, PathKindKey)
		}
	})

	t.Run("String()", func(t *testing.T) {
		s := PathKey("foo").String()

		if s != ".foo" {
			t.Errorf(`PathKey("foo").String() = %q, expected %q`, s, ".foo")
		}
	})

	t.Run("String() escaped", func(t *testing.T) {
		s := PathKey(`foo"`).String()

		if s != `."foo\""` {
			t.Errorf(`PathKey("foo").String() = %q, expected %q`, s, `."foo\""`)
		}
	})
}

func TestPathKind(t *testing.T) {
	t.Run("String()", func(t *testing.T) {

		for _, test := range []struct {
			In       PathKind
			Expected string
		}{
			{
				In:       PathKindIndex,
				Expected: "PathKindIndex",
			},
			{
				In:       PathKindKey,
				Expected: "PathKindKey",
			},
			{
				In:       PathKind(-1),
				Expected: "PathKind(-1)",
			},
		} {
			got := test.In.String()

			if got != test.Expected {
				t.Errorf("PathKind(%d).String() = %q, expected %q", int(test.In), got, test.Expected)
			}
		}
	})
}

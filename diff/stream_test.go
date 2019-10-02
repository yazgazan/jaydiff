package diff

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"testing"
)

func jsonMustDecode(s string) interface{} {
	var v interface{}

	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		panic(err)
	}

	return v
}

func TestJSONStream(t *testing.T) {
	b := []byte(`
{"foo": "bar", "hello": ["world", "!"]}
{"this": "is", "a": "test"}
42
null
"ahoy"
`)

	expected := []interface{}{
		jsonMustDecode(`{"foo": "bar", "hello": ["world", "!"]}`),
		jsonMustDecode(`{"this": "is", "a": "test"}`),
		jsonMustDecode(`42`),
		jsonMustDecode(`null`),
		jsonMustDecode(`"ahoy"`),
	}
	got := []interface{}{}
	stream := JSONStream{
		Decoder: json.NewDecoder(bytes.NewBuffer(b)),
	}
	for {
		v, err := stream.NextValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		got = append(got, v)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("JSONStream(%q) = %+v, expected %+v", b, got, expected)
	}

	v, err := stream.NextValue()
	if err != io.EOF {
		t.Errorf("JSONStream{}.NextValue() = %+v, %v; expected nil, io.EOF", v, err)
	}
}

func TestDiffStreamValues(t *testing.T) {
	t.Run("identical", func(t *testing.T) {
		d, lhs, rhs, err := diffStreamValues(
			config{},
			mockStream("foo"),
			mockStream("foo"),
			&visited{},
		)
		if err != nil {
			t.Errorf("diffStreamValues(...): unexpected error: %v", err)
			return
		}
		if d.Diff() != Identical {
			t.Errorf("diffStreamValues(...).Diff() = %v, expected %v", d.Diff(), Identical)
			return
		}
		if !reflect.DeepEqual(lhs, interface{}("foo")) {
			t.Errorf("diffStreamValues(...).lhs = %v, expected %v", lhs, "foo")
		}
		if !reflect.DeepEqual(rhs, interface{}("foo")) {
			t.Errorf("diffStreamValues(...).rhs = %v, expected %v", rhs, "foo")
		}
	})

	t.Run("content differs", func(t *testing.T) {
		d, lhs, rhs, err := diffStreamValues(
			config{},
			mockStream("foo"),
			mockStream("bar"),
			&visited{},
		)
		if err != nil {
			t.Errorf("diffStreamValues(...): unexpected error: %v", err)
			return
		}
		if d.Diff() != ContentDiffer {
			t.Errorf("diffStreamValues(...).Diff() = %v, expected %v", d.Diff(), ContentDiffer)
			return
		}
		if !reflect.DeepEqual(lhs, interface{}("foo")) {
			t.Errorf("diffStreamValues(...).lhs = %v, expected %v", lhs, "foo")
		}
		if !reflect.DeepEqual(rhs, interface{}("bar")) {
			t.Errorf("diffStreamValues(...).rhs = %v, expected %v", rhs, "bar")
		}
	})

	t.Run("types differs", func(t *testing.T) {
		d, lhs, rhs, err := diffStreamValues(
			config{},
			mockStream("foo"),
			mockStream(42),
			&visited{},
		)
		if err != nil {
			t.Errorf("diffStreamValues(...): unexpected error: %v", err)
			return
		}
		if d.Diff() != TypesDiffer {
			t.Errorf("diffStreamValues(...).Diff() = %v, expected %v", d.Diff(), TypesDiffer)
			return
		}
		if !reflect.DeepEqual(lhs, interface{}("foo")) {
			t.Errorf("diffStreamValues(...).lhs = %v, expected %v", lhs, "foo")
		}
		if !reflect.DeepEqual(rhs, interface{}(42)) {
			t.Errorf("diffStreamValues(...).rhs = %v, expected %v", rhs, 42)
		}
	})

	t.Run("missing", func(t *testing.T) {
		d, lhs, rhs, err := diffStreamValues(
			config{},
			mockStream("foo"),
			mockStream(),
			&visited{},
		)
		if err != nil {
			t.Errorf("diffStreamValues(...): unexpected error: %v", err)
			return
		}
		if d.Diff() != ContentDiffer {
			t.Errorf("diffStreamValues(...).Diff() = %v, expected %v", d.Diff(), ContentDiffer)
			return
		}
		if !reflect.DeepEqual(lhs, interface{}("foo")) {
			t.Errorf("diffStreamValues(...).lhs = %v, expected %v", lhs, "foo")
		}
		if rhs != nil {
			t.Errorf("diffStreamValues(...).rhs = %v, expected %v", rhs, nil)
		}
	})

	t.Run("excess", func(t *testing.T) {
		d, lhs, rhs, err := diffStreamValues(
			config{},
			mockStream(),
			mockStream("foo"),
			&visited{},
		)
		if err != nil {
			t.Errorf("diffStreamValues(...): unexpected error: %v", err)
			return
		}
		if d.Diff() != ContentDiffer {
			t.Errorf("diffStreamValues(...).Diff() = %v, expected %v", d.Diff(), ContentDiffer)
			return
		}
		if lhs != nil {
			t.Errorf("diffStreamValues(...).lhs = %v, expected %v", lhs, nil)
		}
		if !reflect.DeepEqual(rhs, interface{}("foo")) {
			t.Errorf("diffStreamValues(...).rhs = %v, expected %v", rhs, "foo")
		}
	})

	t.Run("erroring stream", func(t *testing.T) {
		_, _, _, err := diffStreamValues(
			config{},
			erroringMockedStream{},
			mockStream(),
			&visited{},
		)
		if err == nil {
			t.Error("diffStreamValues(...): expected error, got nil")
			return
		}

		_, _, _, err = diffStreamValues(
			config{},
			mockStream(),
			erroringMockedStream{},
			&visited{},
		)
		if err == nil {
			t.Error("diffStreamValues(...): expected error, got nil")
			return
		}
	})
}

func TestStream(t *testing.T) {
	c := defaultConfig()

	for _, test := range []stringTest{
		{
			LHS: mockStream(1, 2, 3),
			RHS: mockStream(1, 2, 3),
			Want: [][]string{
				{"int", "1"},
				{"int", "2"},
				{"int", "3"},
			},
			Type: Identical,
		},
		{
			LHS: mockStream(1, 2, 3),
			RHS: mockStream(4, 5, 6),
			Want: [][]string{
				{},
				{"-", "int", "1"},
				{"+", "int", "4"},
				{"-", "int", "2"},
				{"+", "int", "5"},
				{"-", "int", "3"},
				{"+", "int", "6"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: mockStream(1, 2, 3),
			RHS: mockStream(1, 2),
			Want: [][]string{
				{},
				{"int", "1"},
				{"int", "2"},
				{"-", "int", "3"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: mockStream(1, 2),
			RHS: mockStream(1, 2, 3),
			Want: [][]string{
				{},
				{"int", "1"},
				{"int", "2"},
				{"+", "int", "3"},
				{},
			},
			Type: ContentDiffer,
		},
	} {
		typ, err := newStream(c, test.LHS, test.RHS, &visited{})
		if err != nil {
			t.Errorf("newStream(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
			continue
		}
		if typ.Diff() != test.Type {
			t.Errorf("Stream.Diff() = %q, expected %q", typ.Diff(), test.Type)
		}

		ss := typ.Strings()
		indented := typ.StringIndent(testKey, testPrefix, testOutput)
		testStrings("TestStream", t, test.Want, ss, indented)

		indentedJSON := typ.StringIndent(testKey, testPrefix, testJSONOutput)
		testStrings("TestStream", t, test.WantJSON, ss, indentedJSON)
	}

	t.Run("invalid", func(t *testing.T) {
		_, err := newStream(c, nil, mockStream(), &visited{})
		if err == nil {
			t.Errorf("newStream(nil, %+v): expected error, got nil", mockStream())
		}
	})
	t.Run("invalid", func(t *testing.T) {
		_, err := newStream(c, mockStream(), nil, &visited{})
		if err == nil {
			t.Errorf("newStream(%+v, nil): expected error, got nil", mockStream())
		}
	})

	t.Run("erroring", func(t *testing.T) {
		_, err := newStream(c, mockStream(), erroringMockedStream{}, &visited{})
		if err == nil {
			t.Errorf("newStream(%+v, erroring{}): expected error, got nil", mockStream())
		}
	})
	t.Run("erroring", func(t *testing.T) {
		_, err := newStream(c, erroringMockedStream{}, mockStream(), &visited{})
		if err == nil {
			t.Errorf("newStream(erroring{}, %+v): expected error, got nil", mockStream())
		}
	})
}

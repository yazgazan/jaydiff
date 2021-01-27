package diff

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

type mockedStream struct {
	values []interface{}
	i      int
}

func mockStream(ii ...interface{}) *mockedStream {
	return &mockedStream{
		values: ii,
	}
}

func (s *mockedStream) NextValue() (interface{}, error) {
	if s.i >= len(s.values) {
		return nil, io.EOF
	}

	s.i++
	return s.values[s.i-1], nil
}

type erroringMockedStream struct{}

func (s erroringMockedStream) NextValue() (interface{}, error) {
	return nil, errors.New("stream error")
}

func TestDiff(t *testing.T) {
	type CustomType int

	for _, test := range []struct {
		LHS   interface{}
		RHS   interface{}
		Want  Type
		Error bool
	}{
		{LHS: nil, RHS: nil, Want: Identical},
		{LHS: nil, RHS: 32, Want: TypesDiffer},
		{LHS: 23, RHS: nil, Want: TypesDiffer},
		{LHS: 42, RHS: 42, Want: Identical},
		{LHS: 23, RHS: 42, Want: ContentDiffer},
		{LHS: 10.0, RHS: 10, Want: TypesDiffer},
		{LHS: "abc", RHS: "abc", Want: Identical},
		{LHS: "def", RHS: "ghi", Want: ContentDiffer},
		{LHS: "def", RHS: 2, Want: TypesDiffer},
		{LHS: []int{1, 2, 3}, RHS: 23, Want: TypesDiffer},
		{LHS: []int{1, 2, 3}, RHS: []int{1, 2, 3}, Want: Identical},
		{LHS: []int{1, 2, 3, 4}, RHS: []int{1, 2, 3}, Want: ContentDiffer},
		{LHS: []int{1, 2, 3}, RHS: []int{1, 2, 3, 4}, Want: ContentDiffer},
		{LHS: []int{1, 2, 3}, RHS: []int{4, 5}, Want: ContentDiffer},
		{LHS: []int{1, 2, 3}, RHS: []float64{4, 5}, Want: TypesDiffer},
		{LHS: []int{1, 2, 3}, RHS: []float64{4, 5}, Want: TypesDiffer},
		{LHS: []func(){func() {}}, RHS: []func(){func() {}}, Want: ContentDiffer, Error: true},
		{LHS: map[int]int{2: 4, 6: 12}, RHS: map[int]int{2: 4, 6: 12}, Want: Identical},
		{LHS: map[int]int{2: 4, 6: 12, 8: 16}, RHS: map[int]int{2: 4, 6: 12}, Want: ContentDiffer},
		{LHS: map[int]int{2: 4, 6: 12}, RHS: map[int]int{2: 4, 6: 12, 1: 2}, Want: ContentDiffer},
		{LHS: map[int]int{2: 4, 6: 12}, RHS: map[float64]int{2: 4, 6: 12}, Want: TypesDiffer},
		{
			LHS:   map[int]func(){0: func() {}},
			RHS:   map[int]func(){0: func() {}},
			Want:  ContentDiffer,
			Error: true,
		},
		{LHS: map[int]int{2: 4, 6: 12}, RHS: map[int]int{1: 2, 3: 6}, Want: ContentDiffer},
		{LHS: map[int]float32{2: 4, 6: 12}, RHS: map[int]int{1: 2, 3: 6}, Want: TypesDiffer},
		{
			LHS:  map[int][]int{1: {2, 3}, 2: {3, 4}},
			RHS:  map[int][]int{1: {2, 3}, 2: {3, 4}},
			Want: Identical,
		},
		{
			LHS:  map[int][]int{1: {2, 3}, 2: {3, 4}},
			RHS:  map[int][]int{1: {2, 3}, 2: {3, 5}},
			Want: ContentDiffer,
		},
		{LHS: []interface{}{1, 2, 3}, RHS: []interface{}{1, 2, 3}, Want: Identical},
		{LHS: []interface{}{1, 2, 3}, RHS: []interface{}{1, 2, 3.3}, Want: ContentDiffer},
		{LHS: []interface{}(nil), RHS: []interface{}{1, 2, 3.3}, Want: ContentDiffer},
		{LHS: []int(nil), RHS: []int{}, Want: Identical},
		{LHS: func() {}, RHS: func() {}, Want: TypesDiffer, Error: true},
		{
			LHS:  struct{}{},
			RHS:  struct{}{},
			Want: Identical,
		},
		{
			LHS:  struct{ Foo int }{Foo: 42},
			RHS:  struct{ Foo int }{Foo: 21},
			Want: ContentDiffer,
		},
		{
			LHS:  struct{ Foo int }{Foo: 42},
			RHS:  struct{ Bar int }{Bar: 42},
			Want: TypesDiffer,
		},
		{
			LHS:  mockStream(1, 2, 3),
			RHS:  mockStream(1, 2, 3),
			Want: Identical,
		},
		{
			LHS:  mockStream(1, 2, 3),
			RHS:  mockStream(4, 5, 6),
			Want: ContentDiffer,
		},
		{
			LHS:  CustomType(1),
			RHS:  CustomType(1),
			Want: Identical,
		},
		{
			LHS:  CustomType(1),
			RHS:  CustomType(2),
			Want: ContentDiffer,
		},
		{
			LHS:  CustomType(1),
			RHS:  2,
			Want: ContentDiffer,
		},
		{
			LHS:  new(CustomType),
			RHS:  new(CustomType),
			Want: Identical,
		},
		{
			LHS:  time.Time{},
			RHS:  time.Time{},
			Want: Identical,
		},
		{
			LHS:  new(time.Time),
			RHS:  new(time.Time),
			Want: Identical,
		},
		{
			LHS: struct {
				Foo *int
			}{},
			RHS: struct {
				Foo *int
			}{},
			Want: Identical,
		},
		{
			LHS: struct {
				Foo *int
			}{
				Foo: new(int),
			},
			RHS: struct {
				Foo *int
			}{},
			Want: ContentDiffer,
		},
		{
			LHS: struct {
				Foo *int
			}{
				Foo: func() *int {
					i := 42

					return &i
				}(),
			},
			RHS: struct {
				Foo *int
			}{
				Foo: func() *int {
					i := 42

					return &i
				}(),
			},
			Want: Identical,
		},
		{
			LHS: struct {
				Foo *int
			}{
				Foo: func() *int {
					i := 42

					return &i
				}(),
			},
			RHS: struct {
				Foo *int
			}{
				Foo: func() *int {
					i := 84

					return &i
				}(),
			},
			Want: ContentDiffer,
		},
	} {
		diff, err := Diff(test.LHS, test.RHS)

		if err == nil && test.Error {
			t.Errorf("Diff(%#v, %#v) expected an error, got nil instead", test.LHS, test.RHS)
		}
		if err != nil && !test.Error {
			t.Errorf("Diff(%#v, %#v): unexpected error: %q", test.LHS, test.RHS, err)
		}

		if diff.Diff() != test.Want {
			t.Logf("LHS: %+#v\n", test.LHS)
			t.Logf("RHS: %+#v\n", test.RHS)
			t.Errorf("Diff(%#v, %#v) = %q, expected %q", test.LHS, test.RHS, diff.Diff(), test.Want)
		}
	}
}

func TestDiffMyers(t *testing.T) {
	for _, test := range []struct {
		LHS   interface{}
		RHS   interface{}
		Want  Type
		Error bool
	}{
		{LHS: []int{1, 2, 3}, RHS: []int{1, 2, 3}, Want: Identical},
		{LHS: []int{1, 2, 3, 4}, RHS: []int{1, 2, 3}, Want: ContentDiffer},
		{LHS: []int{1, 2, 3}, RHS: []int{1, 2, 3, 4}, Want: ContentDiffer},
		{LHS: []int{1, 2, 3}, RHS: []int{4, 5}, Want: ContentDiffer},
		{LHS: []int{1, 2, 3}, RHS: []float64{4, 5}, Want: TypesDiffer},
		{LHS: []int{1, 2, 3}, RHS: []float64{4, 5}, Want: TypesDiffer},
		{LHS: []func(){func() {}}, RHS: []func(){func() {}}, Want: ContentDiffer, Error: true},
		{
			LHS:  map[int][]int{1: {2, 3}, 2: {3, 4}},
			RHS:  map[int][]int{1: {2, 3}, 2: {3, 4}},
			Want: Identical,
		},
		{
			LHS:  map[int][]int{1: {2, 3}, 2: {3, 4}},
			RHS:  map[int][]int{1: {2, 3}, 2: {3, 5}},
			Want: ContentDiffer,
		},
		{LHS: []interface{}{1, 2, 3}, RHS: []interface{}{1, 2, 3}, Want: Identical},
		{LHS: []interface{}{1, 2, 3}, RHS: []interface{}{1, 2, 3.3}, Want: ContentDiffer},
		{LHS: []interface{}(nil), RHS: []interface{}{1, 2, 3.3}, Want: ContentDiffer},
		{LHS: []int(nil), RHS: []int{}, Want: Identical},
	} {
		diff, err := Diff(test.LHS, test.RHS, UseSliceMyers())

		if err == nil && test.Error {
			t.Errorf("Diff(%#v, %#v) expected an error, got nil instead", test.LHS, test.RHS)
		}
		if err != nil && !test.Error {
			t.Errorf("Diff(%#v, %#v): unexpected error: %q", test.LHS, test.RHS, err)
		}

		if diff.Diff() != test.Want {
			t.Logf("LHS: %+#v\n", test.LHS)
			t.Logf("RHS: %+#v\n", test.RHS)
			t.Errorf("Diff(%#v, %#v) = %q, expected %q", test.LHS, test.RHS, diff.Diff(), test.Want)
		}
	}
}

func TestTypeString(t *testing.T) {
	for _, test := range []struct {
		Input Type
		Want  string
	}{
		{Identical, "identical"},
		{ContentDiffer, "content"},
		{TypesDiffer, "types"},
		{Type(-1), "invalid"},
	} {
		s := test.Input.String()

		if !strings.Contains(s, test.Want) {
			t.Errorf("Type.String() = %q, expected it to contain %q", s, test.Want)
		}
	}
}

type stringTest struct {
	LHS      interface{}
	RHS      interface{}
	Want     [][]string
	WantJSON [][]string
	Type
}

const (
	testKey    = "(key)"
	testPrefix = "(prefix)"
)

var testOutput = Output{ShowTypes: true}
var testJSONOutput = Output{JSON: true}

func TestTypes(t *testing.T) {
	for _, test := range []stringTest{
		{
			LHS: 4,
			RHS: 2.1,
			Want: [][]string{
				{"int", "4"},
				{"float64", "2.1"},
			},
		},
	} {
		typ := types{test.LHS, test.RHS}

		if typ.Diff() != TypesDiffer {
			t.Errorf("Types.Diff() = %q, expected %q", typ.Diff(), TypesDiffer)
		}

		ss := typ.Strings()
		indented := typ.StringIndent(testKey, testPrefix, testOutput)
		testStrings("TestTypes", t, test.Want, ss, indented)
	}
}

func TestScalar(t *testing.T) {
	for _, test := range []stringTest{
		{
			LHS: 4,
			RHS: 4,
			Want: [][]string{
				{"int", "4"},
			},
			Type: Identical,
		},
		{
			LHS: 4,
			RHS: 2,
			Want: [][]string{
				{"int", "4"},
				{"int", "2"},
			},
			Type: ContentDiffer,
		},
		{
			LHS: 4,
			RHS: 2.1,
			Want: [][]string{
				{"int", "4"},
				{"float64", "2.1"},
			},
			Type: TypesDiffer,
		},
		{
			LHS: complex(4, -3),
			RHS: complex(4, -3),
			Want: [][]string{
				{"complex128", "4", "-3"},
			},
			Type: Identical,
		},
		{
			LHS: complex(4, -3),
			RHS: complex(-7, 32),
			Want: [][]string{
				{"complex128", "4", "-3"},
				{"complex128", "-7", "32"},
			},
			Type: ContentDiffer,
		},
		{
			LHS: 2.1,
			RHS: complex(4, -3),
			Want: [][]string{
				{"float64", "2.1"},
				{"complex128", "4", "-3"},
			},
			Type: TypesDiffer,
		},
	} {
		typ := scalar{test.LHS, test.RHS}

		if typ.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", typ.Diff(), test.Type)
		}

		ss := typ.Strings()
		indented := typ.StringIndent(testKey, testPrefix, testOutput)
		testStrings("TestScalar", t, test.Want, ss, indented)
	}
}

type emptyStruct struct{}
type subStruct struct {
	A int
}
type structA struct {
	Foo int
	Bar subStruct
	baz float64
}
type structB struct {
	Foo int
	Bar subStruct
	baz float64
}
type structC struct {
	Foo []int
}
type structInvalid struct {
	A func()
}

func TestTypeStruct(t *testing.T) {
	for i, test := range []stringTest{
		{
			LHS: emptyStruct{},
			RHS: emptyStruct{},
			Want: [][]string{
				{"emptyStruct", "{}"},
			},
			WantJSON: [][]string{
				{"{}"},
			},
			Type: Identical,
		},
		{
			LHS: structA{
				Foo: 42,
				Bar: subStruct{
					A: 2,
				},
				baz: 4.2,
			},
			RHS: structA{
				Foo: 42,
				Bar: subStruct{
					A: 2,
				},
				baz: 1.1,
			},
			Want: [][]string{
				{"structA", "42", "{2}", "4.2"},
			},
			WantJSON: [][]string{
				{"42", "{2}", "4.2"},
			},
			Type: Identical,
		},
		{
			LHS: structA{
				Foo: 42,
				Bar: subStruct{
					A: 2,
				},
				baz: 4.2,
			},
			RHS: structB{
				Foo: 42,
				Bar: subStruct{
					A: 2,
				},
				baz: 1.1,
			},
			Want: [][]string{
				{"structA", "42", "{2}", "4.2"},
			},
			WantJSON: [][]string{
				{"42", "{2}", "4.2"},
			},
			Type: Identical,
		},
		{
			LHS: structA{
				Foo: 42,
				Bar: subStruct{
					A: 2,
				},
				baz: 4.2,
			},
			RHS: structB{
				Foo: 23,
				Bar: subStruct{
					A: 2,
				},
				baz: 1.1,
			},
			Want: [][]string{
				{},
				{"Bar"},
				{"Foo", "-", "int", "42"},
				{"Foo", "+", "int", "23"},
				{},
			},
			WantJSON: [][]string{
				{},
				{"Bar"},
				{"Foo", "-", "42"},
				{"Foo", "+", "23"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: structA{
				Foo: 42,
				Bar: subStruct{
					A: 2,
				},
				baz: 4.2,
			},
			RHS: structC{
				Foo: []int{1, 2},
			},
			Want: [][]string{
				{"-", "structA", "42", "{2}", "4.2"},
				{"+", "structC", "[1 2]"},
			},
			WantJSON: [][]string{
				{"-", "42", "{2}", "4.2"},
				{"+", "{[1 2]}"},
			},
			Type: TypesDiffer,
		},
	} {
		s, err := newStruct(defaultConfig(), test.LHS, test.RHS, &visited{})

		if err != nil {
			t.Errorf("newStruct(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
			continue
		}
		if s.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", s.Diff(), test.Type)
		}

		ss := s.Strings()
		indented := s.StringIndent(testKey, testPrefix, testOutput)
		testStrings(fmt.Sprintf("TestMap[%d]", i), t, test.Want, ss, indented)

		indentedJSON := s.StringIndent(testKey, testPrefix, testJSONOutput)
		testStrings(fmt.Sprintf("TestMap[%d]", i), t, test.WantJSON, ss, indentedJSON)
	}

	invalid, err := newStruct(defaultConfig(), nil, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("newStruct(nil, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("newStruct(nil, nil): expected InvalidType error, got %s", err)
	}
	ss := invalid.Strings()
	if len(ss) != 0 {
		t.Errorf("len(invalidStruct.Strings()) = %d, expected 0", len(ss))
	}
	indented := invalid.StringIndent(testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("invalidStruct.StringIndent(%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
	}

	invalid, err = newStruct(defaultConfig(), structA{}, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("newStruct(structA{}, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("newStruct(structA{}, nil): expected InvalidType error, got %s", err)
	}
	ss = invalid.Strings()
	if len(ss) != 0 {
		t.Errorf("len(invalidStruct.Strings()) = %d, expected 0", len(ss))
	}
	indented = invalid.StringIndent(testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("invalidStruct.StringIndent(%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
	}

	invalid, err = newStruct(defaultConfig(), structInvalid{}, structInvalid{}, &visited{})
	if err == nil {
		t.Errorf("newStruct(structInvalid{}, structInvalid{}): expected error, got nil")
	}
}

func TestIgnore(t *testing.T) {
	ignoreDiff, _ := Ignore()

	if ignoreDiff.Diff() != Identical {
		t.Errorf("NewIgnore().Diff() = %q, expected %q", ignoreDiff.Diff(), Identical)
	}
	if len(ignoreDiff.Strings()) != 0 {
		t.Errorf("len(NewIgnore().Strings()) = %d, expected 0", len(ignoreDiff.Strings()))
	}
	if indented := ignoreDiff.StringIndent(testKey, testPrefix, testOutput); indented != "" {
		t.Errorf("NewIgnore().StringIndent(...) = %q, expected %q", indented, "")
	}
}

func TestLHS(t *testing.T) {
	validLHSTypesGetter := Differ(&types{
		lhs: 42,
		rhs: "hello",
	})
	v, err := LHS(validLHSTypesGetter)
	if err != nil {
		t.Errorf("LHS(%+v): unexpected error: %s", validLHSTypesGetter, err)
	}
	if i, ok := v.(int); !ok || i != 42 {
		t.Errorf("LHS(%+v) = %v, expected %d", validLHSTypesGetter, v, 42)
	}

	validLHSMapGetter := Differ(&mapDiff{
		lhs: 42,
		rhs: "hello",
	})
	v, err = LHS(validLHSMapGetter)
	if err != nil {
		t.Errorf("LHS(%+v): unexpected error: %s", validLHSMapGetter, err)
	}
	if i, ok := v.(int); !ok || i != 42 {
		t.Errorf("LHS(%+v) = %v, expected %d", validLHSMapGetter, v, 42)
	}

	validLHSStructGetter := Differ(&structDiff{
		lhs: structA{Foo: 42},
		rhs: structB{Foo: 23},
	})
	v, err = LHS(validLHSStructGetter)
	if err != nil {
		t.Errorf("LHS(%+v): unexpected error: %s", validLHSStructGetter, err)
	}
	if s, ok := v.(structA); !ok || s.Foo != 42 {
		t.Errorf("LHS(%+v).Foo = %v, expected %d", validLHSStructGetter, s.Foo, 42)
	}

	validLHSSliceGetter := Differ(&slice{
		lhs: 42,
		rhs: "hello",
	})
	v, err = LHS(validLHSSliceGetter)
	if err != nil {
		t.Errorf("LHS(%+v): unexpected error: %s", validLHSSliceGetter, err)
	}
	if i, ok := v.(int); !ok || i != 42 {
		t.Errorf("LHS(%+v) = %v, expected %d", validLHSSliceGetter, v, 42)
	}

	validLHSStreamGetter := Differ(&stream{
		lhs: []interface{}{42},
		rhs: []interface{}{"hello"},
	})
	v, err = LHS(validLHSStreamGetter)
	if err != nil {
		t.Errorf("LHS(%+v): unexpected error: %s", validLHSStreamGetter, err)
	}
	if !reflect.DeepEqual(v, []interface{}{interface{}(42)}) {
		t.Errorf("LHS(%+v) = %v, expected %d", validLHSStreamGetter, v, 42)
	}

	validLHSScalarGetter := Differ(&scalar{
		lhs: 42,
		rhs: "hello",
	})
	v, err = LHS(validLHSScalarGetter)
	if err != nil {
		t.Errorf("LHS(%+v): unexpected error: %s", validLHSScalarGetter, err)
	}
	if i, ok := v.(int); !ok || i != 42 {
		t.Errorf("LHS(%+v) = %v, expected %d", validLHSScalarGetter, v, 42)
	}

	validLHSSliceMissingGetter := Differ(&sliceMissing{
		value: 42,
	})
	v, err = LHS(validLHSSliceMissingGetter)
	if err != nil {
		t.Errorf("LHS(%+v): unexpected error: %s", validLHSSliceMissingGetter, err)
	}
	if i, ok := v.(int); !ok || i != 42 {
		t.Errorf("LHS(%+v) = %v, expected %d", validLHSSliceMissingGetter, v, 42)
	}

	validLHSStreamMissingGetter := Differ(&streamMissing{
		value: 42,
	})
	v, err = LHS(validLHSStreamMissingGetter)
	if err != nil {
		t.Errorf("LHS(%+v): unexpected error: %s", validLHSStreamMissingGetter, err)
	}
	if i, ok := v.(int); !ok || i != 42 {
		t.Errorf("LHS(%+v) = %v, expected %d", validLHSStreamMissingGetter, v, 42)
	}

	validLHSMapMissingGetter := Differ(&mapMissing{
		value: 42,
	})
	v, err = LHS(validLHSMapMissingGetter)
	if err != nil {
		t.Errorf("LHS(%+v): unexpected error: %s", validLHSMapMissingGetter, err)
	}
	if i, ok := v.(int); !ok || i != 42 {
		t.Errorf("LHS(%+v) = %v, expected %d", validLHSMapMissingGetter, v, 42)
	}

	invalidLHSGetter := ignore{}
	_, err = LHS(invalidLHSGetter)
	if err == nil {
		t.Errorf("LHS(%+v): expected error, got nil instead", invalidLHSGetter)
	}
	if _, ok := err.(ErrLHSNotSupported); !ok {
		t.Errorf("LHS(%+v): expected error to be of type %T, got %T instead", invalidLHSGetter, ErrLHSNotSupported{}, err)
	}
	if err.Error() == "" {
		t.Errorf("LHS(%+v): unexpected empty error message", invalidLHSGetter)
	}
}

func TestRHS(t *testing.T) {
	validRHSTypesGetter := Differ(&types{
		lhs: 42,
		rhs: "hello",
	})
	v, err := RHS(validRHSTypesGetter)
	if err != nil {
		t.Errorf("RHS(%+v): unexpected error: %s", validRHSTypesGetter, err)
	}
	if s, ok := v.(string); !ok || s != "hello" {
		t.Errorf("RHS(%+v) = %v, expected %q", validRHSTypesGetter, v, "hello")
	}

	validRHSMapGetter := Differ(&mapDiff{
		lhs: 42,
		rhs: "hello",
	})
	v, err = RHS(validRHSMapGetter)
	if err != nil {
		t.Errorf("RHS(%+v): unexpected error: %s", validRHSMapGetter, err)
	}
	if s, ok := v.(string); !ok || s != "hello" {
		t.Errorf("RHS(%+v) = %v, expected %q", validRHSMapGetter, v, "hello")
	}

	validRHSStructGetter := Differ(&structDiff{
		lhs: structA{Foo: 42},
		rhs: structB{Foo: 23},
	})
	v, err = RHS(validRHSStructGetter)
	if err != nil {
		t.Errorf("RHS(%+v): unexpected error: %s", validRHSStructGetter, err)
	}
	if s, ok := v.(structB); !ok || s.Foo != 23 {
		t.Errorf("RHS(%+v).Foo = %v, expected %d", validRHSStructGetter, s.Foo, 23)
	}

	validRHSSliceGetter := Differ(&slice{
		lhs: 42,
		rhs: "hello",
	})
	v, err = RHS(validRHSSliceGetter)
	if err != nil {
		t.Errorf("RHS(%+v): unexpected error: %s", validRHSSliceGetter, err)
	}
	if s, ok := v.(string); !ok || s != "hello" {
		t.Errorf("RHS(%+v) = %v, expected %q", validRHSSliceGetter, v, "hello")
	}

	validRHSStreamGetter := Differ(&stream{
		lhs: []interface{}{42},
		rhs: []interface{}{"hello"},
	})
	v, err = RHS(validRHSStreamGetter)
	if err != nil {
		t.Errorf("RHS(%+v): unexpected error: %s", validRHSStreamGetter, err)
	}
	if !reflect.DeepEqual(v, []interface{}{interface{}("hello")}) {
		t.Errorf("RHS(%+v) = %v, expected %q", validRHSStreamGetter, v, "hello")
	}

	validRHSScalarGetter := Differ(&scalar{
		lhs: 42,
		rhs: "hello",
	})
	v, err = RHS(validRHSScalarGetter)
	if err != nil {
		t.Errorf("RHS(%+v): unexpected error: %s", validRHSScalarGetter, err)
	}
	if s, ok := v.(string); !ok || s != "hello" {
		t.Errorf("RHS(%+v) = %v, expected %q", validRHSScalarGetter, v, "hello")
	}

	validRHSSliceExcessGetter := Differ(&sliceExcess{
		value: "hello",
	})
	v, err = RHS(validRHSSliceExcessGetter)
	if err != nil {
		t.Errorf("RHS(%+v): unexpected error: %s", validRHSSliceExcessGetter, err)
	}
	if s, ok := v.(string); !ok || s != "hello" {
		t.Errorf("RHS(%+v) = %v, expected %q", validRHSSliceExcessGetter, v, "hello")
	}

	validRHSStreamExcessGetter := Differ(&streamExcess{
		value: "hello",
	})
	v, err = RHS(validRHSStreamExcessGetter)
	if err != nil {
		t.Errorf("RHS(%+v): unexpected error: %s", validRHSStreamExcessGetter, err)
	}
	if s, ok := v.(string); !ok || s != "hello" {
		t.Errorf("RHS(%+v) = %v, expected %q", validRHSStreamExcessGetter, v, "hello")
	}

	validRHSMapExcessGetter := Differ(&mapExcess{
		value: "hello",
	})
	v, err = RHS(validRHSMapExcessGetter)
	if err != nil {
		t.Errorf("RHS(%+v): unexpected error: %s", validRHSMapExcessGetter, err)
	}
	if s, ok := v.(string); !ok || s != "hello" {
		t.Errorf("RHS(%+v) = %v, expected %q", validRHSMapExcessGetter, v, "hello")
	}

	invalidRHSGetter := ignore{}
	_, err = RHS(invalidRHSGetter)
	if err == nil {
		t.Errorf("RHS(%+v): expected error, got nil instead", invalidRHSGetter)
	}
	if _, ok := err.(ErrRHSNotSupported); !ok {
		t.Errorf("RHS(%+v): expected error to be of type %T, got %T instead", invalidRHSGetter, ErrLHSNotSupported{}, err)
	}
	if err.Error() == "" {
		t.Errorf("RHS(%+v): unexpected empty error message", invalidRHSGetter)
	}
}

func TestReport(t *testing.T) {
	want := []string{
		"content",
		"type",
	}

	d, err := Diff(
		map[string]interface{}{
			"match":   5,
			"content": 6,
			"type":    8,
		},
		map[string]interface{}{
			"match":   5,
			"content": 7,
			"type":    9.0,
		},
	)
	if err != nil {
		t.Errorf("Diff(...): unexpected error: %s", err)
		return
	}
	ss, err := Report(d, testOutput)
	if err != nil {
		t.Errorf("Report(Diff(...), %+v): unexpected error: %s", testOutput, err)
		return
	}

	if len(ss) != len(want) {
		t.Errorf("len(Report(Diff(...), %+v)) = %d, expected %d", testOutput, len(ss), len(want))
		return
	}

	for i, s := range ss {
		if !strings.Contains(s, want[i]) {
			t.Errorf("Report(Diff(...), %+v)[%d] = %q, should contain %q", testOutput, i, s, want[i])
		}
	}
}

func testStrings(context string, t *testing.T, wants [][]string, ss []string, indented string) {
	for i, want := range wants {
		s := ss[i]

		for i, needle := range want {
			if !strings.Contains(s, needle) {
				t.Errorf("%s: typ.Strings()[%d] = %q, expected it to contain %q", context, i, ss[i], needle)
			}
			if !strings.Contains(indented, needle) {
				t.Errorf(
					"%s: typ.StringIndent(%q, %q, %+v) = %q, expected it to contain %q",
					context, testKey, testPrefix, testOutput, indented, needle,
				)
			}
		}
	}
}

func TestIsScalar(t *testing.T) {
	d, err := Diff(42, 23)
	if err != nil {
		t.Errorf("Diff(42, 23): unexpected error: %s", err)
		return
	}
	if !IsScalar(d) {
		t.Error("IsScalar(Diff(42, 23)) = false, expected true")
	}
}

func TestIsTypes(t *testing.T) {
	d, err := Diff(42, []string{"hop"})
	if err != nil {
		t.Errorf("Diff(42, \"hop\"): unexpected error: %s", err)
		return
	}
	if !IsTypes(d) {
		t.Error("IsTypes(Diff(42, \"hop\")) = false, expected true")
	}
}

func TestIsIgnore(t *testing.T) {
	d, err := Ignore()
	if err != nil {
		t.Errorf("Ignore(): unexpected error: %s", err)
		return
	}
	if !IsIgnore(d) {
		t.Error("IsIgnore(Ignore()) = false, expected true")
	}
}

func TestIsMap(t *testing.T) {
	d, err := Diff(map[int]bool{0: true, 1: false}, map[int]bool{0: false, 1: true})
	if err != nil {
		t.Errorf("Diff(map[int]bool, map[int]bool): unexpected error: %s", err)
		return
	}
	if !IsMap(d) {
		t.Error("IsMap(Diff(map[int]bool, map[int]bool)) = false, expected true")
	}
}

func TestIsSlice(t *testing.T) {
	d, err := Diff([]int{0, 1}, []int{2, 3})
	if err != nil {
		t.Errorf("Diff([]int, []int): unexpected error: %s", err)
		return
	}
	if !IsSlice(d) {
		t.Error("IsSlice(Diff(map{...}, map{...})) = false, expected true")
	}
}

func TestIsStream(t *testing.T) {
	d, err := Diff(mockStream(0, 1), mockStream(2, 3))
	if err != nil {
		t.Errorf("Diff(Stream, Stream): unexpected error: %s", err)
		return
	}
	if !IsStream(d) {
		t.Error("IsStream(Diff(map{...}, map{...})) = false, expected true")
	}
}

func TestValueIsScalar(t *testing.T) {
	for _, test := range []struct {
		In       interface{}
		Expected bool
	}{
		{int(42), true},
		{int8(23), true},
		{"foo", true},
		{true, true},
		{float32(1.2), true},
		{complex(5, -3), true},

		{[]byte("foo"), false},
		{struct{}{}, false},
		{&struct{}{}, false},
		{[]int{1, 2, 3}, false},
		{[3]int{1, 2, 3}, false},
		{map[string]int{"foo": 22}, false},
		{func() {}, false},
		{make(chan struct{}), false},
	} {
		v := reflect.ValueOf(test.In)
		got := valueIsScalar(v)
		if got != test.Expected {
			t.Errorf("valueIsScalar(%T) = %v, expected %v", test.In, got, test.Expected)
		}
	}
}

func TestValueIsStream(t *testing.T) {
	for _, test := range []struct {
		In       interface{}
		Expected bool
	}{
		{nil, false},
		{42, false},
		{[]int{4, 2}, false},
		{func() {}, false},
		{mockStream(), true},
	} {
		v := reflect.ValueOf(test.In)
		got := valueIsStream(v)
		if got != test.Expected {
			t.Errorf("valueIsStream(%T) = %v, expected %v", test.In, got, test.Expected)
		}
	}
}

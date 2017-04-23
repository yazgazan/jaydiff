package diff

import (
	"fmt"
	"strings"
	"testing"
)

func TestDiff(t *testing.T) {
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
			LHS:  map[int][]int{1: []int{2, 3}, 2: []int{3, 4}},
			RHS:  map[int][]int{1: []int{2, 3}, 2: []int{3, 4}},
			Want: Identical,
		},
		{
			LHS:  map[int][]int{1: []int{2, 3}, 2: []int{3, 4}},
			RHS:  map[int][]int{1: []int{2, 3}, 2: []int{3, 5}},
			Want: ContentDiffer,
		},
		{LHS: []interface{}{1, 2, 3}, RHS: []interface{}{1, 2, 3}, Want: Identical},
		{LHS: []interface{}{1, 2, 3}, RHS: []interface{}{1, 2, 3.3}, Want: ContentDiffer},
		{LHS: []interface{}(nil), RHS: []interface{}{1, 2, 3.3}, Want: ContentDiffer},
		{LHS: []int(nil), RHS: []int{}, Want: Identical},
		{LHS: func() {}, RHS: func() {}, Want: TypesDiffer, Error: true},
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
			t.Logf("LHS: %+#v\n", test.RHS)
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
	LHS  interface{}
	RHS  interface{}
	Want [][]string
	Type
}

const (
	testKey    = "(key)"
	testPrefix = "(prefix)"
)

var testOutput = Output{ShowTypes: true}

func TestTypes(t *testing.T) {
	for _, test := range []stringTest{
		{
			LHS: 4,
			RHS: 2.1,
			Want: [][]string{
				[]string{"int", "4"},
				[]string{"float64", "2.1"},
			},
		},
	} {
		typ := &Types{test.LHS, test.RHS}

		if typ.Diff() != TypesDiffer {
			t.Errorf("Types.Diff() = %q, expected %q", typ.Diff(), TypesDiffer)
		}

		ss := typ.Strings()
		indented := typ.StringIndent(testKey, testPrefix, testOutput)
		testStrings("TestTypes", t, test, ss, indented)
	}
}

func TestScalar(t *testing.T) {
	for _, test := range []stringTest{
		{
			LHS: 4,
			RHS: 4,
			Want: [][]string{
				[]string{"int", "4"},
			},
			Type: Identical,
		},
		{
			LHS: 4,
			RHS: 2,
			Want: [][]string{
				[]string{"int", "4"},
				[]string{"int", "2"},
			},
			Type: ContentDiffer,
		},
		{
			LHS: 4,
			RHS: 2.1,
			Want: [][]string{
				[]string{"int", "4"},
				[]string{"float64", "2.1"},
			},
			Type: TypesDiffer,
		},
	} {
		typ := &Scalar{test.LHS, test.RHS}

		if typ.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", typ.Diff(), test.Type)
		}

		ss := typ.Strings()
		indented := typ.StringIndent(testKey, testPrefix, testOutput)
		testStrings("TestScalar", t, test, ss, indented)
	}
}

func TestSlice(t *testing.T) {
	for _, test := range []stringTest{
		{
			LHS: []int{1, 2},
			RHS: []int{1, 2},
			Want: [][]string{
				[]string{"int", "1", "2"},
			},
			Type: Identical,
		},
		{
			LHS: []int{1},
			RHS: []int{},
			Want: [][]string{
				[]string{},
				[]string{"-", "int", "1"},
				[]string{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: []int{},
			RHS: []int{2},
			Want: [][]string{
				[]string{},
				[]string{"+", "int", "2"},
				[]string{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: []int{1, 2},
			RHS: []float64{1.1, 2.1},
			Want: [][]string{
				[]string{"-", "int", "1", "2"},
				[]string{"+", "float64", "1.1", "2.1"},
			},
			Type: TypesDiffer,
		},
		{
			LHS: []int{1, 3},
			RHS: []int{1, 2},
			Want: [][]string{
				[]string{},
				[]string{"int", "1"},
				[]string{"-", "int", "3"},
				[]string{"+", "int", "2"},
				[]string{},
			},
			Type: ContentDiffer,
		},
	} {
		typ, err := NewSlice(test.LHS, test.RHS)

		if err != nil {
			t.Errorf("NewSlice(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
			continue
		}
		if typ.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", typ.Diff(), test.Type)
		}

		ss := typ.Strings()
		indented := typ.StringIndent(testKey, testPrefix, testOutput)
		testStrings("TestSlice", t, test, ss, indented)
	}

	invalid, err := NewSlice(nil, nil)
	if invalidErr, ok := err.(InvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewSlice(nil, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewSlice(nil, nil): expected InvalidType error, got %s", err)
	}
	ss := invalid.Strings()
	if len(ss) != 0 {
		t.Errorf("len(invalidSlice.Strings()) = %d, expected 0", len(ss))
	}

	indented := invalid.StringIndent(testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("invalidSlice.StringIndent(%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
	}

	invalid, err = NewSlice([]int{}, nil)
	if invalidErr, ok := err.(InvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewSlice([]int{}, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewSlice([]int{}, nil): expected InvalidType error, got %s", err)
	}
	ss = invalid.Strings()
	if len(ss) != 0 {
		t.Errorf("len(invalidSlice.Strings()) = %d, expected 0", len(ss))
	}

	indented = invalid.StringIndent(testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("invalidSlice.StringIndent(%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
	}
}

func TestMap(t *testing.T) {
	for i, test := range []stringTest{
		{
			LHS: map[int]int{1: 2, 3: 4},
			RHS: map[int]int{1: 2, 3: 4},
			Want: [][]string{
				[]string{"int", "1", "2", "3", "4"},
			},
			Type: Identical,
		},
		{
			LHS: map[int]int{1: 2},
			RHS: map[int]float64{1: 3.1},
			Want: [][]string{
				[]string{"-", "int", "1", "2"},
				[]string{"+", "float64", "3", "4"},
			},
			Type: TypesDiffer,
		},
		{
			LHS: map[int]int{1: 2},
			RHS: map[int]int{1: 3},
			Want: [][]string{
				[]string{},
				[]string{"-", "int", "1", "2"},
				[]string{"+", "int", "1", "3"},
				[]string{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: map[int]int{1: 2, 2: 3},
			RHS: map[int]int{1: 3, 2: 3},
			Want: [][]string{
				[]string{},
				[]string{"-", "int", "1", "2"},
				[]string{"+", "int", "1", "3"},
				[]string{"int", "2", "3"},
				[]string{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: map[int]int{1: 2},
			RHS: map[int]int{},
			Want: [][]string{
				[]string{},
				[]string{"-", "int", "1", "2"},
				[]string{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: map[int]int{},
			RHS: map[int]int{1: 2},
			Want: [][]string{
				[]string{},
				[]string{"+", "int", "1", "2"},
				[]string{},
			},
			Type: ContentDiffer,
		},
	} {
		m, err := NewMap(test.LHS, test.RHS)

		if err != nil {
			t.Errorf("NewMap(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
			continue
		}
		if m.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", m.Diff(), test.Type)
		}

		ss := m.Strings()
		indented := m.StringIndent(testKey, testPrefix, testOutput)
		testStrings(fmt.Sprintf("TestMap[%d]", i), t, test, ss, indented)
	}

	invalid, err := NewMap(nil, nil)
	if invalidErr, ok := err.(InvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewMap(nil, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewMap(nil, nil): expected InvalidType error, got %s", err)
	}
	ss := invalid.Strings()
	if len(ss) != 0 {
		t.Errorf("len(invalidMap.Strings()) = %d, expected 0", len(ss))
	}

	indented := invalid.StringIndent(testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("invalidMap.StringIndent(%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
	}

	invalid, err = NewMap(map[int]int{}, nil)
	if invalidErr, ok := err.(InvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewMap(map[int]int{}, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewMap(map[int]int{}, nil): expected InvalidType error, got %s", err)
	}
	ss = invalid.Strings()
	if len(ss) != 0 {
		t.Errorf("len(invalidMap.Strings()) = %d, expected 0", len(ss))
	}

	indented = invalid.StringIndent(testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("invalidMap.StringIndent(%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
	}
}

func TestIgnore(t *testing.T) {
	ignore := Ignore{}

	if ignore.Diff() != Identical {
		t.Errorf("Ignore{}.Diff() = %q, expected %q", ignore.Diff(), Identical)
	}
	if len(ignore.Strings()) != 0 {
		t.Errorf("len(Ignore{}.Strings()) = %d, expected 0", len(ignore.Strings()))
	}
	if indented := ignore.StringIndent(testKey, testPrefix, testOutput); indented != "" {
		t.Errorf("Ignore{}.StringIndent(...) = %q, expected %q", indented, "")
	}
}

func testStrings(context string, t *testing.T, test stringTest, ss []string, indented string) {
	for i, want := range test.Want {
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

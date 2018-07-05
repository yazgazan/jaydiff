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

func TestSlice(t *testing.T) {
	for _, test := range []stringTest{
		{
			LHS: []int{1, 2},
			RHS: []int{1, 2},
			Want: [][]string{
				{"int", "1", "2"},
			},
			WantJSON: [][]string{
				{"1", "2"},
			},
			Type: Identical,
		},
		{
			LHS: []int{1},
			RHS: []int{},
			Want: [][]string{
				{},
				{"-", "int", "1"},
				{},
			},
			WantJSON: [][]string{
				{},
				{"-", "1"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: []int{},
			RHS: []int{2},
			Want: [][]string{
				{},
				{"+", "int", "2"},
				{},
			},
			WantJSON: [][]string{
				{},
				{"+", "2"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: []int{1, 2},
			RHS: []float64{1.1, 2.1},
			Want: [][]string{
				{"-", "int", "1", "2"},
				{"+", "float64", "1.1", "2.1"},
			},
			WantJSON: [][]string{
				{"-", "1", "2"},
				{"+", "1.1", "2.1"},
			},
			Type: TypesDiffer,
		},
		{
			LHS: []int{1, 3},
			RHS: []int{1, 2},
			Want: [][]string{
				{},
				{"int", "1"},
				{"-", "int", "3"},
				{"+", "int", "2"},
				{},
			},
			WantJSON: [][]string{
				{},
				{"1"},
				{"-", "3"},
				{"+", "2"},
				{},
			},
			Type: ContentDiffer,
		},
	} {
		typ, err := newSlice(defaultConfig(), test.LHS, test.RHS, &visited{})

		if err != nil {
			t.Errorf("NewSlice(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
			continue
		}
		if typ.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", typ.Diff(), test.Type)
		}

		ss := typ.Strings()
		indented := typ.StringIndent(testKey, testPrefix, testOutput)
		testStrings("TestSlice", t, test.Want, ss, indented)

		indentedJSON := typ.StringIndent(testKey, testPrefix, testJSONOutput)
		testStrings("TestSlice", t, test.WantJSON, ss, indentedJSON)
	}

	invalid, err := newSlice(defaultConfig(), nil, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
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

	invalid, err = newSlice(defaultConfig(), []int{}, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
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

func TestSliceMyers(t *testing.T) {
	c := defaultConfig()
	c = UseSliceMyers()(c)

	for _, test := range []stringTest{
		{
			LHS: []int{1, 2},
			RHS: []int{1, 2},
			Want: [][]string{
				{"int", "1", "2"},
			},
			Type: Identical,
		},
		{
			LHS: []int{1},
			RHS: []int{},
			Want: [][]string{
				{},
				{"-", "int", "1"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: []int{},
			RHS: []int{2},
			Want: [][]string{
				{},
				{"+", "int", "2"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: []int{1, 2},
			RHS: []float64{1.1, 2.1},
			Want: [][]string{
				{"-", "int", "1", "2"},
				{"+", "float64", "1.1", "2.1"},
			},
			Type: TypesDiffer,
		},
		{
			LHS: []int{1, 3},
			RHS: []int{1, 2},
			Want: [][]string{
				{},
				{"int", "1"},
				{"-", "int", "3"},
				{"+", "int", "2"},
				{},
			},
			Type: ContentDiffer,
		},
	} {
		typ, err := c.sliceFn(c, test.LHS, test.RHS, &visited{})

		if err != nil {
			t.Errorf("NewMyersSlice(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
			continue
		}
		if typ.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", typ.Diff(), test.Type)
		}

		ss := typ.Strings()
		indented := typ.StringIndent(testKey, testPrefix, testOutput)
		testStrings("TestSlice", t, test.Want, ss, indented)
	}

	invalid, err := c.sliceFn(c, nil, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewMyersSlice(nil, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewMyersSlice(nil, nil): expected InvalidType error, got %s", err)
	}
	ss := invalid.Strings()
	if len(ss) != 0 {
		t.Errorf("len(invalidSlice.Strings()) = %d, expected 0", len(ss))
	}

	indented := invalid.StringIndent(testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("invalidSlice.StringIndent(%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
	}

	invalid, err = c.sliceFn(c, []int{}, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewMyersSlice([]int{}, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewMyersSlice([]int{}, nil): expected InvalidType error, got %s", err)
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
				{"int", "1", "2", "3", "4"},
			},
			WantJSON: [][]string{
				{"1", "2", "3", "4"},
			},
			Type: Identical,
		},
		{
			LHS: map[int]int{1: 2},
			RHS: map[int]float64{1: 3.1},
			Want: [][]string{
				{"-", "int", "1", "2"},
				{"+", "float64", "3"},
			},
			WantJSON: [][]string{
				{"-", "1", "2"},
				{"+", "3"},
			},
			Type: TypesDiffer,
		},
		{
			LHS: map[int]int{1: 2},
			RHS: map[int]int{1: 3},
			Want: [][]string{
				{},
				{"-", "int", "1", "2"},
				{"+", "int", "1", "3"},
				{},
			},
			WantJSON: [][]string{
				{},
				{"-", "1", "2"},
				{"+", "1", "3"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: map[int]int{1: 2, 2: 3},
			RHS: map[int]int{1: 3, 2: 3},
			Want: [][]string{
				{},
				{"-", "int", "1", "2"},
				{"+", "int", "1", "3"},
				{"int", "2", "3"},
				{},
			},
			WantJSON: [][]string{
				{},
				{"-", "1", "2"},
				{"+", "1", "3"},
				{"2", "3"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: map[int]int{1: 2},
			RHS: map[int]int{},
			Want: [][]string{
				{},
				{"-", "int", "1", "2"},
				{},
			},
			WantJSON: [][]string{
				{},
				{"-", "1", "2"},
				{},
			},
			Type: ContentDiffer,
		},
		{
			LHS: map[int]int{},
			RHS: map[int]int{1: 2},
			Want: [][]string{
				{},
				{"+", "int", "1", "2"},
				{},
			},
			WantJSON: [][]string{
				{},
				{"+", "1", "2"},
				{},
			},
			Type: ContentDiffer,
		},
	} {
		m, err := newMap(defaultConfig(), test.LHS, test.RHS, &visited{})

		if err != nil {
			t.Errorf("NewMap(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
			continue
		}
		if m.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", m.Diff(), test.Type)
		}

		ss := m.Strings()
		indented := m.StringIndent(testKey, testPrefix, testOutput)
		testStrings(fmt.Sprintf("TestMap[%d]", i), t, test.Want, ss, indented)

		indentedJSON := m.StringIndent(testKey, testPrefix, testJSONOutput)
		testStrings(fmt.Sprintf("TestMap[%d]", i), t, test.WantJSON, ss, indentedJSON)
	}

	invalid, err := newMap(defaultConfig(), nil, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
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

	invalid, err = newMap(defaultConfig(), map[int]int{}, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
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

func TestCircular(t *testing.T) {
	first := map[int]interface{}{}
	second := map[int]interface{}{
		0: first,
	}
	first[0] = second
	notCyclic := map[int]interface{}{
		0: map[int]interface{}{
			0: map[int]interface{}{
				0: "foo",
			},
		},
	}
	emptySlice := map[int]interface{}{
		0: []interface{}{},
	}
	emptySlice[1] = emptySlice[0]
	emptySliceNotRepeating := map[int]interface{}{
		0: []interface{}{},
		1: []interface{}{},
	}
	emptyMap := map[int]interface{}{
		0: map[int]interface{}{},
	}
	emptyMap[1] = emptyMap[0]
	emptyMapNotRepeating := map[int]interface{}{
		0: map[int]interface{}{},
		1: map[int]interface{}{},
	}

	for _, test := range []struct {
		lhs       interface{}
		rhs       interface{}
		wantError bool
	}{
		{lhs: first, rhs: first, wantError: true},
		{lhs: first, rhs: second, wantError: true},
		{lhs: first, rhs: second, wantError: true},
		{lhs: first, rhs: notCyclic, wantError: true},
		{lhs: notCyclic, rhs: first, wantError: true},
		{lhs: notCyclic, rhs: notCyclic},
		{lhs: emptySlice, rhs: emptySliceNotRepeating},
		{lhs: emptySliceNotRepeating, rhs: emptySlice},
		{lhs: emptyMap, rhs: emptyMapNotRepeating},
		{lhs: emptyMapNotRepeating, rhs: emptyMap},
	} {
		d, err := Diff(test.lhs, test.rhs)

		if test.wantError && (err == nil || err != ErrCyclic) {
			t.Errorf("Expected error %q, got %q", ErrCyclic, err)
		}
		if !test.wantError && err != nil {
			t.Errorf("Unexpected error %q", err)
		}

		if test.wantError && d.Diff() != ContentDiffer {
			t.Errorf("Expected Diff() to be %s, got %s", ContentDiffer, d.Diff())
		}
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

package diff

import (
	"fmt"
	"reflect"
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
		indented := stringIndent(typ, testKey, testPrefix, testOutput)
		testStrings("TestTypes", t, test, ss, indented)
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
		indented := stringIndent(typ, testKey, testPrefix, testOutput)
		testStrings("TestScalar", t, test, ss, indented)
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
		typ, err := newSlice(test.LHS, test.RHS, &visited{})

		if err != nil {
			t.Errorf("NewSlice(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
			continue
		}
		if typ.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", typ.Diff(), test.Type)
		}

		ss := getStrings(typ)
		indented := stringIndent(typ, testKey, testPrefix, testOutput)
		testStrings("TestSlice", t, test, ss, indented)
	}

	invalid, err := newSlice(nil, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewSlice(nil, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewSlice(nil, nil): expected InvalidType error, got %s", err)
	}
	ss := getStrings(invalid)
	if len(ss) != 0 {
		t.Errorf("len(invalidSlice.Strings()) = %d, expected 0", len(ss))
	}

	indented := stringIndent(invalid, testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("stringIndent(invalidSlice,%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
	}

	invalid, err = newSlice([]int{}, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewSlice([]int{}, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewSlice([]int{}, nil): expected InvalidType error, got %s", err)
	}
	ss = getStrings(invalid)
	if len(ss) != 0 {
		t.Errorf("len(invalidSlice.Strings()) = %d, expected 0", len(ss))
	}

	indented = stringIndent(invalid, testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("stringIndent(invalidSlice,%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
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
			Type: Identical,
		},
		{
			LHS: map[int]int{1: 2},
			RHS: map[int]float64{1: 3.1},
			Want: [][]string{
				{"-", "int", "1", "2"},
				{"+", "float64", "3", "4"},
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
			Type: ContentDiffer,
		},
	} {
		m, err := newMap(test.LHS, test.RHS, &visited{})

		if err != nil {
			t.Errorf("NewMap(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
			continue
		}
		if m.Diff() != test.Type {
			t.Errorf("Types.Diff() = %q, expected %q", m.Diff(), test.Type)
		}

		ss := getStrings(m)
		indented := stringIndent(m, testKey, testPrefix, testOutput)
		testStrings(fmt.Sprintf("TestMap[%d]", i), t, test, ss, indented)
	}

	invalid, err := newMap(nil, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewMap(nil, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewMap(nil, nil): expected InvalidType error, got %s", err)
	}
	ss := getStrings(invalid)
	if len(ss) != 0 {
		t.Errorf("len(invalidMap.Strings()) = %d, expected 0", len(ss))
	}

	indented := stringIndent(invalid, testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("stringIndent(invalidMap,%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
	}

	invalid, err = newMap(map[int]int{}, nil, &visited{})
	if invalidErr, ok := err.(errInvalidType); ok {
		if !strings.Contains(invalidErr.Error(), "nil") {
			t.Errorf("NewMap(map[int]int{}, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("NewMap(map[int]int{}, nil): expected InvalidType error, got %s", err)
	}
	ss = getStrings(invalid)
	if len(ss) != 0 {
		t.Errorf("len(invalidMap.Strings()) = %d, expected 0", len(ss))
	}

	indented = stringIndent(invalid, testKey, testPrefix, testOutput)
	if indented != "" {
		t.Errorf("stringIndent(invalidMap,%q, %q, %+v) = %q, expected %q", testKey, testPrefix, testOutput, indented, "")
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
	if len(getStrings(ignoreDiff)) != 0 {
		t.Errorf("len(NewIgnore().Strings()) = %d, expected 0", len(getStrings(ignoreDiff)))
	}
	if indented := stringIndent(ignoreDiff, testKey, testPrefix, testOutput); indented != "" {
		t.Errorf("stringIndent(NewIgnore(),...) = %q, expected %q", indented, "")
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

type testDiffStringer string

func (s testDiffStringer) Diff() Type {
	return Identical
}

func (s testDiffStringer) Strings() []string {
	return []string{string(s)}
}

func (s testDiffStringer) StringIndent(key, prefix string, conf Output) string {
	return prefix + key + string(s)
}

type testStringer string

func (s testStringer) String() string {
	return string(s)
}

func TestGetStrings(t *testing.T) {
	for _, test := range []struct {
		input interface{}
		want  interface{}
	}{
		{input: testDiffStringer("foo"), want: []string{"foo"}},
		{input: testStringer("bar"), want: []string{"bar"}},
		{input: "fiz", want: []string{"fiz"}},
	} {
		ss := getStrings(test.input)
		if !reflect.DeepEqual(ss, test.want) {
			t.Errorf("getStrings(%+v) = %v, expected %v", test.input, ss, test.want)
		}
	}
}

func TestStringIndent(t *testing.T) {
	for _, test := range []struct {
		input interface{}
		want  string
	}{
		{input: testDiffStringer("foo"), want: "{prefix}{key}foo"},
		{input: testStringer("bar"), want: " {prefix}{key}bar"},
		{input: "fiz", want: " {prefix}{key}fiz"},
	} {
		s := stringIndent(test.input, "{key}", "{prefix}", Output{})
		if s != test.want {
			t.Errorf("stringIndent(%+v, '{key}', '{prefix}') = %q, expected %q", test.input, s, test.want)
		}
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
					"%s: stringIndent(typ,%q, %q, %+v) = %q, expected it to contain %q",
					context, testKey, testPrefix, testOutput, indented, needle,
				)
			}
		}
	}
}

package diff

import (
	"strings"
	"testing"
)

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

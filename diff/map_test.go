package diff

import (
	"fmt"
	"strings"
	"testing"
)

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
			t.Errorf("newMap(%+v, %+v): unexpected error: %q", test.LHS, test.RHS, err)
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
			t.Errorf("newMap(nil, nil): unexpected format for InvalidType error: got %s", err)
		}
	} else {
		t.Errorf("newMap(nil, nil): expected InvalidType error, got %s", err)
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

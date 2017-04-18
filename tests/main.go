package main

import (
	"fmt"
	"github.com/Pimmr/json-diff/diff"
)

func main() {
	tests := []struct {
		lhs interface{}
		rhs interface{}
	}{
		{42, 4.2},
		{10, "toto"},
		{4, 4},
		{2, 5},
		{[]int{1, 2}, []int{3, 4}},
		{[]int{1, 2}, []int{1, 2}},
		{[]int{1, 2, 4}, []int{1, 2}},
		{[]int{1, 2}, []int{1, 2, 23}},
		{
			[][]int{{1, 2}, {3, 4}},
			[][]int{{1, 2}, {6, 8}},
		},
		{
			[][][]int{
				{{1, 2}, {3, 4}},
				{{5, 6}, {7, 8}},
			},
			[][][]int{
				{{1, 2}, {3, 4}},
				{{5, -4}, {7, 8}},
			},
		},
		{
			map[string]int{"toto": 42},
			map[string]int{"toto": 42},
		},
		{
			map[string]int{"toto": 42},
			map[string]int{"toto": 44},
		},
		{
			map[string][]int{"toto": {1, 2}},
			map[string][]int{"toto": {1, 3}},
		},
		{
			map[string][]int{"toto": {1, 2}},
			map[string][]int{"tata": {1, 3}},
		},
		{
			map[string]interface{}{
				"titi": 42,
				"toto": "hop",
				"yo":   4.2,
				"haha": 1,
			},
			map[string]interface{}{
				"titi": 43,
				"toto": 21,
				"yo":   4.2,
				"plop": []int{1, 2},
			},
		},
		{
			map[string]interface{}{
				"titi": 42,
				"toto": "hop",
				"yo":   4.2,
			},
			map[string][]int{"tata": {1, 3}},
		},
	}

	for _, test := range tests {
		d, _ := diff.Diff(test.lhs, test.rhs)

		fmt.Printf("%v, %v:\n", test.lhs, test.rhs)
		fmt.Println(d.StringIndent("", "", diff.Output{
			Indent:    "    ",
			Colorized: true,
			ShowTypes: true,
		}))
		fmt.Println()
	}
}

package diff_test

import (
	"fmt"
	"github.com/tomertwist/jaydiff/diff"
	"strings"
)

func ExampleDiff() {
	lhs := map[string]interface{}{
		"a": 42,
		"b": []int{1, 2},
		"c": "abc",
	}
	rhs := map[string]interface{}{
		"a": 21,
		"b": []int{1, 2, 3},
		"c": "abc",
	}

	d, _ := diff.Diff(lhs, rhs)

	fmt.Println(d.StringIndent("", "", diff.Output{
		Indent:    "  ",
		ShowTypes: true,
	}))
}

func ExampleReport() {
	lhs := map[string]interface{}{
		"a": 42,
		"b": []int{1, 2},
		"c": "abc",
	}
	rhs := map[string]interface{}{
		"a": 21,
		"b": []int{1, 2, 3},
		"c": "abc",
	}

	d, _ := diff.Diff(lhs, rhs)

	reports, _ := diff.Report(d, diff.Output{
		Indent:    "  ",
		ShowTypes: true,
	})

	for _, report := range reports {
		fmt.Println(report)
	}
}

func ExampleWalk() {
	lhs := map[string]interface{}{
		"a":               42,
		"b":               []int{1, 2},
		"will_be_ignored": []int{3, 4},
		"c":               "abc",
	}
	rhs := map[string]interface{}{
		"a":               41,
		"b":               []int{1, 2},
		"will_be_ignored": 5,
		"c":               "abc",
		"exess_key":       "will be ignored",
	}

	d, _ := diff.Diff(lhs, rhs)
	fmt.Println("Before:")
	fmt.Println(d.StringIndent("", "", diff.Output{
		Indent:    "  ",
		ShowTypes: true,
	}))

	d, _ = diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) (diff.Differ, error) {
		if strings.HasSuffix(path, ".will_be_ignored") || diff.IsExcess(d) {
			return diff.Ignore()
		}

		return nil, nil
	})

	fmt.Println("After:")
	fmt.Println(d.StringIndent("", "", diff.Output{
		Indent:    "  ",
		ShowTypes: true,
	}))
}

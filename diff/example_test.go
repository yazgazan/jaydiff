package diff_test

import (
	"fmt"
	"strings"

	"github.com/yazgazan/jaydiff/diff"
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

	// Output:
	// map[string]interface {} map[
	// -  a: int 42
	// +  a: int 21
	//    b: []int [
	//      int 1
	//      int 2
	// +    int 3
	//    ]
	//    c: string abc
	//  ]
}

func ExampleDiff_struct() {
	type subStruct struct {
		Hello int
		World float64
	}
	type structA struct {
		Foo int
		Bar string
		Baz subStruct

		priv int
	}
	type structB struct {
		Foo int
		Bar string
		Baz subStruct

		priv int
	}

	lhs := structA{
		Foo: 42,
		Bar: "hello",
		Baz: subStruct{
			Hello: 11,
			World: 3.5,
		},
		priv: 0,
	}
	rhs := structB{
		Foo: 21,
		Bar: "hello",
		Baz: subStruct{
			Hello: 11,
			World: 3.5,
		},
		priv: 1,
	}

	d, err := diff.Diff(lhs, rhs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(d.StringIndent("", "", diff.Output{
		Indent:    "  ",
		ShowTypes: true,
	}))

	// Output:
	// diff_test.structA map[
	//    Bar: string hello
	//    Baz: diff_test.subStruct {11 3.5}
	// -  Foo: int 42
	// +  Foo: int 21
	//  ]
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

	// Output:
	// - .a: int 42
	// + .a: int 21
	// + .b[2]: int 3
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

	// Output:
	// Before:
	//  map[string]interface {} map[
	// -  a: int 42
	// +  a: int 41
	//    b: []int [1 2]
	//    c: string abc
	// +  exess_key: string will be ignored
	// -  will_be_ignored: []int [3 4]
	// +  will_be_ignored: int 5
	//  ]
	// After:
	//  map[string]interface {} map[
	// -  a: int 42
	// +  a: int 41
	//    b: []int [1 2]
	//    c: string abc
	//  ]
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Pimmr/json-diff/diff"
)

func main() {
	lhs, err := read("tests/lhs.json")
	if err != nil {
		panic(err)
	}

	rhs, err := read("tests/rhs.json")
	if err != nil {
		panic(err)
	}

	d, err := diff.Diff(lhs, rhs)
	if err != nil {
		panic(err)
	}

	diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) error {
		if shouldIgnore(path) {
			switch t := parent.(type) {
			case diff.Map:
				for k, subd := range t.Diffs {
					if d == subd {
						fmt.Println("ignoring")
						t.Diffs[k] = &diff.Ignore{}
					}
				}
			case diff.Slice:
				for i, subd := range t.Diffs {
					if d == subd {
						fmt.Println("ignoring")
						t.Diffs[i] = &diff.Ignore{}
					}
				}
			}
		}
		fmt.Printf("%s: %s (ignore: %v)\n", path, d.Diff(), shouldIgnore(path))

		return nil
	})

	fmt.Println(d.StringIndent("", "", diff.Output{
		Indent:    "\t",
		Colorized: true,
	}))

	if d.Diff() != diff.Identical {
		os.Exit(1)
	}
}

func shouldIgnore(path string) bool {
	switch path {
	case ".e", ".d", ".c.a", ".c.c":
		return true
	}

	return strings.HasSuffix(path, ".b")
}

func read(fname string) (interface{}, error) {
	var err error
	var v interface{}

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &v)

	return v, err
}

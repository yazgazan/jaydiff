package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/yazgazan/jaydiff/diff"
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

	d, err = diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) (diff.Differ, error) {
		if shouldIgnore(path) {
			fmt.Println("ignoring")
			return &diff.Ignore{}, nil
		}
		fmt.Printf("%s: %s (ignore: %v)\n", path, d.Diff(), shouldIgnore(path))

		return nil, nil
	})
	if err != nil {
		panic(err)
	}

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

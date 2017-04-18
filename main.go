package main

import (
	"encoding/json"
	"fmt"
	"github.com/Pimmr/json-diff/diff"
	"io/ioutil"
	"os"
)

const (
	StatusDiffMismatch   = 1
	StatusReadError      = 3
	StatusUnmarshalError = 4
	StatusDiffError      = 5
)

func main() {
	var err error
	conf := readConfig()

	lhs := parseFile(conf.lhsFile)
	rhs := parseFile(conf.rhsFile)

	d, err := diff.Diff(lhs, rhs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: diff failed: %s", err)
		os.Exit(StatusDiffError)
	}

	fmt.Println(d.StringIndent("", "", conf.Output))
	if d.Diff() != diff.Identical {
		os.Exit(StatusDiffMismatch)
	}
}

func parseFile(fname string) interface{} {
	var err error
	var val interface{}

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read %s\n", fname)
		os.Exit(StatusReadError)
	}
	err = json.Unmarshal(b, &val)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot parse %s: %s\n", fname, err)
		os.Exit(StatusUnmarshalError)
	}

	return val
}

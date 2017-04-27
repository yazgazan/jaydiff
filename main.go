package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yazgazan/jaydiff/diff"
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

	lhs := parseFile(conf.Files.LHS)
	rhs := parseFile(conf.Files.RHS)

	d, err := diff.Diff(lhs, rhs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: diff failed: %s", err)
		os.Exit(StatusDiffError)
	}

	d, err = pruneIgnore(d, conf.Ignore)

	if conf.OutputReport {
		ss, err := diff.Report(d, diff.Output(conf.Output))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to generate report: %s", err)
			os.Exit(StatusDiffError)
		}
		for _, s := range ss {
			fmt.Println(s)
		}
	} else {
		fmt.Println(d.StringIndent("", "", diff.Output(conf.Output)))
	}
	if d.Diff() != diff.Identical {
		os.Exit(StatusDiffMismatch)
	}
}

func pruneIgnore(d diff.Differ, ignore Patterns) (diff.Differ, error) {
	return diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) (diff.Differ, error) {
		if ignore.Match(path) {
			return diff.Ignore()
		}
		return nil, nil
	})
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

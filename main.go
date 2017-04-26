package main

import (
	"encoding/json"
	"fmt"
	"github.com/yazgazan/jaydiff/diff"
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

	d, err = pruneIgnore(d, conf.ignore)

	if conf.outputReport {
		errs, err := diff.Report(d, conf.Output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to generate report: %s", err)
			os.Exit(StatusDiffError)
		}
		for _, e := range errs {
			fmt.Println(e.Error())
		}
	} else {
		fmt.Println(d.StringIndent("", "", conf.Output))
	}
	if d.Diff() != diff.Identical {
		os.Exit(StatusDiffMismatch)
	}
}

func pruneIgnore(d diff.Differ, ignore patterns) (diff.Differ, error) {
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

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

	err = pruneIgnore(d, conf.ignore)

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

func pruneIgnore(d diff.Differ, ignore patterns) error {
	return diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) error {
		if !ignore.Match(path) {
			return nil
		}

		switch t := parent.(type) {
		case diff.Map:
			for k, subd := range t.Diffs {
				if d == subd {
					t.Diffs[k] = &diff.Ignore{}
				}
			}
		case diff.Slice:
			for i, subd := range t.Diffs {
				if d == subd {
					t.Diffs[i] = &diff.Ignore{}
				}
			}
		}

		return nil
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

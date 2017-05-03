package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yazgazan/jaydiff/diff"
)

const (
	statusUsage          = 2
	statusReadError      = 3
	statusUnmarshalError = 4
	statusDiffError      = 5
	statusDiffMismatch   = 6
)

func main() {
	var err error
	conf := readConfig()

	lhs := parseFile(conf.Files.LHS)
	rhs := parseFile(conf.Files.RHS)

	d, err := diff.Diff(lhs, rhs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: diff failed: %s", err)
		os.Exit(statusDiffError)
	}

	d, err = pruneIgnore(d, conf.IgnoreExcess, conf.Ignore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: ignoring failed: %s", err)
		os.Exit(statusDiffError)
	}

	if conf.OutputReport {
		ss, err := diff.Report(d, diff.Output(conf.output))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to generate report: %s", err)
			os.Exit(statusDiffError)
		}
		for _, s := range ss {
			fmt.Println(s)
		}
	} else {
		fmt.Println(d.StringIndent("", "", diff.Output(conf.output)))
	}
	if d.Diff() != diff.Identical {
		os.Exit(statusDiffMismatch)
	}
}

func pruneIgnore(d diff.Differ, ingoreExcess bool, ignore ignorePatterns) (diff.Differ, error) {
	return diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) (diff.Differ, error) {
		if ignore.Match(path) {
			return diff.Ignore()
		}

		if ingoreExcess && diff.IsExcess(d) {
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
		os.Exit(statusReadError)
	}
	err = json.Unmarshal(b, &val)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot parse %s: %s\n", fname, err)
		os.Exit(statusUnmarshalError)
	}

	return val
}

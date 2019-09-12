package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"bufio"
	"log"
	"os"

	"github.com/yazgazan/jaydiff/diff"
)

const (
	statusUsage          = 2
	statusReadError      = 3
	statusUnmarshalError = 4
	statusDiffError      = 5
	statusDiffMismatch   = 6
	statusLinesCountMismatch = 7
)

var (
	// Version is replaced by the tag when creating a new release
	Version = "dev"
)

func main() {
	conf := readConfig()

	if conf.JsonLines {
		os.Exit(compareJsonLines(&conf))
	} else {
		lhs := parseFile(conf.Files.LHS)
		rhs := parseFile(conf.Files.RHS)

		os.Exit(compare(lhs, rhs, &conf))
	}
}

func compareJsonLines(conf *config) int {
	var err error
	var lhs_cnt,rhs_cnt int
	var lhs_file,rhs_file *os.File
	var lhs_val,rhs_val interface{}

	lhs_file, err = os.Open(conf.Files.LHS)
	if err != nil {
		log.Fatal(err)
	}
	defer lhs_file.Close()

	rhs_file, err = os.Open(conf.Files.RHS)
	if err != nil {
		log.Fatal(err)
	}
	defer rhs_file.Close()

	lhs_scanner := bufio.NewScanner(lhs_file)
	rhs_scanner := bufio.NewScanner(rhs_file)

	for {
		if lhs_scanner.Scan() { lhs_cnt+=1 } else { lhs_cnt = -1 }
		if rhs_scanner.Scan() { rhs_cnt+=1 } else { rhs_cnt = -1 }
		if (lhs_cnt < 0 && rhs_cnt < 0) {
			break
		} else if (lhs_cnt < 0 || rhs_cnt < 0) {
			fmt.Fprintf(os.Stderr, "Error: File has different number of liens %d:%d\n", lhs_cnt,rhs_cnt)
			return statusLinesCountMismatch
		}

		err = json.Unmarshal(lhs_scanner.Bytes(), &lhs_val)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot parse %s: %s\n", conf.Files.LHS, err)
			return statusUnmarshalError
		}

		err = json.Unmarshal(rhs_scanner.Bytes(), &rhs_val)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot parse %s: %s\n", conf.Files.RHS, err)
			return statusUnmarshalError
		}

		if rc := compare(lhs_val, rhs_val, conf); rc > 0 {
			return rc
		}
	}
	return 0
}


func compare(lhs, rhs interface{}, conf *config) int {
	d, err := diff.Diff(lhs, rhs, conf.Opts()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: diff failed: %s", err)
		return statusDiffError
	}

	d, err = pruneIgnore(d, conf.IgnoreExcess, conf.IgnoreValues, conf.Ignore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: ignoring failed: %s", err)
		return statusDiffError
	}

	if conf.OutputReport {
		ss, err := diff.Report(d, diff.Output(conf.output))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to generate report: %s", err)
			return statusDiffError
		}
		for _, s := range ss {
			fmt.Println(s)
		}
	} else {
		fmt.Println(d.StringIndent("", "", diff.Output(conf.output)))
	}
	if d.Diff() != diff.Identical {
		 return statusDiffMismatch
	}
	return 0
}

func pruneIgnore(d diff.Differ, ingoreExcess, ignoreValues bool, ignore ignorePatterns) (diff.Differ, error) {
	return diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) (diff.Differ, error) {
		if ignore.Match(path) {
			return diff.Ignore()
		}

		if ingoreExcess && diff.IsExcess(d) {
			return diff.Ignore()
		}

		if ignoreValues && diff.IsScalar(d) && d.Diff() == diff.ContentDiffer {
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

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/yazgazan/jaydiff/diff"
)

const (
	statusUsage           = 2
	statusReadError       = 3
	statusUnexpectedError = 4
	statusDiffError       = 5
	statusDiffMismatch    = 6
)

var (
	// Version is replaced by the tag when creating a new release
	Version = "dev"
)
type nextJson func() (error, interface{}, int)

func main() {
	var (
		lhs_file,rhs_file *os.File
		err error
	)

	conf := readConfig()

	// Open files for reading
	lhs_file, err = os.Open(conf.Files.LHS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read %s: %s\n", conf.Files.LHS, err.Error())
		os.Exit(statusReadError)
	}
	defer lhs_file.Close()
	rhs_file, err = os.Open(conf.Files.RHS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read %s: %s\n", conf.Files.RHS, err.Error())
		os.Exit(statusReadError)
	}
	defer rhs_file.Close()

	//
	if conf.JsonLines {
		os.Exit(compareLoop(getNextJsonByLine(lhs_file), getNextJsonByLine(rhs_file),&conf))
	} else {
		os.Exit(compareLoop(getNextJson(lhs_file), getNextJson(rhs_file),&conf))
	}
}

func compareLoop(lhs_next, rhs_next nextJson, conf *config) int {
	var is_eof bool
	var lhs_err,rhs_err, err error
	var lhs,rhs interface{}
	var lhs_cnt, rhs_cnt, rc, rc_final int

	rc_final = 0
	is_eof = false

	for !is_eof {
		lhs_err, lhs, lhs_cnt = lhs_next()
		rhs_err, rhs, rhs_cnt = rhs_next()

		// If both file reach EOF
		if lhs_err == io.EOF && rhs_err == io.EOF {
			is_eof = true
			continue
		} else if lhs_err != nil && lhs_err != io.EOF {
			fmt.Fprintf(os.Stderr, lhs_err.Error())
			return statusUnexpectedError // TODO: Correct error code?
		} else if rhs_err != nil && rhs_err != io.EOF {
			fmt.Fprintf(os.Stderr, rhs_err.Error())
			return statusUnexpectedError // TODO: Correct error code?
		}

		rc, err = compare(lhs, rhs, lhs_cnt,rhs_cnt, conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return rc
		}
		if rc > 0 {
			rc_final = rc
		}
	}
	return rc_final
}

func getNextJson(file *os.File) nextJson {
	var (
		r *json.Decoder
		cnt int = 0
	)
	r =	json.NewDecoder(file)

	return func() (error, interface{}, int){
		var (
			err error
			i interface{}
		)

		err = r.Decode(&i)
		if err == nil {
			cnt++
		} else {
			i = new(map[string]interface{})
		}

		return err, i, cnt
	}
}

func getNextJsonByLine(file *os.File) nextJson {
	var (
		r *bufio.Scanner
		cnt int = 0
	)
	r = bufio.NewScanner(file)

	return func() (error, interface{}, int) {
		var (
			err error
		  i interface{}
		)
		if r.Scan() {
			cnt++
			err = json.Unmarshal(r.Bytes(), &i)
		} else {
			i = new(map[string]interface{})
			// https://golang.org/pkg/bufio/#Scanner.Scan
			if err = r.Err(); err == nil {
				err = io.EOF
			}
		}

		return err, i, cnt
	}
}

func compare(lhs, rhs interface{},lhs_cnt, rhs_cnt int, conf *config) (int, error) {
	d, err := diff.Diff(lhs, rhs, conf.Opts()...)
	if err != nil {
		return statusDiffError, fmt.Errorf("Error: diff failed: %s", err)
	}

	d, err = pruneIgnore(d, conf.IgnoreExcess, conf.IgnoreValues, conf.Ignore)
	if err != nil {
		return statusDiffError, fmt.Errorf("Error: ignoring failed: %s", err)
	}

	if conf.OutputReport {
		ss, err := diff.Report(d, diff.Output(conf.output))
		if err != nil {
			return statusDiffError, fmt.Errorf("Error: Failed to generate report: %s", err)
		}
		if len(ss) > 0 {
			fmt.Printf("Pos [%d:%d]\n",lhs_cnt,rhs_cnt)
		}
		for _, s := range ss {
			fmt.Println(s)
		}
	} else {
		fmt.Println(d.StringIndent("", "", diff.Output(conf.output)))
	}
	if d.Diff() != diff.Identical {
		 return statusDiffMismatch, nil
	}
	return 0, nil
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

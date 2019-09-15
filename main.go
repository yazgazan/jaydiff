package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

var (
	// Version is replaced by the tag when creating a new release
	Version = "dev"
)

func main() {
	var (
		err                  error
		lhs, rhs             interface{}
		lhsCloser, rhsCloser io.Closer
	)
	conf := readConfig()

	switch conf.Stream {
	case true:
		lhs, lhsCloser = parseStream(conf.Files.LHS, conf.StreamLines)
		defer lhsCloser.Close()
		rhs, rhsCloser = parseStream(conf.Files.RHS, conf.StreamLines)
		defer rhsCloser.Close()
		if conf.StreamValidate {
			lhs = singleValueForValidate(lhs.(diff.Stream), rhs.(HasMore))
		}
	case false:
		lhs = parseFile(conf.Files.LHS)
		rhs = parseFile(conf.Files.RHS)
	}

	d, err := diff.Diff(lhs, rhs, conf.Opts()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: diff failed: %s\n", err)
		os.Exit(statusDiffError)
	}

	d, err = pruneIgnore(
		d,
		conf.IgnoreExcess,
		conf.IgnoreValues,
		conf.StreamIgnoreExcess,
		conf.Ignore,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: ignoring failed: %s\n", err)
		os.Exit(statusDiffError)
	}

	if conf.OutputReport {
		ss, err := diff.Report(d, diff.Output(conf.output))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to generate report: %s\n", err)
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

func pruneIgnore(d diff.Differ, ingoreExcess, ignoreValues, streamIgnoreExcess bool, ignore ignorePatterns) (diff.Differ, error) {
	return diff.Walk(d, func(parent diff.Differ, d diff.Differ, path string) (diff.Differ, error) {
		if ignore.Match(path) {
			return diff.Ignore()
		}

		if ((ingoreExcess && !diff.IsStream(parent)) || (streamIgnoreExcess && diff.IsStream(parent))) && diff.IsExcess(d) {
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

func parseStream(fname string, lineByLine bool) (diff.Stream, io.Closer) {
	f, err := os.Open(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot open %s\n", fname)
		os.Exit(statusReadError)
	}

	if lineByLine {
		return &LineByLineJSONStream{
			Scanner: bufio.NewScanner(f),
		}, f
	}

	return &diff.JSONStream{
		Decoder: json.NewDecoder(f),
	}, f
}

type LineByLineJSONStream struct {
	*bufio.Scanner

	eof bool
	i   int
}

func (s *LineByLineJSONStream) NextValue() (interface{}, error) {
	if s.eof {
		return nil, io.EOF
	}

	if s.Scan() {
		v, err := decodeJSON(s.Bytes())
		if err != nil {
			err = fmt.Errorf("decoding json line %d: %v", s.i, err)
		}

		s.i++
		return v, err
	}

	err := s.Err()
	if err == nil {
		s.eof = true
		return nil, io.EOF
	}

	return nil, err
}

func decodeJSON(b []byte) (interface{}, error) {
	var v interface{}

	err := json.Unmarshal(b, &v)

	return v, err
}

type singleValueStream struct {
	diff.Stream
	partnerStream HasMore

	value    interface{}
	valueSet bool
	eof      bool
}

type HasMore interface {
	More() bool
}

func (s *singleValueStream) NextValue() (interface{}, error) {
	if s.eof {
		return nil, io.EOF
	}
	if !s.partnerStream.More() {
		s.eof = true

		return nil, io.EOF
	}
	if s.valueSet {
		return s.value, nil
	}

	v, err := s.Stream.NextValue()
	if err != nil {
		return nil, err
	}
	s.valueSet = true
	s.value = v

	return v, nil
}

func singleValueForValidate(lhs diff.Stream, rhs HasMore) diff.Stream {
	return &singleValueStream{
		Stream:        lhs,
		partnerStream: rhs,
	}
}

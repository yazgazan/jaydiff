package diff

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type stream struct {
	diffs   []Differ
	indices []int

	lhs []interface{}
	rhs []interface{}
}

type streamMissing struct {
	value interface{}
}

type streamExcess struct {
	value interface{}
}

type Stream interface {
	NextValue() (interface{}, error)
}

type JSONStream struct {
	*json.Decoder

	eof bool
}

func (s *JSONStream) NextValue() (interface{}, error) {
	var v interface{}

	if s.eof {
		return nil, io.EOF
	}

	err := s.Decode(&v)
	if err == io.EOF {
		s.eof = true
	}

	return v, err
}

func newStream(c config, lhs, rhs interface{}, visited *visited) (Differ, error) {
	var (
		diffs   []Differ
		indices []int
		lhsVals []interface{}
		rhsVals []interface{}
	)

	lhsStream, ok := lhs.(Stream)
	if !ok {
		return nil, errInvalidStream{lhs}
	}
	rhsStream, ok := rhs.(Stream)
	if !ok {
		return nil, errInvalidStream{rhs}
	}

	for i := 0; ; i++ {
		indices = append(indices, i)

		d, lhsVal, rhsVal, err := diffStreamValues(c, lhsStream, rhsStream, visited)
		if err == io.EOF {
			break
		}
		lhsVals = append(lhsVals, lhsVal)
		rhsVals = append(rhsVals, rhsVal)
		diffs = append(diffs, d)
		if err != nil {
			return stream{
				diffs:   diffs,
				indices: indices,
				lhs:     lhsVals,
				rhs:     rhsVals,
			}, err
		}
	}
	return stream{
		diffs:   diffs,
		indices: indices,
		lhs:     lhsVals,
		rhs:     rhsVals,
	}, nil
}

func diffStreamValues(c config, lhs, rhs Stream, visited *visited) (d Differ, lhsVal, rhsVal interface{}, err error) {
	var lhsErr, rhsErr error

	lhsVal, lhsErr = lhs.NextValue()
	rhsVal, rhsErr = rhs.NextValue()

	switch {
	default:
		d, err = diff(c, lhsVal, rhsVal, visited)
	case lhsErr == io.EOF && rhsErr == io.EOF:
		err = io.EOF
	case lhsErr != nil && lhsErr != io.EOF:
		err = lhsErr
	case rhsErr != nil && rhsErr != io.EOF:
		err = rhsErr
	case lhsErr == io.EOF:
		d = streamExcess{rhsVal}
	case rhsErr == io.EOF:
		d = streamMissing{lhsVal}
	}

	return d, lhsVal, rhsVal, err
}

func (s stream) Diff() Type {
	for _, d := range s.diffs {
		if d.Diff() != Identical {
			return ContentDiffer
		}
	}

	return Identical
}

func (s stream) Strings() []string {
	switch s.Diff() {
	case Identical:
		return streamStrings(s.lhs)
	default:
		var ss = []string{"["}

		for _, d := range s.diffs {
			ss = append(ss, d.Strings()...)
		}

		return append(ss, "]")
	}
}

func (s stream) StringIndent(key, prefix string, conf Output) string {
	switch s.Diff() {
	case Identical:
		return strings.Join(
			streamStringsIndent(key, prefix, conf, s.lhs),
			newLineSeparatorString(conf),
		)
	default:
		var ss = []string{}

		for _, d := range s.diffs {
			s := d.StringIndent("", prefix+conf.Indent, conf)
			if s != "" {
				ss = append(ss, s)
			}
		}

		return strings.Join(
			[]string{
				s.openString(key, prefix, conf),
				strings.Join(ss, newLineSeparatorString(conf)),
				" " + prefix + "]",
			}, "\n",
		)
	}
}

func (s stream) openString(key, prefix string, conf Output) string {
	if conf.JSON {
		return " " + prefix + key + "["
	}

	return " " + prefix + key + conf.typ(s.lhs) + "["
}

func (s stream) Walk(path string, fn WalkFn) error {
	for i, diff := range s.diffs {
		d, err := walk(s, diff, path+"["+strconv.Itoa(s.lhsIndex(i))+"]", fn)
		if err != nil {
			return err
		}
		if d != nil {
			s.diffs[i] = d
		}
	}

	return nil
}

func (s stream) lhsIndex(i int) int {
	return s.indices[i]
}

func (s stream) LHS() interface{} {
	return s.lhs
}

func (s stream) RHS() interface{} {
	return s.rhs
}

func streamStrings(vv []interface{}) []string {
	ss := make([]string, len(vv))

	for i, v := range vv {
		ss[i] = fmt.Sprintf("  %T %v", v, v)
	}
	return ss
}

func streamStringsIndent(key, prefix string, conf Output, vv []interface{}) []string {
	ss := make([]string, len(vv))

	for i, v := range vv {
		ss[i] = " " + prefix + key + conf.white(v)
	}

	return ss
}

func (m streamMissing) Diff() Type {
	return ContentDiffer
}

func (m streamMissing) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", m.value, m.value),
	}
}

func (m streamMissing) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.red(m.value)
}

func (m streamMissing) LHS() interface{} {
	return m.value
}

func (e streamExcess) Diff() Type {
	return ContentDiffer
}

func (e streamExcess) Strings() []string {
	return []string{
		fmt.Sprintf("+ %T %v", e.value, e.value),
	}
}

func (e streamExcess) StringIndent(key, prefix string, conf Output) string {
	return "+" + prefix + key + conf.green(e.value)
}

func (e streamExcess) RHS() interface{} {
	return e.value
}

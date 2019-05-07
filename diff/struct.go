package diff

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/yazgazan/jaydiff/jpath"
)

type structDiff struct {
	diffs map[string]Differ
	lhs   interface{}
	rhs   interface{}
}

func newStruct(c config, lhs, rhs interface{}, visited *visited) (Differ, error) {
	var diffs = make(map[string]Differ)

	if typesDiffer, err := structTypesDiffer(lhs, rhs); err != nil {
		return structDiff{
			lhs:   lhs,
			rhs:   rhs,
			diffs: diffs,
		}, err
	} else if !typesDiffer {
		lhsVal := reflect.ValueOf(lhs)
		lhsType := lhsVal.Type()
		rhsVal := reflect.ValueOf(rhs)

		for i := 0; i < lhsType.NumField(); i++ {
			fType := lhsType.Field(i)
			lhsFVal := lhsVal.Field(i)
			rhsFVal := rhsVal.Field(i)
			if !lhsFVal.CanInterface() {
				continue
			}

			diff, err := diff(c, lhsFVal.Interface(), rhsFVal.Interface(), visited)
			diffs[fType.Name] = diff

			if err != nil {
				return structDiff{
					lhs:   lhs,
					rhs:   rhs,
					diffs: diffs,
				}, err
			}
		}
	}

	return structDiff{
		lhs:   lhs,
		rhs:   rhs,
		diffs: diffs,
	}, nil
}

func structTypesDiffer(lhs, rhs interface{}) (bool, error) {
	if lhs == nil {
		return true, errInvalidType{Value: lhs, For: "struct"}
	}
	if rhs == nil {
		return true, errInvalidType{Value: rhs, For: "struct"}
	}

	lhsType := reflect.TypeOf(lhs)
	rhsType := reflect.TypeOf(rhs)

	return !lhsType.ConvertibleTo(rhsType), nil
}

func (s structDiff) Diff() Type {
	if ok, err := structTypesDiffer(s.lhs, s.rhs); err != nil {
		return Invalid
	} else if ok {
		return TypesDiffer
	}

	for _, d := range s.diffs {
		if d.Diff() != Identical {
			return ContentDiffer
		}
	}

	return Identical
}

func (s structDiff) Strings() []string {
	switch s.Diff() {
	case Identical:
		return []string{fmt.Sprintf("  %T %v", s.lhs, s.lhs)}
	case TypesDiffer:
		return []string{
			fmt.Sprintf("- %T %v", s.lhs, s.lhs),
			fmt.Sprintf("+ %T %v", s.rhs, s.rhs),
		}
	case ContentDiffer:
		var ss = []string{"{"}
		keys := make([]string, 0, len(s.diffs))

		for key := range s.diffs {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for _, key := range keys {
			d := s.diffs[key]
			for _, s := range d.Strings() {
				ss = append(ss, fmt.Sprintf("%v: %s", key, s))
			}
		}

		return append(ss, "}")
	}

	return []string{}
}

func (s structDiff) StringIndent(keyprefix, prefix string, conf Output) string {
	switch s.Diff() {
	case Identical:
		return " " + prefix + keyprefix + conf.white(s.lhs)
	case TypesDiffer:
		return "-" + prefix + keyprefix + conf.red(s.lhs) + newLineSeparatorString(conf) +
			"+" + prefix + keyprefix + conf.green(s.rhs)
	case ContentDiffer:
		var ss = []string{}
		keys := make([]string, 0, len(s.diffs))

		for key := range s.diffs {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for _, key := range keys {
			d := s.diffs[key]

			s := d.StringIndent(key+": ", prefix+conf.Indent, conf)
			if s != "" {
				ss = append(ss, s)
			}
		}

		return strings.Join([]string{
			s.openString(keyprefix, prefix, conf),
			strings.Join(ss, newLineSeparatorString(conf)),
			s.closeString(prefix, conf),
		}, "\n")
	}

	return ""
}

func (s structDiff) openString(keyprefix, prefix string, conf Output) string {
	if conf.JSON {
		return " " + prefix + keyprefix + "{"
	}
	return " " + prefix + keyprefix + conf.typ(s.lhs) + "map["
}

func (s structDiff) closeString(prefix string, conf Output) string {
	if conf.JSON {
		return " " + prefix + "}"
	}
	return " " + prefix + "]"
}

func (s structDiff) Walk(path string, fn WalkFn) error {
	keys := make([]string, 0, len(s.diffs))

	for k := range s.diffs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		diff := s.diffs[k]
		d, err := walk(s, diff, path+"."+jpath.EscapeKey(k), fn)
		if err != nil {
			return err
		}
		if d != nil {
			s.diffs[k] = d
		}
	}

	return nil
}

func (s structDiff) LHS() interface{} {
	return s.lhs
}

func (s structDiff) RHS() interface{} {
	return s.rhs
}

package diff

import (
	"fmt"
	"reflect"
	"strings"
)

type Map struct {
	Type
	Diffs map[interface{}]Differ
	LHS   interface{}
	RHS   interface{}
}

type MapMissing struct {
	Value interface{}
}

type MapExcess struct {
	Value interface{}
}

func NewMap(lhs, rhs interface{}) (*Map, error) {
	var Type = Identical
	var diffs = make(map[interface{}]Differ)

	lhsVal := reflect.ValueOf(lhs)
	lhsElType := lhsVal.Type().Elem()
	lhsKeyType := lhsVal.Type().Key()
	rhsVal := reflect.ValueOf(rhs)
	rhsElType := rhsVal.Type().Elem()
	rhsKeyType := rhsVal.Type().Key()

	if lhsElType.Kind() != rhsElType.Kind() {
		Type = TypesDiffer
	} else if lhsKeyType.Kind() != rhsKeyType.Kind() {
		Type = TypesDiffer
	} else {
		keys := getKeys(lhsVal, rhsVal)

		for _, key := range keys {
			lhsEl := lhsVal.MapIndex(key)
			rhsEl := rhsVal.MapIndex(key)

			if lhsEl.IsValid() && rhsEl.IsValid() {
				diff, err := Diff(lhsEl.Interface(), rhsEl.Interface())
				if err != nil {
					return &Map{
						Type:  diff.Diff(),
						LHS:   lhs,
						RHS:   rhs,
						Diffs: diffs,
					}, err
				}
				if diff.Diff() != Identical {
					Type = ContentDiffer
				}
				diffs[key.Interface()] = diff
				continue
			}
			if lhsEl.IsValid() {
				missing := &MapMissing{lhsEl.Interface()}
				diffs[key.Interface()] = missing
				Type = missing.Diff()
				continue
			}
			excess := &MapExcess{rhsEl.Interface()}
			diffs[key.Interface()] = excess
			Type = excess.Diff()
		}
	}

	return &Map{
		Type:  Type,
		LHS:   lhs,
		RHS:   rhs,
		Diffs: diffs,
	}, nil
}

func (m Map) Diff() Type {
	return m.Type
}

func (m Map) Strings() []string {
	switch m.Diff() {
	case Identical:
		return []string{fmt.Sprintf("  %T %v", m.LHS, m.LHS)}
	case TypesDiffer:
		return []string{
			fmt.Sprintf("- %T %v", m.LHS, m.LHS),
			fmt.Sprintf("+ %T %v", m.RHS, m.RHS),
		}
	case ContentDiffer:
		var ss = []string{"{"}

		for key, d := range m.Diffs {
			for _, s := range d.Strings() {
				ss = append(ss, fmt.Sprintf("%v: %s", key, s))
			}
		}

		return append(ss, "}")
	}

	return []string{}
}

func (m Map) StringIndent(keyprefix, prefix string, conf Output) string {
	switch m.Diff() {
	case Identical:
		return "  " + prefix + keyprefix + conf.White(m.LHS)
	case TypesDiffer:
		return "-" + prefix + keyprefix + conf.Red(m.LHS) + "\n" +
			"+" + prefix + keyprefix + conf.Green(m.RHS)
	case ContentDiffer:
		var ss = []string{prefix + keyprefix + conf.Type(m.LHS) + "map["}

		for key, d := range m.Diffs {
			keyStr := fmt.Sprintf("%v: ", key)
			ss = append(
				ss,
				d.StringIndent(keyStr, prefix+conf.Indent, conf),
			)
		}

		return strings.Join(append(ss, prefix+"]"), "\n")
	}

	return ""
}

func getKeys(lhs, rhs reflect.Value) []reflect.Value {
	keys := lhs.MapKeys()

	for _, key := range rhs.MapKeys() {
		found := false

		for _, existing := range keys {
			if key.Interface() == existing.Interface() {
				found = true
				break
			}
		}

		if !found {
			keys = append(keys, key)
		}
	}

	return keys
}

func (m MapMissing) Diff() Type {
	return ContentDiffer
}

func (m MapMissing) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", m.Value, m.Value),
	}
}

func (m MapMissing) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.Red(m.Value) +
		"\n+" + prefix + key
}

func (e MapExcess) Diff() Type {
	return ContentDiffer
}

func (e MapExcess) Strings() []string {
	return []string{
		fmt.Sprintf("+ %T %v", e.Value, e.Value),
	}
}

func (e MapExcess) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key +
		"\n+" + prefix + key + conf.Green(e.Value)
}

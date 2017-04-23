package diff

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type Map struct {
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
	var diffs = make(map[interface{}]Differ)

	lhsVal := reflect.ValueOf(lhs)
	rhsVal := reflect.ValueOf(rhs)

	if typesDiffer, err := mapTypesDiffer(lhs, rhs); err != nil {
		return &Map{
			LHS:   lhs,
			RHS:   rhs,
			Diffs: diffs,
		}, err
	} else if !typesDiffer {
		keys := getKeys(lhsVal, rhsVal)

		for _, key := range keys {
			lhsEl := lhsVal.MapIndex(key)
			rhsEl := rhsVal.MapIndex(key)

			if lhsEl.IsValid() && rhsEl.IsValid() {
				diff, err := Diff(lhsEl.Interface(), rhsEl.Interface())
				if diff.Diff() != Identical {
				}
				diffs[key.Interface()] = diff

				if err != nil {
					return &Map{
						LHS:   lhs,
						RHS:   rhs,
						Diffs: diffs,
					}, err
				}
				continue
			}
			if lhsEl.IsValid() {
				diffs[key.Interface()] = &MapMissing{lhsEl.Interface()}
				continue
			}
			diffs[key.Interface()] = &MapExcess{rhsEl.Interface()}
		}
	}

	return &Map{
		LHS:   lhs,
		RHS:   rhs,
		Diffs: diffs,
	}, nil
}

func mapTypesDiffer(lhs, rhs interface{}) (bool, error) {
	if lhs == nil {
		return true, InvalidType{Value: lhs, For: "map"}
	}
	if rhs == nil {
		return true, InvalidType{Value: rhs, For: "map"}
	}

	lhsVal := reflect.ValueOf(lhs)
	lhsElType := lhsVal.Type().Elem()
	lhsKeyType := lhsVal.Type().Key()
	rhsVal := reflect.ValueOf(rhs)
	rhsElType := rhsVal.Type().Elem()
	rhsKeyType := rhsVal.Type().Key()

	if lhsElType.Kind() != rhsElType.Kind() {
		return true, nil
	} else if lhsKeyType.Kind() != rhsKeyType.Kind() {
		return true, nil
	}

	return false, nil
}

func (m Map) Diff() Type {
	if ok, err := mapTypesDiffer(m.LHS, m.RHS); err != nil {
		return Invalid
	} else if ok {
		return TypesDiffer
	}

	for _, d := range m.Diffs {
		if d.Diff() != Identical {
			return ContentDiffer
		}
	}

	return Identical
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
		var keys []interface{}

		for key := range m.Diffs {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool {
			return strings.Compare(fmt.Sprintf("%v", keys[i]), fmt.Sprintf("%v", keys[j])) == -1
		})

		for _, key := range keys {
			d := m.Diffs[key]
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
		var ss = []string{" " + prefix + keyprefix + conf.Type(m.LHS) + "map["}
		var keys []interface{}

		for key := range m.Diffs {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool {
			return strings.Compare(fmt.Sprintf("%v", keys[i]), fmt.Sprintf("%v", keys[j])) == -1
		})

		for _, key := range keys {
			d := m.Diffs[key]

			keyStr := fmt.Sprintf("%v: ", key)
			s := d.StringIndent(keyStr, prefix+conf.Indent, conf)
			if s != "" {
				ss = append(ss, s)
			}
		}

		return strings.Join(append(ss, " "+prefix+"]"), "\n")
	}

	return ""
}

func (m Map) Walk(path string, fn WalkFn) error {
	for k, diff := range m.Diffs {
		err := walk(m, diff, fmt.Sprintf("%s.%v", path, k), fn)
		if err != nil {
			return err
		}
	}

	return nil
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

package diff

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type mapDiff struct {
	diffs map[interface{}]Differ
	lhs   interface{}
	rhs   interface{}
}

type mapMissing struct {
	value interface{}
}

type mapExcess struct {
	value interface{}
}

func newMap(lhs, rhs interface{}) (mapDiff, error) {
	var diffs = make(map[interface{}]Differ)

	lhsVal := reflect.ValueOf(lhs)
	rhsVal := reflect.ValueOf(rhs)

	if typesDiffer, err := mapTypesDiffer(lhs, rhs); err != nil {
		return mapDiff{
			lhs:   lhs,
			rhs:   rhs,
			diffs: diffs,
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
					return mapDiff{
						lhs:   lhs,
						rhs:   rhs,
						diffs: diffs,
					}, err
				}
				continue
			}
			if lhsEl.IsValid() {
				diffs[key.Interface()] = mapMissing{lhsEl.Interface()}
				continue
			}
			diffs[key.Interface()] = mapExcess{rhsEl.Interface()}
		}
	}

	return mapDiff{
		lhs:   lhs,
		rhs:   rhs,
		diffs: diffs,
	}, nil
}

func IsMap(d Differ) bool {
	_, ok := d.(mapDiff)

	return ok
}

func mapTypesDiffer(lhs, rhs interface{}) (bool, error) {
	if lhs == nil {
		return true, ErrInvalidType{Value: lhs, For: "map"}
	}
	if rhs == nil {
		return true, ErrInvalidType{Value: rhs, For: "map"}
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

func (m mapDiff) Diff() Type {
	if ok, err := mapTypesDiffer(m.lhs, m.rhs); err != nil {
		return Invalid
	} else if ok {
		return TypesDiffer
	}

	for _, d := range m.diffs {
		if d.Diff() != Identical {
			return ContentDiffer
		}
	}

	return Identical
}

func (m mapDiff) Strings() []string {
	switch m.Diff() {
	case Identical:
		return []string{fmt.Sprintf("  %T %v", m.lhs, m.lhs)}
	case TypesDiffer:
		return []string{
			fmt.Sprintf("- %T %v", m.lhs, m.lhs),
			fmt.Sprintf("+ %T %v", m.rhs, m.rhs),
		}
	case ContentDiffer:
		var ss = []string{"{"}
		var keys []interface{}

		for key := range m.diffs {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool {
			return strings.Compare(fmt.Sprintf("%v", keys[i]), fmt.Sprintf("%v", keys[j])) == -1
		})

		for _, key := range keys {
			d := m.diffs[key]
			for _, s := range d.Strings() {
				ss = append(ss, fmt.Sprintf("%v: %s", key, s))
			}
		}

		return append(ss, "}")
	}

	return []string{}
}

func (m mapDiff) StringIndent(keyprefix, prefix string, conf Output) string {
	switch m.Diff() {
	case Identical:
		return "  " + prefix + keyprefix + conf.white(m.lhs)
	case TypesDiffer:
		return "-" + prefix + keyprefix + conf.red(m.lhs) + "\n" +
			"+" + prefix + keyprefix + conf.green(m.rhs)
	case ContentDiffer:
		var ss = []string{" " + prefix + keyprefix + conf.typ(m.lhs) + "map["}
		var keys []interface{}

		for key := range m.diffs {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool {
			return strings.Compare(fmt.Sprintf("%v", keys[i]), fmt.Sprintf("%v", keys[j])) == -1
		})

		for _, key := range keys {
			d := m.diffs[key]

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

func (m mapDiff) Walk(path string, fn WalkFn) error {
	var keys []interface{}

	for k := range m.diffs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(fmt.Sprintf("%v", keys[i]), fmt.Sprintf("%v", keys[j])) == -1
	})

	for _, k := range keys {
		diff := m.diffs[k]
		d, err := walk(m, diff, fmt.Sprintf("%s.%v", path, k), fn)
		if err != nil {
			return err
		}
		if d != nil {
			m.diffs[k] = d
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

func IsMapMissing(d Differ) bool {
	_, ok := d.(mapMissing)

	return ok
}

func (m mapMissing) Diff() Type {
	return ContentDiffer
}

func (m mapMissing) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", m.value, m.value),
	}
}

func (m mapMissing) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.red(m.value) +
		"\n+" + prefix + key
}

func IsMapExcess(d Differ) bool {
	_, ok := d.(mapExcess)

	return ok
}

func (e mapExcess) Diff() Type {
	return ContentDiffer
}

func (e mapExcess) Strings() []string {
	return []string{
		fmt.Sprintf("+ %T %v", e.value, e.value),
	}
}

func (e mapExcess) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key +
		"\n+" + prefix + key + conf.green(e.value)
}

package diff

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/yazgazan/jaydiff/jpath"
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

func newMap(c config, lhs, rhs interface{}, visited *visited) (Differ, error) {
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
				diff, err := diff(c, lhsEl.Interface(), rhsEl.Interface(), visited)
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

func mapTypesDiffer(lhs, rhs interface{}) (bool, error) {
	if lhs == nil {
		return true, errInvalidType{Value: lhs, For: "map"}
	}
	if rhs == nil {
		return true, errInvalidType{Value: rhs, For: "map"}
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
		keys := make([]interface{}, 0, len(m.diffs))

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
		return "-" + prefix + keyprefix + conf.red(m.lhs) + newLineSeparatorString(conf) +
			"+" + prefix + keyprefix + conf.green(m.rhs)
	case ContentDiffer:
		var ss = []string{}
		keys := make([]interface{}, 0, len(m.diffs))

		for key := range m.diffs {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool {
			return strings.Compare(fmt.Sprintf("%v", keys[i]), fmt.Sprintf("%v", keys[j])) == -1
		})

		for _, key := range keys {
			d := m.diffs[key]

			keyStr := m.mapKeyString(key, conf)
			s := d.StringIndent(keyStr, prefix+conf.Indent, conf)
			if s != "" {
				ss = append(ss, s)
			}
		}

		return strings.Join([]string{
			m.openString(keyprefix, prefix, conf),
			strings.Join(ss, newLineSeparatorString(conf)),
			m.closeString(prefix, conf),
		}, "\n")
	}

	return ""
}

func (m mapDiff) openString(keyprefix, prefix string, conf Output) string {
	if conf.JSON {
		return " " + prefix + keyprefix + "{"
	}
	return " " + prefix + keyprefix + conf.typ(m.lhs) + "map["
}

func (m mapDiff) closeString(prefix string, conf Output) string {
	if conf.JSON {
		return " " + prefix + "}"
	}
	return " " + prefix + "]"
}

func (m mapDiff) mapKeyString(key interface{}, conf Output) string {
	if conf.JSON {
		return fmt.Sprintf("%q: ", key)
	}

	return fmt.Sprintf("%v: ", key)
}

func (m mapDiff) Walk(path string, fn WalkFn) error {
	keys := make([]interface{}, 0, len(m.diffs))

	for k := range m.diffs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(fmt.Sprintf("%v", keys[i]), fmt.Sprintf("%v", keys[j])) == -1
	})

	for _, k := range keys {
		diff := m.diffs[k]
		d, err := walk(m, diff, path+"."+jpath.EscapeKey(k), fn)
		if err != nil {
			return err
		}
		if d != nil {
			m.diffs[k] = d
		}
	}

	return nil
}

func (m mapDiff) LHS() interface{} {
	return m.lhs
}

func (m mapDiff) RHS() interface{} {
	return m.rhs
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

func (m mapMissing) Diff() Type {
	return ContentDiffer
}

func (m mapMissing) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", m.value, m.value),
	}
}

func (m mapMissing) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.red(m.value)
}

func (m mapMissing) LHS() interface{} {
	return m.value
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
	return "+" + prefix + key + conf.green(e.value)
}

func (e mapExcess) RHS() interface{} {
	return e.value
}

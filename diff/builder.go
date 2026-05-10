package diff

import (
	"fmt"

	"github.com/yazgazan/jaydiff/jpath"
)

type Builder struct {
	diff DiffBuilder
	err  error
}

func (b *Builder) Add(path string, v interface{}) *Builder {
	if b.err != nil {
		return b
	}
	if b.diff == nil {
		b.diff = &value{}
	}
	pp, err := jpath.Parse(path)
	if err != nil {
		b.err = err
		return b
	}

	b.err = b.diff.Add(pp, v)

	return b
}

func (b *Builder) Delete(path string, v interface{}) *Builder {
	if b.err != nil {
		return b
	}
	if b.diff == nil {
		b.diff = &value{}
	}
	pp, err := jpath.Parse(path)
	if err != nil {
		b.err = err
		return b
	}

	b.err = b.diff.Delete(pp, v)

	return b
}

func (b *Builder) Build() (Differ, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.diff == nil {
		return Ignore()
	}

	return b.diff, nil
}

type value struct {
	Differ
}

func (v *value) Add(path jpath.Path, i interface{}) error {
	if len(path) == 0 && v.Differ == nil {
		v.Differ = valueExcess{i}
		return nil
	}
	if len(path) > 0 && v.Differ == nil {
		v.Differ = emptyContainer(path[0])
	}
	if len(path) > 0 {
		diffBuilder, ok := v.Differ.(DiffBuilder)
		if !ok {
			return fmt.Errorf("cannot add value to %T(%+v)", v.Differ, v.Differ)
		}
		return diffBuilder.Add(path, i)
	}

	if t, ok := v.Differ.(valueMissing); ok {
		v.Differ = valueDiffers{
			lhs: t.value,
			rhs: i,
		}

		return nil
	}

	return fmt.Errorf("cannot add value to %T(%+v)", v.Differ, v.Differ)
}

func (v *value) Delete(path jpath.Path, i interface{}) error {
	if len(path) == 0 && v.Differ == nil {
		v.Differ = valueMissing{i}
		return nil
	}
	if len(path) > 0 && v.Differ == nil {
		v.Differ = emptyContainer(path[0])
	}
	if len(path) > 0 {
		diffBuilder, ok := v.Differ.(DiffBuilder)
		if !ok {
			return fmt.Errorf("cannot delete value from %T(%+v)", v.Differ, v.Differ)
		}
		return diffBuilder.Delete(path, i)
	}

	return fmt.Errorf("cannot delete value to %T(%+v)", v.Differ, v.Differ)
}

func (v *value) Walk(path string, fn WalkFn) error {
	d, err := walk(v, v.Differ, path, fn)
	if err != nil {
		return err
	}
	if d != nil {
		v.Differ = d
	}

	return nil
}

func emptyContainer(p jpath.PathPart) DiffBuilder {
	switch p.Kind() {
	default:
		panic(fmt.Errorf("unknown path part %v", p.Kind()))
	case jpath.PathKindIndex:
		return emptySliceDiff()
	case jpath.PathKindKey:
		return emptyMapDiff()
	}
}

func emptyMapDiff() *mapDiff {
	return &mapDiff{
		diffs: map[interface{}]Differ{},
		lhs:   map[string]interface{}{},
		rhs:   map[string]interface{}{},
	}
}

func emptySliceDiff() *slice {
	return &slice{
		diffs:   []Differ{},
		indices: []int{},
		lhs:     []interface{}{},
		rhs:     []interface{}{},
	}
}

type valueDiffers struct {
	lhs interface{}
	rhs interface{}
}

func (v valueDiffers) Diff() Type {
	return ContentDiffer
}

func (v valueDiffers) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", v.lhs, v.lhs),
		fmt.Sprintf("+ %T %v", v.rhs, v.rhs),
	}
}

func (v valueDiffers) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.red(v.lhs) + newLineSeparatorString(conf) +
		"+" + prefix + key + conf.green(v.rhs)
}

func (v valueDiffers) LHS() interface{} {
	return v.lhs
}

func (v valueDiffers) RHS() interface{} {
	return v.rhs
}

type valueMissing struct {
	value interface{}
}

func (v valueMissing) Diff() Type {
	return ContentDiffer
}

func (v valueMissing) Strings() []string {
	return []string{
		fmt.Sprintf("- %T %v", v.value, v.value),
	}
}

func (v valueMissing) StringIndent(key, prefix string, conf Output) string {
	return "-" + prefix + key + conf.red(v.value)
}

func (v valueMissing) LHS() interface{} {
	return v.value
}

type valueExcess struct {
	value interface{}
}

func (v valueExcess) Diff() Type {
	return ContentDiffer
}

func (v valueExcess) Strings() []string {
	return []string{
		fmt.Sprintf("+ %T %v", v.value, v.value),
	}
}

func (v valueExcess) StringIndent(key, prefix string, conf Output) string {
	return "+" + prefix + key + conf.green(v.value)
}

func (v valueExcess) RHS() interface{} {
	return v.value
}

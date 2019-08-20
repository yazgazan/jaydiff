package diff

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
)

// Output is used to configure the output of the Strings and StringIndent functions.
type Output struct {
	Indent     string
	ShowTypes  bool
	Colorized  bool
	JSON       bool
	JSONValues bool
}

type colorFn func(format string, a ...interface{}) string

var whiteFn colorFn = nil

func (o Output) red(v interface{}) string {
	return o.applyColor(v, color.RedString)
}

func (o Output) green(v interface{}) string {
	return o.applyColor(v, color.GreenString)
}

func (o Output) white(v interface{}) string {
	return o.applyColor(v, whiteFn)
}

func (o Output) applyColor(v interface{}, fn colorFn) string {
	var s string

	switch {
	default:
		s = fmt.Sprintf("%v", v)
	case o.ShowTypes:
		s = fmt.Sprintf("%T %v", v, v)
	case o.JSONValues:
		s = jsonString(v)
	}

	if !o.Colorized || fn == nil {
		return s
	}

	return fn("%s", s)
}

func (o Output) typ(v interface{}) string {
	if o.ShowTypes {
		return fmt.Sprintf("%T ", v)
	}

	return ""
}

func newLineSeparatorString(conf Output) string {
	if conf.JSON {
		return ",\n"
	}

	return "\n"
}

func jsonString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("unexpected error marshaling value: %s", err))
	}

	return string(b)
}

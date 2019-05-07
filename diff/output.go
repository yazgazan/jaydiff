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

func (o Output) red(v interface{}) string {
	var s string

	switch {
	default:
		s = fmt.Sprintf("%v", v)
	case o.ShowTypes:
		s = fmt.Sprintf("%T %v", v, v)
	case o.JSONValues:
		s = jsonString(v)
	}

	if !o.Colorized {
		return s
	}

	return color.RedString("%s", s)
}

func (o Output) green(v interface{}) string {
	var s string

	switch {
	default:
		s = fmt.Sprintf("%v", v)
	case o.ShowTypes:
		s = fmt.Sprintf("%T %v", v, v)
	case o.JSONValues:
		s = jsonString(v)
	}

	if !o.Colorized {
		return s
	}

	return color.GreenString("%s", s)
}

func (o Output) white(v interface{}) string {
	var s string

	switch {
	default:
		s = fmt.Sprintf("%v", v)
	case o.ShowTypes:
		s = fmt.Sprintf("%T %v", v, v)
	case o.JSONValues:
		s = jsonString(v)
	}

	return s
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

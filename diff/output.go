package diff

import (
	"fmt"

	"github.com/fatih/color"
)

// Output is used to configure the output of the Strings and StringIndent functions.
type Output struct {
	Indent    string
	ShowTypes bool
	Colorized bool
}

func (o Output) red(v interface{}) string {
	var s string

	if o.ShowTypes {
		s = fmt.Sprintf("%T %v", v, v)
	} else {
		s = fmt.Sprintf("%v", v)
	}

	if !o.Colorized {
		return s
	}

	return color.RedString("%s", s)
}

func (o Output) green(v interface{}) string {
	var s string

	if o.ShowTypes {
		s = fmt.Sprintf("%T %v", v, v)
	} else {
		s = fmt.Sprintf("%v", v)
	}

	if !o.Colorized {
		return s
	}

	return color.GreenString("%s", s)
}

func (o Output) white(v interface{}) string {
	var s string

	if o.ShowTypes {
		s = fmt.Sprintf("%T %v", v, v)
	} else {
		s = fmt.Sprintf("%v", v)
	}

	return s
}

func (o Output) typ(v interface{}) string {
	if o.ShowTypes {
		return fmt.Sprintf("%T ", v)
	}

	return ""
}

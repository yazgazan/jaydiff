package diff

import (
	"fmt"
	"github.com/fatih/color"
)

type Output struct {
	Indent    string
	Colorized bool
	ShowTypes bool
}

func (o Output) Red(v interface{}) string {
	var s string

	if o.ShowTypes {
		s = fmt.Sprintf("%T %v", v, v)
	} else {
		s = fmt.Sprintf("%v", v)
	}

	if !o.Colorized {
		return fmt.Sprintf("%s", s)
	}

	return color.RedString("%s", s)
}

func (o Output) Green(v interface{}) string {
	var s string

	if o.ShowTypes {
		s = fmt.Sprintf("%T %v", v, v)
	} else {
		s = fmt.Sprintf("%v", v)
	}

	if !o.Colorized {
		return fmt.Sprintf("%s", s)
	}

	return color.GreenString("%s", s)
}

func (o Output) White(v interface{}) string {
	var s string

	if o.ShowTypes {
		s = fmt.Sprintf("%T %v", v, v)
	} else {
		s = fmt.Sprintf("%v", v)
	}

	return fmt.Sprintf("%s", s)
}

func (o Output) Type(v interface{}) string {
	if o.ShowTypes {
		return fmt.Sprintf("%T ", v)
	}

	return ""
}

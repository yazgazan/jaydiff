package diff

import (
	"errors"
	"strings"
	"testing"
)

func TestOutput(t *testing.T) {
	for _, test := range []struct {
		Output   Output
		WantVal  []string
		WantType []string
	}{
		{
			Output: Output{
				Indent:    "\t",
				Colorized: false,
				ShowTypes: true,
				JSON:      false,
			},
			WantVal:  []string{"int", "5"},
			WantType: []string{"int"},
		},
		{
			Output: Output{
				Indent:    "\t",
				Colorized: false,
				ShowTypes: false,
				JSON:      false,
			},
			WantVal:  []string{"5"},
			WantType: []string{},
		},
		{
			Output: Output{
				Indent:    "\t",
				Colorized: true,
				ShowTypes: false,
				JSON:      false,
			},
			WantVal:  []string{"5"},
			WantType: []string{},
		},
		{
			Output: Output{
				Indent:    "\t",
				Colorized: true,
				ShowTypes: false,
				JSON:      true,
			},
			WantVal:  []string{"5"},
			WantType: []string{},
		},
		{
			Output: Output{
				Indent:    "\t",
				Colorized: true,
				ShowTypes: true,
				JSON:      false,
			},
			WantVal:  []string{"int", "5"},
			WantType: []string{"int"},
		},
	} {
		red := test.Output.red(5)
		testOut(t, "Output.Red(5)", red, test.WantVal)
		green := test.Output.green(5)
		testOut(t, "Output.Green(5)", green, test.WantVal)
		white := test.Output.white(5)
		testOut(t, "Output.White(5)", white, test.WantVal)
		typ := test.Output.typ(5)
		testOut(t, "Output.Type(5)", typ, test.WantType)
	}
}

type erroringMarshaler struct{}

func (erroringMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("erroringMarshaler error")
}

func TestJSONStringPanic(t *testing.T) {
	defer func() {
		panicV := recover()
		if panicV == nil {
			t.Error("Expected jsonString to panic")
		}
	}()

	jsonString(erroringMarshaler{})
}

func testOut(t *testing.T, expr, result string, wantStrings []string) {
	for _, want := range wantStrings {
		if !strings.Contains(result, want) {
			t.Errorf("%s = %q, expected it to contain %q", expr, result, want)
		}
	}
}

package diff

import (
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
			},
			WantVal:  []string{"int", "5"},
			WantType: []string{"int"},
		},
		{
			Output: Output{
				Indent:    "\t",
				Colorized: false,
				ShowTypes: false,
			},
			WantVal:  []string{"5"},
			WantType: []string{},
		},
		{
			Output: Output{
				Indent:    "\t",
				Colorized: true,
				ShowTypes: false,
			},
			WantVal:  []string{"5"},
			WantType: []string{},
		},
		{
			Output: Output{
				Indent:    "\t",
				Colorized: true,
				ShowTypes: true,
			},
			WantVal:  []string{"int", "5"},
			WantType: []string{"int"},
		},
	} {
		red := test.Output.Red(5)
		testOut(t, "Output.Red(5)", red, test.WantVal)
		green := test.Output.Green(5)
		testOut(t, "Output.Green(5)", green, test.WantVal)
		white := test.Output.White(5)
		testOut(t, "Output.White(5)", white, test.WantVal)
		typ := test.Output.Type(5)
		testOut(t, "Output.Type(5)", typ, test.WantType)
	}
}

func testOut(t *testing.T, expr, result string, wantStrings []string) {
	for _, want := range wantStrings {
		if !strings.Contains(result, want) {
			t.Errorf("%s = %q, expected it to contain %q", expr, result, want)
		}
	}
}

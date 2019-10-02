package diff

import (
	"reflect"
	"strings"
	"testing"
)

func TestUnsupported(t *testing.T) {
	for _, test := range []struct {
		LHS  reflect.Type
		RHS  reflect.Type
		Want string
	}{
		{LHS: reflect.TypeOf(func() {}), RHS: reflect.TypeOf(0), Want: "func()"},
		{LHS: reflect.TypeOf(0), RHS: reflect.TypeOf(struct{}{}), Want: "struct"},
	} {
		s := ErrUnsupported{test.LHS, test.RHS}.Error()

		if !strings.Contains(s, test.Want) {
			t.Errorf("Unsupported.Error() = %q, expected it to contain %q", s, test.Want)
		}
	}
}

func TestInvalidStream(t *testing.T) {
	type invalidStream struct{}
	s := errInvalidStream{invalidStream{}}.Error()

	if !strings.Contains(s, "invalidStream") {
		t.Errorf("errInvalidStream{invalidStream{}}.Error() = %q, expected it to contain %q", s, "invalidStream")
	}
}

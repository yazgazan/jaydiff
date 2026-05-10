package diff

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yazgazan/jaydiff/jpath"
)

func TestBuilder(t *testing.T) {
	t.Run("add value", func(t *testing.T) {
		diff, err := (&Builder{}).Add("", 42).Build()
		if err != nil {
			t.Errorf(`Buidler{}.Add("", 42): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"+", "42"},
		})
	})
	t.Run("delete value", func(t *testing.T) {
		diff, err := (&Builder{}).Delete("", 23).Build()
		if err != nil {
			t.Errorf(`Buidler{}.Delete("", 23): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"-", "23"},
		})
	})
	t.Run("replace value", func(t *testing.T) {
		diff, err := (&Builder{}).Delete("", 23).Add("", 42).Build()
		if err != nil {
			t.Errorf(`Buidler{}.Delete("", 23).Add("", 42): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"-", "23", "+", "42"},
		})
	})

	t.Run("add to map", func(t *testing.T) {
		diff, err := (&Builder{}).Add(".foo", 42).Build()
		if err != nil {
			t.Errorf(`Builder{}.Add(".foo", 42): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"+", ".foo", "42"},
		})
	})

	t.Run("multiple add to map", func(t *testing.T) {
		diff, err := (&Builder{}).Add(".foo", 42).Add(".bar", 23).Build()
		if err != nil {
			t.Errorf(`Builder{}.Add(".foo", 42).Add(".bar", 23): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"+", ".bar", "23"},
			{"+", ".foo", "42"},
		})
	})

	t.Run("delete from map", func(t *testing.T) {
		diff, err := (&Builder{}).Delete(".foo", 23).Build()
		if err != nil {
			t.Errorf(`Builder{}.Delete(".foo", 23): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"-", ".foo", "23"},
		})
	})

	t.Run("multiple delete from map", func(t *testing.T) {
		diff, err := (&Builder{}).Delete(".foo", 23).Delete(".bar", 42).Build()
		if err != nil {
			t.Errorf(`Builder{}.Delete(".foo", 23).Delete(".bar", 42): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"-", ".bar", "42"},
			{"-", ".foo", "23"},
		})
	})

	t.Run("replace in map", func(t *testing.T) {
		diff, err := (&Builder{}).Delete(".foo", 23).Add(".foo", 42).Build()
		if err != nil {
			t.Errorf(`Builder{}.Delete(".foo", 23).Add(".foo", 42): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"-", "foo", "23", "+", ".foo", "42"},
		})
	})

	t.Run("deep add to map", func(t *testing.T) {
		diff, err := (&Builder{}).Add(".foo.bar", 42).Build()
		if err != nil {
			t.Errorf(`Builder{}.Add(".foo.bar", 42): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"+", ".foo.bar", "42"},
		})
	})

	t.Run("add to slice", func(t *testing.T) {
		diff, err := (&Builder{}).Add("[2]", 42).Build()
		if err != nil {
			t.Errorf(`Builder{}.Add("[2]", 42): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"+", "[2]", "42"},
		})
	})

	t.Run("multiple add to slice", func(t *testing.T) {
		diff, err := (&Builder{}).Add("[2]", 42).Add("[3]", 23).Build()
		if err != nil {
			t.Errorf(`Builder{}.Add("[2]", 42).Add("[3]", 23): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"+", "[2]", "42"},
			{"+", "[3]", "23"},
		})
	})

	t.Run("delete from slice", func(t *testing.T) {
		diff, err := (&Builder{}).Delete("[2]", 23).Build()
		if err != nil {
			t.Errorf(`Builder{}.Delete("[2]", 23): unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"-", "[2]", "23"},
		})
	})

	t.Run("complex diff", func(t *testing.T) {
		b := (&Builder{}).Delete(".b[1]", 3).Add(".b[1]", 5).Add(".b[2]", 4)
		b.Delete(".c.a", "toto").Add(".c.a", "titi")
		b.Delete(".c.b", 23).Add(".c.b", "23")
		b.Delete(".e", []interface{}{})
		b.Delete(".f", 42)
		b.Add(".h", 42)

		diff, err := b.Build()
		if err != nil {
			t.Errorf(`Builder{}...: unexpected error: %v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"-", ".b[1]", "3", "+", ".b[1]", "5"},
			{"+", ".b[2]", "4"},
			{"-", ".c.a", `"toto"`, "+", ".c.a", `"titi"`},
			{"-", ".c.b", "23", "+", ".c.b", `"23"`},
			{"-", ".e", "[]"},
			{"-", ".f", "42"},
			{"+", ".h", "42"},
		})
	})

	t.Run("delete same twice from map", func(t *testing.T) {
		_, err := (&Builder{}).Delete(".a", 42).Delete(".a", 23).Add(".b", 0).Delete(".c", 1).Build()

		if err == nil {
			t.Error(`Builder{}.Delete(".a", 42).Delete(".a", 23): expected error, got nil`)
		}
	})

	t.Run("delete same twice from slice", func(t *testing.T) {
		_, err := (&Builder{}).Delete("[1]", 42).Delete("[1]", 23).Add("[2]", 0).Delete("[3]", 1).Build()

		if err == nil {
			t.Error(`Builder{}.Delete("[1]", 42).Delete("[1]", 23): expected error, got nil`)
		}
	})

	t.Run("add invalid path", func(t *testing.T) {
		_, err := (&Builder{}).Add(`."foo`, 42).Build()

		if err == nil {
			t.Error("Builder{}.Add(`.\"foo`, 42): expected error, got nil")
		}
	})

	t.Run("delete invalid path", func(t *testing.T) {
		_, err := (&Builder{}).Delete(`."foo`, 42).Build()

		if err == nil {
			t.Error("Builder{}.Delete(`.\"foo`, 42): expected error, got nil")
		}
	})

	t.Run("nothing", func(t *testing.T) {
		diff, err := (&Builder{}).Build()
		if err != nil {
			t.Errorf(`Builder{}.Build(): unexpected error: %v`, err)
			return
		}

		if !IsIgnore(diff) {
			t.Errorf("Builder{}.Build() = %+v, expected Ignore{}", diff)
		}
	})

	t.Run("path cannot be added (value)", func(t *testing.T) {
		_, err := (&Builder{}).Add("", 4).Add(".bar", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Add("", 4).Add(".bar", 5): expected error, got nil`)
		}
	})

	t.Run("path cannot be deleted (value)", func(t *testing.T) {
		_, err := (&Builder{}).Add("", 4).Delete(".bar", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Add("", 4).Delete(".bar", 5): expected error, got nil`)
		}
	})

	t.Run("path cannot be added (map)", func(t *testing.T) {
		_, err := (&Builder{}).Add(".foo", 4).Add(".foo.bar", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Add(".foo", 4).Add(".foo.bar", 5): expected error, got nil`)
		}
	})

	t.Run("path cannot be deleted (map)", func(t *testing.T) {
		_, err := (&Builder{}).Add(".foo", 4).Delete(".foo.bar", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Add(".foo", 4).Delete(".foo.bar", 5): expected error, got nil`)
		}
	})

	t.Run("add value twice", func(t *testing.T) {
		_, err := (&Builder{}).Add("", 4).Add("", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Add("", 4).Add("", 5): expected error, got nil`)
		}
	})

	t.Run("delete value twice", func(t *testing.T) {
		_, err := (&Builder{}).Delete("", 4).Delete("", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Delete("", 4).Delete("", 5): expected error, got nil`)
		}
	})

	t.Run("add to map twice", func(t *testing.T) {
		_, err := (&Builder{}).Add(".foo", 4).Add(".foo", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Add(".foo", 4).Add(".foo", 5): expected error, got nil`)
		}
	})

	t.Run("del from map twice", func(t *testing.T) {
		_, err := (&Builder{}).Delete(".foo", 4).Delete(".foo", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Delete(".foo", 4).Delete(".foo", 5): expected error, got nil`)
		}
	})

	t.Run("del after add in map", func(t *testing.T) {
		_, err := (&Builder{}).Add(".foo", 4).Delete(".foo", 4).Build()

		if err == nil {
			t.Error(`Builder{}.Add(".foo", 4).Delete(".foo", 4): expected error, got nil`)
		}
	})

	t.Run("add to slice twice", func(t *testing.T) {
		_, err := (&Builder{}).Add("[5]", 4).Add("[5]", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Add("[5]", 4).Add("[5]", 5): expected error, got nil`)
		}
	})

	t.Run("del from slice twice", func(t *testing.T) {
		_, err := (&Builder{}).Delete("[5]", 4).Delete("[5]", 5).Build()

		if err == nil {
			t.Error(`Builder{}.Delete("[5]", 4).Delete("[5]", 5): expected error, got nil`)
		}
	})

	t.Run("del after add in slice", func(t *testing.T) {
		_, err := (&Builder{}).Add("[5]", 4).Delete("[5]", 4).Build()

		if err == nil {
			t.Error(`Builder{}.Add("[5]", 4).Delete("[5]", 4): expected error, got nil`)
		}
	})

	t.Run("add to slice reverse order", func(t *testing.T) {
		diff, err := (&Builder{}).Add("[5]", 4).Add("[2].foo", 1).Build()

		if err != nil {
			t.Errorf(`Builder{}.Add("[5]", 4).Add("[2].foo", 1): unexpected error :%v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"+", "[2].foo", "1"},
			{"+", "[5]", "4"},
		})
	})

	t.Run("add to slice after delete", func(t *testing.T) {
		_, err := (&Builder{}).Delete("[2]", 4).Add("[2].foo", 1).Build()

		if err == nil {
			t.Error(`Builder{}.Delete("[2]", 4).Add("[2].foo", 1): expected error, got nil`)
			return
		}
	})

	t.Run("delete from slice reverse order", func(t *testing.T) {
		diff, err := (&Builder{}).Delete("[5]", 4).Delete("[2].foo", 1).Build()

		if err != nil {
			t.Errorf(`Builder{}.Delete("[5]", 4).Delete("[2].foo", 1): unexpected error :%v`, err)
			return
		}

		testbuilderReport(t, diff, [][]string{
			{"-", "[2].foo", "1"},
			{"-", "[5]", "4"},
		})
	})

	t.Run("delete from slice after add", func(t *testing.T) {
		_, err := (&Builder{}).Add("[2]", 4).Delete("[2].foo", 1).Build()

		if err == nil {
			t.Error(`Builder{}.Add("[2]", 4).Delete("[2].foo", 1): expected error, got nil`)
			return
		}
	})
}

func testbuilderReport(t *testing.T, diff Differ, wants [][]string) {
	t.Helper()

	ss, err := Report(diff, Output{
		JSONValues: true,
	})
	if err != nil {
		t.Errorf("failed to generate report: %v", err)
		t.Fail()
	}

	for i, want := range wants {
		s := ss[i]

		for i, needle := range want {
			if !strings.Contains(s, needle) {
				t.Errorf("Builder{}.Report[%d] = %q, expected it to contain %q", i, s, needle)
			}
		}
	}
}

func testbuilderStrings(t *testing.T, diff Differ, wants [][]string) {
	t.Helper()

	ss := diff.Strings()
	for i, want := range wants {
		s := ss[i]

		for i, needle := range want {
			if !strings.Contains(s, needle) {
				t.Errorf("Builder{}.Strings()[%d] = %q, expected it to contain %q", i, s, needle)
			}
		}
	}
}

func testbuilderStringIndent(t *testing.T, diff Differ, wants []string) {
	t.Helper()

	const (
		key    = "(key)"
		prefix = "(prefix)"
	)

	s := diff.StringIndent(key, prefix, Output{})
	if !strings.Contains(s, key) {
		t.Errorf(".StringIndent() = %q, expected it to contain %q", s, key)
	}
	if !strings.Contains(s, prefix) {
		t.Errorf(".StringIndent() = %q, expected it to contain %q", s, prefix)
	}
	for i, needle := range wants {
		if !strings.Contains(s, needle) {
			t.Errorf("Builder{}.Strings()[%d] = %q, expected it to contain %q", i, s, needle)
		}
	}
}

type invalidPathPart struct{}

func (p invalidPathPart) Kind() jpath.PathKind {
	return -1
}

func (p invalidPathPart) String() string {
	return "<invalid>"
}

func TestEmptyContainer(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("emptyContainer(invalidPathPart{}): expected panic")
		}
	}()

	_ = emptyContainer(invalidPathPart{})
}

func TestMapDiffAdd(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		err := emptyMapDiff().Add(jpath.Path{}, 42)

		if err == nil {
			t.Error("mapDiff{}.Add([], ...): expected error, got nil")
		}
	})
	t.Run("wrong path", func(t *testing.T) {
		err := emptyMapDiff().Add(jpath.Path{jpath.PathIndex(2)}, 42)

		if err == nil {
			t.Error("mapDiff{}.Add(jpath.Path{jpath.PathIndex(2)}, ...): expected error, got nil")
		}
	})
}

func TestMapDiffDelete(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		err := emptyMapDiff().Delete(jpath.Path{}, 42)

		if err == nil {
			t.Error("mapDiff{}.Delete([], ...): expected error, got nil")
		}
	})
	t.Run("wrong path", func(t *testing.T) {
		err := emptyMapDiff().Delete(jpath.Path{jpath.PathIndex(2)}, 42)

		if err == nil {
			t.Error("mapDiff{}.Delete(jpath.Path{jpath.PathIndex(2)}, ...): expected error, got nil")
		}
	})
}

func TestSliceDiffAdd(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		err := emptySliceDiff().Add(jpath.Path{}, 42)

		if err == nil {
			t.Error("slice{}.Add([], ...): expected error, got nil")
		}
	})
	t.Run("wrong path", func(t *testing.T) {
		err := emptySliceDiff().Add(jpath.Path{jpath.PathKey(`foo`)}, 42)

		if err == nil {
			t.Error("slice{}.Add(jpath.Path{jpath.PathKey(`foo`)}, ...): expected error, got nil")
		}
	})
}

func TestSliceDiffDelete(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		err := emptySliceDiff().Delete(jpath.Path{}, 42)

		if err == nil {
			t.Error("slice{}.Delete([], ...): expected error, got nil")
		}
	})
	t.Run("wrong path", func(t *testing.T) {
		err := emptySliceDiff().Delete(jpath.Path{jpath.PathKey(`foo`)}, 42)

		if err == nil {
			t.Error("slice{}.Delete(jpath.Path{jpath.PathKey(`foo`)}, ...): expected error, got nil")
		}
	})
}

func TestValueDiffers(t *testing.T) {
	t.Run("Diff()", func(t *testing.T) {
		dt := valueDiffers{42, 23}.Diff()
		if dt != ContentDiffer {
			t.Errorf("valueDiffers{42, 23}.Diff() = %v, expected %v", dt, ContentDiffer)
		}
	})

	t.Run("Strings()", func(t *testing.T) {
		d := valueDiffers{
			lhs: 42,
			rhs: 23,
		}

		testbuilderStrings(t, d, [][]string{
			{"-", "42"},
			{"+", "23"},
		})
	})

	t.Run("StringIndent()", func(t *testing.T) {
		d := valueDiffers{
			lhs: 42,
			rhs: 23,
		}

		testbuilderStringIndent(t, d, []string{
			"-", "42", "+", "23",
		})
	})

	t.Run("LHS()", func(t *testing.T) {
		d := valueDiffers{
			lhs: 42,
			rhs: 23,
		}

		if !reflect.DeepEqual(d.LHS(), 42) {
			t.Errorf("valueDiffers{42, 23}.LHS() = %v, expected 42", d.LHS())
		}
	})

	t.Run("RHS()", func(t *testing.T) {
		d := valueDiffers{
			lhs: 42,
			rhs: 23,
		}

		if !reflect.DeepEqual(d.RHS(), 23) {
			t.Errorf("valueDiffers{42, 23}.RHS() = %v, expected 23", d.RHS())
		}
	})
}

func TestValueMissing(t *testing.T) {
	t.Run("Diff()", func(t *testing.T) {
		dt := valueMissing{42}.Diff()
		if dt != ContentDiffer {
			t.Errorf("valueMissing{42}.Diff() = %v, expected %v", dt, ContentDiffer)
		}
	})

	t.Run("Strings()", func(t *testing.T) {
		d := valueMissing{
			value: 42,
		}

		testbuilderStrings(t, d, [][]string{
			{"-", "42"},
		})
	})

	t.Run("StringIndent()", func(t *testing.T) {
		d := valueMissing{
			value: 42,
		}

		testbuilderStringIndent(t, d, []string{
			"-", "42",
		})
	})

	t.Run("LHS()", func(t *testing.T) {
		d := valueMissing{
			value: 42,
		}

		if !reflect.DeepEqual(d.LHS(), 42) {
			t.Errorf("valueMissing{42}.LHS() = %v, expected 42", d.LHS())
		}
	})
}

func TestValueExcess(t *testing.T) {
	t.Run("Diff()", func(t *testing.T) {
		dt := valueExcess{42}.Diff()
		if dt != ContentDiffer {
			t.Errorf("valueExcess{42}.Diff() = %v, expected %v", dt, ContentDiffer)
		}
	})

	t.Run("Strings()", func(t *testing.T) {
		d := valueExcess{
			value: 42,
		}

		testbuilderStrings(t, d, [][]string{
			{"+", "42"},
		})
	})

	t.Run("StringIndent()", func(t *testing.T) {
		d := valueExcess{
			value: 42,
		}

		testbuilderStringIndent(t, d, []string{
			"+", "42",
		})
	})

	t.Run("RHS()", func(t *testing.T) {
		d := valueExcess{
			value: 42,
		}

		if !reflect.DeepEqual(d.RHS(), 42) {
			t.Errorf("valueExcess{42}.RHS() = %v, expected 42", d.RHS())
		}
	})
}

package diff

import "testing"

func TestCircular(t *testing.T) {
	first := map[int]interface{}{}
	second := map[int]interface{}{
		0: first,
	}
	first[0] = second
	notCyclic := map[int]interface{}{
		0: map[int]interface{}{
			0: map[int]interface{}{
				0: "foo",
			},
		},
		1: []interface{}{
			"bar", "baz",
		},
	}
	emptySlice := map[int]interface{}{
		0: []interface{}{},
	}
	emptySlice[1] = emptySlice[0]
	emptySliceNotRepeating := map[int]interface{}{
		0: []interface{}{},
		1: []interface{}{},
	}
	emptyMap := map[int]interface{}{
		0: map[int]interface{}{},
	}
	emptyMap[1] = emptyMap[0]
	emptyMapNotRepeating := map[int]interface{}{
		0: map[int]interface{}{},
		1: map[int]interface{}{},
	}

	repeatingNotCyclic := map[int]interface{}{
		0: []interface{}{"foo", "bar"},
	}
	repeatingNotCyclic[1] = repeatingNotCyclic[0]
	repeatingNotCyclic2 := map[int]interface{}{
		0: []interface{}{"foo", "bar"},
	}
	repeatingNotCyclic2[1] = []interface{}{
		"foo",
		repeatingNotCyclic2[0],
		"bar",
	}

	for _, test := range []struct {
		lhs       interface{}
		rhs       interface{}
		wantError bool
	}{
		{lhs: first, rhs: first, wantError: true},
		{lhs: first, rhs: second, wantError: true},
		{lhs: first, rhs: notCyclic, wantError: true},
		{lhs: notCyclic, rhs: first, wantError: true},
		{lhs: notCyclic, rhs: emptySlice},
		{lhs: notCyclic, rhs: emptyMap},
		{lhs: notCyclic, rhs: notCyclic},
		{lhs: emptySlice, rhs: emptySliceNotRepeating},
		{lhs: emptySliceNotRepeating, rhs: emptySlice},
		{lhs: emptyMap, rhs: emptyMapNotRepeating},
		{lhs: emptyMapNotRepeating, rhs: emptyMap},
		{lhs: notCyclic, rhs: repeatingNotCyclic, wantError: false},
		{lhs: repeatingNotCyclic, rhs: notCyclic, wantError: false},
		{lhs: repeatingNotCyclic2, rhs: notCyclic, wantError: false},
		{lhs: repeatingNotCyclic, rhs: repeatingNotCyclic2, wantError: false},
	} {
		d, err := Diff(test.lhs, test.rhs)

		if test.wantError && (err == nil || err != ErrCyclic) {
			t.Errorf("Expected error %q, got %q", ErrCyclic, err)
		}
		if !test.wantError && err != nil {
			t.Errorf("Unexpected error %q", err)
		}

		if test.wantError && d.Diff() != ContentDiffer {
			t.Errorf("Expected Diff() to be %s, got %s", ContentDiffer, d.Diff())
		}
	}
}

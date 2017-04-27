package diff

import (
	"errors"
	"testing"
)

func TestWalk(t *testing.T) {
	for _, test := range []struct {
		LHS  interface{}
		RHS  interface{}
		Want int
	}{
		{42, 42, 1},
		{[]int{42}, []int{42}, 2},
		{map[int]int{1: 2}, map[int]int{1: 2}, 2},
		{
			LHS: map[string][]int{
				"abc": {1, 2},
			},
			RHS: map[string][]int{
				"abc": {1, 4, 5},
			},
			Want: 5,
		},
	} {
		var nCalls int

		d, err := Diff(test.LHS, test.RHS)
		if err != nil {
			t.Errorf("Diff(%+v, %+v): unexpected error: %s", test.LHS, test.RHS, err)
			continue
		}

		_, err = Walk(d, func(_, diff Differ, _ string) (Differ, error) {
			nCalls++
			return nil, nil
		})
		if err != nil {
			t.Errorf("Walk(Diff(%+v, %+v)): unexpected error: %s", test.LHS, test.RHS, err)
			continue
		}

		if nCalls != test.Want {
			t.Errorf(
				"Walk(Diff(%+v, %+v)): expected walk function to be called %d times, not %d",
				test.LHS, test.RHS, test.Want, nCalls,
			)
		}
	}
}

func TestWalkError(t *testing.T) {
	var expectedErr = errors.New("forbidden 42")
	for _, test := range []struct {
		LHS interface{}
		RHS interface{}
	}{
		{42, 43},
		{[]int{42}, []int{44}},
		{map[string]int{"ha": 42}, map[string]int{"ha": 45}},
	} {
		d, err := Diff(test.LHS, test.RHS)
		if err != nil {
			t.Errorf("Diff(%+v, %+v): unexpected error: %s", test.LHS, test.RHS, err)
			continue
		}

		_, err = Walk(d, func(_, diff Differ, _ string) (Differ, error) {
			if _, ok := diff.(scalar); ok {
				return nil, expectedErr
			}

			return nil, nil
		})
		if err != expectedErr {
			t.Errorf("Walk(Diff(%+v, %+v)): expected error %q, got %q", test.LHS, test.RHS, expectedErr, err)
			continue
		}
	}
}

func TestIsExcess(t *testing.T) {
	for _, test := range []struct {
		LHS interface{}
		RHS interface{}
	}{
		{[]int{1, 2}, []int{1, 2, 3}},
		{map[int]int{1: 2, 3: 4}, map[int]int{1: 2, 3: 4, 5: 6}},
	} {
		d, err := Diff(test.LHS, test.RHS)
		if err != nil {
			t.Errorf("Diff(%+v, %+v): unexpected error: %s", test.LHS, test.RHS, err)
			continue
		}
		if d.Diff() != ContentDiffer {
			t.Errorf("Diff(%+v, %+v).Diff() = %s, expected %s", test.LHS, test.RHS, d.Diff(), ContentDiffer)
		}

		d, err = Walk(d, func(_, diff Differ, _ string) (Differ, error) {
			if IsExcess(diff) {
				return Ignore()
			}
			return nil, nil
		})
		if err != nil {
			t.Errorf("Walk(...): unexpected error: %s", err)
			continue
		}
		if d.Diff() != Identical {
			t.Errorf("Walk(...)).Diff() = %s, expected %s", d.Diff(), Identical)
		}
	}
}

func TestIsMissing(t *testing.T) {
	for _, test := range []struct {
		LHS interface{}
		RHS interface{}
	}{
		{[]int{1, 2, 3}, []int{1, 2}},
		{map[int]int{1: 2, 3: 4, 5: 6}, map[int]int{1: 2, 3: 4}},
	} {
		d, err := Diff(test.LHS, test.RHS)
		if err != nil {
			t.Errorf("Diff(%+v, %+v): unexpected error: %s", test.LHS, test.RHS, err)
			continue
		}
		if d.Diff() != ContentDiffer {
			t.Errorf("Diff(%+v, %+v).Diff() = %s, expected %s", test.LHS, test.RHS, d.Diff(), ContentDiffer)
		}

		d, err = Walk(d, func(_, diff Differ, _ string) (Differ, error) {
			if IsMissing(diff) {
				return Ignore()
			}
			return nil, nil
		})
		if err != nil {
			t.Errorf("Walk(...): unexpected error: %s", err)
			continue
		}
		if d.Diff() != Identical {
			t.Errorf("Walk(...)).Diff() = %s, expected %s", d.Diff(), Identical)
		}
	}
}

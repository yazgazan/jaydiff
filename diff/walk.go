package diff

// WalkFn Should be implemented to walk down a diff tree.
// diff and path refer to the current node. If a WalkFn returns
// a non-nil value, the current diff will be replaced (and then walked
// over if possible).
type WalkFn func(parent Differ, diff Differ, path string) (Differ, error)

// Walker is implemented by types that can be walked (such as maps and slices)
type Walker interface {
	// Walk receives its own path and the walking function.
	Walk(path string, fn WalkFn) error
}

// Walk allows to descend a diff tree and replace/edit its leaves and branches.
// When fn returns a non-nil Differ, the current node is replaced
// and the new node is walked over (if walkable).
func Walk(diff Differ, fn WalkFn) (Differ, error) {
	return walk(nil, diff, "", fn)
}

func walk(parent Differ, diff Differ, path string, fn WalkFn) (Differ, error) {
	var err error

	newD, err := fn(parent, diff, path)
	if err != nil {
		return diff, err
	}
	if newD != nil {
		diff = newD
	}

	if walker, ok := diff.(Walker); ok {
		return diff, walker.Walk(path, fn)
	}

	return diff, nil
}

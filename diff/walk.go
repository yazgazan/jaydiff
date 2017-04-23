package diff

type WalkFn func(parent Differ, diff Differ, path string) (Differ, error)

type Walker interface {
	Walk(path string, fn WalkFn) error
}

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

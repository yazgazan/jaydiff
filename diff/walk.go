package diff

type WalkFn func(parent Differ, diff Differ, path string) error

type Walker interface {
	Walk(path string, fn WalkFn) error
}

func Walk(diff Differ, fn WalkFn) error {
	return walk(nil, diff, "", fn)
}

func walk(parent Differ, diff Differ, path string, fn WalkFn) error {
	var err error

	err = fn(parent, diff, path)
	if err != nil {
		return err
	}

	if walker, ok := diff.(Walker); ok {
		return walker.Walk(path, fn)
	}

	return nil
}

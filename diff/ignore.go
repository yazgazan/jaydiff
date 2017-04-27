package diff

type ignore struct{}

// Ignore can be used in a WalkFn to ignore a non-matching diff.
// (See Walk example)
func Ignore() (Differ, error) {
	return ignore{}, nil
}

func (t ignore) Diff() Type {
	return Identical
}

func (t ignore) Strings() []string {
	return []string{}
}

func (t ignore) StringIndent(key, prefix string, conf Output) string {
	return ""
}

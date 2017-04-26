package diff

type ignore struct{}

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

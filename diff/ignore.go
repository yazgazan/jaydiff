package diff

type Ignore struct{}

func (t Ignore) Diff() Type {
	return Identical
}

func (t Ignore) Strings() []string {
	return []string{}
}

func (t Ignore) StringIndent(key, prefix string, conf Output) string {
	return ""
}

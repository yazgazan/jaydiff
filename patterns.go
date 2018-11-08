package main

import (
	"github.com/gobwas/glob"
	"github.com/yazgazan/jaydiff/jpath"
)

type ignorePattern struct {
	glob.Glob
	s string
}

type ignorePatterns []ignorePattern

func (p *ignorePatterns) UnmarshalFlag(s string) error {
	pattern, err := glob.Compile(s)
	if err != nil {
		return err
	}
	*p = append(*p, ignorePattern{
		s:    s,
		Glob: pattern,
	})

	return nil
}

func (p ignorePatterns) Match(s string) bool {
	s = jpath.StripIndices(s)
	for _, pattern := range p {
		if pattern.Match(s) {
			return true
		}
	}

	return false
}

package main

import (
	"strings"

	"github.com/gobwas/glob"
)

type ignorePattern struct {
	glob.Glob
	s string
}

type ignorePatterns []ignorePattern

func (p ignorePatterns) String() string {
	var ss []string

	for _, pattern := range p {
		ss = append(ss, pattern.s)
	}

	return strings.Join(ss, ",")
}

func (p *ignorePatterns) Set(s string) error {
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
	for _, pattern := range p {
		if pattern.Match(s) {
			return true
		}
	}

	return false
}

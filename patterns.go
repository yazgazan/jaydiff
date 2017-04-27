package main

import (
	"strings"

	"github.com/gobwas/glob"
)

type Pattern struct {
	glob.Glob
	s string
}

type Patterns []Pattern

func (p Patterns) String() string {
	var ss []string

	for _, pattern := range p {
		ss = append(ss, pattern.s)
	}

	return strings.Join(ss, ",")
}

func (p *Patterns) Set(s string) error {
	pattern, err := glob.Compile(s)
	if err != nil {
		return err
	}
	*p = append(*p, Pattern{
		s:    s,
		Glob: pattern,
	})

	return nil
}

func (p Patterns) Match(s string) bool {
	for _, pattern := range p {
		if pattern.Match(s) {
			return true
		}
	}

	return false
}

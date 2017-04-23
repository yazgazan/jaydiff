package main

import (
	"github.com/gobwas/glob"
	"strings"
)

type Pattern struct {
	glob.Glob
	s string
}

type patterns []Pattern

func (p patterns) String() string {
	var ss []string

	for _, pattern := range p {
		ss = append(ss, pattern.s)
	}

	return strings.Join(ss, ",")
}

func (p *patterns) Set(s string) error {
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

func (p patterns) Match(s string) bool {
	for _, pattern := range p {
		if pattern.Match(s) {
			return true
		}
	}

	return false
}

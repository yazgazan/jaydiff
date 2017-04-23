package main

import (
	"flag"
	"fmt"
	"github.com/Pimmr/json-diff/diff"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

type config struct {
	diff.Output
	ignore       patterns
	lhsFile      string
	rhsFile      string
	outputReport bool
}

func readConfig() config {
	var c config

	flag.StringVar(&c.Output.Indent, "indent", "    ", "indent string")
	flag.BoolVar(&c.Output.ShowTypes, "show-types", false, "show types")
	flag.BoolVar(&c.outputReport, "report", false, "output report format")
	flag.Var(&c.ignore, "ignore", "paths to ignore (glob)")
	flag.Parse()

	if len(flag.Args()) < 2 {
		fmt.Fprintln(os.Stderr, "Error: missing json files")
		flag.Usage()
		os.Exit(2)
	}

	c.lhsFile = flag.Arg(0)
	c.rhsFile = flag.Arg(1)
	c.Output.Colorized = terminal.IsTerminal(int(os.Stdout.Fd()))

	return c
}

package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/yazgazan/jaydiff/diff"
	"golang.org/x/crypto/ssh/terminal"
)

type files struct {
	LHS string `positional-arg-name:"FILE_1"`
	RHS string `positional-arg-name:"FILE_2"`
}

type config struct {
	Files  files          `positional-args:"yes" required:"yes"`
	Ignore ignorePatterns `long:"ignore" short:"i" description:"paths to ignore (glob)"`
	output
	IgnoreExcess  bool   `long:"ignore-excess" description:"ignore excess keys and arrey elements"`
	IgnoreValues  bool   `long:"ignore-values" description:"ignore scalar's values (only type is compared)"`
	OutputReport  bool   `long:"report" short:"r" description:"output report format"`
	UseSliceMyers bool   `long:"slice-myers" description:"use myers algorithm for slices"`
	Version       func() `long:"version" short:"v" description:"print release version"`
	SortingKeys   string `long:"keys" short:"k" description:"array sorting keys for default slice comparison (comma separated)"`
}

type output struct {
	Indent     string `long:"indent" description:"indent string" default:"\t"`
	ShowTypes  bool   `long:"show-types" short:"t" description:"show types"`
	Colorized  bool
	JSON       bool `long:"json" description:"json-style output"`
	JSONValues bool
}

func readConfig() config {
	var c config
	c.Version = func() {
		fmt.Fprintf(os.Stderr, "%s\n", Version)
		os.Exit(0)
	}

	_, err := flags.Parse(&c)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Failed to parse arguments. See %s --help\n", os.Args[0])
		os.Exit(statusUsage)
	}

	if c.JSON && c.ShowTypes {
		fmt.Fprintf(os.Stderr, "Incompatible options --json and --show-types\n")
		os.Exit(statusUsage)
	}
	if c.JSON {
		c.JSONValues = true
	}
	if c.JSON && c.OutputReport {
		c.JSON = false
	}

	c.output.Colorized = terminal.IsTerminal(int(os.Stdout.Fd()))

	return c
}

func (c config) Opts() []diff.ConfigOpt {
	opts := []diff.ConfigOpt{}

	if c.UseSliceMyers {
		opts = append(opts, diff.UseSliceMyers())
	}
	if len(c.SortingKeys) > 0 {
		opts = append(opts, diff.AddSortingKeys(c.SortingKeys))
	}

	return opts
}

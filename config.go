package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"golang.org/x/crypto/ssh/terminal"
)

type config struct {
	Output
	Ignore       Patterns `long:"ignore" short:"i" description:"paths to ignore (glob)"`
	OutputReport bool     `long:"report" short:"r" description:"output report format"`
	Files        struct {
		LHS string `positional-arg-name:"FILE_1"`
		RHS string `positional-arg-name:"FILE_2"`
	} `positional-args:"yes" required:"yes"`
}

type Output struct {
	Indent    string `long:"indent" description:"indent string" default:"\t"`
	ShowTypes bool   `long:"show-types" short:"t" description:"show types"`
	Colorized bool
}

func readConfig() config {
	var c config

	_, err := flags.Parse(&c)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Failed to parse arguments. See %s --help\n", os.Args[0])
		os.Exit(2)
	}

	c.Output.Colorized = terminal.IsTerminal(int(os.Stdout.Fd()))

	return c
}

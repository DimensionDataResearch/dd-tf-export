package main

import (
	"fmt"
	"go-flags"
	"os"
)

type programOptions struct {
	Region         string   `short:"r" long:"region" required:"yes" description:"The CloudControl region to use."`
	Datacenter     string   `short:"d" long:"datacenter" required:"yes" description:"The CloudControl data centre containing the resource(s) to export."`
	NetworkDomains []string `short:"n" long:"networkdomain" description:"The network domain(s) to export."`
	Servers        []string `short:"s" long:"server" description:"The server(s) to export."`
	NoRecurse      bool     `long:"no-recurse" description:"Don't recursively export child resources."`
	Version        bool     `long:"version" description:"Display the current program version."`
	Verbose        bool     `short:"v" long:"verbose" description:"Display detailed information about the exporter's activities."`
}

func parseOptions() programOptions {
	options := programOptions{}

	parser := flags.NewParser(&options, flags.Default)
	_, err := parser.ParseArgs(os.Args)
	if err != nil {
		switch err.(type) {
		case *flags.Error:
			// Ignore, since the parser will print it out anyway.
		default:
			fmt.Print(err.Error())
		}

		os.Exit(1)
	}

	return options
}

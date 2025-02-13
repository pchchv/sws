package cmd

import (
	"context"
	"flag"
)

type Command interface {
	Setup() error
	// Run and block until context cancel
	Run(context.Context) error
	// Help by printing a usage string. Currently not used anywhere.
	Help() string
	// Describe the command shortly
	Describe() string
	// Flagset which defines the flags for the command
	Flagset() *flag.FlagSet
}

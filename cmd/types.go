package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
)

var (
	ErrHelpful = errors.New("user needs help")
	ErrNoArgs  = errors.New("no arguments found")
)

type UsagePrinter func()

type ArgParser func([]string) (Command, error)

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

type ArgNotFoundError string

func (err ArgNotFoundError) Error() string {
	return fmt.Sprintf("'%s' is not a valid argument\n", err.Error())
}

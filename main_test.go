package main

import (
	"context"
	"flag"
)

type MockCommand struct {
	runFunc      func(context.Context) error
	helpFunc     func() string
	describeFunc func() string
}

func (m MockCommand) Run(ctx context.Context) error {
	return m.runFunc(ctx)
}

func (m MockCommand) Help() string {
	return m.helpFunc()
}

func (m MockCommand) Describe() string {
	return m.describeFunc()
}

func (m MockCommand) Setup() error {
	return nil
}

func (m MockCommand) Flagset() *flag.FlagSet {
	return flag.NewFlagSet("test", flag.ContinueOnError)
}

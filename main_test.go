package main

import (
	"context"
	"errors"
	"flag"
	"testing"

	"github.com/pchchv/sws/cmd"
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

func Test_Run_ExitCodes(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		argParser cmd.ArgParser
		expected  int
	}{
		{
			name: "on invalid arg, it should return exit code 1",
			args: []string{"invalid"},
			argParser: func(s []string) (cmd.Command, error) {
				return nil, errors.New("some error")
			},
			expected: 1,
		},
		{
			name: "on run error, it should return exit code 1",
			args: []string{"valid"},
			argParser: func(s []string) (cmd.Command, error) {
				return MockCommand{
					runFunc: func(ctx context.Context) error {
						return errors.New("whopsidops, error ojoj..!")
					},
				}, nil
			},
			expected: 1,
		},
		{
			name: "on success, error code 0",
			args: []string{"valid"},
			argParser: func(s []string) (cmd.Command, error) {
				return MockCommand{
					runFunc: func(ctx context.Context) error {
						return nil
					},
				}, nil
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := run(context.Background(), tt.args, tt.argParser); result != tt.expected {
				t.Errorf("run() = %v, want %v", result, tt.expected)
			}
		})
	}
}

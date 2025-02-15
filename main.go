package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/pchchv/sws/cmd"
	"github.com/pchchv/sws/helpers/ancli"
	"github.com/pchchv/sws/helpers/shutdown"
)

func printHelp(command cmd.Command, err error, printUsage cmd.UsagePrinter) int {
	var notValidArg cmd.ArgNotFoundError
	if errors.As(err, &notValidArg) {
		ancli.PrintErr(err.Error())
		printUsage()
	} else if errors.Is(err, cmd.ErrNoArgs) {
		printUsage()
	} else if errors.Is(err, cmd.ErrHelpful) {
		if command != nil {
			fmt.Println(command.Help())
		} else {
			printUsage()
		}
		return 0
	} else {
		ancli.PrintfErr("unknown error: %v", err.Error())
	}

	return 1
}

func run(ctx context.Context, args []string, parseArgs cmd.ArgParser) int {
	command, err := parseArgs(args)
	if err != nil {
		return printHelp(command, err, cmd.PrintUsage)
	}

	var cmdArgs []string
	fs := command.Flagset()
	if len(args) > 2 {
		cmdArgs = args[2:]
	}

	if err = fs.Parse(cmdArgs); err != nil {
		ancli.PrintfErr("failed to parse flagset: %e", err)
		return 1
	}

	if err = command.Setup(); err != nil {
		ancli.PrintfErr("failed to setup command: %e", err)
		return 1
	}

	if err = command.Run(ctx); err != nil {
		ancli.PrintfErr("failed to run %e", err)
		return 1
	}

	return 0
}

func main() {
	ancli.Newline = true
	ancli.SlogIt = true
	ancli.SetupSlog()
	ctx, cancel := context.WithCancel(context.Background())
	exitCodeChan := make(chan int, 1)
	go func() {
		exitCodeChan <- run(ctx, os.Args, cmd.Parse)
		cancel()
	}()

	shutdown.Monitor(ctx, cancel)
	os.Exit(<-exitCodeChan)
}

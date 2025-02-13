package main

import (
	"errors"
	"fmt"

	"github.com/pchchv/sws/cmd"
	"github.com/pchchv/sws/helpers/ancli"
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

func main() {
	ancli.Newline = true
	ancli.SlogIt = true
	ancli.SetupSlog()
}

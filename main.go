package main

import (
	"errors"
	"fmt"

	"github.com/pchchv/sws/cmd"
	"github.com/pchchv/sws/helpers"
)

func printHelp(command cmd.Command, err error, printUsage cmd.UsagePrinter) int {
	var notValidArg cmd.ArgNotFoundError
	if errors.As(err, &notValidArg) {
		helpers.PrintErr(err.Error())
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
		helpers.PrintfErr("unknown error: %v", err.Error())
	}

	return 1
}

func main() {
	helpers.Newline = true
	helpers.SlogIt = true
	helpers.SetupSlog()
}

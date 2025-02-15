package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/pchchv/sws/cmd/server"
	"github.com/pchchv/sws/cmd/version"
)

const usage = `== Web Development 41 ==

This tool is designed to enable live reload for statically hosted web development.
It injects a websocket script in a mirrored version of html pages
and uses the fsnotify (cross-platform 'inotify' wrapper) package to detect filechanges.
On filechanges, the websocket will trigger a reload of the page.

The 41 (formerly "40", before I got spooked by potential lawyers) is only
to enable rust-repellant properties.

Commands:
%v`

var commands = map[string]Command{
	"s|serve":   server.Command(),
	"v|version": version.Command(),
}

func PrintUsage() {
	fmt.Printf(usage, formatCommandDescriptions())
}

func Parse(args []string) (Command, error) {
	if len(args) == 1 {
		return nil, ErrNoArgs
	}

	var cmdCandidate string
	for _, arg := range args[1:] {
		if isHelp(arg) {
			return nil, ErrHelpful
		}

		if isFlag := strings.HasPrefix(arg, "-"); isFlag {
			continue
		}

		// break on first non-flag
		cmdCandidate = arg
		break
	}

	for cmdNameWithShortcut, cmd := range commands {
		for _, cmdName := range strings.Split(cmdNameWithShortcut, "|") {
			if cmdName == cmdCandidate {
				return cmd, nil
			}
		}
	}

	return nil, ArgNotFoundError(cmdCandidate)
}

func formatCommandDescriptions() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for name, cmd := range commands {
		fmt.Fprintf(w, "\t%v\t%v\n", name, cmd.Describe())
	}
	w.Flush()
	return buf.String()
}

func isHelp(s string) bool {
	return s == "-h" || s == "-help" || s == "h" || s == "help"
}

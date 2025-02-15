package cmd

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/pchchv/sws/cmd/server"
	"github.com/pchchv/sws/cmd/version"
)

var commands = map[string]Command{
	"s|serve":   server.Command(),
	"v|version": version.Command(),
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

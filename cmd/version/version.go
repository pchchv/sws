package version

import "runtime/debug"

type command struct {
	getVersionCmd func() (*debug.BuildInfo, bool)
}

// Setup sets up the *command.
func (c *command) Setup() error {
	c.getVersionCmd = debug.ReadBuildInfo
	return nil
}

// Help helps by printing out the help.
func (c *command) Help() string {
	return "Print the version of wd-41"
}

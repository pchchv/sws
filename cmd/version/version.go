package version

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"runtime/debug"
)

// Set with buildflag if the build is done in the pipeline, not with go install.
var (
	BUILD_VERSION  = ""
	BUILD_CHECKSUM = ""
)

type command struct {
	getVersionCmd func() (*debug.BuildInfo, bool)
}

func Command() *command {
	return &command{}
}

// Setup sets up the *command.
func (c *command) Setup() error {
	c.getVersionCmd = debug.ReadBuildInfo
	return nil
}

// Help helps by printing out the help.
func (c *command) Help() string {
	return "Print the version of sws"
}

// Run runs the *command, printing the version using either debugbuild or tagged version.
func (c *command) Run(context.Context) error {
	bi, ok := c.getVersionCmd()
	if !ok {
		return errors.New("failed to read build info")
	}

	version := bi.Main.Version
	if version == "" || version == "(devel)" {
		version = BUILD_VERSION
	}

	checksum := bi.Main.Sum
	if checksum == "" {
		checksum = BUILD_CHECKSUM
	}

	fmt.Printf("version: %v, go version: %v, checksum: %v\n", version, bi.GoVersion, checksum)

	return nil
}

// Describe describes the version *command.
func (c *command) Describe() string {
	return "print the version of sws"
}

// Flagset sets the flagset for the version.
// By default is empty.
func (c *command) Flagset() *flag.FlagSet {
	return flag.NewFlagSet("version", flag.ExitOnError)
}

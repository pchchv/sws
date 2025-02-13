package version

import "runtime/debug"

type command struct {
	getVersionCmd func() (*debug.BuildInfo, bool)
}

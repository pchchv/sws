package server

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/gorilla/websocket"
	"github.com/pchchv/sws/internal/wsinject"
)

type Fileserver interface {
	Setup(pathToMaster string) (string, error)
	Start(ctx context.Context) error
	WsHandler(ws *websocket.Conn)
}

type command struct {
	port         *int
	wsPath       *string
	binPath      string
	tlsKeyPath   *string
	masterPath   string
	mirrorPath   string
	tlsCertPath  *string
	cacheControl *string
	forceReload  *bool
	fileserver   Fileserver
	flagset      *flag.FlagSet
}

func Command() *command {
	r, _ := os.Executable()
	return &command{
		binPath: r,
	}
}

func (c *command) Setup() error {
	var relPath string
	if len(c.flagset.Args()) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get exec path: %e", err)
		}
		relPath = wd
	} else {
		relPath = c.flagset.Arg(0)
	}
	c.masterPath = path.Clean(relPath)

	if c.masterPath != "" {
		expectTLS := *c.tlsCertPath != "" && *c.tlsKeyPath != ""
		c.fileserver = wsinject.NewFileServer(*c.port, *c.wsPath, *c.forceReload, expectTLS)
		mirrorPath, err := c.fileserver.Setup(c.masterPath)
		if err != nil {
			return fmt.Errorf("failed to setup websocket injected mirror filesystem: %e", err)
		}
		c.mirrorPath = mirrorPath
	}

	return nil
}

func (c *command) Help() string {
	return "Serve some filesystem. Set the directory as the second argument: sws serve <dir>. If omitted, current wd will be used."
}

func (c *command) Describe() string {
	return fmt.Sprintf("a webserver. Usage: '%v serve <path>'. If <path> is left unfilled, current pwd will be used.", c.binPath)
}

func (c *command) Flagset() *flag.FlagSet {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	c.port = fs.Int("port", 8080, "port to serve http server on")
	c.wsPath = fs.String("wsPort", "/delta-streamer-ws", "the path which the delta streamer websocket should be hosted on")
	c.forceReload = fs.Bool("forceReload", false, "set to true if you wish to reload all attached browser pages on any file change")
	c.cacheControl = fs.String("cacheControl", "no-cache", "set to configure the cache-control header")
	c.tlsCertPath = fs.String("tlsCertPath", "", "set to a path to a cert, requires tlsKeyPath to be set")
	c.tlsKeyPath = fs.String("tlsKeyPath", "", "set to a path to a key, requires tlsCertPath to be set")
	c.flagset = fs
	return fs
}

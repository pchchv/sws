package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/websocket"
	"github.com/pchchv/sws/helpers/ancli"
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
	masterPath   string
	mirrorPath   string
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
		c.fileserver = wsinject.NewFileServer(*c.port, *c.wsPath, *c.forceReload)
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
	c.flagset = fs
	return fs
}

func (c *command) Run(ctx context.Context) (err error) {
	mux := http.NewServeMux()
	fsh := http.FileServer(http.Dir(c.mirrorPath))
	fsh = slogHandler(fsh)
	fsh = cacheHandler(fsh, *c.cacheControl)
	mux.Handle("/", fsh)

	ancli.PrintfOK("setting up websocket host on path: '%v'", *c.wsPath)
	mux.HandleFunc(*c.wsPath, func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade to websocket", http.StatusInternalServerError)
			return
		}
		c.fileserver.WsHandler(conn)
	})

	s := http.Server{
		Addr:        fmt.Sprintf(":%v", *c.port),
		Handler:     mux,
		ReadTimeout: 0,
	}
	serverErrChan := make(chan error, 1)
	fsErrChan := make(chan error, 1)
	go func() {
		ancli.PrintfOK("now serving directory: '%v' on port: '%v', mirror dir is: '%v'", c.masterPath, *c.port, c.mirrorPath)
		err := s.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()
	go func() {
		ancli.PrintOK("starting fsnotify file detector")
		if err := c.fileserver.Start(ctx); err != nil {
			fsErrChan <- err
		}
	}()

	select {
	case <-ctx.Done():
	case serveErr := <-serverErrChan:
		err = serveErr
		break
	case fsErr := <-fsErrChan:
		err = fsErr
		break
	}

	ancli.PrintNotice("initiating webserver graceful shutdown")
	s.Shutdown(ctx)
	ancli.PrintOK("shutdown complete")
	return err
}

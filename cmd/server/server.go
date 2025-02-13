package server

import (
	"context"
	"flag"
	"os"

	"github.com/gorilla/websocket"
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

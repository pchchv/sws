package server

import (
	"context"

	"github.com/gorilla/websocket"
)

type Fileserver interface {
	Setup(pathToMaster string) (string, error)
	Start(ctx context.Context) error
	WsHandler(ws *websocket.Conn)
}

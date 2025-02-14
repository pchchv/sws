package wsinject

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/pchchv/sws/helpers/ancli"
	"golang.org/x/net/websocket"
)

// WsHandler sends page reload notifications to the connected websocket.
func (fs *Fileserver) WsHandler(ws *websocket.Conn) {
	reloadChan := make(chan string)
	killChan := make(chan struct{})
	name := "ws-" + fmt.Sprintf("%v", rand.Int())
	go func() {
		ancli.PrintfOK("new websocket connection: '%v'", ws.Config().Origin)
		for {
			pageToReload, ok := <-reloadChan
			if !ok {
				killChan <- struct{}{}
			}
			err := websocket.Message.Send(ws, pageToReload)
			if err != nil {
				// exit on error
				ancli.PrintfErr("ws: failed to send message via ws: %e", err)
				killChan <- struct{}{}
			}
		}
	}()

	ancli.PrintOK("Listening to file changes on pageReloadChan")
	fs.registerWs(name, reloadChan)
	<-killChan
	ancli.PrintOK("websocket disconnected")
	fs.deregisterWs(name)
	if err := ws.WriteClose(1005); err != nil {
		ancli.PrintfErr("ws-listener: '%s' got err when writeclosing: %e", name, err)
	}

	if err := ws.Close(); err != nil {
		ancli.PrintfErr("ws-listener: '%s' got err when closing: %e", name, err)
	}
}

func (fs *Fileserver) wsDispatcherStart() {
	for {
		pageToReload, ok := <-fs.pageReloadChan
		if !ok {
			ancli.PrintNotice("stopping wsDispatcher")
			fs.wsDispatcher.Range(func(key, value any) bool {
				ancli.PrintfNotice("sending to: '%v'", key)
				wsWriterChan := value.(chan string)
				// close chan to stop the wsRoutine
				close(wsWriterChan)
				return true
			})
			return
		}
		ancli.PrintfNotice("got update: '%v'", pageToReload)
		fs.wsDispatcher.Range(func(key, value any) bool {
			ancli.PrintfNotice("sending to: '%v'", key)
			wsWriterChan := value.(chan string)
			wsWriterChan <- pageToReload
			return true
		})
	}
}

func (fs *Fileserver) registerWs(name string, c chan string) {
	if !read(fs.wsDispatcherStartedMu, fs.wsDispatcherStarted) {
		go fs.wsDispatcherStart()
		write(fs.wsDispatcherStartedMu, true, fs.wsDispatcherStarted)
	}
	ancli.PrintfNotice("registering: '%v'", name)
	fs.wsDispatcher.Store(name, c)
}

func (fs *Fileserver) deregisterWs(name string) {
	fs.wsDispatcher.Delete(name)
}

// Reads by locking the mutex before taking a copy, will then return the copy.
func read[T any](m *sync.Mutex, src *T) T {
	m.Lock()
	defer m.Unlock()
	return *src
}

// Writes by locking the mutex before writing.
func write[T any](m *sync.Mutex, value T, dest *T) {
	m.Lock()
	defer m.Unlock()
	*dest = value
}

package wsinject

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pchchv/sws/helpers/ancli"
)

func TestWsHandler(t *testing.T) {
	ancli.Newline = true
	setup := func(t *testing.T) (*Fileserver, *websocket.Dialer, *httptest.Server) {
		t.Helper()
		started := false
		fs := &Fileserver{
			pageReloadChan:        make(chan string),
			wsDispatcher:          sync.Map{},
			wsDispatcherStarted:   &started,
			wsDispatcherStartedMu: &sync.Mutex{},
		}

		handler := http.NewServeMux()
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		handler.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Fatalf("Failed to upgrade HTTP connection to WebSocket: %v", err)
			}
			fs.WsHandler(ws)
		})

		server := httptest.NewServer(handler)
		dialer := &websocket.Dialer{}
		return fs, dialer, server
	}

	t.Run("it should send messages posted on pageReloadChan", func(t *testing.T) {
		fs, dialer, testServer := setup(t)
		ws, _, err := dialer.Dial(fmt.Sprintf("ws://localhost:%v/ws", testServer.Listener.Addr().(*net.TCPAddr).Port), nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		t.Cleanup(func() {
			testServer.Close()
			ws.Close()
		})

		go func() {
			fs.pageReloadChan <- "test message"
		}()

		var msg []byte
		if _, msg, err = ws.ReadMessage(); err != nil {
			t.Fatalf("Failed to receive message: %v", err)
		}

		if string(msg) != "test message" {
			t.Fatalf("Expected 'test message', got: %v", string(msg))
		}

		close(fs.pageReloadChan)
		select {
		case <-time.After(time.Second):
			t.Fatal("Expected the WebSocket to be closed")
		case <-fs.pageReloadChan:
		}
	})

	t.Run("it should handle multiple connections at once", func(t *testing.T) {
		fs, dialer, testServer := setup(t)
		mockWebClient0, _, err := dialer.Dial(fmt.Sprintf("ws://localhost:%v/ws", testServer.Listener.Addr().(*net.TCPAddr).Port), nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}

		mockWebClient1, _, err := dialer.Dial(fmt.Sprintf("ws://localhost:%v/ws", testServer.Listener.Addr().(*net.TCPAddr).Port), nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}

		t.Cleanup(func() {
			mockWebClient0.Close()
			mockWebClient1.Close()
			testServer.Close()
		})

		mu := &sync.Mutex{}
		go func() {
			mu.Lock()
			defer mu.Unlock()
			fs.pageReloadChan <- "test message"
		}()

		gotMsgChan := make(chan string)
		errChan := make(chan error)
		for _, wsClient := range []*websocket.Conn{mockWebClient0, mockWebClient1} {
			go func(wsClient *websocket.Conn) {
				for {
					_, msg, err := wsClient.ReadMessage()
					if err != nil {
						t.Log(err)
					}
					gotMsgChan <- string(msg)
				}
			}(wsClient)
		}
		want := 0
		for want != 2 {
			select {
			case <-time.After(time.Second):
				t.Fatal("failed to receive data from websocket")
			case got := <-gotMsgChan:
				want += 1
				t.Logf("got message from mocked ws client: %v", got)
			case err := <-errChan:
				t.Fatal(err)
			}
		}

		mu.Lock()
		close(fs.pageReloadChan)
		mu.Unlock()
	})
}

package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type mockFileServer struct{}

func (m *mockFileServer) Setup(pathToMaster string) (string, error) {
	return "/mock/mirror/path", nil
}

func (m *mockFileServer) Start(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (m *mockFileServer) WsHandler(ws *websocket.Conn) {}

func Test_Setup(t *testing.T) {
	tmpDir := t.TempDir()
	t.Run("it should set masterPath to second argument", func(t *testing.T) {
		want := tmpDir
		c := command{
			masterPath: "pre",
		}
		given := []string{want}
		if err := c.Flagset().Parse(given); err != nil {
			t.Fatalf("failed to parse flagset: %e", err)
		}
		c.Setup()
		if got := c.masterPath; got != want {
			t.Fatalf("expected: %s, got: %s", want, got)
		}
	})

	t.Run("it should set port arg", func(t *testing.T) {
		want := 9090
		c := command{}
		givenArgs := []string{"-port", "9090"}
		if err := c.Flagset().Parse(givenArgs); err != nil {
			t.Fatalf("failed to parse flagset: %e", err)
		}

		if err := c.Setup(); err != nil {
			t.Fatalf("failed to setup: %e", err)
		}

		if got := *c.port; got != want {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})

	t.Run("it should set cacheControl arg", func(t *testing.T) {
		want := "test"
		c := command{}
		givenArgs := []string{"-cacheControl", want}
		if err := c.Flagset().Parse(givenArgs); err != nil {
			t.Fatalf("failed to parse flagset: %e", err)
		}

		if err := c.Setup(); err != nil {
			t.Fatalf("failed to setup: %e", err)
		}

		if got := *c.cacheControl; got != want {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})
}

func TestRun(t *testing.T) {
	setup := func() command {
		cmd := command{}
		cmd.fileserver = &mockFileServer{}
		fs := cmd.Flagset()
		fs.Parse([]string{"--port=8081", "--wsPort=/test-ws"})

		if err := cmd.Setup(); err != nil {
			t.Fatalf("Setup failed: %e", err)
		}

		return cmd
	}

	t.Run("it should setup websocket handler on wsPort", func(t *testing.T) {
		cmd := setup()
		ctx, ctxCancel := context.WithCancel(context.Background())
		ready := make(chan struct{})
		go func() {
			close(ready)
			if err := cmd.Run(ctx); err != nil {
				t.Errorf("Run returned error: %e", err)
			}
		}()

		t.Cleanup(ctxCancel)

		<-ready
		// test if the HTTP server is working
		resp, err := http.Get("http://localhost:8081/")
		if err != nil {
			t.Fatalf("Failed to send GET request: %e", err)
		}
		t.Cleanup(func() { resp.Body.Close() })

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status OK, got: %v", resp.Status)
		}

		// test the websocket handler
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}

			// upgrade the HTTP connection to a WebSocket connection
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Failed to upgrade connection: %e", err)
				return
			}
			cmd.fileserver.WsHandler(ws)
		}))

		t.Cleanup(func() { server.Close() })

		wsURL := "ws" + server.URL[len("http"):]
		ws, _, err := websocket.DefaultDialer.Dial(wsURL+"/test-ws", nil)
		if err != nil {
			t.Fatalf("websocket dial failed: %e", err)
		}
		t.Cleanup(func() { ws.Close() })
	})

	t.Run("it should respond with correct cache control", func(t *testing.T) {
		cmd := setup()
		ctx, ctxCancel := context.WithCancel(context.Background())
		t.Cleanup(ctxCancel)
		want := "test"
		port := 13337
		cmd.cacheControl = &want
		cmd.port = &port
		ready := make(chan struct{})
		go func() {
			close(ready)
			if err := cmd.Run(ctx); err != nil {
				t.Errorf("Run returned error: %e", err)
			}
		}()
		<-ready
		time.Sleep(time.Millisecond)
		resp, err := http.Get(fmt.Sprintf("http://localhost:%v", port))
		if err != nil {
			t.Fatal(err)
		}
		if got := resp.Header.Get("Cache-Control"); got != want {
			t.Errorf("Cache-Control: expected %v, got %v", want, got)
		}
	})
}

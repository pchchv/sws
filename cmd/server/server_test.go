package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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

	t.Run("it should respond with correct headers", func(t *testing.T) {
		cmd := setup()
		ctx, ctxCancel := context.WithCancel(context.Background())
		t.Cleanup(ctxCancel)
		wantCacheControl := "test"
		port := 13337
		cmd.cacheControl = &wantCacheControl
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

		t.Run("cache-control", func(t *testing.T) {
			if got := resp.Header.Get("Cache-Control"); got != wantCacheControl {
				t.Errorf("Cache-Control: expected %v, got %v", wantCacheControl, got)
			}
		})

		t.Run("Cross-Origin-Opener-Policy", func(t *testing.T) {
			if got := resp.Header.Get("Cross-Origin-Opener-Policy"); got != "same-origin" {
				t.Errorf("Cross-Origin-Opener-Policy: expected same-origin, got %v", got)
			}
		})

		t.Run("Cross-Origin-Embedder-Policy", func(t *testing.T) {
			if got := resp.Header.Get("Cross-Origin-Embedder-Policy"); got != "require-corp" {
				t.Errorf("Cross-Origin-Embedder-Policy: expected require-corp, got %v", got)
			}
		})
	})

	t.Run("it should serve with tls if cert and key is specified", func(t *testing.T) {
		cmd := setup()
		ctx, ctxCancel := context.WithCancel(context.Background())
		t.Cleanup(ctxCancel)
		testCert := createTestFile(t, "cert.pem")
		testCert.Write([]byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`))
		testKey := createTestFile(t, "key.pem")
		testKey.Write([]byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`))
		port := 13337
		cmd.port = &port
		certPath := testCert.Name()
		cmd.tlsCertPath = &certPath
		keyPath := testKey.Name()
		cmd.tlsKeyPath = &keyPath
		ready := make(chan struct{})
		go func() {
			close(ready)
			if err := cmd.Run(ctx); err != nil {
				t.Errorf("Run returned error: %e", err)
			}
		}()
		<-ready
		time.Sleep(time.Millisecond)

		// cert above expired in 2018
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get(fmt.Sprintf("https://localhost:%v", port))
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code: %v", resp.StatusCode)
		}
	})
}

// createTestFile creates test file or fatal trying.
// Since it t.Fatalf on failure, return value won't matter.
// So the return value can be assumed to never be nil.
func createTestFile(t *testing.T, fileName string) *os.File {
	file, err := os.Create(fmt.Sprintf("%v/%v", t.TempDir(), fileName))
	if err != nil {
		t.Fatalf("failed to create file: %e", err)
	}
	return file
}

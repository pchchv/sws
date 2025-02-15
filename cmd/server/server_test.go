package server

import (
	"context"
	"os"
	"testing"

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

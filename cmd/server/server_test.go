package server

import "testing"

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

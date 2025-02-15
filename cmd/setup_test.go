package cmd

import (
	"errors"
	"testing"

	"github.com/pchchv/sws/cmd/server"
)

func Test_Parse(t *testing.T) {
	t.Run("it should return command if second argument specifies an existing command", func(t *testing.T) {
		want := server.Command()
		if got, err := Parse([]string{"/some/cli/path", "serve"}); err != nil {
			t.Fatalf(": %e", err)
		} else if got.Describe() != want.Describe() {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})

	t.Run("it should return command if second argument specifies shortcut of specific command", func(t *testing.T) {
		want := server.Command()
		if got, err := Parse([]string{"/some/cli/path", "s"}); err != nil {
			t.Fatalf(": %e", err)
		} else if got.Describe() != want.Describe() {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})

	t.Run("it should return error if command doesnt exist", func(t *testing.T) {
		badArg := "blheruh"
		want := ArgNotFoundError(badArg)
		if got, err := Parse([]string{"/some/cli/path", badArg}); got != nil {
			t.Fatalf("expected command to be nil, got: %+v", got)
		} else if err != want {
			t.Fatalf("expected: %v, got: %v", want, err)
		}
	})

	t.Run("it should return NoArgsError on lack of second argument", func(t *testing.T) {
		if _, err := Parse([]string{"/some/cli/path"}); !errors.Is(err, ErrNoArgs) {
			t.Fatalf("expected to get HelpfulError, got: %e", err)
		}
	})
}

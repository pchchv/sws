package ansiprint

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type ANSIPrint struct{}

func (a *ANSIPrint) Enabled(context.Context, slog.Level) bool {
	return true
}

func (a *ANSIPrint) Handle(ctx context.Context, r slog.Record) error {
	var bf bytes.Buffer
	if !r.Time.IsZero() {
		fmt.Fprintf(&bf, "%v %v", r.Time.Format(time.RFC3339), r.Message)
	}

	switch r.Level {
	case slog.LevelDebug, slog.LevelWarn, slog.LevelInfo:
		fmt.Fprint(os.Stdout, bf.String())
	case slog.LevelError:
		fmt.Fprint(os.Stderr, bf.String())
	}

	return nil
}

func (a *ANSIPrint) WithAttrs(attrs []slog.Attr) slog.Handler {
	return a
}

func (a *ANSIPrint) WithGroup(name string) slog.Handler {
	return a
}

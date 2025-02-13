package helpers

import (
	"context"
	"log/slog"
)

type ansiprint struct{}

func (a *ansiprint) Enabled(context.Context, slog.Level) bool {
	return true
}

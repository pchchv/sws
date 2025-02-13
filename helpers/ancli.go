package helpers

import "log/slog"

var (
	SlogIt  = false
	slogger *slog.Logger
)

func SetupSlog() {
	slogger = slog.New(&ansiprint{})
	SlogIt = true
}

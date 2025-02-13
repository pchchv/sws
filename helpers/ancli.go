package helpers

import (
	"log/slog"
	"os"
	"strings"
)

var (
	SlogIt  = false
	slogger *slog.Logger
	Newline = false || strings.ToLower(os.Getenv("ANCLI_NEWLINE")) == "true"
)

func SetupSlog() {
	slogger = slog.New(&ansiprint{})
	SlogIt = true
}

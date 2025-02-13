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

type colorCode int

const (
	RED colorCode = iota + 31
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
)

func SetupSlog() {
	slogger = slog.New(&ansiprint{})
	SlogIt = true
}

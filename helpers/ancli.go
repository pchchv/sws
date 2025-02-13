package helpers

import (
	"fmt"
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

func ColoredMessage(cc colorCode, msg string) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", cc, msg)
}

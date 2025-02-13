package helpers

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

var (
	slogger  *slog.Logger
	SlogIt   = false
	useColor = os.Getenv("NO_COLOR") != "true"
	Newline  = false || strings.ToLower(os.Getenv("ANCLI_NEWLINE")) == "true"
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

func printStatus(out io.Writer, status, msg string, color colorCode) {
	rawStatus := status
	if useColor {
		status = ColoredMessage(color, status)
	}

	var newline string
	if Newline {
		newline = "\n"
	}

	if SlogIt {
		if slogger == nil {
			SlogIt = false
			PrintErr("you have to run ancli.SetupSlog in order to use slog printing, defaulting to normal print")
		} else {
			// always newline slog messages
			fmsg := fmt.Sprintf("%v: %v\n", status, msg)
			switch rawStatus {
			case "ok", "notice":
				slogger.Info(fmsg)
			case "error":
				slogger.Error(fmsg)
			case "warning":
				slogger.Warn(fmsg)
			default:
				slogger.Warn(fmt.Sprintf("failed to find status for: '%v', msg is: %v", status, fmsg))
			}
		}
	} else {
		fmt.Fprintf(out, "%v: %v%v", status, msg, newline)
	}
}

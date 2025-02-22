package ancli

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/pchchv/sws/helpers/ansiprint"
)

var (
	slogger       *slog.Logger
	SlogIt        = false
	useColor      = os.Getenv("NO_COLOR") != "true"
	printWarnings = !truthy(os.Getenv("NO_WARNINGS"))
	Newline       = false || strings.ToLower(os.Getenv("ANCLI_NEWLINE")) == "true"
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
	slogger = slog.New(&ansiprint.ANSIPrint{})
	SlogIt = true
}

func ColoredMessage(cc colorCode, msg string) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", cc, msg)
}

func PrintErr(msg string) {
	printStatus(os.Stderr, "error", msg, RED)
}

func PrintfErr(msg string, a ...any) {
	PrintErr(fmt.Sprintf(msg, a...))
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

func PrintNotice(msg string) {
	printStatus(os.Stdout, "notice", msg, CYAN)
}

func PrintfNotice(msg string, a ...any) {
	PrintNotice(fmt.Sprintf(msg, a...))
}

func PrintOK(msg string) {
	printStatus(os.Stdout, "ok", msg, GREEN)
}

func PrintfOK(msg string, a ...any) {
	PrintOK(fmt.Sprintf(msg, a...))
}

func Okf(msg string, a ...any) {
	PrintOK(fmt.Sprintf(msg, a...))
}

func PrintWarn(msg string) {
	if printWarnings {
		printStatus(os.Stdout, "warning", msg, YELLOW)
	}
}

func PrintfWarn(msg string, a ...any) {
	PrintWarn(fmt.Sprintf(msg, a...))
}

func truthy(v any) bool {
	switch v := v.(type) {
	case bool:
		return v
	case int:
		return v == 1
	case string:
		if v == "" {
			return false
		}

		v = strings.TrimSpace(strings.ToLower(v))
		switch v {
		case "1":
			fallthrough
		case "true":
			return true
		}
	default:
		return v != nil
	}

	return false
}

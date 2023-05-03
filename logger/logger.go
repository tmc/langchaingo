package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

const (
	logformatEnvVarName = "LOG_FORMAT"
	maxMessageLen       = 1024

	logFormatNone = iota
	logFormatCompact
	logFormatFull
	logFormatJSON
)

func logFormat() int {
	// Set log format
	switch strings.ToLower(os.Getenv(logformatEnvVarName)) {
	case "compact":
		return logFormatCompact
	case "full":
		return logFormatFull
	default:
		return logFormatNone
	}
}

// banner displays a banner with the given title.
//
//nolint:forbidigo // This is a debug function. fmt.Print is necessary.
func banner(title string) {
	lf := logFormat()
	if lf == logFormatNone {
		return
	}

	// Full banner
	if lf == logFormatFull {
		sb := strings.Builder{}
		for i := 0; i < len(title)+8; i++ {
			sb.WriteString("-")
		}
		sb.WriteString("\n")
		color.HiBlack(sb.String())
		color.HiBlack("--- " + title + " ---\n")
		color.HiBlack(sb.String())
		return
	}

	// Compact banner
	fmt.Print(color.HiBlackString(title + ">>"))
}

// message displays a message with the given label and message.
//
//nolint:forbidigo // This is a debug function. fmt.Print is necessary.
func message(label string, msg string, colorFunc func(format string, a ...interface{})) {
	lf := logFormat()
	if lf == logFormatNone {
		return
	}

	// Full message
	if lf == logFormatFull {
		colorFunc(label + "\n")
		fmt.Println(msg + "\n")
		return
	}

	// Compact message
	msg = strings.ReplaceAll(msg, "\n", "\\n")
	colorFunc(label + ">>")
	msg = strings.TrimSpace(msg)
	fmt.Print(msg)
	if len(msg) > maxMessageLen {
		fmt.Print(msg[:maxMessageLen] + "... ")
		color.Yellow("(truncated)")
	}
	fmt.Println()
}

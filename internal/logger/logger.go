// Package logger centralizes construction of the application's zerolog logger.
//
// It supports two output formats:
//   - "console" (default): human-readable, colorized output for local CLI use.
//   - "json": structured JSON output suited for aggregation systems such as
//     Grafana Loki.
//
// The format is selected at startup so the rest of the codebase keeps using the
// standard zerolog API unchanged.
package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// FormatConsole renders human-readable, colorized output for the CLI.
const FormatConsole = "console"

// FormatJSON renders structured JSON output for log aggregation (e.g. Loki).
const FormatJSON = "json"

// New builds a zerolog.Logger for the given level and format.
//
// Any format other than "json" falls back to the console writer, so an empty
// or misconfigured value still yields readable output.
func New(level zerolog.Level, format string) zerolog.Logger {
	// RFC3339 timestamps are both readable and what Loki expects, so use them
	// for every format rather than the Unix epoch numbers.
	zerolog.TimeFieldFormat = time.RFC3339

	var out io.Writer = os.Stderr
	if !strings.EqualFold(format, FormatJSON) {
		out = zerolog.ConsoleWriter{
			Out:          os.Stderr,
			TimeFormat:   "15:04:05",
			FormatCaller: shortCaller,
		}
	}

	return zerolog.New(out).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
}

// shortCaller trims the caller to "package/file.go:line" so console lines stay
// compact instead of printing the full absolute path.
func shortCaller(i interface{}) string {
	caller, ok := i.(string)
	if !ok {
		return ""
	}

	dir, file := filepath.Split(caller)
	return filepath.Base(filepath.Clean(dir)) + "/" + file
}

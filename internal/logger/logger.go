// Package logger provides structured logging helpers for arangoadmin.
package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

// NewLogger builds a slog.Logger from CLI log-level and log-format flags.
func NewLogger(cmd *cli.Command) *slog.Logger {
	level := ParseLogLevel(cmd.String("log-level"))
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	switch cmd.String("log-format") {
	case "text":
		h = slog.NewTextHandler(os.Stderr, opts)
	default:
		h = slog.NewJSONHandler(os.Stderr, opts)
	}
	return slog.New(h)
}

// ParseLogLevel converts a log level string to slog.Level.
func ParseLogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

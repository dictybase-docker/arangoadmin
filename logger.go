package main

import (
	"log/slog"
	"os"
	"strings"

	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	"github.com/urfave/cli/v3"
)

func newLogger(cmd *cli.Command) *slog.Logger {
	level := parseLogLevel(cmd.String("log-level"))
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

func parseLogLevel(levelStr string) slog.Level {
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

// Typed logger functions for IOE.ChainFirstIOK
// Each takes all needed values by parameter and returns IO.IO[F.Void]

// nolint:unused // Used in Phase 1 implementation
func logUserCreated(logger *slog.Logger, username string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("user created", "username", username)
		return F.VOID
	}
}

// nolint:unused // Used in Phase 1 implementation
func logUserExists(logger *slog.Logger, username string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("user exists", "username", username)
		return F.VOID
	}
}

// nolint:unused // Used in Phase 2 implementation
func logUserUpdated(logger *slog.Logger, username string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("user updated", "username", username)
		return F.VOID
	}
}

// nolint:unused // Used in Phase 3 implementation
func logDatabaseCreated(logger *slog.Logger, dbname string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("database created", "database", dbname)
		return F.VOID
	}
}

// nolint:unused // Used in Phase 3 implementation
func logDatabaseExists(logger *slog.Logger, dbname string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("database exists", "database", dbname)
		return F.VOID
	}
}

// nolint:unused // Used in Phase 3 implementation
func logGrantAccess(logger *slog.Logger, username, dbname, grant string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("database access granted", "username", username, "database", dbname, "grant", grant)
		return F.VOID
	}
}

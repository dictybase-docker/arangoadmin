package main

import (
	"log/slog"
	"os"
	"strings"

	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/pair"
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

func logUserCreated(logger *slog.Logger, username string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("user status", "username", username, "status", "created")
		return F.VOID
	}
}

func logCreateUserOutcome(logger *slog.Logger, result CreateUserResult) {
	status := F.Pipe2(
		P.First(result),
		O.FromPredicate(F.Identity[bool]),
		O.Fold(
			func() string { return "existing" },
			func(_ bool) string { return "created" },
		),
	)
	logger.Info(
		"user status",
		"username",
		P.Second(result).Name(),
		"status",
		status,
	)
}

func logUserUpdated(logger *slog.Logger, username string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("user status", "username", username, "status", "updated")
		return F.VOID
	}
}

func logDatabaseCreated(logger *slog.Logger, dbname string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("database created", "database", dbname)
		return F.VOID
	}
}

func logDatabaseExists(logger *slog.Logger, dbname string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("database exists", "database", dbname)
		return F.VOID
	}
}

func logGrantAccess(
	logger *slog.Logger,
	username, dbname, grant string,
) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info(
			"database access granted",
			"username",
			username,
			"database",
			dbname,
			"grant",
			grant,
		)
		return F.VOID
	}
}

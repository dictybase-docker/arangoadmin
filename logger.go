package main

import (
	"log/slog"
	"os"
	"strings"

	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
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

func logEnsureUserOutcome(logger *slog.Logger, result EnsureUserResult) {
	logger.Info(
		"user status",
		"username",
		F.Pipe2(result, P.Second, userName),
		"status",
		F.Pipe1(result, P.First),
	)
}

func logEnsureDatabaseOutcome(logger *slog.Logger, result EnsureDatabaseResult) {
	logger.Info(
		"database status",
		"database",
		P.Second(result),
		"status",
		F.Pipe2(result, P.First, statusFromCreated),
	)
}

func logEnsureGrantOutcome(logger *slog.Logger, result EnsureGrantResult) {
	logger.Info(
		"grant status",
		"database",
		P.First(result),
		"grant",
		P.Second(result),
	)
}

func statusFromCreated(created bool) string {
	return F.Pipe2(
		created,
		O.FromPredicate(F.Identity[bool]),
		O.Fold(
			func() string { return "existing" },
			func(_ bool) string { return "created" },
		),
	)
}

func userName(u driver.User) string { return u.Name() }

func logEnsureUser(logger *slog.Logger) func(EnsureUserResult) IO.IO[F.Void] {
	return func(result EnsureUserResult) IO.IO[F.Void] {
		return func() F.Void {
			logEnsureUserOutcome(logger, result)
			return F.VOID
		}
	}
}

func logEnsureDatabase(logger *slog.Logger) func(EnsureDatabaseResult) IO.IO[F.Void] {
	return func(result EnsureDatabaseResult) IO.IO[F.Void] {
		return func() F.Void {
			logEnsureDatabaseOutcome(logger, result)
			return F.VOID
		}
	}
}

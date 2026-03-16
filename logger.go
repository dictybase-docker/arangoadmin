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

func logCreateUserOutcome(logger *slog.Logger, result CreateUserResult) {
	status := statusFromCreated(P.First(result))
	logger.Info(
		"user status",
		"username",
		P.Second(result).Name(),
		"status",
		status,
	)
}

func logCreateDatabaseOutcome(logger *slog.Logger, result CreateDatabaseResult) {
	for _, dbResult := range result.Databases {
		logger.Info(
			"database status",
			"database",
			P.Second(dbResult),
			"status",
			statusFromCreated(P.First(dbResult)),
		)
	}

	if result.User == nil {
		return
	}

	logCreateUserOutcome(logger, *result.User)
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

func logUserUpdated(logger *slog.Logger, username string) IO.IO[F.Void] {
	return func() F.Void {
		logger.Info("user status", "username", username, "status", "updated")
		return F.VOID
	}
}

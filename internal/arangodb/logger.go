package arangodb

import (
	"log/slog"

	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
)

func logEnsureUserOutcome(logger *slog.Logger, result EnsureUserResult) {
	logger.Info(
		"user status",
		"username",
		F.Pipe2(result, P.Second[EnsureUserStatus, driver.User], func(u driver.User) string { return u.Name() }),
		"status",
		F.Pipe1(result, P.First[EnsureUserStatus, driver.User]),
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

// LogEnsureUser returns an IOK action that logs the ensure-user result.
func LogEnsureUser(logger *slog.Logger) func(EnsureUserResult) IO.IO[F.Void] {
	return func(result EnsureUserResult) IO.IO[F.Void] {
		return func() F.Void {
			logEnsureUserOutcome(logger, result)
			return F.VOID
		}
	}
}

// LogEnsureDatabase returns an IOK action that logs the ensure-database result.
func LogEnsureDatabase(logger *slog.Logger) func(EnsureDatabaseResult) IO.IO[F.Void] {
	return func(result EnsureDatabaseResult) IO.IO[F.Void] {
		return func() F.Void {
			logger.Info(
				"database status",
				"database", P.Second(result),
				"created", F.Pipe2(result, P.First, statusFromCreated),
			)
			return F.VOID
		}
	}
}

// LogEnsureGrant returns an IOK action that logs the ensure-grant result.
func LogEnsureGrant(logger *slog.Logger) func(EnsureGrantResult) IO.IO[F.Void] {
	return func(result EnsureGrantResult) IO.IO[F.Void] {
		return func() F.Void {
			logEnsureGrantOutcome(logger, result)
			return F.VOID
		}
	}
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

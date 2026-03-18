package arangodb

import (
	"log/slog"

	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
)

// LogEnsureUserOutcome logs the result of the ensure-user command.
func LogEnsureUserOutcome(logger *slog.Logger, result EnsureUserResult) {
	logger.Info(
		"user status",
		"username",
		F.Pipe2(result, P.Second[EnsureUserStatus, driver.User], func(u driver.User) string { return u.Name() }),
		"status",
		F.Pipe1(result, P.First[EnsureUserStatus, driver.User]),
	)
}

// LogEnsureDatabaseOutcome logs the result of the ensure-database command.
func LogEnsureDatabaseOutcome(logger *slog.Logger, result EnsureDatabaseResult) {
	logger.Info(
		"database status",
		"database", P.Second(result),
		"status", F.Pipe2(result, P.First, statusFromCreated),
	)
}

// LogEnsureGrantOutcome logs the result of the ensure-grant command.
func LogEnsureGrantOutcome(logger *slog.Logger, result EnsureGrantResult) {
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

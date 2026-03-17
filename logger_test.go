package main

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

type testUser struct {
	name                 string
	active               bool
	passwordChangeNeeded bool
}

func (u testUser) Name() string { return u.name }

func (u testUser) IsActive() bool { return u.active }

func (u testUser) IsPasswordChangeNeeded() bool { return u.passwordChangeNeeded }

func (u testUser) Extra(_ interface{}) error { return nil }

func (u testUser) Remove(_ context.Context) error { return nil }

func (u testUser) Update(_ context.Context, _ driver.UserOptions) error { return nil }

func (u testUser) Replace(_ context.Context, _ driver.UserOptions) error { return nil }

func (u testUser) AccessibleDatabases(_ context.Context) ([]driver.Database, error) {
	return nil, nil
}

func (u testUser) SetDatabaseAccess(
	_ context.Context,
	_ driver.Database,
	_ driver.Grant,
) error {
	return nil
}

func (u testUser) GetDatabaseAccess(
	_ context.Context,
	_ driver.Database,
) (driver.Grant, error) {
	return driver.GrantNone, nil
}

func (u testUser) RemoveDatabaseAccess(_ context.Context, _ driver.Database) error {
	return nil
}

func (u testUser) SetCollectionAccess(
	_ context.Context,
	_ driver.AccessTarget,
	_ driver.Grant,
) error {
	return nil
}

func (u testUser) GetCollectionAccess(
	_ context.Context,
	_ driver.AccessTarget,
) (driver.Grant, error) {
	return driver.GrantNone, nil
}

func (u testUser) RemoveCollectionAccess(
	_ context.Context,
	_ driver.AccessTarget,
) error {
	return nil
}

func (u testUser) GrantReadWriteAccess(_ context.Context, _ driver.Database) error {
	return nil
}

func (u testUser) RevokeAccess(_ context.Context, _ driver.Database) error { return nil }

func TestLogCreateUserOutcomeIncludesStatus(t *testing.T) {
	require := require.New(t)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	logCreateUserOutcome(logger, P.MakePair[bool, driver.User](true, testUser{
		name:   "created-user",
		active: true,
	}))
	output := buf.String()
	require.Contains(output, "msg=\"user status\"")
	require.Contains(output, "username=created-user")
	require.Contains(output, "status=created")

	buf.Reset()

	logCreateUserOutcome(logger, P.MakePair[bool, driver.User](false, testUser{
		name:   "existing-user",
		active: true,
	}))
	output = buf.String()
	require.Contains(output, "msg=\"user status\"")
	require.Contains(output, "username=existing-user")
	require.Contains(output, "status=existing")
}

func TestLogUserUpdatedIncludesStatus(t *testing.T) {
	require := require.New(t)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	logUserUpdated(logger, "updated-user")()

	output := buf.String()
	require.Contains(output, "msg=\"user status\"")
	require.Contains(output, "username=updated-user")
	require.Contains(output, "status=updated")
}

func TestLogCreateDatabaseOutcomeIncludesStatuses(t *testing.T) {
	require := require.New(t)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	userResult := P.MakePair[bool, driver.User](true, testUser{
		name:   "db-user",
		active: true,
	})
	result := CreateDatabaseResult{
		Databases: []CreateSingleDBResult{
			P.MakePair(true, "db_created"),
			P.MakePair(false, "db_existing"),
		},
		HasUser: true,
		User:    userResult,
		Grants: []CreateGrantResult{
			P.MakePair("db_created", "rw"),
			P.MakePair("db_existing", "rw"),
		},
	}

	logCreateDatabaseOutcome(logger, result)
	output := buf.String()

	require.Contains(output, "msg=\"database status\"")
	require.Contains(output, "database=db_created")
	require.Contains(output, "status=created")
	require.Contains(output, "database=db_existing")
	require.Contains(output, "status=existing")
	require.Contains(output, "msg=\"user status\"")
	require.Contains(output, "username=db-user")
}

func TestLogEnsureUserOutcome(t *testing.T) {
	require := require.New(t)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	result := P.MakePair[EnsureUserStatus, driver.User](UserCreated, testUser{
		name: "ensure-user",
	})

	logEnsureUserOutcome(logger, result)
	output := buf.String()

	require.Contains(output, "msg=\"user status\"")
	require.Contains(output, "username=ensure-user")
	require.Contains(output, "status=created")
}

func TestParseLogLevel(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(slog.LevelDebug, parseLogLevel("debug"))
	assert.Equal(slog.LevelInfo, parseLogLevel("info"))
	assert.Equal(slog.LevelWarn, parseLogLevel("warn"))
	assert.Equal(slog.LevelError, parseLogLevel("error"))
	assert.Equal(slog.LevelDebug, parseLogLevel("DEBUG"))
	assert.Equal(slog.LevelInfo, parseLogLevel("invalid"))
}

func TestNewLogger(t *testing.T) {
	assert := assert.New(t)

	// Test text format and various levels
	cmd := &cli.Command{
		Flags: globalFlags(),
	}
	ctx := context.Background()

	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		args := []string{"arangoadmin", "--log-format", "text", "--log-level", level}
		err := cmd.Run(ctx, args)
		assert.NoError(err)
		logger := newLogger(cmd)
		assert.NotNil(logger)
	}

	// Test JSON format (default)
	args := []string{"arangoadmin", "--log-format", "json"}
	_ = cmd.Run(ctx, args)
	logger := newLogger(cmd)
	assert.NotNil(logger)
}

func TestLogEnsureGrant(t *testing.T) {
	require := require.New(t)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	result := P.MakePair("grant-db", "ro")
	logEnsureGrant(logger)(result)()

	output := buf.String()
	require.Contains(output, "msg=\"grant status\"")
	require.Contains(output, "database=grant-db")
	require.Contains(output, "grant=ro")
}

package arangodb

import (
	"bytes"
	"log/slog"
	"testing"

	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/require"
)

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

func TestLogEnsureGrant(t *testing.T) {
	require := require.New(t)
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	result := P.MakePair("grant-db", "ro")
	logEnsureGrantOutcome(logger, result)

	output := buf.String()
	require.Contains(output, "msg=\"grant status\"")
	require.Contains(output, "database=grant-db")
	require.Contains(output, "grant=ro")
}

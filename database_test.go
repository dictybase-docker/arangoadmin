package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	driver "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestEnsureDatabase(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			ensureDatabaseCommand(),
		},
	}

	dbName := "ensuredb"
	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-database",
		"--admin-password", arangoPassword,
		"--database", dbName,
	}

	// 1. First run: missing database -> created
	err := cmd.Run(ctx, args)
	assert.NoError(err)

	client, err := getTestClient()
	assert.NoError(err)
	ok, err := client.DatabaseExists(ctx, dbName)
	assert.NoError(err)
	assert.True(ok, "database should be created")

	// 2. Second run: existing database -> existing
	err = cmd.Run(ctx, args)
	assert.NoError(err)
	ok, err = client.DatabaseExists(ctx, dbName)
	assert.NoError(err)
	assert.True(ok, "database should still exist")
}

func TestEnsureDatabaseError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			ensureDatabaseCommand(),
		},
	}

	args := []string{
		"arangoadmin",
		"--host", "nonexistent-host",
		"--port", "8529",
		"ensure-database",
		"--admin-password", "wrong",
		"--database", "error-db",
	}

	err := cmd.Run(ctx, args)
	assert.Error(err, "should fail with invalid connection params")
}

type mockClient struct {
	driver.Client
	userExistsError error
	userError       error
	createUserError error
	userExists      bool
}

func (m *mockClient) UserExists(_ context.Context, _ string) (bool, error) {
	return m.userExists, m.userExistsError
}

func (m *mockClient) User(_ context.Context, _ string) (driver.User, error) {
	return nil, m.userError
}

func (m *mockClient) CreateUser(_ context.Context, _ string, _ *driver.UserOptions) (driver.User, error) {
	return nil, m.createUserError
}

func (m *mockClient) Database(_ context.Context, _ string) (driver.Database, error) {
	return nil, nil // Not used for this test
}


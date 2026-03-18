package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestEnsureDatabase(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: GlobalFlags(),
		Commands: []*cli.Command{
			EnsureDatabaseCommand(),
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
		Flags: GlobalFlags(),
		Commands: []*cli.Command{
			EnsureDatabaseCommand(),
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
	assert.Error(
		err,
		"should fail with invalid connection params",
	)
}

package main

import (
	"context"
	"log/slog"
	"os"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	driver "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestCreateDatabase(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			createDatabaseCommand(),
		},
	}

	dbName := "testdb"
	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"create-database",
		"--admin-password", arangoPassword,
		"--database", dbName,
	}

	err := cmd.Run(ctx, args)
	assert.NoError(err)

	// Verify database exists
	client, err := getTestClient()
	assert.NoError(err)

	ok, err := client.DatabaseExists(ctx, dbName)
	assert.NoError(err)
	assert.True(ok, "database should exist")

	// Test multiple databases
	dbNames := []string{"testdb2", "testdb3"}
	args = []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"create-database",
		"--admin-password", arangoPassword,
		"--database", dbNames[0],
		"--database", dbNames[1],
	}
	err = cmd.Run(ctx, args)
	assert.NoError(err)

	for _, name := range dbNames {
		ok, err := client.DatabaseExists(ctx, name)
		assert.NoError(err)
		assert.True(ok, "database %s should exist", name)
	}
}

func TestCreateDatabaseWithUser(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			createDatabaseCommand(),
		},
	}

	dbName := "testdb_user"
	user := "testuser"
	pass := "testpass"
	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"create-database",
		"--admin-password", arangoPassword,
		"--database", dbName,
		"--user", user,
		"--password", pass,
	}

	err := cmd.Run(ctx, args)
	assert.NoError(err)

	client, err := getTestClient()
	assert.NoError(err)

	// Verify user exists
	ok, err := client.UserExists(ctx, user)
	assert.NoError(err)
	assert.True(ok, "user should exist")

	// Verify database exists
	ok, err = client.DatabaseExists(ctx, dbName)
	assert.NoError(err)
	assert.True(ok, "database should exist")
}

// Phase 3 unit tests for database pipelines
func TestCreateSingleDatabaseNew(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	client, err := getClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
	require.NoError(err)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p := DatabaseParams{
		WithClient: WithClient{Client: client, Logger: logger},
	}
	result := toEither(createSingleDatabase(p)("fptest_newdb"))
	require.True(E.IsRight(result), "createSingleDatabase should succeed for new database")

	ok, err := client.DatabaseExists(ctx, "fptest_newdb")
	require.NoError(err)
	require.True(ok, "database should exist after creation")
}

func TestCreateSingleDatabaseIdempotent(t *testing.T) {
	require := require.New(t)

	client, err := getClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
	require.NoError(err)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p := DatabaseParams{
		WithClient: WithClient{Client: client, Logger: logger},
	}
	r1 := toEither(createSingleDatabase(p)("fptest_idempotentdb"))
	require.True(E.IsRight(r1))
	r2 := toEither(createSingleDatabase(p)("fptest_idempotentdb"))
	require.True(E.IsRight(r2), "creating same database twice should succeed (idempotent)")
}

func TestGrantSingleDatabase(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	client, err := getClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
	require.NoError(err)

	// Pre-create database and user
	_, err = client.CreateDatabase(ctx, "fptest_grantdb", nil)
	require.NoError(err)
	user, err := client.CreateUser(ctx, "fptest_grantusr", &driver.UserOptions{Password: "p"})
	require.NoError(err)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	uwg := UserWithGrant{
		Params: DatabaseParams{
			WithClient: WithClient{Client: client, Logger: logger},
			Grant:      "rw",
			Username:   "fptest_grantusr",
		},
		User: user,
	}
	result := toEither(grantSingleDatabase(uwg)("fptest_grantdb"))
	require.True(E.IsRight(result), "grantSingleDatabase should succeed")
}

func TestCreateDatabasePipelineWithUser(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	client, err := getClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
	require.NoError(err)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p := DatabaseParams{
		WithClient: WithClient{Client: client, Logger: logger},
		Databases:  []string{"fptest_grantdb1", "fptest_grantdb2"},
		Username:   "fptest_grantuser",
		Password:   "pass",
		Grant:      "rw",
	}
	result := toEither(createDatabasePipeline(p))
	require.True(E.IsRight(result), "createDatabasePipeline should succeed")

	// Verify databases exist
	for _, db := range p.Databases {
		ok, err := client.DatabaseExists(ctx, db)
		require.NoError(err)
		require.True(ok, "database %s should exist", db)
	}
	// Verify user exists
	ok, err := client.UserExists(ctx, p.Username)
	require.NoError(err)
	require.True(ok, "user should exist")
}

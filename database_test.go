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

// Unit tests for database pipelines
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
	p := SingleDBParams{
		Client: client,
		Logger: logger,
		Dbname: "fptest_newdb",
	}
	result := toEither(createSingleDatabase(p))
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
	p := SingleDBParams{
		Client: client,
		Logger: logger,
		Dbname: "fptest_idempotentdb",
	}
	r1 := toEither(createSingleDatabase(p))
	require.True(E.IsRight(r1))
	r2 := toEither(createSingleDatabase(p))
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
	g := GrantDBParams{
		Client: client,
		Logger: logger,
		Dbname: "fptest_grantdb",
		Grant:  "rw",
		User:   user,
	}
	result := toEither(grantSingleDatabase(g))
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

	databases := []string{"fptest_grantdb1", "fptest_grantdb2"}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p := DatabaseParams{
		WithClient: WithClient{Client: client, Logger: logger},
		Databases:  databases,
		Username:   "fptest_grantuser",
		Password:   "pass",
		Grant:      "rw",
	}
	result := toEither(createDatabasePipeline(p))
	require.True(E.IsRight(result), "database pipeline should succeed")

	// Verify databases exist
	for _, db := range databases {
		ok, err := client.DatabaseExists(ctx, db)
		require.NoError(err)
		require.True(ok, "database %s should exist", db)
	}
	// Verify user exists
	ok, err := client.UserExists(ctx, p.Username)
	require.NoError(err)
	require.True(ok, "user should exist")
}

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

func TestEnsureDatabaseParallel(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	dbName := "paralleldb"
	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-database",
		"--admin-password", arangoPassword,
		"--database", dbName,
	}

	const workers = 5
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			workerCmd := &cli.Command{
				Flags: globalFlags(),
				Commands: []*cli.Command{
					ensureDatabaseCommand(),
				},
			}
			errs <- workerCmd.Run(ctx, args)
		}()
	}

	for i := 0; i < workers; i++ {
		err := <-errs
		assert.NoError(err, "parallel ensure-database should succeed")
	}

	client, err := getTestClient()
	assert.NoError(err)
	ok, err := client.DatabaseExists(ctx, dbName)
	assert.NoError(err)
	assert.True(ok)
}

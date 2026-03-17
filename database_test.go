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
	dbName := "fptest_grantdb"
	userName := "fptest_grantusr"
	_, _ = client.CreateDatabase(ctx, dbName, nil)
	// If it already exists, that's fine for this test
	user, err := client.CreateUser(ctx, userName, &driver.UserOptions{Password: "p"})
	if err != nil {
		user, err = client.User(ctx, userName)
		require.NoError(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	g := GrantDBParams{
		Client: client,
		Logger: logger,
		Dbname: dbName,
		Grant:  "rw",
		User:   user,
	}
	result := toEither(grantSingleDatabase(g))
	require.True(E.IsRight(result), "grantSingleDatabase should succeed")
}

func TestGrantSingleDatabaseFailure(t *testing.T) {
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

	// Pre-create database
	dbName := "fptest_grantdb_fail"
	_, _ = client.CreateDatabase(ctx, dbName, nil)
	// Ignore error if it already exists

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	g := GrantDBParams{
		Client: client,
		Logger: logger,
		Dbname: dbName,
		Grant:  "rw",
		User:   errorUser{}, // This will fail SetDatabaseAccess
	}
	result := toEither(grantSingleDatabase(g))
	require.True(E.IsLeft(result), "grantSingleDatabase should fail when User.SetDatabaseAccess fails")
	_, err = E.UnwrapError(result)
	require.Contains(err.Error(), "error granting access to database")
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

func TestCreateUserAndGrantErrors(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// 1. UserExists fails
	m1 := &mockClient{userExistsError: fmt.Errorf("user exists fail")}
	p1 := DatabaseParams{
		WithClient: WithClient{Client: m1},
		Username:   "test",
	}
	res1 := toEither(createUserAndGrant(p1))
	assert.True(E.IsLeft(res1))
	_, err1 := E.UnwrapError(res1)
	require.Contains(err1.Error(), "user exists fail")

	// 2. User exists but fetching it fails
	m2 := &mockClient{userExists: true, userError: fmt.Errorf("user fetch fail")}
	p2 := DatabaseParams{
		WithClient: WithClient{Client: m2},
		Username:   "test",
	}
	res2 := toEither(createUserAndGrant(p2))
	assert.True(E.IsLeft(res2))
	_, err2 := E.UnwrapError(res2)
	require.Contains(err2.Error(), "user fetch fail")

	// 3. User does not exist and creation fails
	m3 := &mockClient{userExists: false, createUserError: fmt.Errorf("user create fail")}
	p3 := DatabaseParams{
		WithClient: WithClient{Client: m3},
		Username:   "test",
	}
	res3 := toEither(createUserAndGrant(p3))
	assert.True(E.IsLeft(res3))
	_, err3 := E.UnwrapError(res3)
	require.Contains(err3.Error(), "user create fail")
}

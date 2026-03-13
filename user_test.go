package main

import (
	"context"
	"log/slog"
	"os"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

func TestCreateUser(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			createUserCommand(),
		},
	}

	user := "newuser"
	pass := "newpass"
	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"create-user",
		"--admin-password", arangoPassword,
		"--user", user,
		"--password", pass,
	}

	err := cmd.Run(ctx, args)
	assert.NoError(err)

	client, err := getTestClient()
	assert.NoError(err)

	ok, err := client.UserExists(ctx, user)
	assert.NoError(err)
	assert.True(ok, "user should exist")
}

func TestUpdateUser(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			createDatabaseCommand(),
			updateUserCommand(),
		},
	}

	user := "updateuser"
	pass := "initialpass"
	dbName := "updatetestdb"
	// Create user with database first to ensure it has some permissions
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

	// Update password
	newPass := "updatedpass"
	args = []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"update-user",
		"--admin-password", arangoPassword,
		"--user", user,
		"--password", newPass,
	}
	err = cmd.Run(ctx, args)
	assert.NoError(err)

	// Verify we can connect with new password
	client, err := getClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     user,
		Pass:     newPass,
		IsSecure: false,
	})
	assert.NoError(err)
	// Try to get the database we have access to
	_, err = client.Database(ctx, dbName)
	assert.NoError(err, "should be able to access database with new password")

	// Verify we cannot connect with old password
	clientOld, err := getClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     user,
		Pass:     pass,
		IsSecure: false,
	})
	assert.NoError(err)
	_, err = clientOld.Database(ctx, dbName)
	assert.Error(err, "should not be able to access database with old password")
}

// Phase 1 unit tests for createUserIfNotExists pipeline
func TestCreateUserIfNotExistsNewUser(t *testing.T) {
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
	p := UserParams{
		WithClient: WithClient{Client: client, Logger: logger},
		Username:   "fptest_newuser",
		Password:   "testpass",
	}
	result := toEither(createUserIfNotExists(p))
	require.True(E.IsRight(result), "createUserIfNotExists should succeed for new user")

	// Verify user was actually created
	ok, err := client.UserExists(ctx, "fptest_newuser")
	require.NoError(err)
	require.True(ok, "user should exist after creation")
}

func TestCreateUserIfNotExistsIdempotent(t *testing.T) {
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
	p := UserParams{
		WithClient: WithClient{Client: client, Logger: logger},
		Username:   "fptest_idempotent",
		Password:   "testpass",
	}
	// Create once
	result1 := toEither(createUserIfNotExists(p))
	require.True(E.IsRight(result1))
	// Create again — should succeed (idempotent, logs "exists")
	result2 := toEither(createUserIfNotExists(p))
	require.True(E.IsRight(result2))
}

// Phase 2 unit tests for updateUserPipeline
func TestUpdateUserPipelineSuccess(t *testing.T) {
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

	// Pre-create user
	_, err = client.CreateUser(ctx, "fptest_update", &driver.UserOptions{Password: "oldpass"})
	require.NoError(err)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p := UserParams{
		WithClient: WithClient{Client: client, Logger: logger},
		Username:   "fptest_update",
		Password:   "newpass",
	}
	result := toEither(updateUserPipeline(p))
	require.True(E.IsRight(result), "updateUserPipeline should succeed for existing user")
}

func TestUpdateUserPipelineNonExistent(t *testing.T) {
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
	p := UserParams{
		WithClient: WithClient{Client: client, Logger: logger},
		Username:   "fptest_ghost",
		Password:   "pass",
	}
	result := toEither(updateUserPipeline(p))
	require.True(E.IsLeft(result), "updateUserPipeline should fail for non-existent user")
}

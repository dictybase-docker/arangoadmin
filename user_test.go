package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestCreateUserPipelineNewUser(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	client, err := getTestClient()
	assert.NoError(err)

	result, err := createOrGetUser(ctx, client, "pipelinenewuser", "pipelinepass")
	assert.NoError(err)
	assert.False(result.Existed, "user should have been newly created")
	assert.Equal("pipelinenewuser", result.Username)

	ok, err := client.UserExists(ctx, "pipelinenewuser")
	assert.NoError(err)
	assert.True(ok, "user should exist in database")
}

func TestCreateUserPipelineIdempotent(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	client, err := getTestClient()
	assert.NoError(err)

	// First call — user is created
	first, err := createOrGetUser(ctx, client, "idempotentuser", "idempotentpass")
	assert.NoError(err)
	assert.False(first.Existed, "first call should create user")
	assert.Equal("idempotentuser", first.Username)

	// Second call — user already exists
	second, err := createOrGetUser(ctx, client, "idempotentuser", "idempotentpass")
	assert.NoError(err)
	assert.True(second.Existed, "second call should report user already exists")
	assert.Equal("idempotentuser", second.Username)
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

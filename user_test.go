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

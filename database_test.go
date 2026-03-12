package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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

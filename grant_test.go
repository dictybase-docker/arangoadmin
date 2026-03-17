package main

import (
	"context"
	"fmt"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	driver "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestEnsureGrant(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			ensureGrantCommand(),
		},
	}

	dbName := "ensure-grant-db"
	userName := "ensure-grant-user"
	grantLevel := "ro"

	// Pre-requisites: database and user must exist
	client, err := getTestClient()
	assert.NoError(err)

	_, err = client.CreateDatabase(ctx, dbName, nil)
	assert.NoError(err)

	_, err = client.CreateUser(ctx, userName, &driver.UserOptions{Password: "password"})
	assert.NoError(err)

	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-grant",
		"--admin-password", arangoPassword,
		"--database", dbName,
		"--user", userName,
		"--grant", grantLevel,
	}

	// 1. Run ensure-grant
	err = cmd.Run(ctx, args)
	assert.NoError(err)

	// 2. Verify grant level
	user, err := client.User(ctx, userName)
	assert.NoError(err)

	db, err := client.Database(ctx, dbName)
	assert.NoError(err)

	access, err := user.GetDatabaseAccess(ctx, db)
	assert.NoError(err)
	assert.Equal(driver.GrantReadOnly, access, "grant level should be read-only")

	// 3. Update grant level
	grantLevel = "rw"
	args[len(args)-1] = grantLevel
	err = cmd.Run(ctx, args)
	assert.NoError(err)

	access, err = user.GetDatabaseAccess(ctx, db)
	assert.NoError(err)
	assert.Equal(driver.GrantReadWrite, access, "grant level should be read-write")
}

func TestGetGrantValues(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(driver.GrantReadWrite, getGrant("rw"))
	assert.Equal(driver.GrantReadOnly, getGrant("ro"))
	assert.Equal(driver.GrantNone, getGrant("none"))
	assert.Equal(driver.GrantNone, getGrant("invalid"))
	assert.Equal(driver.GrantNone, getGrant(""))
}

func TestEnsureGrantInvalidLevel(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			ensureGrantCommand(),
		},
	}

	dbName := "ensure-grant-invalid-db"
	userName := "ensure-grant-invalid-user"

	client, err := getTestClient()
	assert.NoError(err)
	_, _ = client.CreateDatabase(ctx, dbName, nil)
	_, _ = client.CreateUser(ctx, userName, &driver.UserOptions{Password: "password"})

	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-grant",
		"--admin-password", arangoPassword,
		"--database", dbName,
		"--user", userName,
		"--grant", "completely-invalid",
	}

	err = cmd.Run(ctx, args)
	assert.NoError(err, "should not fail for invalid grant level, just apply 'none'")

	user, _ := client.User(ctx, userName)
	db, _ := client.Database(ctx, dbName)
	access, _ := user.GetDatabaseAccess(ctx, db)
	assert.Equal(driver.GrantNone, access, "should default to GrantNone for invalid inputs")
}

type errorUser struct {
	testUser
}

func (u errorUser) SetDatabaseAccess(_ context.Context, _ driver.Database, _ driver.Grant) error {
	return fmt.Errorf("forced error")
}

func TestApplyGrantFailure(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	state := GrantState{
		Params: EnsureGrantParams{
			Context:  ctx,
			Database: "test-db",
		},
		User: errorUser{},
	}

	result := toEither(applyGrant(state))
	assert.True(E.IsLeft(result))
	_, err := E.UnwrapError(result)
	assert.Contains(err.Error(), "error granting access to database test-db")
	assert.Contains(err.Error(), "forced error")
}

func TestEnsureGrantMissingUser(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			ensureGrantCommand(),
		},
	}

	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-grant",
		"--admin-password", arangoPassword,
		"--database", "any-db",
		"--user", "non-existent-user",
	}

	err := cmd.Run(ctx, args)
	assert.Error(err, "should fail for missing user")
	assert.Contains(err.Error(), "error fetching user non-existent-user")
}

func TestEnsureGrantMissingDatabase(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			ensureGrantCommand(),
		},
	}

	// Pre-create user
	client, err := getTestClient()
	assert.NoError(err)
	userName := "grant-test-user-no-db"
	_, err = client.CreateUser(ctx, userName, &driver.UserOptions{Password: "password"})
	assert.NoError(err)

	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-grant",
		"--admin-password", arangoPassword,
		"--database", "non-existent-db",
		"--user", userName,
	}

	err = cmd.Run(ctx, args)
	assert.Error(err, "should fail for missing database")
	assert.Contains(err.Error(), "error fetching database non-existent-db")
}

func TestEnsureGrantEmptyInputs(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			ensureGrantCommand(),
		},
	}

	// 1. Test empty user
	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-grant",
		"--admin-password", arangoPassword,
		"--database", "any-db",
		"--user", "",
	}
	err := cmd.Run(ctx, args)
	assert.Error(err, "should fail for empty user")

	// 2. Test empty database
	args = []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-grant",
		"--admin-password", arangoPassword,
		"--database", "",
		"--user", "any-user",
	}
	err = cmd.Run(ctx, args)
	assert.Error(err, "should fail for empty database")
}

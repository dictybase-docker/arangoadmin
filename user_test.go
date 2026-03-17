package main

import (
	"context"
	"log/slog"
	"os"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Unit tests for createUserPipeline
func TestCreateUserPipelineNewUser(t *testing.T) {
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

	p := CreateUserParams{
		Client:   client,
		Username: "fptest_newuser",
		Password: "testpass",
	}
	result, err := toTuple(createUserPipeline(p))
	require.NoError(err, "createUserPipeline should succeed for new user")
	require.True(P.First(result), "true should mean newly created")
	require.Equal("fptest_newuser", P.Second(result).Name())

	// Verify user was actually created
	ok, err := client.UserExists(ctx, "fptest_newuser")
	require.NoError(err)
	require.True(ok, "user should exist after creation")
}

func TestCreateUserPipelineIdempotent(t *testing.T) {
	require := require.New(t)

	client, err := getClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
	require.NoError(err)

	p := CreateUserParams{
		Client:   client,
		Username: "fptest_idempotent",
		Password: "testpass",
	}
	// Create once
	result1, err := toTuple(createUserPipeline(p))
	require.NoError(err)
	require.True(P.First(result1), "first creation should report created=true")
	require.Equal("fptest_idempotent", P.Second(result1).Name())
	// Create again — should succeed (idempotent, logs "exists")
	result2, err := toTuple(createUserPipeline(p))
	require.NoError(err)
	require.False(P.First(result2), "second creation should report created=false")
	require.Equal("fptest_idempotent", P.Second(result2).Name())
}

// Unit tests for updateUserPipeline
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

func TestEnsureUser(t *testing.T) {
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
	user := "ensure_user_test"

	// 1. Create missing user
	p1 := EnsureUserParams{
		Client:   client,
		Logger:   logger,
		Username: user,
		Password: "pass1",
		Policy:   "never",
	}
	res1, err := toTuple(ensureUserPipeline(p1))
	require.NoError(err)
	require.Equal(UserCreated, P.First(res1))
	require.Equal(user, P.Second(res1).Name())

	// 2. Existing + never -> existing
	p2 := EnsureUserParams{
		Client:   client,
		Logger:   logger,
		Username: user,
		Password: "pass2",
		Policy:   "never",
	}
	res2, err := toTuple(ensureUserPipeline(p2))
	require.NoError(err)
	require.Equal(UserExisting, P.First(res2))

	// 3. Existing + always -> updated
	p3 := EnsureUserParams{
		Client:   client,
		Logger:   logger,
		Username: user,
		Password: "pass3",
		Policy:   "always",
	}
	res3, err := toTuple(ensureUserPipeline(p3))
	require.NoError(err)
	require.Equal(UserUpdated, P.First(res3))

	// 4. Existing + if-provided + no password -> existing
	p4 := EnsureUserParams{
		Client:   client,
		Logger:   logger,
		Username: user,
		Password: "",
		Policy:   "if-provided",
	}
	res4, err := toTuple(ensureUserPipeline(p4))
	require.NoError(err)
	require.Equal(UserExisting, P.First(res4))

	// 5. Existing + if-provided + password -> updated
	p5 := EnsureUserParams{
		Client:   client,
		Logger:   logger,
		Username: user,
		Password: "pass5",
		Policy:   "if-provided",
	}
	res5, err := toTuple(ensureUserPipeline(p5))
	require.NoError(err)
	require.Equal(UserUpdated, P.First(res5))

	// 6. Invalid policy -> existing
	p6 := EnsureUserParams{
		Client:   client,
		Logger:   logger,
		Username: user,
		Password: "pass6",
		Policy:   "invalid",
	}
	res6, err := toTuple(ensureUserPipeline(p6))
	require.NoError(err)
	require.Equal(UserExisting, P.First(res6))
}

func TestEnsureUserCLI(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			ensureUserCommand(),
		},
	}

	user := "cli_ensure_user"
	pass := "clipass"
	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-user",
		"--admin-password", arangoPassword,
		"--user", user,
		"--password", pass,
		"--password-policy", "always",
	}

	err := cmd.Run(ctx, args)
	require.NoError(err)

	client, err := getTestClient()
	require.NoError(err)

	ok, err := client.UserExists(ctx, user)
	require.NoError(err)
	require.True(ok, "user should exist after CLI ensure-user")
}

func TestGetGrant(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(driver.GrantReadWrite, getGrant("rw"))
	assert.Equal(driver.GrantReadOnly, getGrant("ro"))
	assert.Equal(driver.GrantNone, getGrant("none"))
	assert.Equal(driver.GrantNone, getGrant("invalid"))
}

func TestCreateUserAuthFailure(t *testing.T) {
	require := require.New(t)
	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			createUserCommand(),
		},
	}
	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"create-user",
		"--admin-password", "wrong-password",
		"--user", "failuser",
		"--password", "failpass",
	}
	err := cmd.Run(context.Background(), args)
	require.Error(err)
	require.Contains(err.Error(), "not authorized")
}

func TestEnsureUserAuthFailure(t *testing.T) {
	require := require.New(t)
	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			ensureUserCommand(),
		},
	}
	args := []string{
		"arangoadmin",
		"--host", arangoHost,
		"--port", arangoPort,
		"ensure-user",
		"--admin-password", "wrong-password",
		"--user", "failuser",
		"--password", "failpass",
	}
	err := cmd.Run(context.Background(), args)
	require.Error(err)
	require.Contains(err.Error(), "not authorized")
}

func TestConnectionFailure(t *testing.T) {
	require := require.New(t)
	cmd := &cli.Command{
		Flags: globalFlags(),
		Commands: []*cli.Command{
			createUserCommand(),
		},
	}
	args := []string{
		"arangoadmin",
		"--host", "nonexistent-host",
		"--port", "8529",
		"create-user",
		"--admin-password", arangoPassword,
		"--user", "failuser",
		"--password", "failpass",
	}
	err := cmd.Run(context.Background(), args)
	require.Error(err)
	// Connection errors occur at API calls in arangodb driver too
	require.Contains(err.Error(), "error checking for user")
}

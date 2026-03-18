package main

import (
	"context"
	"testing"

	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

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

	user := "ensure_user_test"

	// 1. Create missing user
	p1 := EnsureUserParams{
		Client:   client,
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
			ensureUserCommand(),
		},
	}
	args := []string{
		"arangoadmin",
		"--host", "nonexistent-host",
		"--port", "8529",
		"ensure-user",
		"--admin-password", arangoPassword,
		"--user", "failuser",
		"--password", "failpass",
	}
	err := cmd.Run(context.Background(), args)
	require.Error(err)
	// Connection errors occur at API calls in arangodb driver too
	require.Contains(err.Error(), "error checking for user")
}

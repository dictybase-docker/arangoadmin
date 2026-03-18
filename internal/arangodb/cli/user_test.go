package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestEnsureUserCLI(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: GlobalFlags(),
		Commands: []*cli.Command{
			EnsureUserCommand(),
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

func TestEnsureUserAuthFailure(t *testing.T) {
	require := require.New(t)
	cmd := &cli.Command{
		Flags: GlobalFlags(),
		Commands: []*cli.Command{
			EnsureUserCommand(),
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
		Flags: GlobalFlags(),
		Commands: []*cli.Command{
			EnsureUserCommand(),
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
	require.Contains(err.Error(), "error checking for user")
}

func TestNewLogger(t *testing.T) {
	assert := require.New(t)

	cmd := &cli.Command{
		Flags: GlobalFlags(),
	}
	ctx := context.Background()

	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		args := []string{"arangoadmin", "--log-format", "text", "--log-level", level}
		err := cmd.Run(ctx, args)
		assert.NoError(err)
	}

	// Test JSON format (default)
	args := []string{"arangoadmin", "--log-format", "json"}
	err := cmd.Run(ctx, args)
	assert.NoError(err)
}

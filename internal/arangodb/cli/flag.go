// Package cli provides CLI command definitions and action functions for arangoadmin.
package cli

import (
	"github.com/dictybase-docker/arangoadmin/internal/arangodb"
	"github.com/urfave/cli/v3"
)

// GlobalFlags returns the top-level flags shared by all commands.
func GlobalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Usage:   "arangodb host address",
			Sources: cli.EnvVars("ARANGODB_SERVICE_HOST"),
			Value:   "arangodb",
		},
		&cli.StringFlag{
			Name:    "port",
			Usage:   "arangodb port",
			Sources: cli.EnvVars("ARANGODB_SERVICE_PORT"),
			Value:   "8529",
		},
		&cli.StringFlag{
			Name:  "log-level",
			Usage: "log level for the application",
			Value: "info",
		},
		&cli.StringFlag{
			Name:  "log-format",
			Usage: "format of the logging out, either of json or text",
			Value: "json",
		},
		&cli.BoolFlag{
			Name:  "is-secure",
			Usage: "connect through a secure endpoint",
		},
	}
}

// connParamsFromCmd extracts connection parameters from CLI flags.
func connParamsFromCmd(cmd *cli.Command) arangodb.ConnectionParams {
	return arangodb.ConnectionParams{
		Host:     cmd.String("host"),
		Port:     cmd.String("port"),
		User:     cmd.String("admin-user"),
		Pass:     cmd.String("admin-password"),
		IsSecure: cmd.Bool("is-secure"),
	}
}

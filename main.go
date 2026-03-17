package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "arangoadmin",
		Usage:   "cli for creating databases and users in arangodb",
		Version: "1.0.0",
		Flags:   globalFlags(),
		Commands: []*cli.Command{
			ensureUserCommand(),
			ensureDatabaseCommand(),
			ensureGrantCommand(),
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func globalFlags() []cli.Flag {
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

func ensureUserCommand() *cli.Command {
	return &cli.Command{
		Name:   "ensure-user",
		Usage:  "create or update a user for accessing arangodb",
		Action: EnsureUser,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "admin-user",
				Aliases: []string{"au"},
				Usage:   "arangodb admin user",
				Value:   "root",
			},
			&cli.StringFlag{
				Name:    "admin-password",
				Aliases: []string{"ap"},
				Usage:   "arangodb admin password",
			},
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"u"},
				Usage:    "arangodb user",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "password",
				Aliases:  []string{"pw"},
				Usage:    "arangodb password for user",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "password-policy",
				Usage: "policy for updating password (never, if-provided, always)",
				Value: "never",
			},
		},
	}
}

func ensureDatabaseCommand() *cli.Command {
	return &cli.Command{
		Name:   "ensure-database",
		Usage:  "create a single arangodb database if missing",
		Action: EnsureDatabase,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "admin-user",
				Aliases: []string{"au"},
				Usage:   "arangodb admin user",
				Value:   "root",
			},
			&cli.StringFlag{
				Name:    "admin-password",
				Aliases: []string{"ap"},
				Usage:   "arangodb admin password",
			},
			&cli.StringFlag{
				Name:     "database",
				Aliases:  []string{"db"},
				Usage:    "name of arangodb database",
				Required: true,
			},
		},
	}
}

func ensureGrantCommand() *cli.Command {
	return &cli.Command{
		Name:   "ensure-grant",
		Usage:  "ensure a user has a specific grant level on a database",
		Action: EnsureGrant,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "admin-user",
				Aliases: []string{"au"},
				Usage:   "arangodb admin user",
				Value:   "root",
			},
			&cli.StringFlag{
				Name:    "admin-password",
				Aliases: []string{"ap"},
				Usage:   "arangodb admin password",
			},
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"u"},
				Usage:    "arangodb user",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "database",
				Aliases:  []string{"db"},
				Usage:    "name of arangodb database",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "grant",
				Aliases: []string{"g"},
				Usage:   "level of access for arangodb user (rw, ro, none)",
				Value:   "rw",
			},
		},
	}
}

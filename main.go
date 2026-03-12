package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "arangoadmin",
		Usage:   "cli for creating databases and users in arangodb",
		Version: "1.0.0",
		Flags:   globalFlags(),
		Commands: []*cli.Command{
			createDatabaseCommand(),
			{
				Name:   "create-user",
				Usage:  "create a new user for accessing arangodb",
				Action: CreateUser,
				Before: ValidateUserArgs,
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
						Name:    "user",
						Aliases: []string{"u"},
						Usage:   "arangodb user",
					},
					&cli.StringFlag{
						Name:    "password",
						Aliases: []string{"pw"},
						Usage:   "arangodb password for new user",
					},
				},
			},
			{
				Name:   "update-user",
				Usage:  "update an existing user's password for accessing arangodb",
				Action: UpdateUser,
				Before: ValidateUserArgs,
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
						Name:    "user",
						Aliases: []string{"u"},
						Usage:   "arangodb user",
					},
					&cli.StringFlag{
						Name:    "password",
						Aliases: []string{"pw"},
						Usage:   "new arangodb password for the user",
					},
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getLogger(cmd *cli.Command) *logrus.Entry {
	log := logrus.New()
	log.Out = os.Stderr
	switch cmd.String("log-format") {
	case "text":
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	case "json":
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	}
	l := cmd.String("log-level")
	switch l {
	case "debug":
		log.Level = logrus.DebugLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "error":
		log.Level = logrus.ErrorLevel
	case "fatal":
		log.Level = logrus.FatalLevel
	case "panic":
		log.Level = logrus.PanicLevel
	}
	return logrus.NewEntry(log)
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

func createDatabaseCommand() *cli.Command {
	return &cli.Command{
		Name:   "create-database",
		Usage:  "create a new arangodb database",
		Action: CreateDatabase,
		Before: ValidateDatabaseArgs,
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
			&cli.StringSliceFlag{
				Name:    "database",
				Aliases: []string{"db"},
				Usage:   "name of arangodb database",
			},
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "arangodb user",
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"pw"},
				Usage:   "arangodb password for new user",
			},
			&cli.StringFlag{
				Name:    "grant",
				Aliases: []string{"g"},
				Usage:   "level of access for arangodb user",
				Value:   "rw",
			},
		},
	}
}

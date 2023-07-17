package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "arangoadmin"
	app.Usage = "cli for creating databases and users in arangodb"
	app.Version = "1.0.0"
	app.Flags = globalFlags()
	app.Commands = []cli.Command{
		{
			Name:   "create-database",
			Usage:  "create a new arangodb database",
			Action: CreateDatabase,
			Before: ValidateDatabaseArgs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "admin-user,au",
					Usage: "arangodb admin user",
					Value: "root",
				},
				cli.StringFlag{
					Name:  "admin-password,ap",
					Usage: "arangodb admin password",
					Value: "",
				},
				cli.StringSliceFlag{
					Name:  "database,db",
					Usage: "name of arangodb database",
					Value: &cli.StringSlice{},
				},
				cli.StringFlag{
					Name:  "user,u",
					Usage: "arangodb user",
				},
				cli.StringFlag{
					Name:  "password,pw",
					Usage: "arangodb password for new user",
				},
				cli.StringFlag{
					Name:  "grant,g",
					Usage: "level of access for arangodb user",
					Value: "rw",
				},
			},
		},
		{
			Name:   "create-user",
			Usage:  "create a new user for accessing arangodb",
			Action: CreateUser,
			Before: ValidateUserArgs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "admin-user,au",
					Usage: "arangodb admin user",
					Value: "root",
				},
				cli.StringFlag{
					Name:  "admin-password,ap",
					Usage: "arangodb admin password",
					Value: "",
				},
				cli.StringFlag{
					Name:  "user,u",
					Usage: "arangodb user",
				},
				cli.StringFlag{
					Name:  "password,pw",
					Usage: "arangodb password for new user",
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getLogger(c *cli.Context) *logrus.Entry {
	log := logrus.New()
	log.Out = os.Stderr
	switch c.GlobalString("log-format") {
	case "text":
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	case "json":
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	}
	l := c.GlobalString("log-level")
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
		cli.StringFlag{
			Name:   "host",
			Usage:  "arangodb host address",
			EnvVar: "ARANGODB_SERVICE_HOST",
			Value:  "arangodb",
		},
		cli.StringFlag{
			Name:   "port",
			Usage:  "arangodb port",
			EnvVar: "ARANGODB_SERVICE_PORT",
			Value:  "8529",
		},
		cli.StringFlag{
			Name:  "log-level",
			Usage: "log level for the application",
			Value: "info",
		},
		cli.StringFlag{
			Name:  "log-format",
			Usage: "format of the logging out, either of json or text",
			Value: "json",
		},
		cli.BoolTFlag{
			Name:  "is-secure",
			Usage: "connect through a secure endpoint",
		},
	}
}

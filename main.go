package main

import (
	"os"

	"github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "arangoadmin"
	app.Usage = "cli for creating databases and users in arangodb"
	app.Flags = []cli.Flag{
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
		cli.BoolFlag{
			Name:  "is-secure",
			Usage: "connect through a secure endpoint",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "create-database",
			Usage:  "create a new arangodb database",
			Action: CreateDatabase,
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
					Value: "test",
				},
				cli.StringFlag{
					Name:  "password,pw",
					Usage: "arangodb password for new user",
					Value: "",
				},
				cli.StringFlag{
					Name:  "database,db",
					Usage: "name of arangodb database",
				},
				cli.StringFlag{
					Name:  "grant",
					Usage: "level of access for arangodb user",
				},
			},
		},
		{
			Name:   "create-user",
			Usage:  "create a new user for accessing arangodb",
			Action: CreateUser,
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
					Value: "test",
				},
				cli.StringFlag{
					Name:  "password,pw",
					Usage: "arangodb password for new user",
					Value: "",
				},
				cli.StringFlag{
					Name:  "grant",
					Usage: "level of access for arangodb user",
				},
			},
		},
	}
	app.Run(os.Args)
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

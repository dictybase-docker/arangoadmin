package main

import (
	"os"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "arangomanager"
	app.Usage = "cli for managing databases, users and collection in arangodb"
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
	}
	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "run the action defined in the yaml file",
			Action: Run,
			Before: validateRun,
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
				cli.BoolFlag{
					Name:  "is-secure",
					Usage: "connect through a secure endpoint",
					Value: true,
				},
			},
		},
	}
	app.Run(os.Args)
}

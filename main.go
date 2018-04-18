package main

import (
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
			Name:   "create-db-user",
			Usage:  "create database, users and database level access",
			Action: createDatabase,
			Before: validateCreateDatabase,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "admin-user,au",
					Usage: "arangodb admin user",
				},
				cli.StringFlag{
					Name:  "admin-password,ap",
					Usage: "arangodb admin password",
					Value: "",
				},
				cli.StringFlag{
					Name:  "database,db",
					Usage: "database to create, skip if it already exists",
				},
				cli.StringFlag{
					Name:  "user,u",
					Usage: "user to create, skip if the user already exists",
				},
				cli.StringFlag{
					Name:  "password,pass",
					Usage: "password for the user",
				},
				cli.StringFlag{
					Name:  "grant,g",
					Usage: "access level of user, could be one of ro,rw or none",
					Value: "rw",
				},
			},
		},
		{
			Name:   "create-collection",
			Usage:  "create collections in a database",
			Action: createCollection,
			Before: validateCreateCollection,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "database,db",
					Usage: "database where the collections will be created",
				},
				cli.StringFlag{
					Name:  "collection,c",
					Usage: "name of collection to create, skip if it already exist",
				},
				cli.BoolFlag{
					Name:  "is-edge",
					Usage: "flag to create an edge collection",
				},
				cli.StringFlag{
					Name:  "user,u",
					Usage: "user to create collection, should have required grant to create collection",
				},
				cli.StringFlag{
					Name:  "password,pass",
					Usage: "password for the user",
				},
			},
		},
	}
}

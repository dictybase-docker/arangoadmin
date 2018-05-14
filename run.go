package main

import (
	"context"
	"fmt"
	"os"

	"github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

func createDatabase(c *cli.Context) error {
	logger := getLogger(c)
	client, err := getClient(
		c.GlobalString("host"),
		c.GlobalString("port"),
		c.String("admin-user"),
		c.String("admin-password"),
		true,
	)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	v, err := client.Version(context.Background())
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("unable to get version %s", err),
			2,
		)
	}
	logger.Infof("Got version %s", v.Version)
	var db driver.Database
	ok, err := client.DatabaseExists(context.Background(), c.String("database"))
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("error in checking for database %s %s", c.String("database"), err),
			2,
		)
	}
	if ok {
		db, err = client.Database(context.Background(), c.String("database"))
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in retrieving database %s %s", c.String("database"), err),
				2,
			)
		}
		logger.Infof("database %s exists", c.String("database"))
	} else {
		db, err = client.CreateDatabase(context.Background(), c.String("database"), nil)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in creating database %s %s", c.String("database"), err),
				2,
			)
		}
		logger.Infof("successfully created database %s", c.String("database"))
	}
	var u driver.User
	ok, err = client.UserExists(context.Background(), c.String("user"))
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("error in checking for user %s %s", c.String("user"), err),
			2,
		)
	}
	if ok {
		u, err = client.User(context.Background(), c.String("user"))
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("cannot retrieve the user %s %s", c.String("user"), err),
				2,
			)
		}
		logger.Infof("user %s exists nothing to create", c.String("user"))
	} else {
		u, err = client.CreateUser(
			context.Background(),
			c.String("user"),
			&driver.UserOptions{
				Password: c.String("password"),
			})
		logger.Infof("successfully created user %s", c.String("user"))
	}
	err = u.SetDatabaseAccess(context.Background(), db, getGrant(c.String("grant")))
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf(
				"error in granting permission %s for user %s %s",
				c.String("grant"),
				c.String("user"),
				err,
			), 2)
	}
	logger.Infof("successfully granted permission to user %s", c.String("user"))
	return nil
}

func createCollection(c *cli.Context) error {
	logger := getLogger(c)
	client, err := getClient(
		c.GlobalString("host"),
		c.GlobalString("port"),
		c.String("user"),
		c.String("password"),
		false,
	)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	var db driver.Database
	ok, err := client.DatabaseExists(context.Background(), c.String("database"))
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("error in checking for database %s %s", c.String("database"), err),
			2,
		)
	}
	if !ok {
		logger.Errorf("database %s does not exists", c.String("database"))
		return cli.NewExitError(
			fmt.Sprintf("database %s does not exist, cannot create collection", c.String("database")),
			2,
		)
	}
	db, err = client.Database(context.Background(), c.String("database"))
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("error in retrieving database %s %s", c.String("database"), err),
			2,
		)
	}
	coll := c.String("collection")
	ok, err = db.CollectionExists(context.Background(), coll)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("error in collection lookup %s", err), 2)
	}
	if ok {
		logger.Infof("collection %s exist, nothing to create", coll)
		return nil
	}
	opt := &driver.CreateCollectionOptions{}
	if c.Bool("is-edge") {
		opt.Type = driver.CollectionTypeEdge
	}
	_, err = db.CreateCollection(context.Background(), coll, opt)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("error in creating collection %s", err), 2)
	}
	logger.Infof("successfully created collection %s", coll)
	return nil
}

func getGrant(g string) driver.Grant {
	var grnt driver.Grant
	switch g {
	case "rw":
		grnt = driver.GrantReadWrite
	case "ro":
		grnt = driver.GrantReadOnly
	default:
		grnt = driver.GrantNone
	}
	return grnt
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

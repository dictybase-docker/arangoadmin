package main

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	cli "gopkg.in/urfave/cli.v1"
)

// CreateDatabase creates a new ArangoDB database
func CreateDatabase(c *cli.Context) error {
	logger := getLogger(c)
	db := c.StringSlice("database")
	client, err := getClient(&ClientParams{
		Host:     c.GlobalString("host"),
		Port:     c.GlobalString("port"),
		User:     c.String("admin-user"),
		Pass:     c.String("admin-password"),
		IsSecure: c.GlobalBool("is-secure"),
	},
	)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("unable to get client %s", err), 2)
	}
	for _, n := range db {
		ok, err := client.DatabaseExists(context.Background(), n)
		if err != nil {
			return fmt.Errorf("error in checking existence of database %s %s", n, err)
		}
		if ok {
			logger.Infof("database %s exists, nothing to create", n)
		} else {
			_, err = client.CreateDatabase(context.Background(), n, nil)
			if err != nil {
				return fmt.Errorf("error in creating database %s %s", n, err)
			}
			logger.Infof("created database %s", n)
		}
	}
	if len(c.String("user")) > 0 {
		user := c.String("user")
		pass := c.String("pass")
		grant := c.String("grant")
		db := c.String("database")
		ok, err := client.UserExists(context.Background(), user)
		if err != nil {
			return fmt.Errorf("error in checking for user %s", err)
		}
		if !ok {
			dbuser, err := client.CreateUser(context.Background(), user, &driver.UserOptions{Password: pass})
			if err != nil {
				return fmt.Errorf("error in creating user %s %s", user, err)
			}
			logger.Infof("successfully created user %s", user)
			dbh, err := client.Database(context.Background(), db)
			if err != nil {
				return fmt.Errorf("cannot get a database instance for %s %s", db, err)
			}
			err = dbuser.SetDatabaseAccess(context.Background(), dbh, getGrant(grant))
			if err != nil {
				return fmt.Errorf(
					"error in granting permission %s for user %s %s",
					grant,
					user,
					err,
				)
			}
			logger.Infof("successfully granted permission %s existing user %s", grant, user)
		} else {
			logger.Infof("the user %s already exists", user)
		}
	}
	return nil
}

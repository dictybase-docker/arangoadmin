package main

import (
	"context"
	"fmt"

	cli "gopkg.in/urfave/cli.v1"
)

// CreateDatabase creates a new ArangoDB database
func CreateDatabase(c *cli.Context) error {
	logger := getLogger(c)
	db := c.StringSlice("database")
	client, err := getClient(
		c.GlobalString("host"),
		c.GlobalString("port"),
		c.String("admin-user"),
		c.String("admin-password"),
		c.GlobalBool("is-secure"),
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
	return nil
}

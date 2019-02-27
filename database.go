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
	db := c.String("database")
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
	ok, err := client.DatabaseExists(context.Background(), db)
	if err != nil {
		return fmt.Errorf("error in checking existence of database %s %s", db, err)
	}
	if ok {
		logger.Infof("database %s exists, nothing to create", db)
	} else {
		_, err = client.CreateDatabase(context.Background(), db, nil)
		if err != nil {
			return fmt.Errorf("error in creating database %s %s", db, err)
		}
		logger.Infof("created database %s", db)
	}
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

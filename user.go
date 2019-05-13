package main

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	cli "gopkg.in/urfave/cli.v1"
)

// CreateUser adds a new user with pre-specified privileges to ArangoDB
func CreateUser(c *cli.Context) error {
	logger := getLogger(c)
	user := c.String("user")
	pass := c.String("password")
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
	ok, err := client.UserExists(context.Background(), user)
	if err != nil {
		return fmt.Errorf("error in checking for user %s", err)
	}
	if ok {
		logger.Infof("user %s already exists", user)
		return nil
	}
	u, err := client.CreateUser(context.Background(), user, &driver.UserOptions{Password: pass})
	if err != nil {
		return fmt.Errorf("error in creating user %s %s", user, err)
	}
	logger.Infof("successfully created user %s", u.Name())
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

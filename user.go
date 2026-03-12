package main

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

// CreateUser adds a new user with pre-specified privileges to ArangoDB
func CreateUser(ctx context.Context, cmd *cli.Command) error {
	logger := getLogger(cmd)
	user := cmd.String("user")
	pass := cmd.String("password")
	client, err := getClient(&ClientParams{
		Host:     cmd.String("host"),
		Port:     cmd.String("port"),
		User:     cmd.String("admin-user"),
		Pass:     cmd.String("admin-password"),
		IsSecure: cmd.Bool("is-secure"),
	},
	)
	if err != nil {
		return cli.Exit(fmt.Sprintf("unable to get client %s", err), 2)
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

// UpdateUser updates the password of an existing user in ArangoDB
func UpdateUser(ctx context.Context, cmd *cli.Command) error {
	logger := getLogger(cmd)
	user := cmd.String("user")
	pass := cmd.String("password")
	client, err := getClient(&ClientParams{
		Host:     cmd.String("host"),
		Port:     cmd.String("port"),
		User:     cmd.String("admin-user"),
		Pass:     cmd.String("admin-password"),
		IsSecure: cmd.Bool("is-secure"),
	},
	)
	if err != nil {
		return cli.Exit(fmt.Sprintf("unable to get client %s", err), 2)
	}

	ok, err := client.UserExists(context.Background(), user)
	if err != nil {
		return fmt.Errorf("error in checking for user %s: %s", user, err)
	}
	if !ok {
		logger.Errorf("user %s does not exist", user)
		return cli.Exit(fmt.Sprintf("user %s does not exist", user), 2)
	}

	dbuser, err := client.User(context.Background(), user)
	if err != nil {
		return fmt.Errorf("error fetching user %s: %s", user, err)
	}

	err = dbuser.Update(context.Background(), driver.UserOptions{Password: pass})
	if err != nil {
		return fmt.Errorf("error updating user %s: %s", user, err)
	}

	logger.Infof("successfully updated password for user %s", user)
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

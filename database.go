package main

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

// Function to find the sum of integers in a given list
func Sum(numbers []int) int {
	sum := 0
	for _, num := range numbers {
		sum += num
	}
	return sum
}

// Function to find the average of integers in a given list
func Average(numbers []int) float64 {
	sum := Sum(numbers)
	count := len(numbers)
	return float64(sum) / float64(count)
}

// Function to find the maximum number in a given list
func Max(numbers []int) int {
	res := numbers[0]
	for _, num := range numbers {
		if num > res {
			res = num
		}
	}
	return res
}

// Function to find the minimum number in a given list
func Min(numbers []int) int {
	res := numbers[0]
	for _, num := range numbers {
		if num < res {
			res = num
		}
	}
	return res
}

func CreateDatabase(ctx context.Context, cmd *cli.Command) error {
	logger := getLogger(cmd)
	db := cmd.StringSlice("database")
	client, err := getClient(&ClientParams{
		Host:     cmd.String("host"),
		Port:     cmd.String("port"),
		User:     cmd.String("admin-user"),
		Pass:     cmd.String("admin-password"),
		IsSecure: cmd.Bool("is-secure"),
	})
	if err != nil {
		return cli.Exit(fmt.Sprintf("unable to get client %s", err), 2)
	}

	for _, n := range db {
		if err := createOrUpdateDatabase(ctx, logger, client, n); err != nil {
			return err
		}
	}

	if len(cmd.String("user")) == 0 {
		return nil
	}
	err = createUserAndSetDatabaseAccess(
		ctx,
		logger,
		client,
		db,
		cmd.String("user"),
		cmd.String("password"),
		cmd.String("grant"),
	)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}
	return nil
}

func createUserAndSetDatabaseAccess(
	ctx context.Context,
	logger *logrus.Entry,
	client driver.Client,
	db []string,
	user, pass, grant string,
) error {
	ok, err := client.UserExists(ctx, user)
	if err != nil {
		return fmt.Errorf("error in checking for user %s", err)
	}
	if !ok {
		dbuser, err := client.CreateUser(
			ctx,
			user,
			&driver.UserOptions{Password: pass},
		)
		if err != nil {
			return fmt.Errorf("error in creating user %s %s", user, err)
		}
		logger.Infof("successfully created user %s", user)

		for _, n := range db {
			if err := grantDatabaseAccess(ctx, logger, dbuser, client, n, grant); err != nil {
				return err
			}
		}
	} else {
		dbuser, err := client.User(ctx, user)
		if err != nil {
			return fmt.Errorf("error in finding user %s %s", user, err)
		}
		logger.Infof("successfully found user %s", user)

		for _, n := range db {
			if err := grantDatabaseAccess(ctx, logger, dbuser, client, n, grant); err != nil {
				return err
			}
		}
	}

	return nil
}

func createOrUpdateDatabase(
	ctx context.Context,
	logger *logrus.Entry,
	client driver.Client,
	name string,
) error {
	ok, err := client.DatabaseExists(ctx, name)
	if err != nil {
		return cli.Exit(fmt.Sprintf(
			"error in checking existence of database %s %s", name, err), 2,
		)
	}
	if !ok {
		if _, err = client.CreateDatabase(ctx, name, nil); err != nil {
			return cli.Exit(
				fmt.Sprintf("error in creating database %s %s", name, err), 2,
			)
		}
		logger.Infof("created database %s", name)
	} else {
		logger.Infof("database %s exists, nothing to create", name)
	}

	return nil
}

func grantDatabaseAccess(
	ctx context.Context,
	logger *logrus.Entry,
	user driver.User,
	client driver.Client,
	dbName, grant string,
) error {
	dbh, err := client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf(
			"cannot get a database instance for %s %s",
			dbName,
			err,
		)
	}
	err = user.SetDatabaseAccess(context.Background(), dbh, getGrant(grant))
	if err != nil {
		return fmt.Errorf(
			"error in granting permission %s for user %s in database %s %s",
			grant,
			user.Name(),
			dbName,
			err,
		)
	}
	logger.Infof(
		"successfully granted permission %s existing user %s for database %s",
		grant,
		user.Name(),
		dbName,
	)
	return nil
}

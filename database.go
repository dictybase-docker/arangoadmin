package main

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
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
	max := numbers[0]
	for _, num := range numbers {
		if num > max {
			max = num
		}
	}
	return max
}

// Function to find the minimum number in a given list
func Min(numbers []int) int {
	min := numbers[0]
	for _, num := range numbers {
		if num < min {
			min = num
		}
	}
	return min
}

func CreateDatabase(c *cli.Context) error {
	logger := getLogger(c)
	db := c.StringSlice("database")
	client, err := getClient(&ClientParams{
		Host:     c.GlobalString("host"),
		Port:     c.GlobalString("port"),
		User:     c.String("admin-user"),
		Pass:     c.String("admin-password"),
		IsSecure: c.GlobalBool("is-secure"),
	})
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("unable to get client %s", err), 2)
	}

	for _, n := range db {
		if err := createOrUpdateDatabase(logger, client, n); err != nil {
			return err
		}
	}

	if len(c.String("user")) == 0 {
		return nil
	}
	err = createUserAndSetDatabaseAccess(
		logger,
		client,
		db,
		c.String("user"),
		c.String("password"),
		c.String("grant"),
	)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	return nil
}

func createUserAndSetDatabaseAccess(
	logger *logrus.Entry,
	client driver.Client,
	db []string,
	user, pass, grant string,
) error {
	ok, err := client.UserExists(context.Background(), user)
	if err != nil {
		return fmt.Errorf("error in checking for user %s", err)
	}
	if !ok {
		dbuser, err := client.CreateUser(
			context.Background(),
			user,
			&driver.UserOptions{Password: pass},
		)
		if err != nil {
			return fmt.Errorf("error in creating user %s %s", user, err)
		}
		logger.Infof("successfully created user %s", user)

		for _, n := range db {
			if err := grantDatabaseAccess(logger, dbuser, client, n, grant); err != nil {
				return err
			}
		}
	} else {
		dbuser, err := client.User(context.Background(), user)
		if err != nil {
			return fmt.Errorf("error in finding user %s %s", user, err)
		}
		logger.Infof("successfully found user %s", user)

		for _, n := range db {
			if err := grantDatabaseAccess(logger, dbuser, client, n, grant); err != nil {
				return err
			}
		}
	}

	return nil
}

func createOrUpdateDatabase(
	logger *logrus.Entry,
	client driver.Client,
	name string,
) error {
	ok, err := client.DatabaseExists(context.Background(), name)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf(
			"error in checking existence of database %s %s", name, err), 2,
		)
	}
	if !ok {
		if _, err = client.CreateDatabase(context.Background(), name, nil); err != nil {
			return cli.NewExitError(
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
	logger *logrus.Entry,
	user driver.User,
	client driver.Client,
	dbName, grant string,
) error {
	dbh, err := client.Database(context.Background(), dbName)
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

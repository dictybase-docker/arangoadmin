package main

import (
	"context"
	"fmt"
	"log/slog"

	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

func CreateDatabase(ctx context.Context, cmd *cli.Command) error {
	logger := newLogger(cmd)
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
	logger *slog.Logger,
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
		logger.Info("successfully created user", "user", user)

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
		logger.Info("successfully found user", "user", user)

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
	logger *slog.Logger,
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
		logger.Info("created database", "database", name)
	} else {
		logger.Info("database exists", "database", name)
	}

	return nil
}

func grantDatabaseAccess(
	ctx context.Context,
	logger *slog.Logger,
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
	logger.Info(
		"successfully granted permission",
		"grant", grant,
		"user", user.Name(),
		"database", dbName,
	)
	return nil
}

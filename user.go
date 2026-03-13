package main

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	fperrors "github.com/IBM/fp-go/v2/errors"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

// CreateUser adds a new user with pre-specified privileges to ArangoDB
func CreateUser(_ context.Context, cmd *cli.Command) error {
	logger := newLogger(cmd)
	connParams := connParamsFromCmd(cmd)
	username := cmd.String("user")
	password := cmd.String("password")

	result := F.Pipe1(
		createArangoClient(connParams),
		IOE.Chain(func(client driver.Client) IOE.IOEither[error, struct{}] {
			return createUserIfNotExists(UserParams{
				WithClient: WithClient{Client: client, Logger: logger},
				Username:   username,
				Password:   password,
			})
		}),
	)

	either := toEither(result)
	return E.Fold(
		F.Identity[error],
		func(_ struct{}) error { return nil },
	)(either)
}

// createUserIfNotExists creates a user if they don't exist, otherwise logs that they exist
// nolint:unused // Used by CreateUser action
func createUserIfNotExists(p UserParams) IOE.IOEither[error, struct{}] {
	return F.Pipe2(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool, error, error](fperrors.OnError(fmt.Sprintf("error checking for user %s", p.Username))),
		IOE.Chain(func(exists bool) IOE.IOEither[error, struct{}] {
			if exists {
				return IOE.FromIO[error](logUserExists(p.Logger, p.Username))
			}
			return F.Pipe3(
				IOE.TryCatchError(func() (driver.User, error) {
					return p.Client.CreateUser(context.Background(), p.Username, &driver.UserOptions{Password: p.Password})
				}),
				IOE.MapLeft[driver.User, error, error](fperrors.OnError(fmt.Sprintf("error creating user %s", p.Username))),
				IOE.ChainFirstIOK[error](func(_ driver.User) IO.IO[struct{}] {
					return logUserCreated(p.Logger, p.Username)
				}),
				IOE.Map[error](func(_ driver.User) struct{} { return struct{}{} }),
			)
		}),
	)
}

// UpdateUser updates the password of an existing user in ArangoDB
func UpdateUser(ctx context.Context, cmd *cli.Command) error {
	logger := newLogger(cmd)
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

	ok, err := client.UserExists(ctx, user)
	if err != nil {
		return fmt.Errorf("error in checking for user %s: %s", user, err)
	}
	if !ok {
		logger.Error("user does not exist", "user", user)
		return cli.Exit(fmt.Sprintf("user %s does not exist", user), 2)
	}

	dbuser, err := client.User(ctx, user)
	if err != nil {
		return fmt.Errorf("error fetching user %s: %s", user, err)
	}

	err = dbuser.Update(ctx, driver.UserOptions{Password: pass})

	if err != nil {
		return fmt.Errorf("error updating user %s: %s", user, err)
	}

	logger.Info("successfully updated password for user", "user", user)
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

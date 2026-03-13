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
		IOE.Chain(func(client driver.Client) IOE.IOEither[error, F.Void] {
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
		func(_ F.Void) error { return nil },
	)(either)
}

// createUserIfNotExists creates a user if they don't exist, otherwise logs that they exist
func createUserIfNotExists(p UserParams) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool, error, error](fperrors.OnError(fmt.Sprintf("error checking for user %s", p.Username))),
		IOE.Chain(F.Ternary(
			F.Identity[bool],
			func(_ bool) IOE.IOEither[error, F.Void] {
				return IOE.FromIO[error](logUserExists(p.Logger, p.Username))
			},
			func(_ bool) IOE.IOEither[error, F.Void] {
				return F.Pipe3(
					IOE.TryCatchError(func() (driver.User, error) {
						return p.Client.CreateUser(context.Background(), p.Username, &driver.UserOptions{Password: p.Password})
					}),
					IOE.MapLeft[driver.User, error, error](fperrors.OnError(fmt.Sprintf("error creating user %s", p.Username))),
					IOE.ChainFirstIOK[error](func(_ driver.User) IO.IO[F.Void] {
						return logUserCreated(p.Logger, p.Username)
					}),
					IOE.MapTo[error, driver.User](F.VOID),
				)
			},
		)),
	)
}

// UpdateUser updates the password of an existing user in ArangoDB
func UpdateUser(_ context.Context, cmd *cli.Command) error {
	logger := newLogger(cmd)
	connParams := connParamsFromCmd(cmd)
	username := cmd.String("user")
	password := cmd.String("password")

	result := F.Pipe1(
		createArangoClient(connParams),
		IOE.Chain(func(client driver.Client) IOE.IOEither[error, F.Void] {
			return updateUserPipeline(UserParams{
				WithClient: WithClient{Client: client, Logger: logger},
				Username:   username,
				Password:   password,
			})
		}),
	)

	either := toEither(result)
	return E.Fold(
		F.Identity[error],
		func(_ F.Void) error { return nil },
	)(either)
}

// updateUserPipeline updates a user's password if they exist
func updateUserPipeline(p UserParams) IOE.IOEither[error, F.Void] {
	return F.Pipe4(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool, error, error](fperrors.OnError(fmt.Sprintf("error checking for user %s", p.Username))),
		IOE.ChainEitherK(E.FromPredicate(
			F.Identity[bool],
			func(_ bool) error { return fmt.Errorf("user %s does not exist", p.Username) },
		)),
		IOE.Chain(func(_ bool) IOE.IOEither[error, driver.User] {
			return F.Pipe1(
				IOE.TryCatchError(func() (driver.User, error) {
					return p.Client.User(context.Background(), p.Username)
				}),
				IOE.MapLeft[driver.User, error, error](fperrors.OnError(fmt.Sprintf("error fetching user %s", p.Username))),
			)
		}),
		IOE.Chain(func(user driver.User) IOE.IOEither[error, F.Void] {
			return F.Pipe2(
				IOE.TryCatchError(func() (F.Void, error) {
					return F.VOID, user.Update(context.Background(), driver.UserOptions{Password: p.Password})
				}),
				IOE.MapLeft[F.Void, error, error](fperrors.OnError(fmt.Sprintf("error updating user %s", p.Username))),
				IOE.ChainFirstIOK[error](func(_ F.Void) IO.IO[F.Void] {
					return logUserUpdated(p.Logger, p.Username)
				}),
			)
		}),
	)
}

// getGrant converts a grant string to the driver.Grant type
func getGrant(g string) driver.Grant {
	return F.Pipe1(
		g,
		F.Ternary(
			func(s string) bool { return s == "rw" },
			func(_ string) driver.Grant { return driver.GrantReadWrite },
			F.Ternary(
				func(s string) bool { return s == "ro" },
				func(_ string) driver.Grant { return driver.GrantReadOnly },
				func(_ string) driver.Grant { return driver.GrantNone },
			),
		),
	)
}

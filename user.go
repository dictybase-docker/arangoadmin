package main

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

type UserForUpdate struct {
	Params UserParams
	User   driver.User
}

// CreateUser adds a new user with pre-specified privileges to ArangoDB
func CreateUser(_ context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) UserParams {
			return UserParams{
				WithClient: WithClient{
					Client: client,
					Logger: newLogger(cmd),
				},
				Username: cmd.String("user"),
				Password: cmd.String("password"),
			}
		}),
		IOE.Chain(createUserIfNotExists),
		foldIOE[F.Void],
	)
}

// createUserIfNotExists creates a user if they don't exist, otherwise logs that they exist
func createUserIfNotExists(params UserParams) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (bool, error) {
			return params.Client.UserExists(
				context.Background(),
				params.Username,
			)
		}),
		IOE.MapLeft[bool](
			fperrors.OnError(
				fmt.Sprintf("error checking for user %s", params.Username),
			),
		),
		IOE.Chain(F.Ternary(
			F.Identity[bool],
			func(_ bool) IOE.IOEither[error, F.Void] {
				return IOE.FromIO[error](
					logUserExists(params.Logger, params.Username),
				)
			},
			func(_ bool) IOE.IOEither[error, F.Void] {
				return F.Pipe3(
					IOE.TryCatchError(func() (driver.User, error) {
						return params.Client.CreateUser(
							context.Background(),
							params.Username,
							&driver.UserOptions{Password: params.Password},
						)
					}),
					IOE.MapLeft[driver.User](
						fperrors.OnError(
							fmt.Sprintf(
								"error creating user %s",
								params.Username,
							),
						),
					),
					IOE.ChainFirstIOK[error](func(_ driver.User) IO.IO[F.Void] {
						return logUserCreated(params.Logger, params.Username)
					}),
					IOE.MapTo[error, driver.User](F.VOID),
				)
			},
		)),
	)
}

// UpdateUser updates the password of an existing user in ArangoDB
func UpdateUser(_ context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) UserParams {
			return UserParams{
				WithClient: WithClient{
					Client: client,
					Logger: newLogger(cmd),
				},
				Username: cmd.String("user"),
				Password: cmd.String("password"),
			}
		}),
		IOE.Chain(updateUserPipeline),
		foldIOE[F.Void],
	)
}

// updateUserPipeline updates a user's password if they exist
func updateUserPipeline(params UserParams) IOE.IOEither[error, F.Void] {
	return F.Pipe3(
		params,
		checkUserExists,
		IOE.Chain(getExistingUser),
		IOE.Chain(updateExistingUser),
	)
}

func checkUserExists(params UserParams) IOE.IOEither[error, UserParams] {
	return F.Pipe3(
		IOE.TryCatchError(func() (bool, error) {
			return params.Client.UserExists(
				context.Background(),
				params.Username,
			)
		}),
		IOE.MapLeft[bool](
			fperrors.OnError(
				fmt.Sprintf("error checking for user %s", params.Username),
			),
		),
		IOE.ChainEitherK(E.FromPredicate(
			F.Identity[bool],
			func(_ bool) error { return fmt.Errorf("user %s does not exist", params.Username) },
		)),
		IOE.Map[error](F.Constant1[bool](params)),
	)
}

func getExistingUser(params UserParams) IOE.IOEither[error, UserForUpdate] {
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.User, error) {
			return params.Client.User(context.Background(), params.Username)
		}),
		IOE.MapLeft[driver.User](
			fperrors.OnError(
				fmt.Sprintf("error fetching user %s", params.Username),
			),
		),
		IOE.Map[error](withExistingUser(params)),
	)
}

func withExistingUser(params UserParams) func(driver.User) UserForUpdate {
	return func(user driver.User) UserForUpdate {
		return UserForUpdate{Params: params, User: user}
	}
}

func updateExistingUser(u UserForUpdate) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (F.Void, error) {
			return F.VOID, u.User.Update(
				context.Background(),
				driver.UserOptions{Password: u.Params.Password},
			)
		}),
		IOE.MapLeft[F.Void](
			fperrors.OnError(
				fmt.Sprintf("error updating user %s", u.Params.Username),
			),
		),
		IOE.ChainFirstIOK[error](logUserUpdatedStep(u)),
	)
}

func logUserUpdatedStep(u UserForUpdate) func(F.Void) IO.IO[F.Void] {
	return func(_ F.Void) IO.IO[F.Void] {
		return logUserUpdated(u.Params.Logger, u.Params.Username)
	}
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

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
	userParams := UserParams{
		WithClient: WithClient{Logger: newLogger(cmd)},
		Username:   cmd.String("user"),
		Password:   cmd.String("password"),
	}
	return F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](withUserClient(userParams)),
		IOE.Chain(createUserIfNotExists),
		toEither,
		E.Fold(
			F.Identity[error],
			func(_ F.Void) error { return nil },
		),
	)
}

// createUserIfNotExists creates a user if they don't exist, otherwise logs that they exist
func createUserIfNotExists(p UserParams) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool](
			fperrors.OnError(
				fmt.Sprintf("error checking for user %s", p.Username),
			),
		),
		IOE.Chain(F.Ternary(
			F.Identity[bool],
			func(_ bool) IOE.IOEither[error, F.Void] {
				return IOE.FromIO[error](logUserExists(p.Logger, p.Username))
			},
			func(_ bool) IOE.IOEither[error, F.Void] {
				return F.Pipe3(
					IOE.TryCatchError(func() (driver.User, error) {
						return p.Client.CreateUser(
							context.Background(),
							p.Username,
							&driver.UserOptions{Password: p.Password},
						)
					}),
					IOE.MapLeft[driver.User](
						fperrors.OnError(
							fmt.Sprintf("error creating user %s", p.Username),
						),
					),
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
	userParams := UserParams{
		WithClient: WithClient{Logger: newLogger(cmd)},
		Username:   cmd.String("user"),
		Password:   cmd.String("password"),
	}
	return F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](withUserClient(userParams)),
		IOE.Chain(updateUserPipeline),
		toEither,
		E.Fold(
			F.Identity[error],
			func(_ F.Void) error { return nil },
		),
	)
}

// updateUserPipeline updates a user's password if they exist
func updateUserPipeline(p UserParams) IOE.IOEither[error, F.Void] {
	return F.Pipe3(
		p,
		checkUserExists,
		IOE.Chain(getExistingUser),
		IOE.Chain(updateExistingUser),
	)
}

func withUserClient(p UserParams) func(driver.Client) UserParams {
	return func(client driver.Client) UserParams {
		p.Client = client
		return p
	}
}

func checkUserExists(p UserParams) IOE.IOEither[error, UserParams] {
	return F.Pipe3(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool](
			fperrors.OnError(
				fmt.Sprintf("error checking for user %s", p.Username),
			),
		),
		IOE.ChainEitherK(E.FromPredicate(
			F.Identity[bool],
			func(_ bool) error { return fmt.Errorf("user %s does not exist", p.Username) },
		)),
		IOE.Map[error](F.Constant1[bool](p)),
	)
}

func getExistingUser(p UserParams) IOE.IOEither[error, UserForUpdate] {
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.User, error) {
			return p.Client.User(context.Background(), p.Username)
		}),
		IOE.MapLeft[driver.User](
			fperrors.OnError(fmt.Sprintf("error fetching user %s", p.Username)),
		),
		IOE.Map[error](withExistingUser(p)),
	)
}

func withExistingUser(p UserParams) func(driver.User) UserForUpdate {
	return func(user driver.User) UserForUpdate {
		return UserForUpdate{Params: p, User: user}
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

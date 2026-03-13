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

// CreateDatabase creates one or more databases with optional user and grants
func CreateDatabase(_ context.Context, cmd *cli.Command) error {
	logger := newLogger(cmd)
	connParams := connParamsFromCmd(cmd)
	databases := cmd.StringSlice("database")
	username := cmd.String("user")
	password := cmd.String("password")
	grant := cmd.String("grant")

	result := F.Pipe1(
		createArangoClient(connParams),
		IOE.Chain(func(client driver.Client) IOE.IOEither[error, struct{}] {
			return createDatabasePipeline(DatabaseParams{
				WithClient: WithClient{Client: client, Logger: logger},
				Databases:  databases,
				Username:   username,
				Password:   password,
				Grant:      grant,
			})
		}),
	)

	either := toEither(result)
	return E.Fold(
		F.Identity[error],
		func(_ struct{}) error { return nil },
	)(either)
}

// createDatabasePipeline creates all databases, then optionally creates user and grants access
func createDatabasePipeline(p DatabaseParams) IOE.IOEither[error, struct{}] {
	return F.Pipe2(
		p.Databases,
		IOE.TraverseArraySeq(createSingleDatabase(p)),
		IOE.Chain(F.Ternary(
			func(_ []struct{}) bool { return len(p.Username) > 0 },
			func(_ []struct{}) IOE.IOEither[error, struct{}] { return createUserAndGrant(p) },
			func(_ []struct{}) IOE.IOEither[error, struct{}] { return IOE.Of[error](struct{}{}) },
		)),
	)
}

// createSingleDatabase creates a single database if it doesn't exist, logs the result
func createSingleDatabase(p DatabaseParams) func(string) IOE.IOEither[error, struct{}] {
	return func(dbname string) IOE.IOEither[error, struct{}] {
		return F.Pipe2(
			IOE.TryCatchError(func() (bool, error) {
				return p.Client.DatabaseExists(context.Background(), dbname)
			}),
			IOE.MapLeft[bool, error, error](fperrors.OnError(fmt.Sprintf("error checking for database %s", dbname))),
			IOE.Chain(F.Ternary(
				F.Identity[bool],
				func(_ bool) IOE.IOEither[error, struct{}] {
					return IOE.FromIO[error](logDatabaseExists(p.Logger, dbname))
				},
				func(_ bool) IOE.IOEither[error, struct{}] {
					return F.Pipe2(
						IOE.TryCatchError(func() (struct{}, error) {
							_, err := p.Client.CreateDatabase(context.Background(), dbname, nil)
							return struct{}{}, err
						}),
						IOE.MapLeft[struct{}, error, error](fperrors.OnError(fmt.Sprintf("error creating database %s", dbname))),
						IOE.ChainFirstIOK[error](func(_ struct{}) IO.IO[struct{}] {
							return logDatabaseCreated(p.Logger, dbname)
						}),
					)
				},
			)),
		)
	}
}

// createUserAndGrant creates a user if they don't exist, then grants access to all databases
func createUserAndGrant(p DatabaseParams) IOE.IOEither[error, struct{}] {
	return F.Pipe3(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool, error, error](fperrors.OnError(fmt.Sprintf("error checking for user %s", p.Username))),
		IOE.Chain(F.Ternary(
			F.Identity[bool],
			func(_ bool) IOE.IOEither[error, driver.User] {
				return F.Pipe1(
					IOE.TryCatchError(func() (driver.User, error) {
						return p.Client.User(context.Background(), p.Username)
					}),
					IOE.MapLeft[driver.User, error, error](fperrors.OnError(fmt.Sprintf("error fetching user %s", p.Username))),
				)
			},
			func(_ bool) IOE.IOEither[error, driver.User] {
				return F.Pipe2(
					IOE.TryCatchError(func() (driver.User, error) {
						return p.Client.CreateUser(context.Background(), p.Username, &driver.UserOptions{Password: p.Password})
					}),
					IOE.MapLeft[driver.User, error, error](fperrors.OnError(fmt.Sprintf("error creating user %s", p.Username))),
					IOE.ChainFirstIOK[error](func(_ driver.User) IO.IO[struct{}] {
						return logUserCreated(p.Logger, p.Username)
					}),
				)
			},
		)),
		IOE.Chain(func(user driver.User) IOE.IOEither[error, struct{}] {
			return F.Pipe2(
				p.Databases,
				IOE.TraverseArraySeq(grantSingleDatabase(UserWithGrant{Params: p, User: user})),
				IOE.Map[error](func(_ []struct{}) struct{} { return struct{}{} }),
			)
		}),
	)
}

// grantSingleDatabase grants access to a single database for a user
func grantSingleDatabase(u UserWithGrant) func(string) IOE.IOEither[error, struct{}] {
	return func(dbname string) IOE.IOEither[error, struct{}] {
		return F.Pipe2(
			IOE.TryCatchError(func() (driver.Database, error) {
				return u.Params.Client.Database(context.Background(), dbname)
			}),
			IOE.MapLeft[driver.Database, error, error](fperrors.OnError(fmt.Sprintf("error getting database %s", dbname))),
			IOE.Chain(func(db driver.Database) IOE.IOEither[error, struct{}] {
				return F.Pipe2(
					IOE.TryCatchError(func() (struct{}, error) {
						return struct{}{}, u.User.SetDatabaseAccess(context.Background(), db, getGrant(u.Params.Grant))
					}),
					IOE.MapLeft[struct{}, error, error](fperrors.OnError(fmt.Sprintf("error granting access to database %s", dbname))),
					IOE.ChainFirstIOK[error](func(_ struct{}) IO.IO[struct{}] {
						return logGrantAccess(u.Params.Logger, u.User.Name(), dbname, u.Params.Grant)
					}),
				)
			}),
		)
	}
}

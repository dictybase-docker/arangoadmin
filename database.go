package main

import (
	"context"
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

// CreateDatabase creates one or more databases with optional user and grants
func CreateDatabase(_ context.Context, cmd *cli.Command) error {
	databases := cmd.StringSlice("database")
	dbparams := DatabaseParams{
		WithClient: WithClient{Logger: newLogger(cmd)},
		Username:   cmd.String("user"),
		Password:   cmd.String("password"),
		Grant:      cmd.String("grant"),
	}
	return F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) DatabaseParams {
			dbparams.Client = client
			return dbparams
		}),
		IOE.Chain(func(p DatabaseParams) IOE.IOEither[error, struct{}] {
			return F.Pipe2(
				F.Pipe1(
					databases,
					A.Map(func(dbname string) DatabaseParams {
						q := p
						q.Dbname = dbname
						return q
					}),
				),
				IOE.TraverseArraySeq(createSingleDatabase),
				IOE.Chain(F.Ternary(
					func(_ []struct{}) bool { return len(p.Username) > 0 },
					F.Constant1[[]struct{}](createUserAndGrant(p, databases)),
					F.Constant1[[]struct{}](IOE.Of[error](struct{}{})),
				)),
			)
		}),
		toEither,
		E.Fold(
			F.Identity[error],
			func(_ struct{}) error { return nil },
		),
	)
}

// createSingleDatabase creates a single database if it doesn't exist, logs the result
func createSingleDatabase(p DatabaseParams) IOE.IOEither[error, struct{}] {
	return F.Pipe2(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.DatabaseExists(context.Background(), p.Dbname)
		}),
		IOE.MapLeft[bool](fperrors.OnError(
			fmt.Sprintf("error checking for database %s", p.Dbname),
		)),
		IOE.Chain(F.Ternary(
			F.Identity[bool],
			F.Constant1[bool](
				IOE.FromIO[error](logDatabaseExists(p.Logger, p.Dbname)),
			),
			F.Constant1[bool](createDatabase(p)),
		)),
	)
}

// createDatabase creates a single database and logs the result
func createDatabase(p DatabaseParams) IOE.IOEither[error, struct{}] {
	return F.Pipe2(
		IOE.TryCatchError(func() (struct{}, error) {
			_, err := p.Client.CreateDatabase(
				context.Background(),
				p.Dbname,
				nil,
			)
			return struct{}{}, err
		}),
		IOE.MapLeft[struct{}](fperrors.OnError(
			fmt.Sprintf("error creating database %s", p.Dbname),
		)),
		IOE.ChainFirstIOK[error](func(_ struct{}) IO.IO[struct{}] {
			return logDatabaseCreated(p.Logger, p.Dbname)
		}),
	)
}

// createUserAndGrant creates a user if they don't exist, then grants access to all databases
func createUserAndGrant(p DatabaseParams, databases []string) IOE.IOEither[error, struct{}] {
	return F.Pipe3(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool](fperrors.OnError(
			fmt.Sprintf("error checking for user %s", p.Username),
		)),
		IOE.Chain(F.Ternary(
			F.Identity[bool],
			func(_ bool) IOE.IOEither[error, driver.User] {
				return F.Pipe1(
					IOE.TryCatchError(func() (driver.User, error) {
						return p.Client.User(context.Background(), p.Username)
					}),
					IOE.MapLeft[driver.User](fperrors.OnError(
						fmt.Sprintf("error fetching user %s", p.Username),
					)),
				)
			},
			func(_ bool) IOE.IOEither[error, driver.User] {
				return F.Pipe2(
					IOE.TryCatchError(func() (driver.User, error) {
						return p.Client.CreateUser(
							context.Background(),
							p.Username,
							&driver.UserOptions{Password: p.Password},
						)
					}),
					IOE.MapLeft[driver.User](fperrors.OnError(
						fmt.Sprintf("error creating user %s", p.Username),
					)),
					IOE.ChainFirstIOK[error](
						func(_ driver.User) IO.IO[struct{}] {
							return logUserCreated(p.Logger, p.Username)
						},
					),
				)
			},
		)),
		IOE.Chain(func(user driver.User) IOE.IOEither[error, struct{}] {
			return F.Pipe2(
				F.Pipe1(
					databases,
					A.Map(func(dbname string) UserWithGrant {
						q := p
						q.Dbname = dbname
						return UserWithGrant{Params: q, User: user}
					}),
				),
				IOE.TraverseArraySeq(grantSingleDatabase),
				IOE.Map[error](
					func(_ []struct{}) struct{} { return struct{}{} },
				),
			)
		}),
	)
}

// grantSingleDatabase grants access to a single database for a user
func grantSingleDatabase(u UserWithGrant) IOE.IOEither[error, struct{}] {
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.Database, error) {
			return u.Params.Client.Database(
				context.Background(),
				u.Params.Dbname,
			)
		}),
		IOE.MapLeft[driver.Database](fperrors.OnError(
			fmt.Sprintf("error getting database %s", u.Params.Dbname),
		)),
		IOE.Chain(func(db driver.Database) IOE.IOEither[error, struct{}] {
			return F.Pipe2(
				IOE.TryCatchError(func() (struct{}, error) {
					return struct{}{}, u.User.SetDatabaseAccess(
						context.Background(),
						db,
						getGrant(u.Params.Grant),
					)
				}),
				IOE.MapLeft[struct{}](fperrors.OnError(
					fmt.Sprintf(
						"error granting access to database %s",
						u.Params.Dbname,
					),
				)),
				IOE.ChainFirstIOK[error](func(_ struct{}) IO.IO[struct{}] {
					return logGrantAccess(
						u.Params.Logger,
						u.User.Name(),
						u.Params.Dbname,
						u.Params.Grant,
					)
				}),
			)
		}),
	)
}

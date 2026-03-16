package main

import (
	"context"
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

// CreateDatabase creates one or more databases with optional user and grants
func CreateDatabase(_ context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) DatabaseParams {
			databases := cmd.StringSlice("database")
			return DatabaseParams{
				WithClient: WithClient{
					Client: client,
					Logger: newLogger(cmd),
				},
				Dbname:    "",
				Username:  cmd.String("user"),
				Password:  cmd.String("password"),
				Grant:     cmd.String("grant"),
				Databases: databases,
			}
		}),
		IOE.Chain(createDatabasePipeline),
		foldIOE[F.Void],
	)
}

// createDatabasePipeline creates databases and optionally creates a user with grants.
func createDatabasePipeline(p DatabaseParams) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		F.Pipe1(
			p.Databases,
			A.Map(func(dbname string) DatabaseParams {
				q := p
				q.Dbname = dbname
				return q
			}),
		),
		IOE.TraverseArraySeq(createSingleDatabase),
		IOE.Chain(optionalCreateUserAndGrant(UserGrantParams{
			Params: p, Databases: p.Databases,
		})),
	)
}

// createSingleDatabase creates a single database if it doesn't exist, logs the result
func createSingleDatabase(p DatabaseParams) IOE.IOEither[error, F.Void] {
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
func createDatabase(p DatabaseParams) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (F.Void, error) {
			_, err := p.Client.CreateDatabase(
				context.Background(),
				p.Dbname,
				nil,
			)
			return F.VOID, err
		}),
		IOE.MapLeft[F.Void](fperrors.OnError(
			fmt.Sprintf("error creating database %s", p.Dbname),
		)),
		IOE.ChainFirstIOK[error](func(_ F.Void) IO.IO[F.Void] {
			return logDatabaseCreated(p.Logger, p.Dbname)
		}),
	)
}

// optionalCreateUserAndGrant returns a Kleisli arrow that creates a user and
// grants access if Username is non-empty, otherwise succeeds with F.VOID
func optionalCreateUserAndGrant(g UserGrantParams) func([]F.Void) IOE.IOEither[error, F.Void] {
	return F.Constant1[[]F.Void](F.Pipe2(
		g.Params.Username,
		O.FromPredicate(func(s string) bool { return len(s) > 0 }),
		O.Fold(
			func() IOE.IOEither[error, F.Void] {
				return IOE.Of[error](F.VOID)
			},
			func(_ string) IOE.IOEither[error, F.Void] {
				return createUserAndGrant(g)
			},
		),
	))
}

// createUserAndGrant creates a user if they don't exist, then grants access to all databases
func createUserAndGrant(g UserGrantParams) IOE.IOEither[error, F.Void] {
	p := g.Params
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
						func(_ driver.User) IO.IO[F.Void] {
							return logUserCreated(p.Logger, p.Username)
						},
					),
				)
			},
		)),
		IOE.Chain(func(user driver.User) IOE.IOEither[error, F.Void] {
			return F.Pipe2(
				F.Pipe1(
					g.Databases,
					A.Map(func(dbname string) UserWithGrant {
						q := p
						q.Dbname = dbname
						return UserWithGrant{Params: q, User: user}
					}),
				),
				IOE.TraverseArraySeq(grantSingleDatabase),
				IOE.MapTo[error, []F.Void](F.VOID),
			)
		}),
	)
}

// grantSingleDatabase grants access to a single database for a user
func grantSingleDatabase(u UserWithGrant) IOE.IOEither[error, F.Void] {
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
		IOE.Chain(func(db driver.Database) IOE.IOEither[error, F.Void] {
			return F.Pipe2(
				IOE.TryCatchError(func() (F.Void, error) {
					return F.VOID, u.User.SetDatabaseAccess(
						context.Background(),
						db,
						getGrant(u.Params.Grant),
					)
				}),
				IOE.MapLeft[F.Void](fperrors.OnError(
					fmt.Sprintf(
						"error granting access to database %s",
						u.Params.Dbname,
					),
				)),
				IOE.ChainFirstIOK[error](func(_ F.Void) IO.IO[F.Void] {
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

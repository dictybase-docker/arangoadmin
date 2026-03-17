package main

import (
	"context"
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

// CreateDatabase creates one or more databases with optional user and grants
func CreateDatabase(_ context.Context, cmd *cli.Command) error {
	output := F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) DatabaseParams {
			return DatabaseParams{
				WithClient: WithClient{
					Client: client,
					Logger: newLogger(cmd),
				},
				Username:  cmd.String("user"),
				Password:  cmd.String("password"),
				Grant:     cmd.String("grant"),
				Databases: cmd.StringSlice("database"),
			}
		}),
		IOE.Chain(createDatabasePipeline),
		toEither,
		E.Fold(
			func(err error) P.Pair[CreateDatabaseResult, error] {
				var zero CreateDatabaseResult
				return P.MakePair(zero, err)
			},
			func(result CreateDatabaseResult) P.Pair[CreateDatabaseResult, error] {
				return P.MakePair[CreateDatabaseResult, error](result, nil)
			},
		),
	)
	if err := P.Second(output); err != nil {
		return err
	}

	logCreateDatabaseOutcome(newLogger(cmd), P.First(output))
	return nil
}

func ensureDatabasePipeline(
	p EnsureDatabaseParams,
) IOE.IOEither[error, EnsureDatabaseResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.DatabaseExists(context.Background(), p.Database)
		}),
		IOE.MapLeft[bool](fperrors.OnError(
			fmt.Sprintf("error checking for database %s", p.Database),
		)),
		IOE.Chain(F.Ternary(
			F.Identity[bool],
			F.Constant1[bool](IOE.Of[error](P.MakePair(false, p.Database))),
			F.Constant1[bool](createDatabaseForEnsure(p)),
		)),
	)
}

func createDatabaseForEnsure(
	p EnsureDatabaseParams,
) IOE.IOEither[error, EnsureDatabaseResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (F.Void, error) {
			_, err := p.Client.CreateDatabase(context.Background(), p.Database, nil)
			return F.VOID, err
		}),
		IOE.MapLeft[F.Void](fperrors.OnError(
			fmt.Sprintf("error creating database %s", p.Database),
		)),
		IOE.Map[error](func(_ F.Void) EnsureDatabaseResult {
			return P.MakePair(true, p.Database)
		}),
	)
}

// createDatabasePipeline creates databases and optionally creates a user with grants.
func createDatabasePipeline(
	p DatabaseParams,
) IOE.IOEither[error, CreateDatabaseResult] {
	return F.Pipe2(
		F.Pipe1(
			p.Databases,
			A.Map(func(dbname string) SingleDBParams {
				return SingleDBParams{
					Client: p.Client,
					Logger: p.Logger,
					Dbname: dbname,
				}
			}),
		),
		IOE.TraverseArraySeq(createSingleDatabase),
		IOE.Chain(maybeCreateUserAndGrant(p)),
	)
}

// createSingleDatabase creates a single database if needed and returns creation status.
func createSingleDatabase(p SingleDBParams) IOE.IOEither[error, CreateSingleDBResult] {
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
				IOE.Of[error](P.MakePair(false, p.Dbname)),
			),
			F.Constant1[bool](createDatabase(p)),
		)),
	)
}

// createDatabase creates a single database and returns creation status.
func createDatabase(p SingleDBParams) IOE.IOEither[error, CreateSingleDBResult] {
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
		IOE.Map[error](func(_ F.Void) CreateSingleDBResult {
			return P.MakePair(true, p.Dbname)
		}),
	)
}

// maybeCreateUserAndGrant returns a Kleisli arrow that creates a user and
// grants access if Username is non-empty, otherwise returns database outcomes.
func maybeCreateUserAndGrant(
	p DatabaseParams,
) func([]CreateSingleDBResult) IOE.IOEither[error, CreateDatabaseResult] {
	return func(
		dbResults []CreateSingleDBResult,
	) IOE.IOEither[error, CreateDatabaseResult] {
		return F.Pipe2(
			p.Username,
			O.FromPredicate(func(s string) bool { return len(s) > 0 }),
			O.Fold(
				func() IOE.IOEither[error, CreateDatabaseResult] {
					return IOE.Of[error](CreateDatabaseResult{
						Databases: dbResults,
						HasUser:   false,
					})
				},
				func(_ string) IOE.IOEither[error, CreateDatabaseResult] {
					return F.Pipe1(
						createUserAndGrant(p),
						IOE.Map[error](func(
							result P.Pair[CreateUserResult, []CreateGrantResult],
						) CreateDatabaseResult {
							userResult := P.First(result)
							return CreateDatabaseResult{
								Databases: dbResults,
								HasUser:   true,
								User:      userResult,
								Grants:    P.Second(result),
							}
						}),
					)
				},
			),
		)
	}
}

// createUserAndGrant creates/gets a user and grants access to all databases.
func createUserAndGrant(
	p DatabaseParams,
) IOE.IOEither[error, P.Pair[CreateUserResult, []CreateGrantResult]] {
	return F.Pipe3(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool](fperrors.OnError(
			fmt.Sprintf("error checking for user %s", p.Username),
		)),
		IOE.Chain(F.Ternary(
			F.Identity[bool],
			func(_ bool) IOE.IOEither[error, CreateUserResult] {
				return F.Pipe2(
					IOE.TryCatchError(func() (driver.User, error) {
						return p.Client.User(context.Background(), p.Username)
					}),
					IOE.MapLeft[driver.User](fperrors.OnError(
						fmt.Sprintf("error fetching user %s", p.Username),
					)),
					IOE.Map[error](withCreateStatus(false)),
				)
			},
			func(_ bool) IOE.IOEither[error, CreateUserResult] {
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
					IOE.Map[error](withCreateStatus(true)),
				)
			},
		)),
		IOE.Chain(func(
			userResult CreateUserResult,
		) IOE.IOEither[error, P.Pair[CreateUserResult, []CreateGrantResult]] {
			return F.Pipe2(
				F.Pipe1(
					p.Databases,
					A.Map(func(dbname string) GrantDBParams {
						return GrantDBParams{
							Client: p.Client,
							Logger: p.Logger,
							Dbname: dbname,
							Grant:  p.Grant,
							User:   P.Second(userResult),
						}
					}),
				),
				IOE.TraverseArraySeq(grantSingleDatabase),
				IOE.Map[error](func(
					grants []CreateGrantResult,
				) P.Pair[CreateUserResult, []CreateGrantResult] {
					return P.MakePair(userResult, grants)
				}),
			)
		}),
	)
}

// grantSingleDatabase grants access to a single database for a user
func grantSingleDatabase(g GrantDBParams) IOE.IOEither[error, CreateGrantResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.Database, error) {
			return g.Client.Database(
				context.Background(),
				g.Dbname,
			)
		}),
		IOE.MapLeft[driver.Database](fperrors.OnError(
			fmt.Sprintf("error getting database %s", g.Dbname),
		)),
		IOE.Chain(func(
			db driver.Database,
		) IOE.IOEither[error, CreateGrantResult] {
			return F.Pipe1(
				IOE.TryCatchError(func() (CreateGrantResult, error) {
					err := g.User.SetDatabaseAccess(
						context.Background(),
						db,
						getGrant(g.Grant),
					)
					if err != nil {
						var zero CreateGrantResult
						return zero, err
					}
					return P.MakePair(g.Dbname, g.Grant), nil
				}),
				IOE.MapLeft[CreateGrantResult](fperrors.OnError(
					fmt.Sprintf(
						"error granting access to database %s",
						g.Dbname,
					),
				)),
			)
		}),
	)
}

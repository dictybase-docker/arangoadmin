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

// EnsureDatabase ensures a single database exists, creating it if missing.
func EnsureDatabase(_ context.Context, cmd *cli.Command) error {
	output := F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) EnsureDatabaseParams {
			return EnsureDatabaseParams{
				Client:   client,
				Logger:   newLogger(cmd),
				Database: cmd.String("database"),
			}
		}),
		IOE.Chain(ensureDatabasePipeline),
		toEither,
		E.Fold(
			func(err error) P.Pair[EnsureDatabaseResult, error] {
				var zero EnsureDatabaseResult
				return P.MakePair(zero, err)
			},
			func(result EnsureDatabaseResult) P.Pair[EnsureDatabaseResult, error] {
				return P.MakePair[EnsureDatabaseResult, error](result, nil)
			},
		),
	)
	if err := P.Second(output); err != nil {
		return err
	}
	logEnsureDatabaseOutcome(newLogger(cmd), P.First(output))
	return nil
}

func ensureDatabasePipeline(
	params EnsureDatabaseParams,
) IOE.IOEither[error, EnsureDatabaseResult] {
	return F.Pipe3(
		params,
		checkDatabaseExistenceForEnsure,
		IOE.Map[error](func(exists bool) P.Pair[bool, EnsureDatabaseParams] {
			return P.MakePair(exists, params)
		}),
		IOE.Chain(routeEnsureDatabase),
	)
}

func checkDatabaseExistenceForEnsure(
	params EnsureDatabaseParams,
) IOE.IOEither[error, bool] {
	return F.Pipe1(
		IOE.TryCatchError(func() (bool, error) {
			return params.Client.DatabaseExists(
				context.Background(),
				params.Database,
			)
		}),
		IOE.MapLeft[bool](fperrors.OnError(
			fmt.Sprintf("error checking for database %s", params.Database),
		)),
	)
}

func routeEnsureDatabase(
	params P.Pair[bool, EnsureDatabaseParams],
) IOE.IOEither[error, EnsureDatabaseResult] {
	return F.Pipe1(
		params,
		F.Ternary(
			P.First[bool, EnsureDatabaseParams],
			handleExistingDatabase,
			handleNewDatabase,
		),
	)
}

func handleExistingDatabase(
	params P.Pair[bool, EnsureDatabaseParams],
) IOE.IOEither[error, EnsureDatabaseResult] {
	p := P.Second(params)
	return IOE.Of[error](P.MakePair(false, p.Database))
}

func handleNewDatabase(
	params P.Pair[bool, EnsureDatabaseParams],
) IOE.IOEither[error, EnsureDatabaseResult] {
	p := P.Second(params)
	return F.Pipe1(
		IOE.TryCatchError(func() (EnsureDatabaseResult, error) {
			_, err := p.Client.CreateDatabase(context.Background(), p.Database, nil)
			if driver.IsConflict(err) {
				return P.MakePair(false, p.Database), nil
			}
			return P.MakePair(true, p.Database), err
		}),
		IOE.MapLeft[EnsureDatabaseResult](fperrors.OnError(
			fmt.Sprintf("error creating database %s", p.Database),
		)),
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
	return F.Pipe1(
		IOE.TryCatchError(func() (CreateSingleDBResult, error) {
			_, err := p.Client.CreateDatabase(
				context.Background(),
				p.Dbname,
				nil,
			)
			return P.MakePair(true, p.Dbname), err
		}),
		IOE.MapLeft[CreateSingleDBResult](fperrors.OnError(
			fmt.Sprintf("error creating database %s", p.Dbname),
		)),
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

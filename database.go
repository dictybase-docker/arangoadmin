package main

import (
	"context"
	"fmt"

	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

// EnsureDatabase ensures a single database exists, creating it if missing.
func EnsureDatabase(ctx context.Context, cmd *cli.Command) error {
	logger := newLogger(cmd)
	return F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) EnsureDatabaseParams {
			return EnsureDatabaseParams{
				Context:  ctx,
				Client:   client,
				Database: cmd.String("database"),
			}
		}),
		IOE.Chain(ensureDatabasePipeline),
		IOE.ChainFirstIOK[error](logEnsureDatabase(logger)),
		foldIOE[EnsureDatabaseResult],
	)
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
				params.Context,
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
			_, err := p.Client.CreateDatabase(p.Context, p.Database, nil)
			return P.MakePair(true, p.Database), err
		}),
		IOE.MapLeft[EnsureDatabaseResult](fperrors.OnError(
			fmt.Sprintf("error creating database %s", p.Database),
		)),
	)
}


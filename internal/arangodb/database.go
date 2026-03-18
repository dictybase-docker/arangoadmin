package arangodb

import (
	"fmt"

	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/pair"
)

// EnsureDatabasePipeline ensures a single database exists, creating it if missing.
func EnsureDatabasePipeline(
	params EnsureDatabaseParams,
) IOE.IOEither[error, EnsureDatabaseResult] {
	return F.Pipe3(
		params,
		checkDatabaseExistenceForEnsure,
		IOE.Map[error](
			func(exists bool) P.Pair[bool, EnsureDatabaseParams] {
				return P.MakePair(exists, params)
			},
		),
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
			fmt.Sprintf(
				"error checking for database %s",
				params.Database,
			),
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
			_, err := p.Client.CreateDatabase(
				p.Context,
				p.Database,
				nil,
			)
			return P.MakePair(true, p.Database), err
		}),
		IOE.MapLeft[EnsureDatabaseResult](fperrors.OnError(
			fmt.Sprintf(
				"error creating database %s",
				p.Database,
			),
		)),
	)
}

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

// EnsureGrant ensures that a user has a specific grant level on a database.
func EnsureGrant(ctx context.Context, cmd *cli.Command) error {
	logger := newLogger(cmd)
	return F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) EnsureGrantParams {
			return EnsureGrantParams{
				Context:  ctx,
				Client:   client,
				Logger:   logger,
				Username: cmd.String("user"),
				Database: cmd.String("database"),
				Grant:    cmd.String("grant"),
			}
		}),
		IOE.Chain(ensureGrantPipeline),
		IOE.ChainFirstIOK[error](logEnsureGrant(logger)),
		foldIOE[EnsureGrantResult],
	)
}

func validateGrantParams(p EnsureGrantParams) IOE.IOEither[error, EnsureGrantParams] {
	if p.Username == "" {
		return IOE.Left[EnsureGrantParams](fmt.Errorf("username cannot be empty"))
	}
	if p.Database == "" {
		return IOE.Left[EnsureGrantParams](fmt.Errorf("database cannot be empty"))
	}
	return IOE.Of[error](p)
}

func ensureGrantPipeline(p EnsureGrantParams) IOE.IOEither[error, EnsureGrantResult] {
	return F.Pipe4(
		p,
		validateGrantParams,
		IOE.Chain(fetchGrantUser),
		IOE.Chain(fetchGrantDatabase),
		IOE.Chain(applyGrant),
	)
}

func fetchGrantUser(p EnsureGrantParams) IOE.IOEither[error, GrantState] {
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.User, error) {
			return p.Client.User(p.Context, p.Username)
		}),
		IOE.MapLeft[driver.User](fperrors.OnError(
			fmt.Sprintf("error fetching user %s", p.Username),
		)),
		IOE.Map[error](func(u driver.User) GrantState {
			return GrantState{Params: p, User: u}
		}),
	)
}

func fetchGrantDatabase(s GrantState) IOE.IOEither[error, GrantState] {
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.Database, error) {
			return s.Params.Client.Database(s.Params.Context, s.Params.Database)
		}),
		IOE.MapLeft[driver.Database](fperrors.OnError(
			fmt.Sprintf("error fetching database %s", s.Params.Database),
		)),
		IOE.Map[error](func(db driver.Database) GrantState {
			s.DB = db
			return s
		}),
	)
}

func applyGrant(s GrantState) IOE.IOEither[error, EnsureGrantResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (F.Void, error) {
			return F.VOID, s.User.SetDatabaseAccess(
				s.Params.Context,
				s.DB,
				getGrant(s.Params.Grant),
			)
		}),
		IOE.MapLeft[F.Void](fperrors.OnError(
			fmt.Sprintf("error granting access to database %s", s.Params.Database),
		)),
		IOE.Map[error](func(_ F.Void) EnsureGrantResult {
			return P.MakePair(s.Params.Database, s.Params.Grant)
		}),
	)
}

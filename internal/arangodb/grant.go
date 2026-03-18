package arangodb

import (
	"fmt"

	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/pair"
	Pred "github.com/IBM/fp-go/v2/predicate"
	Str "github.com/IBM/fp-go/v2/string"
	driver "github.com/arangodb/go-driver"
)

var (
	nonEmptyUsername = F.Pipe1(
		Str.IsNonEmpty,
		Pred.ContraMap(func(p EnsureGrantParams) string {
			return p.Username
		}),
	)
	nonEmptyDatabase = F.Pipe1(
		Str.IsNonEmpty,
		Pred.ContraMap(func(p EnsureGrantParams) string {
			return p.Database
		}),
	)
)

// EnsureGrantPipeline ensures one user has one grant level on one database.
func EnsureGrantPipeline(
	p EnsureGrantParams,
) IOE.IOEither[error, EnsureGrantResult] {
	return F.Pipe4(
		p,
		validateGrantParams,
		IOE.Chain(fetchGrantUser),
		IOE.Chain(fetchGrantDatabase),
		IOE.Chain(ApplyGrant),
	)
}

func validateGrantParams(
	p EnsureGrantParams,
) IOE.IOEither[error, EnsureGrantParams] {
	return F.Pipe3(
		p,
		O.FromPredicate(nonEmptyUsername),
		IOE.FromOption[EnsureGrantParams](func() error {
			return fmt.Errorf("username cannot be empty")
		}),
		IOE.Chain(F.Flow2(
			O.FromPredicate(nonEmptyDatabase),
			IOE.FromOption[EnsureGrantParams](func() error {
				return fmt.Errorf("database cannot be empty")
			}),
		)),
	)
}

func fetchGrantUser(
	p EnsureGrantParams,
) IOE.IOEither[error, GrantState] {
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

func fetchGrantDatabase(
	s GrantState,
) IOE.IOEither[error, GrantState] {
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.Database, error) {
			return s.Params.Client.Database(
				s.Params.Context,
				s.Params.Database,
			)
		}),
		IOE.MapLeft[driver.Database](fperrors.OnError(
			fmt.Sprintf(
				"error fetching database %s",
				s.Params.Database,
			),
		)),
		IOE.Map[error](func(db driver.Database) GrantState {
			s.DB = db
			return s
		}),
	)
}

// ApplyGrant applies the grant level from GrantState to the database.
func ApplyGrant(
	s GrantState,
) IOE.IOEither[error, EnsureGrantResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (F.Void, error) {
			return F.VOID, s.User.SetDatabaseAccess(
				s.Params.Context,
				s.DB,
				GetGrant(s.Params.Grant),
			)
		}),
		IOE.MapLeft[F.Void](fperrors.OnError(
			fmt.Sprintf(
				"error granting access to database %s",
				s.Params.Database,
			),
		)),
		IOE.Map[error](func(_ F.Void) EnsureGrantResult {
			return P.MakePair(s.Params.Database, s.Params.Grant)
		}),
	)
}

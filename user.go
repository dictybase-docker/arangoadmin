package main

import (
	"context"
	"fmt"

	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/pair"
	PR "github.com/IBM/fp-go/v2/predicate"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

type EnsureExistingUserPolicyInput struct {
	Params EnsureUserParams
	User   driver.User
}

// EnsureUser adds a new user or updates an existing one based on the policy.
func EnsureUser(ctx context.Context, cmd *cli.Command) error {
	logger := newLogger(cmd)
	return F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) EnsureUserParams {
			return EnsureUserParams{
				Context:  ctx,
				Client:   client,
				Username: cmd.String("user"),
				Password: cmd.String("password"),
				Policy:   cmd.String("password-policy"),
			}
		}),
		IOE.Chain(ensureUserPipeline),
		IOE.ChainFirstIOK[error](logEnsureUser(logger)),
		foldIOE[EnsureUserResult],
	)
}

func ensureUserPipeline(params EnsureUserParams) IOE.IOEither[error, EnsureUserResult] {
	return F.Pipe3(
		params,
		checkUserExistenceForEnsure,
		IOE.Map[error](func(exists bool) P.Pair[bool, EnsureUserParams] {
			return P.MakePair(exists, params)
		}),
		IOE.Chain(routeEnsureUser),
	)
}

func checkUserExistenceForEnsure(params EnsureUserParams) IOE.IOEither[error, bool] {
	return F.Pipe1(
		IOE.TryCatchError(func() (bool, error) {
			return params.Client.UserExists(
				params.Context,
				params.Username,
			)
		}),
		IOE.MapLeft[bool](fperrors.OnError(
			fmt.Sprintf("error checking for user %s", params.Username),
		)),
	)
}

func routeEnsureUser(
	params P.Pair[bool, EnsureUserParams],
) IOE.IOEither[error, EnsureUserResult] {
	return F.Pipe1(
		params,
		F.Ternary(
			P.First[bool, EnsureUserParams],
			ensureExistingUserFlow,
			ensureNewUserFlow,
		),
	)
}

func ensureNewUserFlow(
	params P.Pair[bool, EnsureUserParams],
) IOE.IOEither[error, EnsureUserResult] {
	p := P.Second(params)
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.User, error) {
			return p.Client.CreateUser(
				p.Context,
				p.Username,
				&driver.UserOptions{Password: p.Password},
			)
		}),
		IOE.MapLeft[driver.User](fperrors.OnError(
			fmt.Sprintf("error creating user %s", p.Username),
		)),
		IOE.Map[error](func(user driver.User) EnsureUserResult {
			return P.MakePair(UserCreated, user)
		}),
	)
}

func ensureExistingUserFlow(
	params P.Pair[bool, EnsureUserParams],
) IOE.IOEither[error, EnsureUserResult] {
	return F.Pipe3(
		P.Second(params),
		fetchExistingEnsureUser,
		IOE.Map[error](func(user driver.User) EnsureExistingUserPolicyInput {
			return EnsureExistingUserPolicyInput{
				Params: P.Second(params),
				User:   user,
			}
		}),
		IOE.Chain(applyExistingUserPolicy),
	)
}

func fetchExistingEnsureUser(p EnsureUserParams) IOE.IOEither[error, driver.User] {
	return F.Pipe1(
		IOE.TryCatchError(func() (driver.User, error) {
			return p.Client.User(p.Context, p.Username)
		}),
		IOE.MapLeft[driver.User](fperrors.OnError(
			fmt.Sprintf("error fetching user %s", p.Username),
		)),
	)
}

func applyExistingUserPolicy(
	input EnsureExistingUserPolicyInput,
) IOE.IOEither[error, EnsureUserResult] {
	isPasswordProvided := F.Pipe2(
		input.Params.Password,
		PR.IsNonZero[string](),
		F.Constant1[string],
	)

	shouldUpdatePassword := F.Pipe2(
		PR.IsStrictEqual[string]()("if-provided"),
		PR.And(isPasswordProvided),
		PR.Or(PR.IsStrictEqual[string]()("always")),
	)

	return F.Pipe1(
		input.Params.Policy,
		F.Ternary(
			shouldUpdatePassword,
			func(_ string) IOE.IOEither[error, EnsureUserResult] {
				return updateEnsureUserPassword(
					input.Params,
					input.User,
				)
			},
			func(_ string) IOE.IOEither[error, EnsureUserResult] {
				return IOE.Of[error](
					P.MakePair(
						UserExisting,
						input.User,
					),
				)
			},
		),
	)
}

func updateEnsureUserPassword(
	p EnsureUserParams,
	user driver.User,
) IOE.IOEither[error, EnsureUserResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.User, error) {
			return user, user.Update(
				p.Context,
				driver.UserOptions{Password: p.Password},
			)
		}),
		IOE.MapLeft[driver.User](fperrors.OnError(
			fmt.Sprintf("error updating user %s", p.Username),
		)),
		IOE.Map[error](func(u driver.User) EnsureUserResult {
			return P.MakePair(UserUpdated, u)
		}),
	)
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

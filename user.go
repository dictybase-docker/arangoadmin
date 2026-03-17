package main

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/pair"
	PR "github.com/IBM/fp-go/v2/predicate"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

type UserForUpdate struct {
	Params UserParams
	User   driver.User
}

// CreateUser adds a new user with pre-specified privileges to ArangoDB
func CreateUser(_ context.Context, cmd *cli.Command) error {
	output := F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) CreateUserParams {
			return CreateUserParams{
				Client:   client,
				Username: cmd.String("user"),
				Password: cmd.String("password"),
			}
		}),
		IOE.Chain(createUserPipeline),
		toEither,
		E.Fold(
			func(err error) P.Pair[CreateUserResult, error] {
				var zero CreateUserResult
				return P.MakePair(zero, err)
			},
			func(result CreateUserResult) P.Pair[CreateUserResult, error] {
				return P.MakePair[CreateUserResult, error](result, nil)
			},
		),
	)
	if err := P.Second(output); err != nil {
		return err
	}

	logCreateUserOutcome(newLogger(cmd), P.First(output))
	return nil
}

// createUserPipeline creates a user if they don't exist or returns the existing user.
func createUserPipeline(
	p CreateUserParams,
) IOE.IOEither[error, CreateUserResult] {
	return F.Pipe3(
		p,
		checkUserExistence,
		IOE.Map[error](func(exists bool) CreateUserRouteParams {
			return P.MakePair(exists, p)
		}),
		IOE.Chain(routeUserCreation),
	)
}

// checkUserExistence checks whether the user exists in ArangoDB
func checkUserExistence(p CreateUserParams) IOE.IOEither[error, bool] {
	return F.Pipe1(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool](fperrors.OnError(
			fmt.Sprintf("error checking for user %s", p.Username),
		)),
	)
}

// routeUserCreation routes to either fetching an existing user or creating a new one.
func routeUserCreation(
	params CreateUserRouteParams,
) IOE.IOEither[error, CreateUserResult] {
	return F.Pipe1(
		params,
		F.Ternary(
			P.First[bool, CreateUserParams],
			handleExistingUser,
			handleNewUser,
		),
	)
}

func handleExistingUser(
	params CreateUserRouteParams,
) IOE.IOEither[error, CreateUserResult] {
	return F.Pipe2(
		P.Second(params),
		fetchExistingUser,
		IOE.Map[error](withCreateStatus(false)),
	)
}

func handleNewUser(
	params CreateUserRouteParams,
) IOE.IOEither[error, CreateUserResult] {
	return F.Pipe2(
		P.Second(params),
		createNewUser,
		IOE.Map[error](withCreateStatus(true)),
	)
}

func fetchExistingUser(p CreateUserParams) IOE.IOEither[error, driver.User] {
	return F.Pipe1(
		IOE.TryCatchError(func() (driver.User, error) {
			return p.Client.User(context.Background(), p.Username)
		}),
		IOE.MapLeft[driver.User](fperrors.OnError(
			fmt.Sprintf("error fetching user %s", p.Username),
		)),
	)
}

func withCreateStatus(created bool) func(driver.User) CreateUserResult {
	return func(user driver.User) CreateUserResult {
		return P.MakePair(created, user)
	}
}

// createNewUser creates a new user in ArangoDB.
func createNewUser(p CreateUserParams) IOE.IOEither[error, driver.User] {
	return F.Pipe1(
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
	)
}

// UpdateUser updates the password of an existing user in ArangoDB
func UpdateUser(_ context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) UserParams {
			return UserParams{
				WithClient: WithClient{
					Client: client,
					Logger: newLogger(cmd),
				},
				Username: cmd.String("user"),
				Password: cmd.String("password"),
			}
		}),
		IOE.Chain(updateUserPipeline),
		foldIOE[F.Void],
	)
}

// updateUserPipeline updates a user's password if they exist
func updateUserPipeline(p UserParams) IOE.IOEither[error, F.Void] {
	return F.Pipe3(
		p,
		checkUserExists,
		IOE.Chain(getExistingUser),
		IOE.Chain(updateExistingUser),
	)
}

func checkUserExists(p UserParams) IOE.IOEither[error, UserParams] {
	return F.Pipe3(
		IOE.TryCatchError(func() (bool, error) {
			return p.Client.UserExists(context.Background(), p.Username)
		}),
		IOE.MapLeft[bool](fperrors.OnError(
			fmt.Sprintf("error checking for user %s", p.Username),
		)),
		IOE.ChainEitherK(E.FromPredicate(
			F.Identity[bool],
			func(_ bool) error { return fmt.Errorf("user %s does not exist", p.Username) },
		)),
		IOE.Map[error](F.Constant1[bool](p)),
	)
}

func getExistingUser(p UserParams) IOE.IOEither[error, UserForUpdate] {
	return F.Pipe2(
		IOE.TryCatchError(func() (driver.User, error) {
			return p.Client.User(context.Background(), p.Username)
		}),
		IOE.MapLeft[driver.User](
			fperrors.OnError(fmt.Sprintf("error fetching user %s", p.Username)),
		),
		IOE.Map[error](withExistingUser(p)),
	)
}

func withExistingUser(p UserParams) func(driver.User) UserForUpdate {
	return func(user driver.User) UserForUpdate {
		return UserForUpdate{Params: p, User: user}
	}
}

func updateExistingUser(u UserForUpdate) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (F.Void, error) {
			return F.VOID, u.User.Update(
				context.Background(),
				driver.UserOptions{Password: u.Params.Password},
			)
		}),
		IOE.MapLeft[F.Void](fperrors.OnError(
			fmt.Sprintf("error updating user %s", u.Params.Username),
		)),
		IOE.ChainFirstIOK[error](logUserUpdatedStep(u)),
	)
}

func logUserUpdatedStep(u UserForUpdate) func(F.Void) IO.IO[F.Void] {
	return func(_ F.Void) IO.IO[F.Void] {
		return logUserUpdated(u.Params.Logger, u.Params.Username)
	}
}

type EnsureExistingUserPolicyInput struct {
	Params EnsureUserParams
	User   driver.User
}

// EnsureUser adds a new user or updates an existing one based on the policy.
func EnsureUser(_ context.Context, cmd *cli.Command) error {
	output := F.Pipe6(
		cmd,
		connParamsFromCmd,
		createArangoClient,
		IOE.Map[error](func(client driver.Client) EnsureUserParams {
			return EnsureUserParams{
				Client:   client,
				Logger:   newLogger(cmd),
				Username: cmd.String("user"),
				Password: cmd.String("password"),
				Policy:   cmd.String("password-policy"),
			}
		}),
		IOE.Chain(ensureUserPipeline),
		toEither,
		E.Fold(
			func(err error) P.Pair[EnsureUserResult, error] {
				var zero EnsureUserResult
				return P.MakePair(zero, err)
			},
			func(result EnsureUserResult) P.Pair[EnsureUserResult, error] {
				return P.MakePair[EnsureUserResult, error](result, nil)
			},
		),
	)
	if err := P.Second(output); err != nil {
		return err
	}
	logEnsureUserOutcome(newLogger(cmd), P.First(output))
	return nil
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
				context.Background(),
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
				context.Background(),
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
			return p.Client.User(context.Background(), p.Username)
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
				context.Background(),
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

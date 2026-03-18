package cli

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/pair"
	"github.com/dictybase-docker/arangoadmin/internal/arangodb"
	"github.com/dictybase-docker/arangoadmin/internal/logger"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

// EnsureUserCommand returns the CLI command definition for ensure-user.
func EnsureUserCommand() *cli.Command {
	return &cli.Command{
		Name:   "ensure-user",
		Usage:  "create or update a user for accessing arangodb",
		Action: EnsureUser,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "admin-user",
				Aliases: []string{"au"},
				Usage:   "arangodb admin user",
				Value:   "root",
			},
			&cli.StringFlag{
				Name:    "admin-password",
				Aliases: []string{"ap"},
				Usage:   "arangodb admin password",
			},
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"u"},
				Usage:    "arangodb user",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "password",
				Aliases:  []string{"pw"},
				Usage:    "arangodb password for user",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "password-policy",
				Usage: "policy for updating password (never, if-provided, always)",
				Value: "never",
			},
		},
	}
}

// EnsureUser is the CLI action for the ensure-user command.
func EnsureUser(ctx context.Context, cmd *cli.Command) error {
	output := F.Pipe6(
		cmd,
		connParamsFromCmd,
		arangodb.CreateArangoClient,
		IOE.Map[error](func(client driver.Client) arangodb.EnsureUserParams {
			return arangodb.EnsureUserParams{
				Context:  ctx,
				Client:   client,
				Username: cmd.String("user"),
				Password: cmd.String("password"),
				Policy:   cmd.String("password-policy"),
			}
		}),
		IOE.Chain(arangodb.EnsureUserPipeline),
		arangodb.ToEither[error, arangodb.EnsureUserResult],
		E.Fold(
			func(err error) P.Pair[arangodb.EnsureUserResult, error] {
				var zero arangodb.EnsureUserResult
				return P.MakePair(zero, err)
			},
			func(result arangodb.EnsureUserResult) P.Pair[arangodb.EnsureUserResult, error] {
				return P.MakePair[arangodb.EnsureUserResult, error](result, nil)
			},
		),
	)
	if err := P.Second(output); err != nil {
		return err
	}
	arangodb.LogEnsureUserOutcome(logger.NewLogger(cmd), P.First(output))
	return nil
}

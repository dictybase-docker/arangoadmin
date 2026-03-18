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

// EnsureGrantCommand returns the CLI command definition for ensure-grant.
func EnsureGrantCommand() *cli.Command {
	return &cli.Command{
		Name:   "ensure-grant",
		Usage:  "ensure a user has a specific grant level on a database",
		Action: EnsureGrant,
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
				Name:     "database",
				Aliases:  []string{"db"},
				Usage:    "name of arangodb database",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "grant",
				Aliases: []string{"g"},
				Usage:   "level of access for arangodb user (rw, ro, none)",
				Value:   "rw",
			},
		},
	}
}

// EnsureGrant is the CLI action for the ensure-grant command.
func EnsureGrant(ctx context.Context, cmd *cli.Command) error {
	output := F.Pipe6(
		cmd,
		connParamsFromCmd,
		arangodb.CreateArangoClient,
		IOE.Map[error](func(client driver.Client) arangodb.EnsureGrantParams {
			return arangodb.EnsureGrantParams{
				Context:  ctx,
				Client:   client,
				Username: cmd.String("user"),
				Database: cmd.String("database"),
				Grant:    cmd.String("grant"),
			}
		}),
		IOE.Chain(arangodb.EnsureGrantPipeline),
		arangodb.ToEither[error, arangodb.EnsureGrantResult],
		E.Fold(
			func(err error) P.Pair[arangodb.EnsureGrantResult, error] {
				var zero arangodb.EnsureGrantResult
				return P.MakePair(zero, err)
			},
			func(result arangodb.EnsureGrantResult) P.Pair[arangodb.EnsureGrantResult, error] {
				return P.MakePair[arangodb.EnsureGrantResult, error](result, nil)
			},
		),
	)
	if err := P.Second(output); err != nil {
		return err
	}
	arangodb.LogEnsureGrantOutcome(logger.NewLogger(cmd), P.First(output))
	return nil
}

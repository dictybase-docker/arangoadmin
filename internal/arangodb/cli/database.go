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

var (
	ensureDatabaseError   = F.Bind1st(P.MakePair[arangodb.EnsureDatabaseResult, error], arangodb.EnsureDatabaseResult{})
	ensureDatabaseSuccess = F.Bind2nd(P.MakePair[arangodb.EnsureDatabaseResult, error], (error)(nil))
)

// EnsureDatabaseCommand returns the CLI command definition for ensure-database.
func EnsureDatabaseCommand() *cli.Command {
	return &cli.Command{
		Name:   "ensure-database",
		Usage:  "create a single arangodb database if missing",
		Action: EnsureDatabase,
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
				Name:     "database",
				Aliases:  []string{"db"},
				Usage:    "name of arangodb database",
				Required: true,
			},
		},
	}
}

// EnsureDatabase is the CLI action for the ensure-database command.
func EnsureDatabase(ctx context.Context, cmd *cli.Command) error {
	output := F.Pipe6(
		cmd,
		connParamsFromCmd,
		arangodb.CreateArangoClient,
		IOE.Map[error](func(client driver.Client) arangodb.EnsureDatabaseParams {
			return arangodb.EnsureDatabaseParams{
				Context:  ctx,
				Client:   client,
				Database: cmd.String("database"),
			}
		}),
		IOE.Chain(arangodb.EnsureDatabasePipeline),
		arangodb.ToEither[error, arangodb.EnsureDatabaseResult],
		E.Fold(ensureDatabaseError, ensureDatabaseSuccess),
	)
	if err := P.Second(output); err != nil {
		return err
	}
	arangodb.LogEnsureDatabaseOutcome(logger.NewLogger(cmd), P.First(output))
	return nil
}

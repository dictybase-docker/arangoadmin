package main

import (
	"context"
	"fmt"
	"os"

	arangocli "github.com/dictybase-docker/arangoadmin/internal/arangodb/cli"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "arangoadmin",
		Usage:   "cli for creating databases and users in arangodb",
		Version: "1.0.0",
		Flags:   arangocli.GlobalFlags(),
		Commands: []*cli.Command{
			arangocli.EnsureUserCommand(),
			arangocli.EnsureDatabaseCommand(),
			arangocli.EnsureGrantCommand(),
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

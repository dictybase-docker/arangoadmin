package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func ValidateDatabaseArgs(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	for _, p := range []string{
		"admin-user",
		"database",
	} {
		if len(cmd.String(p)) == 0 {
			return ctx, cli.Exit(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return ctx, nil
}

func ValidateUserArgs(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	for _, p := range []string{
		"admin-user",
		"user",
		"password",
	} {
		if len(cmd.String(p)) == 0 {
			return ctx, cli.Exit(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return ctx, nil
}

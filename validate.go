package main

import (
	"fmt"

	cli "gopkg.in/urfave/cli.v1"
)

func ValidateDatabaseArgs(c *cli.Context) error {
	for _, p := range []string{
		"admin-user",
		"admin-password",
		"database",
	} {
		if len(c.String(p)) == 0 {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return nil
}

func ValidateUserArgs(c *cli.Context) error {
	for _, p := range []string{
		"admin-user",
		"database",
		"user",
		"grant",
	} {
		if len(c.String(p)) == 0 {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return nil
}

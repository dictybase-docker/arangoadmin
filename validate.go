package main

import (
	"fmt"

	"gopkg.in/urfave/cli.v1"
)

func validateCreateDatabase(c *cli.Context) error {
	for _, p := range []string{
		"admin-user",
		"admin-password",
		"database",
		"user",
		"password",
	} {
		if !c.IsSet(p) {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return nil
}

func validateCreateCollection(c *cli.Context) error {
	for _, p := range []string{
		"database",
		"user",
		"password",
		"collection",
	} {
		if !c.IsSet(p) {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return nil
}

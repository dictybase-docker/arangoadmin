package main

import (
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/vst"
	"gopkg.in/urfave/cli.v1"
)

func getClient(host, port, user, pass string) (driver.Client, error) {
	var c driver.Client
	conn, err := vst.NewConnection(
		vst.ConnectionConfig{
			Endpoints: []string{
				fmt.Sprintf("vst://%s:%s", host, port),
			},
		})
	if err != nil {
		return c, cli.NewExitError(
			fmt.Sprintf("could not connect %s", err),
			2,
		)
	}
	client, err := driver.NewClient(
		driver.ClientConfig{
			Connection: conn,
			Authentication: driver.BasicAuthentication(
				user,
				pass,
			),
		})
	if err != nil {
		return c, cli.NewExitError(
			fmt.Sprintf("could not get a client instance %s", err),
			2,
		)
	}
	return client, nil
}

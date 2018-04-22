package main

import (
	"crypto/tls"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

func getClient(host, port, user, pass string) (driver.Client, error) {
	var c driver.Client
	conn, err := http.NewConnection(
		http.ConnectionConfig{
			Endpoints: []string{
				fmt.Sprintf("https://%s:%s", host, port),
			},
			TLSConfig: &tls.Config{InsecureSkipVerify: true},
		})
	if err != nil {
		return c, fmt.Errorf("could not connect %s", err)
	}
	client, err := driver.NewClient(
		driver.ClientConfig{
			Connection:     conn,
			Authentication: driver.BasicAuthentication(user, pass),
		})
	if err != nil {
		return c, fmt.Errorf("could not get a client instance %s", err)
	}
	return client, nil
}

package main

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

func getClient(host, port, user, pass string, isSync bool) (driver.Client, error) {
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
	config := driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(user, pass),
	}
	if isSync {
		d, _ := time.ParseDuration("1s")
		config.SynchronizeEndpointsInterval = d
	}
	client, err := driver.NewClient(config)
	if err != nil {
		return c, fmt.Errorf("could not get a client instance %s", err)
	}
	return client, nil
}

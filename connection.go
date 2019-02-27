package main

import (
	"crypto/tls"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

type ClientParams struct {
	Host     string
	Port     string
	User     string
	Pass     string
	IsSecure bool
}

func getClient(p *ClientParams) (driver.Client, error) {
	var client driver.Client
	var conn driver.Connection
	host := p.Host
	port := p.Port
	user := p.User
	pass := p.Pass
	isSecure := p.IsSecure

	if isSecure {
		c, err := http.NewConnection(
			http.ConnectionConfig{
				Endpoints: []string{
					fmt.Sprintf("https://%s:%s", host, port),
				},
				TLSConfig: &tls.Config{InsecureSkipVerify: true},
			})
		if err != nil {
			return client, fmt.Errorf("could not connect %s", err)
		}
		conn = c
	} else {
		c, err := http.NewConnection(
			http.ConnectionConfig{
				Endpoints: []string{
					fmt.Sprintf("http://%s:%s", host, port),
				},
			})
		if err != nil {
			return client, fmt.Errorf("could not connect %s", err)
		}
		conn = c
	}
	client, err := driver.NewClient(
		driver.ClientConfig{
			Connection:     conn,
			Authentication: driver.BasicAuthentication(user, pass),
		})
	if err != nil {
		return client, fmt.Errorf("could not get a client instance %s", err)
	}
	return client, nil
}

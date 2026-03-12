package main

import (
	driver "github.com/arangodb/go-driver"
)

func getTestClient() (driver.Client, error) {
	return getClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
}

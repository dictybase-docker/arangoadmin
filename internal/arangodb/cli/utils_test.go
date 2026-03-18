package cli

import (
	driver "github.com/arangodb/go-driver"
	"github.com/dictybase-docker/arangoadmin/internal/arangodb"
)

func getTestClient() (driver.Client, error) {
	return arangodb.GetClient(&arangodb.ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
}

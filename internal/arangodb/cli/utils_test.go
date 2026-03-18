package cli

import (
	"github.com/dictybase-docker/arangoadmin/internal/arangodb"
	driver "github.com/arangodb/go-driver"
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

package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/arangodb"
)

var (
	arangoEndpoint string
	arangoHost     string
	arangoPort     string
)

const arangoPassword = "password"

func TestMain(m *testing.M) {
	ctx := context.Background()
	arangoContainer, err := arangodb.Run(ctx, "arangodb:3.11.6", arangodb.WithRootPassword(arangoPassword))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	arangoEndpoint, err = arangoContainer.HTTPEndpoint(ctx)
	if err != nil {
		log.Fatalf("failed to get endpoint: %s", err)
	}

	u, err := url.Parse(arangoEndpoint)
	if err != nil {
		log.Fatalf("failed to parse endpoint: %s", err)
	}
	arangoHost = u.Hostname()
	arangoPort = u.Port()

	code := m.Run()

	if err := arangoContainer.Terminate(ctx); err != nil {
		log.Fatalf("failed to terminate container: %s", err)
	}

	os.Exit(code)
}

package main

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/require"
)

func TestCreateArangoClientSuccess(t *testing.T) {
	require := require.New(t)
	params := ConnectionParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	}
	result := toEither(createArangoClient(params))
	require.True(E.IsRight(result))
}

func TestCreateArangoClientReturnsClient(t *testing.T) {
	require := require.New(t)
	// Test that the pipeline returns a non-nil client
	params := ConnectionParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	}
	result := toEither(createArangoClient(params))
	require.True(E.IsRight(result))
	// Unwrap the client
	client, err := E.UnwrapError(result)
	require.NoError(err)
	require.NotNil(client)
}

func TestCreateArangoClientBackwardCompat(t *testing.T) {
	require := require.New(t)
	// getClient wrapper must still work for existing tests
	client, err := getClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
	require.NoError(err)
	require.NotNil(client)
}

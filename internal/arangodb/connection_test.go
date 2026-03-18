package arangodb

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
	result := ToEither(CreateArangoClient(params))
	require.True(E.IsRight(result))
}

func TestCreateArangoClientReturnsClient(t *testing.T) {
	require := require.New(t)
	params := ConnectionParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	}
	result := ToEither(CreateArangoClient(params))
	require.True(E.IsRight(result))
	client, err := E.UnwrapError(result)
	require.NoError(err)
	require.NotNil(client)
}

func TestCreateArangoClientBackwardCompat(t *testing.T) {
	require := require.New(t)
	client, err := GetClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
	require.NoError(err)
	require.NotNil(client)
}

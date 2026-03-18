package arangodb

import (
	"testing"

	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureUser(t *testing.T) {
	require := require.New(t)

	client, err := getTestClient()
	require.NoError(err)

	user := "ensure_user_test"

	// 1. Create missing user
	p1 := EnsureUserParams{
		Client:   client,
		Username: user,
		Password: "pass1",
		Policy:   "never",
	}
	res1, err := ToTuple(EnsureUserPipeline(p1))
	require.NoError(err)
	require.Equal(UserCreated, P.First(res1))
	require.Equal(user, P.Second(res1).Name())

	// 2. Existing + never -> existing
	p2 := EnsureUserParams{
		Client:   client,
		Username: user,
		Password: "pass2",
		Policy:   "never",
	}
	res2, err := ToTuple(EnsureUserPipeline(p2))
	require.NoError(err)
	require.Equal(UserExisting, P.First(res2))

	// 3. Existing + always -> updated
	p3 := EnsureUserParams{
		Client:   client,
		Username: user,
		Password: "pass3",
		Policy:   "always",
	}
	res3, err := ToTuple(EnsureUserPipeline(p3))
	require.NoError(err)
	require.Equal(UserUpdated, P.First(res3))

	// 4. Existing + if-provided + no password -> existing
	p4 := EnsureUserParams{
		Client:   client,
		Username: user,
		Password: "",
		Policy:   "if-provided",
	}
	res4, err := ToTuple(EnsureUserPipeline(p4))
	require.NoError(err)
	require.Equal(UserExisting, P.First(res4))

	// 5. Existing + if-provided + password -> updated
	p5 := EnsureUserParams{
		Client:   client,
		Username: user,
		Password: "pass5",
		Policy:   "if-provided",
	}
	res5, err := ToTuple(EnsureUserPipeline(p5))
	require.NoError(err)
	require.Equal(UserUpdated, P.First(res5))

	// 6. Invalid policy -> existing
	p6 := EnsureUserParams{
		Client:   client,
		Username: user,
		Password: "pass6",
		Policy:   "invalid",
	}
	res6, err := ToTuple(EnsureUserPipeline(p6))
	require.NoError(err)
	require.Equal(UserExisting, P.First(res6))
}

func TestGetGrant(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(driver.GrantReadWrite, GetGrant("rw"))
	assert.Equal(driver.GrantReadOnly, GetGrant("ro"))
	assert.Equal(driver.GrantNone, GetGrant("none"))
	assert.Equal(driver.GrantNone, GetGrant("invalid"))
}

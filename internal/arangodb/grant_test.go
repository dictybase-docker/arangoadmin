package arangodb

import (
	"context"
	"fmt"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	driver "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/assert"
)

type errorUser struct {
	testUser
}

func (u errorUser) SetDatabaseAccess(
	_ context.Context,
	_ driver.Database,
	_ driver.Grant,
) error {
	return fmt.Errorf("forced error")
}

func TestApplyGrantFailure(t *testing.T) {
	assert := assert.New(t)

	state := GrantState{
		Params: EnsureGrantParams{
			Database: "test-db",
		},
		User: errorUser{},
	}

	result := ToEither(ApplyGrant(state))
	assert.True(E.IsLeft(result))
	_, err := E.UnwrapError(result)
	assert.Contains(
		err.Error(),
		"error granting access to database test-db",
	)
	assert.Contains(err.Error(), "forced error")
}

func TestGetGrantValues(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(driver.GrantReadWrite, GetGrant("rw"))
	assert.Equal(driver.GrantReadOnly, GetGrant("ro"))
	assert.Equal(driver.GrantNone, GetGrant("none"))
	assert.Equal(driver.GrantNone, GetGrant("invalid"))
	assert.Equal(driver.GrantNone, GetGrant(""))
}

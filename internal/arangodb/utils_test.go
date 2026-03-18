package arangodb

import (
	"context"

	driver "github.com/arangodb/go-driver"
)

func getTestClient() (driver.Client, error) {
	return GetClient(&ClientParams{
		Host:     arangoHost,
		Port:     arangoPort,
		User:     "root",
		Pass:     arangoPassword,
		IsSecure: false,
	})
}

// testUser is a minimal driver.User implementation for unit tests.
type testUser struct {
	name                 string
	active               bool
	passwordChangeNeeded bool
}

func (u testUser) Name() string { return u.name }

func (u testUser) IsActive() bool { return u.active }

func (u testUser) IsPasswordChangeNeeded() bool { return u.passwordChangeNeeded }

func (u testUser) Extra(_ interface{}) error { return nil }

func (u testUser) Remove(_ context.Context) error { return nil }

func (u testUser) Update(
	_ context.Context,
	_ driver.UserOptions,
) error {
	return nil
}

func (u testUser) Replace(
	_ context.Context,
	_ driver.UserOptions,
) error {
	return nil
}

func (u testUser) AccessibleDatabases(
	_ context.Context,
) ([]driver.Database, error) {
	return nil, nil
}

func (u testUser) SetDatabaseAccess(
	_ context.Context,
	_ driver.Database,
	_ driver.Grant,
) error {
	return nil
}

func (u testUser) GetDatabaseAccess(
	_ context.Context,
	_ driver.Database,
) (driver.Grant, error) {
	return driver.GrantNone, nil
}

func (u testUser) RemoveDatabaseAccess(
	_ context.Context,
	_ driver.Database,
) error {
	return nil
}

func (u testUser) SetCollectionAccess(
	_ context.Context,
	_ driver.AccessTarget,
	_ driver.Grant,
) error {
	return nil
}

func (u testUser) GetCollectionAccess(
	_ context.Context,
	_ driver.AccessTarget,
) (driver.Grant, error) {
	return driver.GrantNone, nil
}

func (u testUser) RemoveCollectionAccess(
	_ context.Context,
	_ driver.AccessTarget,
) error {
	return nil
}

func (u testUser) GrantReadWriteAccess(
	_ context.Context,
	_ driver.Database,
) error {
	return nil
}

func (u testUser) RevokeAccess(
	_ context.Context,
	_ driver.Database,
) error {
	return nil
}

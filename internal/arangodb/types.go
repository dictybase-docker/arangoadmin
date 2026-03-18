// Package arangodb provides domain logic for managing ArangoDB users, databases, and grants.
package arangodb

import (
	"context"

	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
)

// ConnectionParams contains connection values parsed from CLI flags.
type ConnectionParams struct {
	Host, Port, User, Pass string
	IsSecure               bool
}

// WithConnection enriches ConnectionParams with a driver.Connection
type WithConnection struct {
	ConnectionParams
	Conn driver.Connection
}

// CreateUserResult carries whether the user was newly created and the user itself.
type CreateUserResult = P.Pair[bool, driver.User]

// EnsureUserParams carries inputs for the ensure-user command.
type EnsureUserParams struct {
	Context  context.Context
	Client   driver.Client
	Username string
	Password string
	Policy   string
}

// EnsureUserStatus is an enum representing the outcome of the ensure-user command.
type EnsureUserStatus string

const (
	UserCreated  EnsureUserStatus = "created"
	UserExisting EnsureUserStatus = "existing"
	UserUpdated  EnsureUserStatus = "updated"
)

// EnsureUserResult carries the final outcome status and the driver.User.
type EnsureUserResult = P.Pair[EnsureUserStatus, driver.User]

// EnsureDatabaseParams carries inputs for the ensure-database command.
type EnsureDatabaseParams struct {
	Context  context.Context
	Client   driver.Client
	Database string
}

// EnsureDatabaseResult carries whether the database was newly created and its name.
type EnsureDatabaseResult = P.Pair[bool, string]

// EnsureGrantParams carries inputs for the ensure-grant command.
type EnsureGrantParams struct {
	Context  context.Context
	Client   driver.Client
	Username string
	Database string
	Grant    string
}

// EnsureGrantResult carries database name and applied grant.
type EnsureGrantResult = P.Pair[string, string] // dbname, grant

// GrantState carries parameters and resolved ArangoDB objects for granting access.
type GrantState struct {
	Params EnsureGrantParams
	User   driver.User
	DB     driver.Database
}

// ClientParams is a type alias for backward compat with tests.
type ClientParams = ConnectionParams

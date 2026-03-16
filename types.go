package main

import (
	"log/slog"

	P "github.com/IBM/fp-go/v2/pair"
	driver "github.com/arangodb/go-driver"
	"github.com/urfave/cli/v3"
)

// Base connection params (parsed from CLI flags)
type ConnectionParams struct {
	Host, Port, User, Pass string
	IsSecure               bool
}

// WithConnection enriches ConnectionParams with a driver.Connection
type WithConnection struct {
	ConnectionParams
	Conn driver.Connection
}

// Enriched with ArangoDB client + logger
type WithClient struct {
	ConnectionParams
	Client driver.Client
	Logger *slog.Logger
}

// --- create-user / update-user ---
type UserParams struct {
	WithClient
	Username, Password string
}

// CreateUserParams is a narrower input for the create-user pipeline.
type CreateUserParams struct {
	Client             driver.Client
	Username, Password string
}

// CreateUserResult carries whether the user was newly created and the user itself.
type CreateUserResult = P.Pair[bool, driver.User]

// CreateUserRouteParams carries user existence state alongside create-user params.
type CreateUserRouteParams = P.Pair[bool, CreateUserParams]

// --- create-database ---
type DatabaseParams struct {
	WithClient
	Databases          []string
	Username, Password string // optional user creation
	Grant              string
}

// SingleDBParams carries what's needed to create one database
type SingleDBParams struct {
	Client driver.Client
	Logger *slog.Logger
	Dbname string
}

// GrantDBParams carries what's needed to grant a user access to one database
type GrantDBParams struct {
	Client driver.Client
	Logger *slog.Logger
	Dbname string
	Grant  string
	User   driver.User
}

// CreateSingleDBResult carries whether a database was newly created and its name.
type CreateSingleDBResult = P.Pair[bool, string]

// CreateGrantResult carries database name and applied grant.
type CreateGrantResult = P.Pair[string, string]

// CreateDatabaseResult carries outcomes for database creation, optional user creation,
// and grants applied during create-database command execution.
type CreateDatabaseResult struct {
	Databases []CreateSingleDBResult
	User      *CreateUserResult
	Grants    []CreateGrantResult
}

// Pure helper to extract connection params from CLI
func connParamsFromCmd(cmd *cli.Command) ConnectionParams {
	return ConnectionParams{
		Host:     cmd.String("host"),
		Port:     cmd.String("port"),
		User:     cmd.String("admin-user"),
		Pass:     cmd.String("admin-password"),
		IsSecure: cmd.Bool("is-secure"),
	}
}

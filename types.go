package main

import (
	"log/slog"

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

// --- create-database ---
type DatabaseParams struct {
	WithClient
	Dbname             string
	Username, Password string // optional user creation
	Grant              string
}

// Intermediate type for create-database grant pipeline
type UserWithGrant struct {
	Params DatabaseParams
	User   driver.User
}

// Pure helper to extract connection params from CLI
// nolint:unused // Used in Phase 1-3 implementation
func connParamsFromCmd(cmd *cli.Command) ConnectionParams {
	return ConnectionParams{
		Host:     cmd.String("host"),
		Port:     cmd.String("port"),
		User:     cmd.String("admin-user"),
		Pass:     cmd.String("admin-password"),
		IsSecure: cmd.Bool("is-secure"),
	}
}

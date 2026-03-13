// Package main provides the arangoadmin CLI for creating databases and users in ArangoDB.
package main

import (
	"crypto/tls"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	fperrors "github.com/IBM/fp-go/v2/errors"
	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// createConnection builds the HTTP connection
func createConnection(p ConnectionParams) IOE.IOEither[error, driver.Connection] {
	return IOE.TryCatchError(func() (driver.Connection, error) {
		scheme := "http"
		if p.IsSecure {
			scheme = "https"
		}
		cfg := http.ConnectionConfig{
			Endpoints: []string{fmt.Sprintf("%s://%s:%s", scheme, p.Host, p.Port)},
		}
		if p.IsSecure {
			cfg.TLSConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
		}
		return http.NewConnection(cfg)
	})
}

// newClientFromConn creates a driver client from a connection (extracted helper — hookify: no nested closures)
func newClientFromConn(p ConnectionParams) func(driver.Connection) IOE.IOEither[error, driver.Client] {
	return func(conn driver.Connection) IOE.IOEither[error, driver.Client] {
		return F.Pipe1(
			IOE.TryCatchError(func() (driver.Client, error) {
				return driver.NewClient(driver.ClientConfig{
					Connection:     conn,
					Authentication: driver.BasicAuthentication(p.User, p.Pass),
				})
			}),
			IOE.MapLeft[driver.Client, error, error](fperrors.OnError("could not create client")),
		)
	}
}

// createArangoClient is the main pipeline: connection → driver client (flat, no nested closures)
func createArangoClient(p ConnectionParams) IOE.IOEither[error, driver.Client] {
	return F.Pipe2(
		createConnection(p),
		IOE.MapLeft[driver.Connection, error, error](fperrors.OnError("could not connect")),
		IOE.Chain(newClientFromConn(p)),
	)
}

// Backward compat for tests
func getClient(p *ClientParams) (driver.Client, error) {
	return toTuple(createArangoClient(*p))
}

// ClientParams is a type alias for backward compat
type ClientParams = ConnectionParams

// Shared utilities
func toEither[ER, A any](ioe IOE.IOEither[ER, A]) E.Either[ER, A] {
	return ioe()
}

// toTuple converts IOEither to tuple
func toTuple[A any](ioe IOE.IOEither[error, A]) (A, error) {
	return E.UnwrapError(ioe())
}

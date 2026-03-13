// Package main provides the arangoadmin CLI for creating databases and users in ArangoDB.
package main

import (
	"crypto/tls"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// createConnection builds the HTTP connection
func createConnection(
	p ConnectionParams,
) IOE.IOEither[error, driver.Connection] {
	return IOE.TryCatchError(func() (driver.Connection, error) {
		return http.NewConnection(http.ConnectionConfig{
			Endpoints: []string{fmt.Sprintf("%s://%s:%s",
				F.Pipe1(p.IsSecure, F.Ternary(
					F.Identity[bool],
					func(_ bool) string { return "https" },
					func(_ bool) string { return "http" },
				)),
				p.Host, p.Port)},
			TLSConfig: F.Pipe1(p.IsSecure, F.Ternary(
				F.Identity[bool],
				func(_ bool) *tls.Config { return &tls.Config{InsecureSkipVerify: true} }, // #nosec G402
				func(_ bool) *tls.Config { return nil },
			)),
		})
	})
}

// newClientFromConn creates a driver client from a connection
func newClientFromConn(wc WithConnection) IOE.IOEither[error, driver.Client] {
	return F.Pipe1(
		IOE.TryCatchError(func() (driver.Client, error) {
			return driver.NewClient(driver.ClientConfig{
				Connection:     wc.Conn,
				Authentication: driver.BasicAuthentication(wc.User, wc.Pass),
			})
		}),
		IOE.MapLeft[driver.Client](
			fperrors.OnError("could not create client"),
		),
	)
}

// createArangoClient is the main pipeline: connection → driver client
func createArangoClient(p ConnectionParams) IOE.IOEither[error, driver.Client] {
	return F.Pipe4(
		p,
		createConnection,
		IOE.MapLeft[driver.Connection](fperrors.OnError("could not connect")),
		IOE.Map[error](func(conn driver.Connection) WithConnection {
			return WithConnection{ConnectionParams: p, Conn: conn}
		}),
		IOE.Chain(newClientFromConn),
	)
}

// getClient wraps createArangoClient for backward compat with tests
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

package arangodb

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

// CreateConnection builds the HTTP connection.
func CreateConnection(
	params ConnectionParams,
) IOE.IOEither[error, driver.Connection] {
	return IOE.TryCatchError(func() (driver.Connection, error) {
		return http.NewConnection(http.ConnectionConfig{
			Endpoints: []string{fmt.Sprintf("%s://%s:%s",
				F.Pipe1(params.IsSecure, F.Ternary(
					F.Identity[bool],
					func(_ bool) string { return "https" },
					func(_ bool) string { return "http" },
				)),
				params.Host, params.Port)},
			TLSConfig: F.Pipe1(params.IsSecure, F.Ternary(
				F.Identity[bool],
				func(_ bool) *tls.Config { return &tls.Config{InsecureSkipVerify: true} }, // #nosec G402
				func(_ bool) *tls.Config { return nil },
			)),
		})
	})
}

// newClientFromConn creates a driver client from a connection.
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

// CreateArangoClient is the main pipeline: connection → driver client.
func CreateArangoClient(
	params ConnectionParams,
) IOE.IOEither[error, driver.Client] {
	return F.Pipe4(
		params,
		CreateConnection,
		IOE.MapLeft[driver.Connection](
			fperrors.OnError("could not connect"),
		),
		IOE.Map[error](func(conn driver.Connection) WithConnection {
			return WithConnection{
				ConnectionParams: params,
				Conn:             conn,
			}
		}),
		IOE.Chain(newClientFromConn),
	)
}

// GetClient wraps CreateArangoClient for use in tests.
func GetClient(params *ClientParams) (driver.Client, error) {
	return ToTuple(CreateArangoClient(*params))
}

// ToEither executes an IOEither and returns the Either result.
func ToEither[ER, A any](ioe IOE.IOEither[ER, A]) E.Either[ER, A] {
	return ioe()
}

// ToTuple converts IOEither to a (value, error) tuple.
func ToTuple[A any](ioe IOE.IOEither[error, A]) (A, error) {
	return E.UnwrapError(ioe())
}

// FoldIOE executes an IOEither and folds the result into an error (nil on success).
func FoldIOE[A any](ma IOE.IOEither[error, A]) error {
	return F.Pipe2(
		ma,
		ToEither[error, A],
		E.Fold(
			F.Identity[error],
			func(_ A) error { return nil },
		),
	)
}

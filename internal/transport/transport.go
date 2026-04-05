// Package transport provides shared HTTP types used by all OBIE service packages.
// It exists to break the import cycle: service packages import transport (not obie),
// while obie imports both transport and all service packages without creating a cycle.
package transport

import (
	"context"
	"net/http"
)

// DoOptions carries optional per-request HTTP settings.
type DoOptions struct {
	// IdempotencyKey is sent as x-idempotency-key. Required for all PIS POST requests.
	IdempotencyKey string
	// JWSSignature is a detached JWS sent as x-jws-signature. Required for
	// payment initiation and VRP POST requests.
	JWSSignature string
	// ExtraHeaders are additional HTTP headers to inject for this request.
	ExtraHeaders map[string]string
}

// HTTPDoer is the exported HTTP capability interface implemented by obie.Client
// and accepted by every service constructor. Using an exported interface in a
// shared internal package is the only way to allow cross-package interface
// satisfaction in Go (unexported interface methods cannot be satisfied from
// outside the package that declares them).
//
// Note: raw HTTP (file upload/download) uses a separate rawHTTPDoer interface
// local to the filepayments package to avoid requiring Do() from all services.
type HTTPDoer interface {
	// Get performs a GET request, JSON-decoding the response body into out.
	Get(ctx context.Context, url string, out any) error
	// Post performs a POST request with a JSON body, decoding the response into out.
	Post(ctx context.Context, url string, body, out any, opts DoOptions) error
	// Put performs a PUT request with a JSON body, decoding the response into out.
	Put(ctx context.Context, url string, body, out any, opts DoOptions) error
	// Delete performs a DELETE request.
	Delete(ctx context.Context, url string) error
}

// RawDoer is the interface for raw (non-JSON) HTTP calls such as file upload
// and download. Implemented by obie.Client; used only by filepayments.
type RawDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

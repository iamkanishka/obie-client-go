// Package apierror defines the APIError type returned by all OBIE HTTP calls.
// It lives in an internal package so service packages (filepayments, etc.) can
// reference it without importing the obie package and creating an import cycle.
package apierror

import (
	"fmt"
	"net/http"
)

// APIError represents an error response from the OBIE API.
type APIError struct {
	// StatusCode is the HTTP status code returned by the ASPSP.
	StatusCode int

	// Body is the raw response body (may be JSON or plain text).
	Body string

	// InteractionID is the x-fapi-interaction-id echoed from the response.
	InteractionID string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("obie: API error %d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("obie: API error %d: %s", e.StatusCode, http.StatusText(e.StatusCode))
}

// IsStatus returns true if the error has the given HTTP status code.
func (e *APIError) IsStatus(code int) bool {
	return e.StatusCode == code
}

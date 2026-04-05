package obie

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iamkanishka/obie-client-go/models"
)

// APIError represents an error returned by the OBIE API.
type APIError struct {
	StatusCode int
	OBError    *models.OBError
	Body       string
	// InteractionID is the x-fapi-interaction-id echoed from the response, if present.
	InteractionID string
}

func (e *APIError) Error() string {
	if e.OBError != nil {
		return fmt.Sprintf("obie: API error %d: %s", e.StatusCode, models.OBErrorSummary(e.OBError))
	}
	return fmt.Sprintf("obie: API error %d: %s", e.StatusCode, e.Body)
}

// IsErrorCode returns true if this APIError wraps an OBIE error detail with
// the given typed error code.
func (e *APIError) IsErrorCode(code models.OBIEErrorCode) bool {
	return models.IsOBIEErrorCode(e.OBError, code)
}

// parseAPIError attempts to decode an OBIE error body from resp.
func parseAPIError(resp *http.Response, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode:    resp.StatusCode,
		Body:          string(body),
		InteractionID: resp.Header.Get("x-fapi-interaction-id"),
	}
	var obErr models.OBError
	if err := json.Unmarshal(body, &obErr); err == nil && obErr.Code != "" {
		apiErr.OBError = &obErr
	}
	return apiErr
}

// ErrTokenExpired is returned when the cached token has expired and refresh failed.
var ErrTokenExpired = fmt.Errorf("obie: access token expired")

// ErrInvalidConfig is returned when the configuration is incomplete or invalid.
type ErrInvalidConfig struct {
	Field   string
	Message string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("obie: invalid config field %q: %s", e.Field, e.Message)
}

// ErrSigningFailed is returned when JWS/JWT signing fails.
type ErrSigningFailed struct {
	Cause error
}

func (e *ErrSigningFailed) Error() string {
	return fmt.Sprintf("obie: signing failed: %v", e.Cause)
}

func (e *ErrSigningFailed) Unwrap() error { return e.Cause }

package obie

import (
	"net/http"
	"testing"
)

func TestAPIError_Error_WithOBError(t *testing.T) {
	resp := &http.Response{StatusCode: 400}
	body := []byte(`{"Code":"UK.OBIE.Field.Missing","Message":"field is required","Errors":[]}`)

	apiErr := parseAPIError(resp, body)
	if apiErr == nil {
		t.Fatal("expected non-nil APIError")
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode: got %d, want 400", apiErr.StatusCode)
	}
	if apiErr.OBError == nil {
		t.Fatal("expected OBError to be populated")
	}
	if apiErr.OBError.Code != "UK.OBIE.Field.Missing" {
		t.Errorf("OBError.Code: got %q, want %q", apiErr.OBError.Code, "UK.OBIE.Field.Missing")
	}

	msg := apiErr.Error()
	if msg == "" {
		t.Error("Error() returned empty string")
	}
}

func TestAPIError_Error_NonJSONBody(t *testing.T) {
	resp := &http.Response{StatusCode: 500}
	body := []byte("Internal Server Error")

	apiErr := parseAPIError(resp, body)
	if apiErr.OBError != nil {
		t.Error("expected OBError to be nil for non-JSON body")
	}
	if apiErr.Body != "Internal Server Error" {
		t.Errorf("Body: got %q, want %q", apiErr.Body, "Internal Server Error")
	}
}

func TestErrInvalidConfig_Error(t *testing.T) {
	err := ErrInvalidConfig{Field: "ClientID", Message: "must not be empty"}
	want := `obie: invalid config field "ClientID": must not be empty`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestErrSigningFailed_Unwrap(t *testing.T) {
	inner := ErrTokenExpired
	err := &ErrSigningFailed{Cause: inner}
	if err.Unwrap() != inner {
		t.Error("Unwrap should return the inner error")
	}
}

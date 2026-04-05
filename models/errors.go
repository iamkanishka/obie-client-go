package models

import "fmt"

// ────────────────────────────────────────────────────────────────────────────
// OBIE Error response helpers
// ────────────────────────────────────────────────────────────────────────────

// IsOBIEErrorCode checks whether an OBError contains a specific OBIE error code
// in any of its detail entries.
func IsOBIEErrorCode(err *OBError, code OBIEErrorCode) bool {
	if err == nil {
		return false
	}
	for _, detail := range err.Errors {
		if detail.ErrorCode == code {
			return true
		}
	}
	return false
}

// OBErrorSummary returns a human-readable summary of the error and all its details.
func OBErrorSummary(err *OBError) string {
	if err == nil {
		return ""
	}
	if len(err.Errors) == 0 {
		return fmt.Sprintf("[%s] %s", err.Code, err.Message)
	}
	s := fmt.Sprintf("[%s] %s", err.Code, err.Message)
	for _, d := range err.Errors {
		s += fmt.Sprintf(" | %s: %s", d.ErrorCode, d.Message)
		if d.Path != "" {
			s += fmt.Sprintf(" (path: %s)", d.Path)
		}
	}
	return s
}

// NewOBError constructs a minimal OBError with a single detail entry.
func NewOBError(code OBIEErrorCode, message, path string) *OBError {
	return &OBError{
		Code:    string(code),
		Message: message,
		Errors: []OBErrorDetail{
			{ErrorCode: code, Message: message, Path: path},
		},
	}
}

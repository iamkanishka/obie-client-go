package events

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// EventNotification is the top-level structure of an OBIE webhook event.
type EventNotification struct {
	Iss   string                 `json:"iss"`
	Iat   int64                  `json:"iat"`
	Jti   string                 `json:"jti"`
	Aud   string                 `json:"aud"`
	Sub   string                 `json:"sub"`
	Txn   string                 `json:"txn"`
	Toe   int64                  `json:"toe"`
	Events map[string]any `json:"events"`
}

// Handler provides helpers for receiving and verifying OBIE webhook events.
type Handler struct {
	// PublicKey is the ASPSP's public key used to verify JWS signatures on events.
	// When nil, signature verification is skipped (not recommended for production).
	PublicKey *rsa.PublicKey
}

// NewHandler creates a Handler with the supplied ASPSP public key.
func NewHandler(pubKey *rsa.PublicKey) *Handler {
	return &Handler{PublicKey: pubKey}
}

// ParseRequest reads an HTTP request body and parses it as an OBIE event notification.
// If the Handler has a PublicKey set, the x-jws-signature header is verified.
func (h *Handler) ParseRequest(r *http.Request) (*EventNotification, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("events: read body: %w", err)
	}

	if h.PublicKey != nil {
		jwsSig := r.Header.Get("x-jws-signature")
		if jwsSig == "" {
			return nil, fmt.Errorf("events: missing x-jws-signature header")
		}
		if err := h.verifySignature(jwsSig, body); err != nil {
			return nil, fmt.Errorf("events: invalid signature: %w", err)
		}
	}

	var notification EventNotification
	if err := json.Unmarshal(body, &notification); err != nil {
		return nil, fmt.Errorf("events: decode notification: %w", err)
	}
	return &notification, nil
}

// verifySignature reconstructs the detached JWS and verifies the RS256 signature.
func (h *Handler) verifySignature(jwsSig string, payload []byte) error {
	// Detached JWS format: <header>..<signature>
	// Reconstruct as: <header>.<base64url(payload)>.<signature>
	parts := strings.Split(jwsSig, "..")
	if len(parts) != 2 {
		return fmt.Errorf("invalid detached JWS format")
	}
	header := parts[0]
	signature := parts[1]

	// Re-attach the payload to form a verifiable compact JWS.
	// Per RFC 7797, the payload is base64url-encoded when b64=true (default).
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	compact := header + "." + encodedPayload + "." + signature
	_, err := jwt.Parse(compact, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.PublicKey, nil
	})
	return err
}

// EventType constants for well-known OBIE event types.
const (
	EventTypeResourceUpdate            = "urn:uk:org:openbanking:events:resource-update"
	EventTypeConsentAuthorizationRevoked = "urn:uk:org:openbanking:events:consent-authorization-revoked"
	EventTypeAccountAccessConsentLinkedAccountUpdate = "urn:uk:org:openbanking:events:account-access-consent-linked-account-update"
)

// ResourceUpdateEvent carries the standard resource-update event payload.
type ResourceUpdateEvent struct {
	Subject ResourceUpdateSubject `json:"subject"`
}

type ResourceUpdateSubject struct {
	SubjectType    string            `json:"subject_type"`
	HTTPStatusCode int               `json:"http_status_code"`
	Links          map[string]string `json:"links"`
	Version        string            `json:"version"`
}

// ExtractResourceUpdate pulls a typed ResourceUpdateEvent from an EventNotification.
func ExtractResourceUpdate(n *EventNotification) (*ResourceUpdateEvent, error) {
	raw, ok := n.Events[EventTypeResourceUpdate]
	if !ok {
		return nil, fmt.Errorf("events: resource-update event not present")
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var ev ResourceUpdateEvent
	if err := json.Unmarshal(b, &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

// HTTPHandlerFunc returns a standard http.HandlerFunc that parses and dispatches
// OBIE webhook events to the provided callback.
func (h *Handler) HTTPHandlerFunc(callback func(*EventNotification) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notification, err := h.ParseRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := callback(notification); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

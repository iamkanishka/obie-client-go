// Package consent provides a state machine that models the OBIE consent
// lifecycle and helpers for the authorisation-code (PKCE) redirect flow.
//
// OBIE consent lifecycle:
//
//	Created → AwaitingAuthorisation → Authorised → [Consumed | Rejected | Revoked]
//	                                             ↓
//	                                          Expired
package consent

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// State machine
// ────────────────────────────────────────────────────────────────────────────

// State represents a consent lifecycle state.
type State string

const (
	StateAwaitingAuthorisation State = "AwaitingAuthorisation"
	StateAuthorised            State = "Authorised"
	StateConsumed              State = "Consumed"
	StateRejected              State = "Rejected"
	StateRevoked               State = "Revoked"
)

// Event triggers a state transition.
type Event string

const (
	EventAuthorise Event = "Authorise"
	EventConsume   Event = "Consume"
	EventReject    Event = "Reject"
	EventRevoke    Event = "Revoke"
)

// ErrInvalidTransition is returned when an event is not valid for the current state.
type ErrInvalidTransition struct {
	From  State
	Event Event
}

func (e *ErrInvalidTransition) Error() string {
	return fmt.Sprintf("consent: invalid transition: %s -[%s]-> (no valid target state)", e.From, e.Event)
}

// transitions maps (currentState, event) → nextState.
var transitions = map[State]map[Event]State{
	StateAwaitingAuthorisation: {
		EventAuthorise: StateAuthorised,
		EventReject:    StateRejected,
	},
	StateAuthorised: {
		EventConsume: StateConsumed,
		EventRevoke:  StateRevoked,
	},
}

// Machine tracks the state of a single consent object.
type Machine struct {
	ConsentID  string
	State      State
	History    []Transition
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Transition records a single state change.
type Transition struct {
	From      State
	To        State
	Event     Event
	Timestamp time.Time
}

// NewMachine creates a Machine starting in AwaitingAuthorisation.
func NewMachine(consentID string) *Machine {
	return &Machine{
		ConsentID: consentID,
		State:     StateAwaitingAuthorisation,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Apply attempts to apply event to the machine. Returns ErrInvalidTransition
// if the event is not valid for the current state.
func (m *Machine) Apply(event Event) error {
	targets, ok := transitions[m.State]
	if !ok {
		return &ErrInvalidTransition{From: m.State, Event: event}
	}
	next, ok := targets[event]
	if !ok {
		return &ErrInvalidTransition{From: m.State, Event: event}
	}

	m.History = append(m.History, Transition{
		From:      m.State,
		To:        next,
		Event:     event,
		Timestamp: time.Now(),
	})
	m.State = next
	m.UpdatedAt = time.Now()
	return nil
}

// IsTerminal returns true when the consent is in a terminal state from which
// no further transitions are possible.
func (m *Machine) IsTerminal() bool {
	_, hasTargets := transitions[m.State]
	return !hasTargets
}

// CanTransition reports whether event is valid from the current state.
func (m *Machine) CanTransition(event Event) bool {
	targets, ok := transitions[m.State]
	if !ok {
		return false
	}
	_, ok = targets[event]
	return ok
}

// SyncFromASPSP updates the machine's state to match the status string returned
// by the ASPSP API, recording the synthetic event in history.
func (m *Machine) SyncFromASPSP(aspspStatus string) error {
	s := State(aspspStatus)
	if s == m.State {
		return nil // no change
	}

	// Determine which event caused this transition.
	var event Event
	switch s {
	case StateAuthorised:
		event = EventAuthorise
	case StateConsumed:
		event = EventConsume
	case StateRejected:
		event = EventReject
	case StateRevoked:
		event = EventRevoke
	default:
		return fmt.Errorf("consent: unknown ASPSP status %q", aspspStatus)
	}

	return m.Apply(event)
}

// ────────────────────────────────────────────────────────────────────────────
// PKCE (Proof Key for Code Exchange) helpers
// ────────────────────────────────────────────────────────────────────────────

// PKCEPair holds a PKCE code verifier and its derived challenge.
type PKCEPair struct {
	// Verifier is the random secret sent with the token exchange request.
	Verifier string
	// Challenge is the SHA-256 hash of Verifier, base64url-encoded.
	// Send this in the authorisation redirect URL.
	Challenge string
	// Method is always "S256".
	Method string
}

// GeneratePKCE generates a cryptographically random PKCE pair.
func GeneratePKCE() (*PKCEPair, error) {
	b := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, fmt.Errorf("consent: generate PKCE verifier: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])
	return &PKCEPair{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Authorisation URL builder
// ────────────────────────────────────────────────────────────────────────────

// AuthURLParams carries the parameters needed to build an OBIE authorisation URL.
type AuthURLParams struct {
	// AuthorisationEndpoint is the ASPSP's /authorize URL.
	AuthorisationEndpoint string
	// ClientID is the registered OAuth2 client ID.
	ClientID string
	// RedirectURI is the registered callback URL.
	RedirectURI string
	// ConsentID is the ID returned by the consent creation endpoint.
	ConsentID string
	// Scope is the OAuth2 scope string (e.g. "openid accounts").
	Scope string
	// State is a random string to protect against CSRF.
	State string
	// PKCE holds the code challenge. When nil, PKCE is omitted.
	PKCE *PKCEPair
	// Nonce is a random string embedded in the id_token to prevent replay attacks.
	Nonce string
	// ResponseType defaults to "code id_token".
	ResponseType string
}

// BuildAuthURL constructs the redirect URL a TPP must send the PSU to in order
// to authorise the consent.
func BuildAuthURL(p AuthURLParams) (string, error) {
	if p.AuthorisationEndpoint == "" {
		return "", fmt.Errorf("consent: AuthorisationEndpoint is required")
	}
	if p.ClientID == "" {
		return "", fmt.Errorf("consent: ClientID is required")
	}
	if p.ConsentID == "" {
		return "", fmt.Errorf("consent: ConsentID is required")
	}

	responseType := p.ResponseType
	if responseType == "" {
		responseType = "code id_token"
	}

	q := url.Values{}
	q.Set("response_type", responseType)
	q.Set("client_id", p.ClientID)
	q.Set("scope", p.Scope)
	q.Set("redirect_uri", p.RedirectURI)
	q.Set("state", p.State)
	q.Set("nonce", p.Nonce)
	// OBIE: intent ID is passed as request parameter or request JWT claim.
	// Simple implementation: pass as claims parameter.
	claims := map[string]any{
		"id_token": map[string]any{
			"acr": map[string]any{
				"essential": true,
				"values":    []string{"urn:openbanking:psd2:sca"},
			},
			"openbanking_intent_id": map[string]any{
				"value":     p.ConsentID,
				"essential": true,
			},
		},
		"userinfo": map[string]any{
			"openbanking_intent_id": map[string]any{
				"value":     p.ConsentID,
				"essential": true,
			},
		},
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("consent: marshal claims: %w", err)
	}
	q.Set("claims", string(claimsJSON))

	if p.PKCE != nil {
		q.Set("code_challenge", p.PKCE.Challenge)
		q.Set("code_challenge_method", p.PKCE.Method)
	}

	return p.AuthorisationEndpoint + "?" + q.Encode(), nil
}

// ────────────────────────────────────────────────────────────────────────────
// Token exchange (authorisation code → access token)
// ────────────────────────────────────────────────────────────────────────────

// TokenExchangeRequest carries parameters for exchanging an auth code.
type TokenExchangeRequest struct {
	TokenEndpoint string
	ClientID      string
	RedirectURI   string
	Code          string
	// PKCEVerifier is required when PKCE was used in the redirect.
	PKCEVerifier string
	// ClientAssertion is a signed JWT for private_key_jwt auth.
	// When set, client_secret is not required.
	ClientAssertion string
}

// TokenResponse holds the token endpoint response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// ExchangeCode exchanges an authorisation code for tokens.
func ExchangeCode(ctx context.Context, client *http.Client, req TokenExchangeRequest) (*TokenResponse, error) {
	if client == nil {
		client = http.DefaultClient
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", req.Code)
	form.Set("redirect_uri", req.RedirectURI)
	form.Set("client_id", req.ClientID)

	if req.PKCEVerifier != "" {
		form.Set("code_verifier", req.PKCEVerifier)
	}
	if req.ClientAssertion != "" {
		form.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		form.Set("client_assertion", req.ClientAssertion)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.TokenEndpoint,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("consent: build token exchange request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("consent: token exchange request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("consent: read token exchange response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consent: token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("consent: decode token response: %w", err)
	}
	return &tr, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Consent poller
// ────────────────────────────────────────────────────────────────────────────

// StatusFetcher retrieves the latest consent status from the ASPSP.
type StatusFetcher func(ctx context.Context, consentID string) (string, error)

// PollUntilAuthorised polls the consent status until it reaches Authorised or
// a terminal failure state, or ctx is cancelled.
func PollUntilAuthorised(ctx context.Context, machine *Machine, fetch StatusFetcher, interval time.Duration) error {
	if interval <= 0 {
		interval = 3 * time.Second
	}
	for {
		status, err := fetch(ctx, machine.ConsentID)
		if err != nil {
			return fmt.Errorf("consent: poll status: %w", err)
		}

		if err := machine.SyncFromASPSP(status); err != nil {
			return err
		}

		switch machine.State {
		case StateAuthorised:
			return nil
		case StateRejected, StateRevoked:
			return fmt.Errorf("consent: reached terminal state %s", machine.State)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("consent: polling cancelled: %w", ctx.Err())
		case <-time.After(interval):
		}
	}
}

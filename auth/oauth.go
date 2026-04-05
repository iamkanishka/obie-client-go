package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// assertionTTL is the lifetime used when building client assertion JWTs.
	assertionTTL = 5 * time.Minute
	// tokenExpiryBuffer is how early (before actual expiry) we consider a token stale.
	tokenExpiryBuffer = 30 * time.Second
)

// tokenResponse models the JSON body returned by an OAuth2 token endpoint.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// TokenManager handles OAuth2 client-credentials token lifecycle including
// automatic refresh using JWT client assertion (private_key_jwt).
type TokenManager struct {
	mu         sync.Mutex
	httpClient *http.Client

	tokenURL  string
	clientID  string
	keyID     string
	scopes    []string
	keyPEM    []byte

	cachedToken  string
	tokenExpiry  time.Time
}

// TokenManagerConfig carries parameters for TokenManager construction.
type TokenManagerConfig struct {
	// TokenURL is the full URL of the OAuth2 token endpoint.
	TokenURL string
	// ClientID is the registered client identifier.
	ClientID string
	// KeyID is the kid value embedded in JWT headers.
	KeyID string
	// PrivateKeyPEM is the PEM-encoded RSA private key.
	PrivateKeyPEM []byte
	// Scopes is the list of OAuth2 scopes to request.
	Scopes []string
	// HTTPClient is used for token endpoint requests.
	// When nil, http.DefaultClient is used.
	HTTPClient *http.Client
}

// NewTokenManager constructs a TokenManager from the given config.
func NewTokenManager(cfg TokenManagerConfig) (*TokenManager, error) {
	if cfg.TokenURL == "" {
		return nil, fmt.Errorf("auth: TokenURL is required")
	}
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("auth: ClientID is required")
	}
	hc := cfg.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	return &TokenManager{
		httpClient: hc,
		tokenURL:   cfg.TokenURL,
		clientID:   cfg.ClientID,
		keyID:      cfg.KeyID,
		scopes:     cfg.Scopes,
		keyPEM:     cfg.PrivateKeyPEM,
	}, nil
}

// AccessToken returns a valid access token, fetching a new one if necessary.
// It is safe to call concurrently from multiple goroutines.
func (m *TokenManager) AccessToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cachedToken != "" && time.Now().Before(m.tokenExpiry.Add(-tokenExpiryBuffer)) {
		return m.cachedToken, nil
	}

	tok, expiry, err := m.fetchToken(ctx)
	if err != nil {
		return "", err
	}

	m.cachedToken = tok
	m.tokenExpiry = expiry
	return tok, nil
}

// Invalidate clears the cached token, forcing a fresh fetch on the next call.
func (m *TokenManager) Invalidate() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cachedToken = ""
	m.tokenExpiry = time.Time{}
}

func (m *TokenManager) fetchToken(ctx context.Context) (string, time.Time, error) {
	assertion, err := m.buildAssertion()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: build client assertion: %w", err)
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	form.Set("client_assertion", assertion)
	if len(m.scopes) > 0 {
		form.Set("scope", strings.Join(m.scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("auth: token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", time.Time{}, fmt.Errorf("auth: decode token response: %w", err)
	}

	if tr.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("auth: empty access_token in response")
	}

	expiry := time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	return tr.AccessToken, expiry, nil
}

func (m *TokenManager) buildAssertion() (string, error) {
	key, err := ParseRSAPrivateKeyFromPEM(m.keyPEM)
	if err != nil {
		return "", err
	}
	return BuildClientAssertion(m.clientID, m.tokenURL, m.keyID, key, assertionTTL)
}

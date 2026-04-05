package dcr

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// RegistrationRequest models the JWT claims sent to a DCR endpoint.
// Per OBIE DCR spec (based on RFC 7591 + FAPI).
type RegistrationRequest struct {
	// JWT standard claims.
	Issuer    string `json:"iss"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	JTI       string `json:"jti"`

	// DCR-specific claims.
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	SoftwareID              string   `json:"software_id"`
	Scope                   string   `json:"scope"`
	SoftwareStatement       string   `json:"software_statement"` // SSA from directory.
	ApplicationType         string   `json:"application_type"`
	IDTokenSignedResponseAlg string  `json:"id_token_signed_response_alg"`
	RequestObjectSigningAlg string   `json:"request_object_signing_alg"`
	TokenEndpointAuthSigningAlg string `json:"token_endpoint_auth_signing_alg"`
	TLSClientAuthSubjectDN  string   `json:"tls_client_auth_subject_dn,omitempty"`
}

// RegistrationResponse is the JSON response returned by the ASPSP's DCR endpoint.
type RegistrationResponse struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret,omitempty"`
	ClientIDIssuedAt        int64    `json:"client_id_issued_at,omitempty"`
	ClientSecretExpiresAt   int64    `json:"client_secret_expires_at,omitempty"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	SoftwareID              string   `json:"software_id"`
	Scope                   string   `json:"scope"`
}

// Client performs Dynamic Client Registration against an ASPSP.
type Client struct {
	httpClient *http.Client
	privateKey *rsa.PrivateKey
	keyID      string
}

// New creates a DCR Client.
func New(privateKey *rsa.PrivateKey, keyID string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		httpClient: httpClient,
		privateKey: privateKey,
		keyID:      keyID,
	}
}

// Register sends a DCR request to the given endpoint URL and returns the registered client details.
// The RegistrationRequest is signed as a JWT (RS256) before transmission.
func (c *Client) Register(ctx context.Context, dcrEndpointURL string, req *RegistrationRequest) (*RegistrationResponse, error) {
	if req.IssuedAt == 0 {
		req.IssuedAt = time.Now().Unix()
	}
	if req.ExpiresAt == 0 {
		req.ExpiresAt = time.Now().Add(5 * time.Minute).Unix()
	}
	if req.JTI == "" {
		req.JTI = uuid.New().String()
	}

	claims := jwt.MapClaims{}
	// Populate claims from the struct via JSON round-trip for simplicity.
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("dcr: marshal request: %w", err)
	}
	if err := json.Unmarshal(b, &claims); err != nil {
		return nil, fmt.Errorf("dcr: unmarshal to claims: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = c.keyID

	signed, err := token.SignedString(c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("dcr: sign registration JWT: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, dcrEndpointURL,
		bytes.NewBufferString(signed))
	if err != nil {
		return nil, fmt.Errorf("dcr: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/jwt")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("dcr: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("dcr: read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dcr: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result RegistrationResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("dcr: decode response: %w", err)
	}
	return &result, nil
}

// Delete removes a dynamic client registration.
func (c *Client) Delete(ctx context.Context, dcrEndpointURL, clientID, accessToken string) error {
	url := fmt.Sprintf("%s/%s", dcrEndpointURL, clientID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("dcr: build delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dcr: delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("dcr: delete unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

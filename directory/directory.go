package directory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// SandboxDirectoryURL is the OBIE sandbox directory base URL.
	SandboxDirectoryURL = "https://keystore.openbanking.org.uk"
	// ProductionDirectoryURL is the OBIE production directory base URL.
	ProductionDirectoryURL = "https://keystore.openbanking.org.uk"
)

// Participant represents an entry in the Open Banking directory.
type Participant struct {
	OrganisationID   string    `json:"OrganisationId"`
	Name             string    `json:"Name"`
	Status           string    `json:"Status"`
	Roles            []string  `json:"Roles"`
	RegistrationDate time.Time `json:"RegistrationDate"`
}

// ParticipantsResponse wraps a list of directory participants.
type ParticipantsResponse struct {
	Participants []Participant `json:"Participants"`
}

// Client interacts with the Open Banking Directory.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// New creates a directory Client targeting the given base URL.
// Pass SandboxDirectoryURL or ProductionDirectoryURL as appropriate.
func New(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{httpClient: httpClient, baseURL: baseURL}
}

// GetParticipants fetches the list of participants from the Open Banking directory.
func (c *Client) GetParticipants(ctx context.Context) (*ParticipantsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/participants", nil)
	if err != nil {
		return nil, fmt.Errorf("directory: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("directory: request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("directory: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("directory: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result ParticipantsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("directory: decode response: %w", err)
	}
	return &result, nil
}

// GetJWKS fetches the JSON Web Key Set for a given organisation and software statement.
func (c *Client) GetJWKS(ctx context.Context, orgID, softwareID string) (map[string]any, error) {
	url := fmt.Sprintf("%s/%s/%s/application.jwks", c.baseURL, orgID, softwareID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("directory: build JWKS request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("directory: JWKS request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("directory: read JWKS: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("directory: JWKS unexpected status %d", resp.StatusCode)
	}

	var jwks map[string]any
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("directory: decode JWKS: %w", err)
	}
	return jwks, nil
}

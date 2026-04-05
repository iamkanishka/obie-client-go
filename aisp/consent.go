// Package aisp implements the OBIE Account Information Service Provider (AISP)
// API endpoints defined in v3.1.3 of the Read/Write Data API specification.
//
// This package covers Account Access Consents (POST/GET/DELETE) which are the
// prerequisite for all AIS resource reads.
package aisp

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

// ConsentService manages OBIE Account Access Consents.
//
// OBIE spec: POST/GET/DELETE /account-access-consents
// Ref: https://openbankinguk.github.io/read-write-api-site2/standards/v3.1.3/resources-and-data-models/aisp/account-access-consents/
type ConsentService struct {
	http    transport.HTTPDoer
	baseURL string
}

// NewConsentService creates a ConsentService.
func NewConsentService(h transport.HTTPDoer, baseURL string) *ConsentService {
	return &ConsentService{http: h, baseURL: baseURL}
}

const aisBasePath = "/open-banking/v3.1/aisp"

// CreateAccountAccessConsent creates a new account-access-consent resource.
//
// The AISP must call this before redirecting the PSU to authorise.
// Returns a ConsentId that is used in the authorisation redirect.
//
// POST /account-access-consents
func (s *ConsentService) CreateAccountAccessConsent(
	ctx context.Context,
	req *models.OBReadConsent1,
) (*models.OBReadConsentResponse1, error) {
	var resp models.OBReadConsentResponse1
	if err := s.http.Post(ctx,
		s.baseURL+aisBasePath+"/account-access-consents",
		req, &resp, transport.DoOptions{}); err != nil {
		return nil, fmt.Errorf("aisp: CreateAccountAccessConsent: %w", err)
	}
	return &resp, nil
}

// GetAccountAccessConsent retrieves the status of an existing account-access-consent.
//
// GET /account-access-consents/{ConsentId}
func (s *ConsentService) GetAccountAccessConsent(
	ctx context.Context,
	consentID string,
) (*models.OBReadConsentResponse1, error) {
	var resp models.OBReadConsentResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s%s/account-access-consents/%s", s.baseURL, aisBasePath, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("aisp: GetAccountAccessConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// DeleteAccountAccessConsent revokes an account-access-consent.
//
// The AISP MUST call this when the PSU revokes consent.
//
// DELETE /account-access-consents/{ConsentId}
func (s *ConsentService) DeleteAccountAccessConsent(
	ctx context.Context,
	consentID string,
) error {
	if err := s.http.Delete(ctx,
		fmt.Sprintf("%s%s/account-access-consents/%s", s.baseURL, aisBasePath, consentID)); err != nil {
		return fmt.Errorf("aisp: DeleteAccountAccessConsent(%s): %w", consentID, err)
	}
	return nil
}

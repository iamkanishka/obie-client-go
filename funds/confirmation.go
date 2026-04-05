package funds

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

// Service exposes OBIE Confirmation of Funds (CBPII) endpoints.
type Service struct {
	http    transport.HTTPDoer
	baseURL string
}

// New creates a funds confirmation Service.
func New(h transport.HTTPDoer, baseURL string) *Service {
	return &Service{http: h, baseURL: baseURL}
}

// CreateConsent creates a funds confirmation consent.
func (s *Service) CreateConsent(
	ctx context.Context,
	req *models.OBFundsConfirmationConsent1,
) (*models.OBFundsConfirmationConsentResponse1, error) {
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String()}
	var resp models.OBFundsConfirmationConsentResponse1
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/cbpii/funds-confirmation-consents",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("funds: CreateConsent: %w", err)
	}
	return &resp, nil
}

// GetConsent retrieves a funds confirmation consent by ID.
func (s *Service) GetConsent(
	ctx context.Context, consentID string,
) (*models.OBFundsConfirmationConsentResponse1, error) {
	var resp models.OBFundsConfirmationConsentResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/cbpii/funds-confirmation-consents/%s", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("funds: GetConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// DeleteConsent revokes a funds confirmation consent.
func (s *Service) DeleteConsent(ctx context.Context, consentID string) error {
	if err := s.http.Delete(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/cbpii/funds-confirmation-consents/%s", s.baseURL, consentID)); err != nil {
		return fmt.Errorf("funds: DeleteConsent(%s): %w", consentID, err)
	}
	return nil
}

// ConfirmFundsAvailability checks whether the PSU account has sufficient funds.
func (s *Service) ConfirmFundsAvailability(
	ctx context.Context,
	req *models.OBFundsConfirmation1,
) (*models.OBFundsConfirmationResponse1, error) {
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String()}
	var resp models.OBFundsConfirmationResponse1
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/cbpii/funds-confirmations",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("funds: ConfirmFundsAvailability: %w", err)
	}
	return &resp, nil
}

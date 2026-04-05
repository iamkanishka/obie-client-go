package vrp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

// Service exposes OBIE Variable Recurring Payments (VRP) endpoints.
type Service struct {
	http    transport.HTTPDoer
	signer  jwsSigner
	baseURL string
}

// jwsSigner signs a payload for JWS header injection.
type jwsSigner interface {
	SignJSON(v any) (string, error)
}

// New creates a VRP Service.
func New(h transport.HTTPDoer, signer jwsSigner, baseURL string) *Service {
	return &Service{http: h, signer: signer, baseURL: baseURL}
}

// CreateConsent creates a VRP consent.
func (s *Service) CreateConsent(
	ctx context.Context,
	req *models.OBDomesticVRPConsentRequest,
) (*models.OBDomesticVRPConsentResponse, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("vrp: sign consent: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBDomesticVRPConsentResponse
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/domestic-vrp-consents",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("vrp: CreateConsent: %w", err)
	}
	return &resp, nil
}

// GetConsent retrieves a VRP consent by ID.
func (s *Service) GetConsent(
	ctx context.Context, consentID string,
) (*models.OBDomesticVRPConsentResponse, error) {
	var resp models.OBDomesticVRPConsentResponse
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-vrp-consents/%s", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("vrp: GetConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// DeleteConsent revokes (deletes) a VRP consent at the ASPSP.
func (s *Service) DeleteConsent(ctx context.Context, consentID string) error {
	if err := s.http.Delete(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-vrp-consents/%s", s.baseURL, consentID)); err != nil {
		return fmt.Errorf("vrp: DeleteConsent(%s): %w", consentID, err)
	}
	return nil
}

// GetConsentFundsConfirmation checks whether sufficient funds are available for a VRP payment.
func (s *Service) GetConsentFundsConfirmation(
	ctx context.Context,
	consentID string,
	req *models.OBVRPFundsConfirmationRequest,
) (*models.OBVRPFundsConfirmationResponse, error) {
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String()}
	var resp models.OBVRPFundsConfirmationResponse
	if err := s.http.Post(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-vrp-consents/%s/funds-confirmation", s.baseURL, consentID),
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("vrp: GetConsentFundsConfirmation(%s): %w", consentID, err)
	}
	return &resp, nil
}

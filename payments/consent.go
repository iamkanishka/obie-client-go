package payments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

// Service exposes the OBIE Payment Initiation Service (PIS) endpoints.
type Service struct {
	http    transport.HTTPDoer
	signer  jwsSigner
	baseURL string
}

// jwsSigner signs a payload for JWS header injection.
type jwsSigner interface {
	SignJSON(v any) (string, error)
}

// New creates a payments Service.
func New(h transport.HTTPDoer, signer jwsSigner, baseURL string) *Service {
	return &Service{http: h, signer: signer, baseURL: baseURL}
}

// ── Domestic Payment Consent ─────────────────────────────────────────────

// CreateDomesticPaymentConsent creates a domestic payment consent.
func (s *Service) CreateDomesticPaymentConsent(
	ctx context.Context,
	req *models.OBWriteDomesticConsent5,
) (*models.OBWriteDomesticConsentResponse5, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign consent request: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteDomesticConsentResponse5
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/domestic-payment-consents",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: CreateDomesticPaymentConsent: %w", err)
	}
	return &resp, nil
}

// GetDomesticPaymentConsent retrieves a domestic payment consent by ID.
func (s *Service) GetDomesticPaymentConsent(
	ctx context.Context, consentID string,
) (*models.OBWriteDomesticConsentResponse5, error) {
	var resp models.OBWriteDomesticConsentResponse5
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-payment-consents/%s", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticPaymentConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// GetDomesticPaymentConsentFundsConfirmation checks whether sufficient funds exist for the consent.
func (s *Service) GetDomesticPaymentConsentFundsConfirmation(
	ctx context.Context, consentID string,
) (*models.OBFundsConfirmationResponse1, error) {
	var resp models.OBFundsConfirmationResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-payment-consents/%s/funds-confirmation", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticPaymentConsentFundsConfirmation(%s): %w", consentID, err)
	}
	return &resp, nil
}

package payments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

// ── Domestic Scheduled Payment ────────────────────────────────────────────

// CreateDomesticScheduledPaymentConsent creates a domestic scheduled payment consent.
func (s *Service) CreateDomesticScheduledPaymentConsent(
	ctx context.Context,
	req *models.OBWriteDomesticScheduledConsent4,
) (*models.OBWriteDomesticScheduledConsentResponse4, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign scheduled consent: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteDomesticScheduledConsentResponse4
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/domestic-scheduled-payment-consents",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: CreateDomesticScheduledPaymentConsent: %w", err)
	}
	return &resp, nil
}

// GetDomesticScheduledPaymentConsent retrieves a domestic scheduled payment consent by ID.
func (s *Service) GetDomesticScheduledPaymentConsent(
	ctx context.Context, consentID string,
) (*models.OBWriteDomesticScheduledConsentResponse4, error) {
	var resp models.OBWriteDomesticScheduledConsentResponse4
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-scheduled-payment-consents/%s", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticScheduledPaymentConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// DeleteDomesticScheduledPaymentConsent deletes an awaiting-authorisation domestic scheduled payment consent.
func (s *Service) DeleteDomesticScheduledPaymentConsent(
	ctx context.Context, consentID string,
) error {
	if err := s.http.Delete(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-scheduled-payment-consents/%s", s.baseURL, consentID)); err != nil {
		return fmt.Errorf("payments: DeleteDomesticScheduledPaymentConsent(%s): %w", consentID, err)
	}
	return nil
}

// SubmitDomesticScheduledPayment submits a domestic scheduled payment.
func (s *Service) SubmitDomesticScheduledPayment(
	ctx context.Context,
	req *models.OBWriteDomesticScheduled3,
) (*models.OBWriteDomesticScheduledResponse5, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign scheduled payment: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteDomesticScheduledResponse5
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/domestic-scheduled-payments",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: SubmitDomesticScheduledPayment: %w", err)
	}
	return &resp, nil
}

// GetDomesticScheduledPayment retrieves a domestic scheduled payment by ID.
func (s *Service) GetDomesticScheduledPayment(
	ctx context.Context, paymentID string,
) (*models.OBWriteDomesticScheduledResponse5, error) {
	var resp models.OBWriteDomesticScheduledResponse5
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-scheduled-payments/%s", s.baseURL, paymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticScheduledPayment(%s): %w", paymentID, err)
	}
	return &resp, nil
}

// GetDomesticScheduledPaymentDetails retrieves detailed status for a domestic scheduled payment.
func (s *Service) GetDomesticScheduledPaymentDetails(
	ctx context.Context, paymentID string,
) (*models.OBWritePaymentDetailsResponse1, error) {
	var resp models.OBWritePaymentDetailsResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-scheduled-payments/%s/payment-details", s.baseURL, paymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticScheduledPaymentDetails(%s): %w", paymentID, err)
	}
	return &resp, nil
}

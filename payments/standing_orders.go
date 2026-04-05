package payments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

// ── Domestic Standing Order ───────────────────────────────────────────────

// CreateDomesticStandingOrderConsent creates a domestic standing order consent.
func (s *Service) CreateDomesticStandingOrderConsent(
	ctx context.Context,
	req *models.OBWriteDomesticStandingOrderConsent5,
) (*models.OBWriteDomesticStandingOrderConsentResponse6, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign standing order consent: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteDomesticStandingOrderConsentResponse6
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/domestic-standing-order-consents",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: CreateDomesticStandingOrderConsent: %w", err)
	}
	return &resp, nil
}

// GetDomesticStandingOrderConsent retrieves a domestic standing order consent by ID.
func (s *Service) GetDomesticStandingOrderConsent(
	ctx context.Context, consentID string,
) (*models.OBWriteDomesticStandingOrderConsentResponse6, error) {
	var resp models.OBWriteDomesticStandingOrderConsentResponse6
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-standing-order-consents/%s", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticStandingOrderConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// SubmitDomesticStandingOrder submits a domestic standing order payment.
func (s *Service) SubmitDomesticStandingOrder(
	ctx context.Context,
	req *models.OBWriteDomesticStandingOrder4,
) (*models.OBWriteDomesticStandingOrderResponse6, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign standing order: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteDomesticStandingOrderResponse6
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/domestic-standing-orders",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: SubmitDomesticStandingOrder: %w", err)
	}
	return &resp, nil
}

// GetDomesticStandingOrder retrieves a domestic standing order by ID.
func (s *Service) GetDomesticStandingOrder(
	ctx context.Context, standingOrderID string,
) (*models.OBWriteDomesticStandingOrderResponse6, error) {
	var resp models.OBWriteDomesticStandingOrderResponse6
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-standing-orders/%s", s.baseURL, standingOrderID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticStandingOrder(%s): %w", standingOrderID, err)
	}
	return &resp, nil
}

// GetDomesticStandingOrderDetails retrieves detailed status for a standing order.
func (s *Service) GetDomesticStandingOrderDetails(
	ctx context.Context, standingOrderID string,
) (*models.OBWritePaymentDetailsResponse1, error) {
	var resp models.OBWritePaymentDetailsResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-standing-orders/%s/payment-details", s.baseURL, standingOrderID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticStandingOrderDetails(%s): %w", standingOrderID, err)
	}
	return &resp, nil
}

// DeleteDomesticStandingOrderConsent deletes a domestic standing order consent.
// This is optional per the spec but recommended for clean lifecycle management.
// Spec: DELETE /domestic-standing-order-consents/{ConsentId}
func (s *Service) DeleteDomesticStandingOrderConsent(ctx context.Context, consentID string) error {
	if err := s.http.Delete(ctx,
		fmt.Sprintf("%s/domestic-standing-order-consents/%s", s.baseURL, consentID)); err != nil {
		return fmt.Errorf("payments: DeleteDomesticStandingOrderConsent(%s): %w", consentID, err)
	}
	return nil
}

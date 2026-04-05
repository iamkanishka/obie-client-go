package payments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

// ── Domestic Payment ──────────────────────────────────────────────────────

// SubmitDomesticPayment submits a domestic payment against an authorised consent.
func (s *Service) SubmitDomesticPayment(
	ctx context.Context,
	req *models.OBWriteDomestic2,
) (*models.OBWriteDomesticResponse5, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign payment request: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteDomesticResponse5
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/domestic-payments",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: SubmitDomesticPayment: %w", err)
	}
	return &resp, nil
}

// GetDomesticPayment retrieves a domestic payment by its payment ID.
func (s *Service) GetDomesticPayment(
	ctx context.Context, paymentID string,
) (*models.OBWriteDomesticResponse5, error) {
	var resp models.OBWriteDomesticResponse5
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-payments/%s", s.baseURL, paymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticPayment(%s): %w", paymentID, err)
	}
	return &resp, nil
}

// GetDomesticPaymentDetails retrieves detailed status information for a domestic payment.
func (s *Service) GetDomesticPaymentDetails(
	ctx context.Context, paymentID string,
) (*models.OBWritePaymentDetailsResponse1, error) {
	var resp models.OBWritePaymentDetailsResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-payments/%s/payment-details", s.baseURL, paymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetDomesticPaymentDetails(%s): %w", paymentID, err)
	}
	return &resp, nil
}

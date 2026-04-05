package payments

import (
	"time"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

// ── International Payment ─────────────────────────────────────────────────

// CreateInternationalPaymentConsent creates an international payment consent.
func (s *Service) CreateInternationalPaymentConsent(
	ctx context.Context,
	req *models.OBWriteInternationalConsent5,
) (*models.OBWriteInternationalConsentResponse6, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign international consent: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteInternationalConsentResponse6
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/international-payment-consents",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: CreateInternationalPaymentConsent: %w", err)
	}
	return &resp, nil
}

// GetInternationalPaymentConsent retrieves an international payment consent by ID.
func (s *Service) GetInternationalPaymentConsent(
	ctx context.Context, consentID string,
) (*models.OBWriteInternationalConsentResponse6, error) {
	var resp models.OBWriteInternationalConsentResponse6
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-payment-consents/%s", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalPaymentConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// GetInternationalPaymentConsentFundsConfirmation checks funds availability for an international payment consent.
func (s *Service) GetInternationalPaymentConsentFundsConfirmation(
	ctx context.Context, consentID string,
) (*models.OBFundsConfirmationResponse1, error) {
	var resp models.OBFundsConfirmationResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-payment-consents/%s/funds-confirmation", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalPaymentConsentFundsConfirmation(%s): %w", consentID, err)
	}
	return &resp, nil
}

// SubmitInternationalPayment submits an international payment.
func (s *Service) SubmitInternationalPayment(
	ctx context.Context,
	req *models.OBWriteInternational3,
) (*models.OBWriteInternationalResponse5, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign international payment: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteInternationalResponse5
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/international-payments",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: SubmitInternationalPayment: %w", err)
	}
	return &resp, nil
}

// GetInternationalPayment retrieves an international payment by its payment ID.
func (s *Service) GetInternationalPayment(
	ctx context.Context, paymentID string,
) (*models.OBWriteInternationalResponse5, error) {
	var resp models.OBWriteInternationalResponse5
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-payments/%s", s.baseURL, paymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalPayment(%s): %w", paymentID, err)
	}
	return &resp, nil
}

// GetInternationalPaymentDetails retrieves detailed status information for an international payment.
func (s *Service) GetInternationalPaymentDetails(
	ctx context.Context, paymentID string,
) (*models.OBWritePaymentDetailsResponse1, error) {
	var resp models.OBWritePaymentDetailsResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-payments/%s/payment-details", s.baseURL, paymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalPaymentDetails(%s): %w", paymentID, err)
	}
	return &resp, nil
}

// ── International Scheduled Payment ──────────────────────────────────────

// CreateInternationalScheduledPaymentConsent creates an international scheduled payment consent.
func (s *Service) CreateInternationalScheduledPaymentConsent(
	ctx context.Context,
	req *models.OBWriteInternationalScheduledConsent5,
) (*models.OBWriteInternationalScheduledConsentResponse6, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign international scheduled consent: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteInternationalScheduledConsentResponse6
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/international-scheduled-payment-consents",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: CreateInternationalScheduledPaymentConsent: %w", err)
	}
	return &resp, nil
}

// GetInternationalScheduledPaymentConsent retrieves an international scheduled payment consent.
func (s *Service) GetInternationalScheduledPaymentConsent(
	ctx context.Context, consentID string,
) (*models.OBWriteInternationalScheduledConsentResponse6, error) {
	var resp models.OBWriteInternationalScheduledConsentResponse6
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-scheduled-payment-consents/%s", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalScheduledPaymentConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// DeleteInternationalScheduledPaymentConsent deletes an international scheduled payment consent.
func (s *Service) DeleteInternationalScheduledPaymentConsent(
	ctx context.Context, consentID string,
) error {
	if err := s.http.Delete(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-scheduled-payment-consents/%s", s.baseURL, consentID)); err != nil {
		return fmt.Errorf("payments: DeleteInternationalScheduledPaymentConsent(%s): %w", consentID, err)
	}
	return nil
}

// SubmitInternationalScheduledPayment submits an international scheduled payment.
func (s *Service) SubmitInternationalScheduledPayment(
	ctx context.Context,
	req *models.OBWriteInternationalScheduled3,
) (*models.OBWriteInternationalScheduledResponse6, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign international scheduled payment: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteInternationalScheduledResponse6
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/international-scheduled-payments",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: SubmitInternationalScheduledPayment: %w", err)
	}
	return &resp, nil
}

// GetInternationalScheduledPayment retrieves an international scheduled payment by ID.
func (s *Service) GetInternationalScheduledPayment(
	ctx context.Context, paymentID string,
) (*models.OBWriteInternationalScheduledResponse6, error) {
	var resp models.OBWriteInternationalScheduledResponse6
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-scheduled-payments/%s", s.baseURL, paymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalScheduledPayment(%s): %w", paymentID, err)
	}
	return &resp, nil
}

// GetInternationalScheduledPaymentDetails retrieves detailed status for an international scheduled payment.
func (s *Service) GetInternationalScheduledPaymentDetails(
	ctx context.Context, paymentID string,
) (*models.OBWritePaymentDetailsResponse1, error) {
	var resp models.OBWritePaymentDetailsResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-scheduled-payments/%s/payment-details", s.baseURL, paymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalScheduledPaymentDetails(%s): %w", paymentID, err)
	}
	return &resp, nil
}

// ── International Standing Order ──────────────────────────────────────────

// CreateInternationalStandingOrderConsent creates an international standing order consent.
func (s *Service) CreateInternationalStandingOrderConsent(
	ctx context.Context,
	req *models.OBWriteInternationalStandingOrderConsent6,
) (*models.OBWriteInternationalStandingOrderConsentResponse7, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign international standing order consent: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteInternationalStandingOrderConsentResponse7
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/international-standing-order-consents",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: CreateInternationalStandingOrderConsent: %w", err)
	}
	return &resp, nil
}

// GetInternationalStandingOrderConsent retrieves an international standing order consent.
func (s *Service) GetInternationalStandingOrderConsent(
	ctx context.Context, consentID string,
) (*models.OBWriteInternationalStandingOrderConsentResponse7, error) {
	var resp models.OBWriteInternationalStandingOrderConsentResponse7
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-standing-order-consents/%s", s.baseURL, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalStandingOrderConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// SubmitInternationalStandingOrder submits an international standing order.
func (s *Service) SubmitInternationalStandingOrder(
	ctx context.Context,
	req *models.OBWriteInternationalStandingOrder6,
) (*models.OBWriteInternationalStandingOrderResponse7, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("payments: sign international standing order: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBWriteInternationalStandingOrderResponse7
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/international-standing-orders",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("payments: SubmitInternationalStandingOrder: %w", err)
	}
	return &resp, nil
}

// GetInternationalStandingOrder retrieves a submitted international standing order by ID.
func (s *Service) GetInternationalStandingOrder(
	ctx context.Context, standingOrderID string,
) (*models.OBWriteInternationalStandingOrderResponse7, error) {
	var resp models.OBWriteInternationalStandingOrderResponse7
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-standing-orders/%s", s.baseURL, standingOrderID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalStandingOrder(%s): %w", standingOrderID, err)
	}
	return &resp, nil
}

// GetInternationalStandingOrderDetails retrieves detailed status for an international standing order.
func (s *Service) GetInternationalStandingOrderDetails(
	ctx context.Context, standingOrderID string,
) (*models.OBWritePaymentDetailsResponse1, error) {
	var resp models.OBWritePaymentDetailsResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/international-standing-orders/%s/payment-details", s.baseURL, standingOrderID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetInternationalStandingOrderDetails(%s): %w", standingOrderID, err)
	}
	return &resp, nil
}

// DeleteInternationalStandingOrderConsent deletes an international standing order consent.
// Spec: DELETE /international-standing-order-consents/{ConsentId}
func (s *Service) DeleteInternationalStandingOrderConsent(ctx context.Context, consentID string) error {
	if err := s.http.Delete(ctx,
		fmt.Sprintf("%s/international-standing-order-consents/%s", s.baseURL, consentID)); err != nil {
		return fmt.Errorf("payments: DeleteInternationalStandingOrderConsent(%s): %w", consentID, err)
	}
	return nil
}

// PollInternationalScheduledPaymentUntilTerminal polls until the international
// scheduled payment reaches a terminal status or the context is cancelled.
// Spec: GET /international-scheduled-payments/{InternationalScheduledPaymentId}
func (s *Service) PollInternationalScheduledPaymentUntilTerminal(
	ctx context.Context, paymentID string, interval time.Duration,
) (*models.OBWriteInternationalScheduledResponse6, error) {
	for {
		resp, err := s.GetInternationalScheduledPayment(ctx, paymentID)
		if err != nil {
			return nil, err
		}
		switch resp.Data.Status {
		case models.PaymentStatusAcceptedCreditSettlementCompleted,
			models.PaymentStatusAcceptedSettlementCompleted,
			models.PaymentStatusRejected:
			return resp, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// PollInternationalStandingOrderUntilTerminal polls until the international
// standing order reaches a terminal status or the context is cancelled.
// Spec: GET /international-standing-orders/{InternationalStandingOrderPaymentId}
func (s *Service) PollInternationalStandingOrderUntilTerminal(
	ctx context.Context, paymentID string, interval time.Duration,
) (*models.OBWriteInternationalStandingOrderResponse7, error) {
	for {
		resp, err := s.GetInternationalStandingOrder(ctx, paymentID)
		if err != nil {
			return nil, err
		}
		switch resp.Data.Status {
		case models.PaymentStatusAcceptedCreditSettlementCompleted,
			models.PaymentStatusAcceptedSettlementCompleted,
			models.PaymentStatusRejected:
			return resp, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}

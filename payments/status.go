package payments

import (
	"context"
	"fmt"
	"time"

	"github.com/iamkanishka/obie-client-go/models"
)

// PaymentType identifies which payment resource type to query.
type PaymentType string

const (
	PaymentTypeDomestic                   PaymentType = "domestic-payments"
	PaymentTypeDomesticScheduled          PaymentType = "domestic-scheduled-payments"
	PaymentTypeDomesticStandingOrder      PaymentType = "domestic-standing-orders"
	PaymentTypeInternational              PaymentType = "international-payments"
	PaymentTypeInternationalScheduled     PaymentType = "international-scheduled-payments"
	PaymentTypeInternationalStandingOrder PaymentType = "international-standing-orders"
)

// GetPaymentStatus fetches the current raw status of any payment type by its ID.
// The returned map is intentionally generic to support all payment variants.
func (s *Service) GetPaymentStatus(
	ctx context.Context,
	paymentType PaymentType,
	paymentID string,
) (map[string]any, error) {
	var resp map[string]any
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/%s/%s", s.baseURL, paymentType, paymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("payments: GetPaymentStatus(%s, %s): %w", paymentType, paymentID, err)
	}
	return resp, nil
}

// ── Terminal-state polling helpers ────────────────────────────────────────

// PollDomesticPaymentUntilTerminal polls a domestic payment until it reaches
// a terminal status or ctx is cancelled.
func (s *Service) PollDomesticPaymentUntilTerminal(
	ctx context.Context, paymentID string, interval time.Duration,
) (*models.OBWriteDomesticResponse5, error) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	for {
		resp, err := s.GetDomesticPayment(ctx, paymentID)
		if err != nil {
			return nil, err
		}
		if isTerminalPaymentStatus(resp.Data.Status) {
			return resp, nil
		}
		if err := sleep(ctx, interval); err != nil {
			return nil, err
		}
	}
}

// PollInternationalPaymentUntilTerminal polls an international payment until terminal.
func (s *Service) PollInternationalPaymentUntilTerminal(
	ctx context.Context, paymentID string, interval time.Duration,
) (*models.OBWriteInternationalResponse5, error) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	for {
		resp, err := s.GetInternationalPayment(ctx, paymentID)
		if err != nil {
			return nil, err
		}
		if isTerminalPaymentStatus(resp.Data.Status) {
			return resp, nil
		}
		if err := sleep(ctx, interval); err != nil {
			return nil, err
		}
	}
}

// PollDomesticScheduledPaymentUntilTerminal polls a domestic scheduled payment until terminal.
func (s *Service) PollDomesticScheduledPaymentUntilTerminal(
	ctx context.Context, paymentID string, interval time.Duration,
) (*models.OBWriteDomesticScheduledResponse5, error) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	for {
		resp, err := s.GetDomesticScheduledPayment(ctx, paymentID)
		if err != nil {
			return nil, err
		}
		if isTerminalPaymentStatus(resp.Data.Status) {
			return resp, nil
		}
		if err := sleep(ctx, interval); err != nil {
			return nil, err
		}
	}
}

// PollDomesticStandingOrderUntilTerminal polls a domestic standing order until terminal.
func (s *Service) PollDomesticStandingOrderUntilTerminal(
	ctx context.Context, standingOrderID string, interval time.Duration,
) (*models.OBWriteDomesticStandingOrderResponse6, error) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	for {
		resp, err := s.GetDomesticStandingOrder(ctx, standingOrderID)
		if err != nil {
			return nil, err
		}
		if isTerminalPaymentStatus(resp.Data.Status) {
			return resp, nil
		}
		if err := sleep(ctx, interval); err != nil {
			return nil, err
		}
	}
}

// isTerminalPaymentStatus returns true when status is a final, non-retryable state.
func isTerminalPaymentStatus(status models.PaymentStatus) bool {
	switch status {
	case models.PaymentStatusAcceptedCreditSettlementCompleted,
		models.PaymentStatusAcceptedSettlementCompleted,
		models.PaymentStatusRejected:
		return true
	}
	return false
}

// sleep blocks for d, returning ctx.Err() if the context is cancelled first.
func sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("payments: polling cancelled: %w", ctx.Err())
	case <-time.After(d):
		return nil
	}
}

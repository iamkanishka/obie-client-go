package vrp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

// SubmitPayment submits a VRP payment under an authorised consent.
func (s *Service) SubmitPayment(
	ctx context.Context,
	req *models.OBDomesticVRPRequest,
) (*models.OBDomesticVRPResponse, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("vrp: sign payment: %w", err)
	}
	opts := transport.DoOptions{IdempotencyKey: uuid.New().String(), JWSSignature: sig}
	var resp models.OBDomesticVRPResponse
	if err := s.http.Post(ctx,
		s.baseURL+"/open-banking/v3.1/pisp/domestic-vrps",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("vrp: SubmitPayment: %w", err)
	}
	return &resp, nil
}

// GetPayment retrieves a VRP payment by its ID.
func (s *Service) GetPayment(
	ctx context.Context, vrpID string,
) (*models.OBDomesticVRPResponse, error) {
	var resp models.OBDomesticVRPResponse
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-vrps/%s", s.baseURL, vrpID),
		&resp); err != nil {
		return nil, fmt.Errorf("vrp: GetPayment(%s): %w", vrpID, err)
	}
	return &resp, nil
}

// GetPaymentDetails retrieves detailed multi-step status information for a VRP payment.
func (s *Service) GetPaymentDetails(
	ctx context.Context, vrpID string,
) (*models.OBWritePaymentDetailsResponse1, error) {
	var resp models.OBWritePaymentDetailsResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/open-banking/v3.1/pisp/domestic-vrps/%s/payment-details", s.baseURL, vrpID),
		&resp); err != nil {
		return nil, fmt.Errorf("vrp: GetPaymentDetails(%s): %w", vrpID, err)
	}
	return &resp, nil
}

// PollPaymentUntilTerminal polls a VRP payment until it reaches a terminal
// status or ctx is cancelled.
func (s *Service) PollPaymentUntilTerminal(
	ctx context.Context, vrpID string, interval time.Duration,
) (*models.OBDomesticVRPResponse, error) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	for {
		resp, err := s.GetPayment(ctx, vrpID)
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
			return nil, fmt.Errorf("vrp: polling cancelled: %w", ctx.Err())
		case <-time.After(interval):
		}
	}
}

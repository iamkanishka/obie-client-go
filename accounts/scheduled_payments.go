package accounts

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/models"
)

// GetScheduledPayments returns all scheduled payments across all accounts.
func (s *Service) GetScheduledPayments(ctx context.Context) (*models.GetScheduledPaymentsResponse, error) {
	var resp models.GetScheduledPaymentsResponse
	if err := s.http.Get(ctx, s.baseURL+"/open-banking/v3.1/aisp/scheduled-payments", &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetScheduledPayments: %w", err)
	}
	return &resp, nil
}

// GetAccountScheduledPayments returns scheduled payments for a specific account.
func (s *Service) GetAccountScheduledPayments(ctx context.Context, accountID string) (*models.GetScheduledPaymentsResponse, error) {
	var resp models.GetScheduledPaymentsResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/scheduled-payments", s.baseURL, accountID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccountScheduledPayments(%s): %w", accountID, err)
	}
	return &resp, nil
}

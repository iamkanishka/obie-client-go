package accounts

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/models"
)

// GetDirectDebits returns all direct debits across all accounts.
func (s *Service) GetDirectDebits(ctx context.Context) (*models.GetDirectDebitsResponse, error) {
	var resp models.GetDirectDebitsResponse
	if err := s.http.Get(ctx, s.baseURL+"/open-banking/v3.1/aisp/direct-debits", &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetDirectDebits: %w", err)
	}
	return &resp, nil
}

// GetAccountDirectDebits returns direct debits for a specific account.
func (s *Service) GetAccountDirectDebits(ctx context.Context, accountID string) (*models.GetDirectDebitsResponse, error) {
	var resp models.GetDirectDebitsResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/direct-debits", s.baseURL, accountID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccountDirectDebits(%s): %w", accountID, err)
	}
	return &resp, nil
}

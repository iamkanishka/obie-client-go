package accounts

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/models"
)

// GetBalances returns balances for all accounts under the current consent.
func (s *Service) GetBalances(ctx context.Context) (*models.GetBalancesResponse, error) {
	var resp models.GetBalancesResponse
	if err := s.http.Get(ctx, s.baseURL+"/open-banking/v3.1/aisp/balances", &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetBalances: %w", err)
	}
	return &resp, nil
}

// GetAccountBalances returns balances for a specific account.
func (s *Service) GetAccountBalances(ctx context.Context, accountID string) (*models.GetBalancesResponse, error) {
	var resp models.GetBalancesResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/balances", s.baseURL, accountID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccountBalances(%s): %w", accountID, err)
	}
	return &resp, nil
}

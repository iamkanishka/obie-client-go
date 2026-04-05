package accounts

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/models"
)

// GetParty returns the party associated with the authorised PSU.
func (s *Service) GetParty(ctx context.Context) (*models.GetPartiesResponse, error) {
	var resp models.GetPartiesResponse
	if err := s.http.Get(ctx, s.baseURL+"/open-banking/v3.1/aisp/party", &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetParty: %w", err)
	}
	return &resp, nil
}

// GetAccountParty returns the party associated with a specific account.
func (s *Service) GetAccountParty(ctx context.Context, accountID string) (*models.GetPartiesResponse, error) {
	var resp models.GetPartiesResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/party", s.baseURL, accountID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccountParty(%s): %w", accountID, err)
	}
	return &resp, nil
}

// GetAccountParties returns all parties associated with a specific account.
func (s *Service) GetAccountParties(ctx context.Context, accountID string) (*models.GetPartiesResponse, error) {
	var resp models.GetPartiesResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/parties", s.baseURL, accountID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccountParties(%s): %w", accountID, err)
	}
	return &resp, nil
}

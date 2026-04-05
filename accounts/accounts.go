package accounts

import (
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/models"
)

// Service exposes the OBIE Account Information Service (AIS) endpoints.
type Service struct {
	http   transport.HTTPDoer
	baseURL string
}

// New creates an accounts Service.
func New(h transport.HTTPDoer, baseURL string) *Service {
	return &Service{http: h, baseURL: baseURL}
}

// GetAccounts returns all accounts accessible under the current consent.
func (s *Service) GetAccounts(ctx context.Context) (*models.GetAccountsResponse, error) {
	var resp models.GetAccountsResponse
	if err := s.http.Get(ctx, s.baseURL+"/open-banking/v3.1/aisp/accounts", &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccounts: %w", err)
	}
	return &resp, nil
}

// GetAccount returns a single account by ID.
func (s *Service) GetAccount(ctx context.Context, accountID string) (*models.GetAccountResponse, error) {
	var resp models.GetAccountResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s", s.baseURL, accountID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccount(%s): %w", accountID, err)
	}
	return &resp, nil
}

package accounts

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/models"
)

// GetBeneficiaries returns beneficiaries across all accounts.
func (s *Service) GetBeneficiaries(ctx context.Context) (*models.GetBeneficiariesResponse, error) {
	var resp models.GetBeneficiariesResponse
	if err := s.http.Get(ctx, s.baseURL+"/open-banking/v3.1/aisp/beneficiaries", &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetBeneficiaries: %w", err)
	}
	return &resp, nil
}

// GetAccountBeneficiaries returns beneficiaries for a specific account.
func (s *Service) GetAccountBeneficiaries(ctx context.Context, accountID string) (*models.GetBeneficiariesResponse, error) {
	var resp models.GetBeneficiariesResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/beneficiaries", s.baseURL, accountID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccountBeneficiaries(%s): %w", accountID, err)
	}
	return &resp, nil
}

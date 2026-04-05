package accounts

import (
	"context"
	"fmt"
)

// ProductsResponse is a generic container for OBIE product responses.
// The product schema is complex and bank-specific; we use a raw map for
// maximum compatibility and let callers unmarshal further if needed.
type ProductsResponse struct {
	Data  ProductsData  `json:"Data"`
	Links any   `json:"Links"`
	Meta  any   `json:"Meta"`
}

type ProductsData struct {
	Product []map[string]any `json:"Product"`
}

// GetProducts returns products across all accounts.
func (s *Service) GetProducts(ctx context.Context) (*ProductsResponse, error) {
	var resp ProductsResponse
	if err := s.http.Get(ctx, s.baseURL+"/open-banking/v3.1/aisp/products", &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetProducts: %w", err)
	}
	return &resp, nil
}

// GetAccountProducts returns products for a specific account.
func (s *Service) GetAccountProducts(ctx context.Context, accountID string) (*ProductsResponse, error) {
	var resp ProductsResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/product", s.baseURL, accountID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccountProducts(%s): %w", accountID, err)
	}
	return &resp, nil
}

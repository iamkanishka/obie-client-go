package aisp

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/models"
)

// GetStandingOrders returns all standing orders across authorised accounts.
//
// GET /standing-orders
func (s *ConsentService) GetStandingOrders(ctx context.Context) (*models.OBReadStandingOrder6, error) {
	var resp models.OBReadStandingOrder6
	if err := s.http.Get(ctx, s.baseURL+aisBasePath+"/standing-orders", &resp); err != nil {
		return nil, fmt.Errorf("aisp: GetStandingOrders: %w", err)
	}
	return &resp, nil
}

// GetAccountStandingOrders returns standing orders for a specific account.
//
// GET /accounts/{AccountId}/standing-orders
func (s *ConsentService) GetAccountStandingOrders(ctx context.Context, accountID string) (*models.OBReadStandingOrder6, error) {
	var resp models.OBReadStandingOrder6
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s%s/accounts/%s/standing-orders", s.baseURL, aisBasePath, accountID),
		&resp); err != nil {
		return nil, fmt.Errorf("aisp: GetAccountStandingOrders(%s): %w", accountID, err)
	}
	return &resp, nil
}

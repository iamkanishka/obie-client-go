package aisp

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/models"
)

// GetOffers returns offers across all authorised accounts (bulk).
//
// GET /offers
func (s *ConsentService) GetOffers(ctx context.Context) (*models.OBReadOffer1, error) {
	var resp models.OBReadOffer1
	if err := s.http.Get(ctx, s.baseURL+aisBasePath+"/offers", &resp); err != nil {
		return nil, fmt.Errorf("aisp: GetOffers: %w", err)
	}
	return &resp, nil
}

// GetAccountOffers returns offers for a specific account.
//
// GET /accounts/{AccountId}/offers
func (s *ConsentService) GetAccountOffers(ctx context.Context, accountID string) (*models.OBReadOffer1, error) {
	var resp models.OBReadOffer1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s%s/accounts/%s/offers", s.baseURL, aisBasePath, accountID),
		&resp); err != nil {
		return nil, fmt.Errorf("aisp: GetAccountOffers(%s): %w", accountID, err)
	}
	return &resp, nil
}

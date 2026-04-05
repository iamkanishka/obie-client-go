package accounts

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/iamkanishka/obie-client-go/models"
)

// TransactionFilter carries optional query parameters for transaction requests.
type TransactionFilter struct {
	FromBookingDateTime *time.Time
	ToBookingDateTime   *time.Time
}

func (f TransactionFilter) toQuery() string {
	q := url.Values{}
	if f.FromBookingDateTime != nil {
		q.Set("fromBookingDateTime", f.FromBookingDateTime.Format(time.RFC3339))
	}
	if f.ToBookingDateTime != nil {
		q.Set("toBookingDateTime", f.ToBookingDateTime.Format(time.RFC3339))
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

// GetTransactions returns transactions across all accounts under the current consent.
func (s *Service) GetTransactions(ctx context.Context, filter TransactionFilter) (*models.GetTransactionsResponse, error) {
	var resp models.GetTransactionsResponse
	rawURL := s.baseURL + "/open-banking/v3.1/aisp/transactions" + filter.toQuery()
	if err := s.http.Get(ctx, rawURL, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetTransactions: %w", err)
	}
	return &resp, nil
}

// GetAccountTransactions returns transactions for a specific account.
func (s *Service) GetAccountTransactions(ctx context.Context, accountID string, filter TransactionFilter) (*models.GetTransactionsResponse, error) {
	var resp models.GetTransactionsResponse
	rawURL := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/transactions%s",
		s.baseURL, accountID, filter.toQuery())
	if err := s.http.Get(ctx, rawURL, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccountTransactions(%s): %w", accountID, err)
	}
	return &resp, nil
}

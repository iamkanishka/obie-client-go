package accounts

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/models"
)

// GetStatements returns all statements across all accounts.
func (s *Service) GetStatements(ctx context.Context) (*models.GetStatementsResponse, error) {
	var resp models.GetStatementsResponse
	if err := s.http.Get(ctx, s.baseURL+"/open-banking/v3.1/aisp/statements", &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetStatements: %w", err)
	}
	return &resp, nil
}

// GetAccountStatements returns statements for a specific account.
func (s *Service) GetAccountStatements(ctx context.Context, accountID string) (*models.GetStatementsResponse, error) {
	var resp models.GetStatementsResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/statements", s.baseURL, accountID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetAccountStatements(%s): %w", accountID, err)
	}
	return &resp, nil
}

// GetStatement returns a specific statement by ID.
func (s *Service) GetStatement(ctx context.Context, accountID, statementID string) (*models.GetStatementsResponse, error) {
	var resp models.GetStatementsResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/statements/%s", s.baseURL, accountID, statementID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetStatement(%s, %s): %w", accountID, statementID, err)
	}
	return &resp, nil
}

// GetStatementTransactions returns transactions for a specific statement.
func (s *Service) GetStatementTransactions(ctx context.Context, accountID, statementID string) (*models.GetTransactionsResponse, error) {
	var resp models.GetTransactionsResponse
	url := fmt.Sprintf("%s/open-banking/v3.1/aisp/accounts/%s/statements/%s/transactions",
		s.baseURL, accountID, statementID)
	if err := s.http.Get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("accounts: GetStatementTransactions(%s, %s): %w", accountID, statementID, err)
	}
	return &resp, nil
}

// GetStatementTransactionsBulk retrieves all transactions associated with a
// specific statement across all accounts (no AccountId in the path).
// Spec: GET /statements/{StatementId}/transactions
func (s *Service) GetStatementTransactionsBulk(ctx context.Context, statementID string) (*models.GetTransactionsResponse, error) {
	var resp models.GetTransactionsResponse
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s/statements/%s/transactions", s.baseURL, statementID),
		&resp); err != nil {
		return nil, fmt.Errorf("accounts: GetStatementTransactionsBulk(%s): %w", statementID, err)
	}
	return &resp, nil
}

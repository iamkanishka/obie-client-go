package accounts_test

import (
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"strings"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/accounts"
	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/obie"
)

// ─── helpers shared across both account test files ───────────────────────
// (accounts_test.go already defines testDoer and newAccountsService)

func jsonH(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(v); err != nil {
  	http.Error(w, err.Error(), http.StatusInternalServerError)
  }
	}
}

type fullDoer struct{ client *http.Client }

func (d *fullDoer) Get(ctx context.Context, url string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return &obie.APIError{StatusCode: resp.StatusCode}
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func newFullSvc(t *testing.T, mux *http.ServeMux) (*accounts.Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return accounts.New(&fullDoer{client: srv.Client()}, srv.URL), srv
}

// ─── Balances ────────────────────────────────────────────────────────────

func TestGetBalances(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/balances", jsonH(models.GetBalancesResponse{
		Data: models.GetBalancesData{Balance: []models.OBBalance1{
			{AccountId: "acc-1", CreditDebitIndicator: "Credit", Type: "InterimAvailable",
				Amount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "1500.00", Currency: "GBP"},
				DateTime: time.Now()},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetBalances(context.Background())
	if err != nil {
		t.Fatalf("GetBalances: %v", err)
	}
	if len(resp.Data.Balance) != 1 {
		t.Errorf("balance count: got %d, want 1", len(resp.Data.Balance))
	}
	if resp.Data.Balance[0].Amount.Amount != "1500.00" {
		t.Errorf("Amount: got %q", resp.Data.Balance[0].Amount.Amount)
	}
}

func TestGetAccountBalances(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/balances", jsonH(models.GetBalancesResponse{
		Data: models.GetBalancesData{Balance: []models.OBBalance1{
			{AccountId: "acc-1", CreditDebitIndicator: "Credit", Type: "InterimBooked",
				Amount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "800.00", Currency: "GBP"},
				DateTime: time.Now()},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetAccountBalances(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountBalances: %v", err)
	}
	if len(resp.Data.Balance) != 1 {
		t.Errorf("balance count: got %d", len(resp.Data.Balance))
	}
}

// ─── Transactions ────────────────────────────────────────────────────────

func TestGetTransactions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/transactions", jsonH(models.GetTransactionsResponse{
		Data: models.GetTransactionsData{Transaction: []models.OBTransaction6{
			{AccountId: "acc-1", TransactionId: "tx-1", Status: "Booked",
				CreditDebitIndicator: "Credit",
				Amount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "50.00", Currency: "GBP"},
				BookingDateTime: time.Now()},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetTransactions(context.Background(), accounts.TransactionFilter{})
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(resp.Data.Transaction) != 1 {
		t.Errorf("transaction count: got %d, want 1", len(resp.Data.Transaction))
	}
}

func TestGetTransactionsWithFilter(t *testing.T) {
	var gotURL string
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/transactions", func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.RawQuery
  if err := json.NewEncoder(w).Encode(models.GetTransactionsResponse{}); err != nil {
  	http.Error(w, err.Error(), http.StatusInternalServerError)
  }
	})
	svc, _ := newFullSvc(t, mux)
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	svc.GetTransactions(context.Background(), accounts.TransactionFilter{ //nolint:errcheck
		FromBookingDateTime: &from,
		ToBookingDateTime:   &to,
	})
	if gotURL == "" {
		t.Error("expected query string with date filter params")
	}
}

func TestGetAccountTransactions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/transactions", jsonH(models.GetTransactionsResponse{
		Data: models.GetTransactionsData{Transaction: []models.OBTransaction6{
			{AccountId: "acc-1", TransactionId: "tx-2", Status: "Booked",
				CreditDebitIndicator: "Debit",
				Amount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "20.00", Currency: "GBP"},
				BookingDateTime: time.Now()},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetAccountTransactions(context.Background(), "acc-1", accounts.TransactionFilter{})
	if err != nil {
		t.Fatalf("GetAccountTransactions: %v", err)
	}
	if resp.Data.Transaction[0].TransactionId != "tx-2" {
		t.Errorf("TransactionId: got %q", resp.Data.Transaction[0].TransactionId)
	}
}

// ─── Beneficiaries ───────────────────────────────────────────────────────

func TestGetBeneficiaries(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/beneficiaries", jsonH(models.GetBeneficiariesResponse{
		Data: models.GetBeneficiariesData{Beneficiary: []models.OBBeneficiary5{
			{AccountId: "acc-1", BeneficiaryId: "ben-1",
				CreditorAccount: &models.OBCashAccount3{SchemeName: "UK.OBIE.SortCodeAccountNumber", Identification: "20000319825731", Name: "Beneficiary One"}},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetBeneficiaries(context.Background())
	if err != nil {
		t.Fatalf("GetBeneficiaries: %v", err)
	}
	if len(resp.Data.Beneficiary) != 1 {
		t.Errorf("beneficiary count: got %d, want 1", len(resp.Data.Beneficiary))
	}
}

func TestGetAccountBeneficiaries(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/beneficiaries", jsonH(models.GetBeneficiariesResponse{
		Data: models.GetBeneficiariesData{Beneficiary: []models.OBBeneficiary5{
			{BeneficiaryId: "ben-2"},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetAccountBeneficiaries(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountBeneficiaries: %v", err)
	}
	if resp.Data.Beneficiary[0].BeneficiaryId != "ben-2" {
		t.Errorf("BeneficiaryId: got %q", resp.Data.Beneficiary[0].BeneficiaryId)
	}
}

// ─── Direct Debits ───────────────────────────────────────────────────────

func TestGetDirectDebits(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/direct-debits", jsonH(models.GetDirectDebitsResponse{
		Data: models.GetDirectDebitsData{DirectDebit: []models.OBDirectDebit2{
			{AccountId: "acc-1", DirectDebitId: "dd-1", MandateIdentification: "MANDATE-001", Name: "Broadband DD"},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetDirectDebits(context.Background())
	if err != nil {
		t.Fatalf("GetDirectDebits: %v", err)
	}
	if len(resp.Data.DirectDebit) != 1 {
		t.Errorf("direct debit count: got %d, want 1", len(resp.Data.DirectDebit))
	}
	if resp.Data.DirectDebit[0].Name != "Broadband DD" {
		t.Errorf("Name: got %q", resp.Data.DirectDebit[0].Name)
	}
}

func TestGetAccountDirectDebits(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/direct-debits", jsonH(models.GetDirectDebitsResponse{
		Data: models.GetDirectDebitsData{DirectDebit: []models.OBDirectDebit2{
			{AccountId: "acc-1", DirectDebitId: "dd-2", MandateIdentification: "MANDATE-002", Name: "Energy DD"},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetAccountDirectDebits(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountDirectDebits: %v", err)
	}
	if resp.Data.DirectDebit[0].DirectDebitId != "dd-2" {
		t.Errorf("DirectDebitId: got %q", resp.Data.DirectDebit[0].DirectDebitId)
	}
}

// ─── Scheduled Payments ──────────────────────────────────────────────────

func TestGetScheduledPayments(t *testing.T) {
	mux := http.NewServeMux()
	execTime := time.Now().Add(7 * 24 * time.Hour)
	mux.HandleFunc("/open-banking/v3.1/aisp/scheduled-payments", jsonH(models.GetScheduledPaymentsResponse{
		Data: models.GetScheduledPaymentsData{ScheduledPayment: []models.OBScheduledPayment3{
			{AccountId: "acc-1", ScheduledPaymentId: "sp-1", ScheduledType: "Execution",
				ScheduledPaymentDateTime: execTime,
				InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "200.00", Currency: "GBP"},
				CreditorAccount: &models.OBCashAccount3{SchemeName: "UK.OBIE.SortCodeAccountNumber", Identification: "20000319825731"}},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetScheduledPayments(context.Background())
	if err != nil {
		t.Fatalf("GetScheduledPayments: %v", err)
	}
	if resp.Data.ScheduledPayment[0].ScheduledPaymentId != "sp-1" {
		t.Errorf("ScheduledPaymentId: got %q", resp.Data.ScheduledPayment[0].ScheduledPaymentId)
	}
}

func TestGetAccountScheduledPayments(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/scheduled-payments", jsonH(models.GetScheduledPaymentsResponse{
		Data: models.GetScheduledPaymentsData{ScheduledPayment: []models.OBScheduledPayment3{
			{AccountId: "acc-1", ScheduledPaymentId: "sp-2", ScheduledType: "Arrival",
				ScheduledPaymentDateTime: time.Now().Add(48 * time.Hour),
				InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "75.00", Currency: "GBP"}},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetAccountScheduledPayments(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountScheduledPayments: %v", err)
	}
	if resp.Data.ScheduledPayment[0].ScheduledPaymentId != "sp-2" {
		t.Errorf("ScheduledPaymentId: got %q", resp.Data.ScheduledPayment[0].ScheduledPaymentId)
	}
}

// ─── Statements ───────────────────────────────────────────────────────────

func TestGetStatements(t *testing.T) {
	mux := http.NewServeMux()
	now := time.Now()
	mux.HandleFunc("/open-banking/v3.1/aisp/statements", jsonH(models.GetStatementsResponse{
		Data: models.GetStatementsData{Statement: []models.OBStatement2{
			{AccountId: "acc-1", StatementId: "stmt-1", Type: "RegularPeriodic",
				StartDateTime: now.AddDate(0, -1, 0), EndDateTime: now, CreationDateTime: now},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetStatements(context.Background())
	if err != nil {
		t.Fatalf("GetStatements: %v", err)
	}
	if resp.Data.Statement[0].StatementId != "stmt-1" {
		t.Errorf("StatementId: got %q", resp.Data.Statement[0].StatementId)
	}
}

func TestGetAccountStatements(t *testing.T) {
	mux := http.NewServeMux()
	now := time.Now()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/statements", jsonH(models.GetStatementsResponse{
		Data: models.GetStatementsData{Statement: []models.OBStatement2{
			{AccountId: "acc-1", StatementId: "stmt-2", Type: "RegularPeriodic",
				StartDateTime: now.AddDate(0, -2, 0), EndDateTime: now.AddDate(0, -1, 0), CreationDateTime: now},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetAccountStatements(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountStatements: %v", err)
	}
	if resp.Data.Statement[0].StatementId != "stmt-2" {
		t.Errorf("StatementId: got %q", resp.Data.Statement[0].StatementId)
	}
}

func TestGetStatement(t *testing.T) {
	mux := http.NewServeMux()
	now := time.Now()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/statements/stmt-1", jsonH(models.GetStatementsResponse{
		Data: models.GetStatementsData{Statement: []models.OBStatement2{
			{AccountId: "acc-1", StatementId: "stmt-1", Type: "RegularPeriodic",
				StartDateTime: now.AddDate(0, -1, 0), EndDateTime: now, CreationDateTime: now,
				StatementAmount: []models.OBStatementAmount1{
					{Amount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "3400.00", Currency: "GBP"},
						CreditDebitIndicator: "Credit", Type: "ClosingAvailable"},
				}},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetStatement(context.Background(), "acc-1", "stmt-1")
	if err != nil {
		t.Fatalf("GetStatement: %v", err)
	}
	if len(resp.Data.Statement[0].StatementAmount) == 0 {
		t.Error("expected StatementAmount to be populated")
	}
}

func TestGetStatementTransactions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/statements/stmt-1/transactions",
		jsonH(models.GetTransactionsResponse{
			Data: models.GetTransactionsData{Transaction: []models.OBTransaction6{
				{AccountId: "acc-1", TransactionId: "tx-stmt-1", Status: "Booked",
					CreditDebitIndicator: "Debit",
					Amount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "12.99", Currency: "GBP"},
					BookingDateTime: time.Now()},
			}},
		}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetStatementTransactions(context.Background(), "acc-1", "stmt-1")
	if err != nil {
		t.Fatalf("GetStatementTransactions: %v", err)
	}
	if resp.Data.Transaction[0].TransactionId != "tx-stmt-1" {
		t.Errorf("TransactionId: got %q", resp.Data.Transaction[0].TransactionId)
	}
}

// ─── Parties ─────────────────────────────────────────────────────────────

func TestGetParty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/party", jsonH(models.GetPartiesResponse{
		Data: models.GetPartiesData{Party: []models.OBParty2{
			{PartyId: "party-1", Name: "John Doe", PartyType: "Individual"},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetParty(context.Background())
	if err != nil {
		t.Fatalf("GetParty: %v", err)
	}
	if resp.Data.Party[0].PartyId != "party-1" {
		t.Errorf("PartyId: got %q", resp.Data.Party[0].PartyId)
	}
}

func TestGetAccountParty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/party", jsonH(models.GetPartiesResponse{
		Data: models.GetPartiesData{Party: []models.OBParty2{
			{PartyId: "party-1", Name: "John Doe", AccountRole: "Principal"},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetAccountParty(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountParty: %v", err)
	}
	if resp.Data.Party[0].AccountRole != "Principal" {
		t.Errorf("AccountRole: got %q", resp.Data.Party[0].AccountRole)
	}
}

func TestGetAccountParties(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/parties", jsonH(models.GetPartiesResponse{
		Data: models.GetPartiesData{Party: []models.OBParty2{
			{PartyId: "party-1", Name: "John Doe", AccountRole: "Principal"},
			{PartyId: "party-2", Name: "Jane Doe", AccountRole: "SecondaryOwner"},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetAccountParties(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountParties: %v", err)
	}
	if len(resp.Data.Party) != 2 {
		t.Errorf("party count: got %d, want 2", len(resp.Data.Party))
	}
}

// ─── Products ────────────────────────────────────────────────────────────

func TestGetProducts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/products", jsonH(accounts.ProductsResponse{
		Data: accounts.ProductsData{Product: []map[string]any{
			{"ProductType": "PersonalCurrentAccount", "ProductId": "prod-1", "AccountId": "acc-1"},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetProducts(context.Background())
	if err != nil {
		t.Fatalf("GetProducts: %v", err)
	}
	if len(resp.Data.Product) != 1 {
		t.Errorf("product count: got %d, want 1", len(resp.Data.Product))
	}
}

func TestGetAccountProducts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/product", jsonH(accounts.ProductsResponse{
		Data: accounts.ProductsData{Product: []map[string]any{
			{"ProductType": "BusinessCurrentAccount", "ProductId": "prod-2", "AccountId": "acc-1",
				"MarketingStateId": "M001"},
		}},
	}))
	svc, _ := newFullSvc(t, mux)
	resp, err := svc.GetAccountProducts(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountProducts: %v", err)
	}
	if resp.Data.Product[0]["ProductType"] != "BusinessCurrentAccount" {
		t.Errorf("ProductType: got %v", resp.Data.Product[0]["ProductType"])
	}
}

// ─── Error propagation ────────────────────────────────────────────────────

func TestGetAccounts_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"Code":    "UK.OBIE.Unexpected.Error",
			"Message": "Internal server error",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	svc, _ := newFullSvc(t, mux)
	_, err := svc.GetAccounts(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	apiErr, ok := err.(*obie.APIError)
	if !ok {
		t.Fatalf("expected *obie.APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode: got %d, want 500", apiErr.StatusCode)
	}
}

func (d *fullDoer) Delete(ctx context.Context, url string) error {
	return nil
}

func (d *fullDoer) Post(ctx context.Context, url string, body, out any, _ transport.DoOptions) error {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func (d *fullDoer) Put(ctx context.Context, url string, body, out any, _ transport.DoOptions) error {
	return nil
}

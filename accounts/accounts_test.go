package accounts_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iamkanishka/obie-client-go/internal/transport"

	"github.com/iamkanishka/obie-client-go/accounts"
	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/obie"
)

func newAccountsService(t *testing.T, handler http.Handler) (*accounts.Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	// Use a real httptest server so URLs resolve correctly.
	doer := &testDoer{client: srv.Client()}
	return accounts.New(doer, srv.URL), srv
}

// testDoer uses an httptest.Server's client to make real HTTP calls to it.
type testDoer struct {
	client *http.Client
}

func (d *testDoer) Get(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer test-token")
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

func TestGetAccounts_ReturnsAccounts(t *testing.T) {
	expected := models.GetAccountsResponse{
		Data: models.GetAccountsData{
			Account: []models.OBAccount6{
				{AccountId: "acc-1", Currency: "GBP", AccountType: "Personal", AccountSubType: "CurrentAccount"},
				{AccountId: "acc-2", Currency: "USD", AccountType: "Business", AccountSubType: "CurrentAccount"},
			},
		},
		Links: models.Links{Self: "/open-banking/v3.1/aisp/accounts"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-banking/v3.1/aisp/accounts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expected); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	svc, _ := newAccountsService(t, handler)

	resp, err := svc.GetAccounts(context.Background())
	if err != nil {
		t.Fatalf("GetAccounts: %v", err)
	}
	if len(resp.Data.Account) != 2 {
		t.Errorf("accounts: got %d, want 2", len(resp.Data.Account))
	}
	if resp.Data.Account[0].AccountId != "acc-1" {
		t.Errorf("first account ID: got %q, want %q", resp.Data.Account[0].AccountId, "acc-1")
	}
}

func TestGetAccount_SingleAccount(t *testing.T) {
	expected := models.GetAccountResponse{
		Data: models.GetAccountData{
			Account: []models.OBAccount6{
				{AccountId: "acc-42", Currency: "GBP", AccountType: "Personal", AccountSubType: "Savings"},
			},
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expected); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	svc, _ := newAccountsService(t, handler)

	resp, err := svc.GetAccount(context.Background(), "acc-42")
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if resp.Data.Account[0].AccountId != "acc-42" {
		t.Errorf("account ID: got %q, want %q", resp.Data.Account[0].AccountId, "acc-42")
	}
}

func TestGetAccounts_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]string{"Code": "UK.OBIE.Unauthorized"}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	svc, _ := newAccountsService(t, handler)

	_, err := svc.GetAccounts(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func (d *testDoer) Delete(ctx context.Context, url string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	return nil
}

func (d *testDoer) Post(ctx context.Context, url string, body, out any, _ transport.DoOptions) error {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (d *testDoer) Put(ctx context.Context, url string, body, out any, _ transport.DoOptions) error {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

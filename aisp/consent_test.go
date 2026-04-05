package aisp_test

import (
	"fmt"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/aisp"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/obie"
)

// ─── test infrastructure ─────────────────────────────────────────────────

type testDoer struct{ client *http.Client }

func (d *testDoer) Get(ctx context.Context, url string, out any) error {
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
		return &obie.APIError{StatusCode: resp.StatusCode}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (d *testDoer) Delete(ctx context.Context, url string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return &obie.APIError{StatusCode: resp.StatusCode}
	}
	return nil
}

func newSvc(t *testing.T, mux *http.ServeMux) (*aisp.ConsentService, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return aisp.NewConsentService(&testDoer{client: srv.Client()}, srv.URL), srv
}

func jsonH(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(v); err != nil {
  	http.Error(w, err.Error(), http.StatusInternalServerError)
  }
	}
}

// ─── Account Access Consent tests ────────────────────────────────────────

func TestCreateAccountAccessConsent_AllPermissions(t *testing.T) {
	expiry := time.Now().Add(90 * 24 * time.Hour)
	txFrom := time.Now().AddDate(-1, 0, 0)
	txTo := time.Now()

	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/account-access-consents",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req models.OBReadConsent1
   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
   	http.Error(w, err.Error(), http.StatusBadRequest)
   	return
   }
			if len(req.Data.Permissions) == 0 {
				t.Error("expected permissions to be set")
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBReadConsentResponse1{
				Data: models.OBReadDataConsentResponse1{
					ConsentId:              "urn-alphabank-intent-88379",
					Status:                 "AwaitingAuthorisation",
					CreationDateTime:       time.Now(),
					StatusUpdateDateTime:   time.Now(),
					Permissions:            req.Data.Permissions,
					ExpirationDateTime:     req.Data.ExpirationDateTime,
					TransactionFromDateTime: req.Data.TransactionFromDateTime,
					TransactionToDateTime:  req.Data.TransactionToDateTime,
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateAccountAccessConsent(context.Background(), &models.OBReadConsent1{
		Data: models.OBReadData1{
			Permissions: []models.Permission{
				models.PermissionReadAccountsDetail,
				models.PermissionReadBalances,
				models.PermissionReadBeneficiariesDetail,
				models.PermissionReadDirectDebits,
				models.PermissionReadProducts,
				models.PermissionReadStandingOrdersDetail,
				models.PermissionReadTransactionsCredits,
				models.PermissionReadTransactionsDebits,
				models.PermissionReadTransactionsDetail,
				models.PermissionReadOffers,
				models.PermissionReadPAN,
				models.PermissionReadParty,
				models.PermissionReadPartyPSU,
				models.PermissionReadScheduledPaymentsDetail,
				models.PermissionReadStatementsDetail,
			},
			ExpirationDateTime:      &expiry,
			TransactionFromDateTime: &txFrom,
			TransactionToDateTime:   &txTo,
		},
	})
	if err != nil {
		t.Fatalf("CreateAccountAccessConsent: %v", err)
	}
	if resp.Data.ConsentId != "urn-alphabank-intent-88379" {
		t.Errorf("ConsentId: got %q", resp.Data.ConsentId)
	}
	if resp.Data.Status != "AwaitingAuthorisation" {
		t.Errorf("Status: got %q", resp.Data.Status)
	}
	if len(resp.Data.Permissions) == 0 {
		t.Error("expected permissions in response")
	}
}

func TestCreateAccountAccessConsent_LimitedPermissions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/account-access-consents",
		func(w http.ResponseWriter, r *http.Request) {
			var req models.OBReadConsent1
   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
   	http.Error(w, err.Error(), http.StatusBadRequest)
   	return
   }
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBReadConsentResponse1{
				Data: models.OBReadDataConsentResponse1{
					ConsentId:   "consent-limited",
					Status:      "AwaitingAuthorisation",
					Permissions: req.Data.Permissions,
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateAccountAccessConsent(context.Background(), &models.OBReadConsent1{
		Data: models.OBReadData1{
			Permissions: []models.Permission{
				models.PermissionReadAccountsBasic,
				models.PermissionReadBalances,
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateAccountAccessConsent limited: %v", err)
	}
	if len(resp.Data.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(resp.Data.Permissions))
	}
}

func TestGetAccountAccessConsent_AwaitingAuthorisation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/account-access-consents/consent-abc",
		jsonH(models.OBReadConsentResponse1{
			Data: models.OBReadDataConsentResponse1{
				ConsentId: "consent-abc",
				Status:    "AwaitingAuthorisation",
				Permissions: []models.Permission{
					models.PermissionReadAccountsDetail,
					models.PermissionReadBalances,
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetAccountAccessConsent(context.Background(), "consent-abc")
	if err != nil {
		t.Fatalf("GetAccountAccessConsent: %v", err)
	}
	if resp.Data.Status != "AwaitingAuthorisation" {
		t.Errorf("Status: got %q", resp.Data.Status)
	}
}

func TestGetAccountAccessConsent_Authorised(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/account-access-consents/consent-abc",
		jsonH(models.OBReadConsentResponse1{
			Data: models.OBReadDataConsentResponse1{
				ConsentId: "consent-abc",
				Status:    "Authorised",
				Permissions: []models.Permission{
					models.PermissionReadAccountsDetail,
					models.PermissionReadBalances,
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetAccountAccessConsent(context.Background(), "consent-abc")
	if err != nil {
		t.Fatalf("GetAccountAccessConsent: %v", err)
	}
	if resp.Data.Status != "Authorised" {
		t.Errorf("Status: got %q", resp.Data.Status)
	}
}

func TestDeleteAccountAccessConsent(t *testing.T) {
	deleted := false
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/account-access-consents/consent-del",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		})

	svc, _ := newSvc(t, mux)
	if err := svc.DeleteAccountAccessConsent(context.Background(), "consent-del"); err != nil {
		t.Fatalf("DeleteAccountAccessConsent: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to reach server")
	}
}

func TestDeleteAccountAccessConsent_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/account-access-consents/missing",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

	svc, _ := newSvc(t, mux)
	err := svc.DeleteAccountAccessConsent(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

// ─── Offers tests ────────────────────────────────────────────────────────

func TestGetOffers_Bulk(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/offers",
		jsonH(models.OBReadOffer1{
			Data: models.OBReadDataOffer1{
				Offer: []models.OBOffer1{
					{AccountId: "acc-1", OfferId: "offer-1", OfferType: models.OfferTypeLimitIncrease,
						Description: "Credit limit increase up to £10000",
						Amount:      &models.OBActiveOrHistoricCurrencyAndAmount{Amount: "10000.00", Currency: "GBP"}},
					{AccountId: "acc-2", OfferId: "offer-2", OfferType: models.OfferTypeBalanceTransfer,
						Description: "Balance transfer 0% for 18 months"},
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetOffers(context.Background())
	if err != nil {
		t.Fatalf("GetOffers: %v", err)
	}
	if len(resp.Data.Offer) != 2 {
		t.Errorf("offer count: got %d, want 2", len(resp.Data.Offer))
	}
	if resp.Data.Offer[0].OfferType != models.OfferTypeLimitIncrease {
		t.Errorf("OfferType: got %q", resp.Data.Offer[0].OfferType)
	}
}

func TestGetAccountOffers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/offers",
		jsonH(models.OBReadOffer1{
			Data: models.OBReadDataOffer1{
				Offer: []models.OBOffer1{
					{AccountId: "acc-1", OfferId: "offer-1", OfferType: models.OfferTypePromotionalRate,
						Description: "Promotional interest rate 1.5% for 12 months",
						Rate:        "1.5"},
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetAccountOffers(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountOffers: %v", err)
	}
	if len(resp.Data.Offer) != 1 {
		t.Errorf("offer count: got %d, want 1", len(resp.Data.Offer))
	}
	if resp.Data.Offer[0].Rate != "1.5" {
		t.Errorf("Rate: got %q", resp.Data.Offer[0].Rate)
	}
}

// ─── Standing Orders (AIS) tests ─────────────────────────────────────────

func TestGetStandingOrders(t *testing.T) {
	now := time.Now()
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/standing-orders",
		jsonH(models.OBReadStandingOrder6{
			Data: models.OBReadDataStandingOrder6{
				StandingOrder: []models.OBStandingOrder6{
					{AccountId: "acc-1", StandingOrderId: "so-1",
						Frequency:            "IntrvlMnthDay:01:3",
						Reference:            "Rent payment",
						FirstPaymentDateTime: &now,
						NextPaymentAmount: &models.OBActiveOrHistoricCurrencyAndAmount{Amount: "1200.00", Currency: "GBP"},
						CreditorAccount: &models.OBCashAccount3{
							SchemeName:     "UK.OBIE.SortCodeAccountNumber",
							Identification: "20000319825731",
							Name:           "Landlord Ltd",
						}},
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetStandingOrders(context.Background())
	if err != nil {
		t.Fatalf("GetStandingOrders: %v", err)
	}
	if len(resp.Data.StandingOrder) != 1 {
		t.Errorf("count: got %d, want 1", len(resp.Data.StandingOrder))
	}
	if resp.Data.StandingOrder[0].Reference != "Rent payment" {
		t.Errorf("Reference: got %q", resp.Data.StandingOrder[0].Reference)
	}
}

func TestGetAccountStandingOrders(t *testing.T) {
	now := time.Now()
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/aisp/accounts/acc-1/standing-orders",
		jsonH(models.OBReadStandingOrder6{
			Data: models.OBReadDataStandingOrder6{
				StandingOrder: []models.OBStandingOrder6{
					{AccountId: "acc-1", StandingOrderId: "so-2",
						Frequency:            "EvryWorkgDay",
						NextPaymentDateTime:  &now,
						NextPaymentAmount: &models.OBActiveOrHistoricCurrencyAndAmount{Amount: "50.00", Currency: "GBP"},
					},
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetAccountStandingOrders(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("GetAccountStandingOrders: %v", err)
	}
	if resp.Data.StandingOrder[0].StandingOrderId != "so-2" {
		t.Errorf("StandingOrderId: got %q", resp.Data.StandingOrder[0].StandingOrderId)
	}
}

// ─── Full AIS consent lifecycle test ─────────────────────────────────────

func TestFullConsentLifecycle(t *testing.T) {
	mux := http.NewServeMux()

	// POST
	mux.HandleFunc("/open-banking/v3.1/aisp/account-access-consents",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBReadConsentResponse1{
				Data: models.OBReadDataConsentResponse1{
					ConsentId: "lifecycle-consent",
					Status:    "AwaitingAuthorisation",
					Permissions: []models.Permission{
						models.PermissionReadAccountsDetail,
						models.PermissionReadBalances,
					},
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	calls := map[string]int{}
	// GET + DELETE
	mux.HandleFunc("/open-banking/v3.1/aisp/account-access-consents/lifecycle-consent",
		func(w http.ResponseWriter, r *http.Request) {
			calls[r.Method]++
			switch r.Method {
			case http.MethodGet:
				status := "AwaitingAuthorisation"
				if calls[http.MethodGet] > 1 {
					status = "Authorised"
				}
				if err := json.NewEncoder(w).Encode(models.OBReadConsentResponse1{
					Data: models.OBReadDataConsentResponse1{
						ConsentId: "lifecycle-consent",
						Status:    models.ConsentStatus(status),
					},
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			}
		})

	svc, _ := newSvc(t, mux)
	ctx := context.Background()

	// Step 1: Create
	created, err := svc.CreateAccountAccessConsent(ctx, &models.OBReadConsent1{
		Data: models.OBReadData1{
			Permissions: []models.Permission{models.PermissionReadAccountsDetail, models.PermissionReadBalances},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Data.Status != "AwaitingAuthorisation" {
		t.Errorf("Create status: %q", created.Data.Status)
	}

	// Step 2: GET before auth
	before, err := svc.GetAccountAccessConsent(ctx, "lifecycle-consent")
	if err != nil {
		t.Fatalf("Get (before auth): %v", err)
	}
	if before.Data.Status != "AwaitingAuthorisation" {
		t.Errorf("Before auth status: %q", before.Data.Status)
	}

	// Step 3: GET after auth (simulated)
	after, err := svc.GetAccountAccessConsent(ctx, "lifecycle-consent")
	if err != nil {
		t.Fatalf("Get (after auth): %v", err)
	}
	if after.Data.Status != "Authorised" {
		t.Errorf("After auth status: %q", after.Data.Status)
	}

	// Step 4: DELETE
	if err := svc.DeleteAccountAccessConsent(ctx, "lifecycle-consent"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if calls[http.MethodDelete] != 1 {
		t.Errorf("DELETE calls: got %d, want 1", calls[http.MethodDelete])
	}
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

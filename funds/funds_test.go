package funds_test

import (
	"fmt"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/funds"
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

func newSvc(t *testing.T, mux *http.ServeMux) (*funds.Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return funds.New(&testDoer{client: srv.Client()}, srv.URL), srv
}

func jsonResp(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(v); err != nil {
  	http.Error(w, err.Error(), http.StatusInternalServerError)
  }
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────

func TestCreateConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmation-consents",
		jsonResp(models.OBFundsConfirmationConsentResponse1{
			Data: models.OBFundsConfirmationConsentResponseData1{
				ConsentId:            "cbpii-cid-1",
				Status:               "AwaitingAuthorisation",
				CreationDateTime:     time.Now(),
				StatusUpdateDateTime: time.Now(),
				DebtorAccount: models.OBCashAccount3{
					SchemeName:     "UK.OBIE.SortCodeAccountNumber",
					Identification: "20000319825731",
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateConsent(context.Background(), &models.OBFundsConfirmationConsent1{
		Data: models.OBFundsConfirmationConsentData1{
			DebtorAccount: models.OBCashAccount3{
				SchemeName:     "UK.OBIE.SortCodeAccountNumber",
				Identification: "20000319825731",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateConsent: %v", err)
	}
	if resp.Data.ConsentId != "cbpii-cid-1" {
		t.Errorf("ConsentId: got %q, want %q", resp.Data.ConsentId, "cbpii-cid-1")
	}
	if resp.Data.Status != "AwaitingAuthorisation" {
		t.Errorf("Status: got %q, want AwaitingAuthorisation", resp.Data.Status)
	}
}

func TestGetConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmation-consents/cbpii-cid-1",
		jsonResp(models.OBFundsConfirmationConsentResponse1{
			Data: models.OBFundsConfirmationConsentResponseData1{
				ConsentId: "cbpii-cid-1",
				Status:    "Authorised",
				DebtorAccount: models.OBCashAccount3{
					SchemeName:     "UK.OBIE.SortCodeAccountNumber",
					Identification: "20000319825731",
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetConsent(context.Background(), "cbpii-cid-1")
	if err != nil {
		t.Fatalf("GetConsent: %v", err)
	}
	if resp.Data.ConsentId != "cbpii-cid-1" {
		t.Errorf("ConsentId: got %q", resp.Data.ConsentId)
	}
	if resp.Data.Status != "Authorised" {
		t.Errorf("Status: got %q, want Authorised", resp.Data.Status)
	}
}

func TestDeleteConsent(t *testing.T) {
	deleted := false
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmation-consents/cbpii-cid-1",
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
	if err := svc.DeleteConsent(context.Background(), "cbpii-cid-1"); err != nil {
		t.Fatalf("DeleteConsent: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to reach server")
	}
}

func TestDeleteConsent_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmation-consents/missing-id",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(map[string]string{
				"Code":    "UK.OBIE.Resource.NotFound",
				"Message": "Consent not found",
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	err := svc.DeleteConsent(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	apiErr, ok := err.(*obie.APIError)
	if !ok {
		t.Fatalf("expected *obie.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode: got %d, want 404", apiErr.StatusCode)
	}
}

func TestConfirmFundsAvailability_Available(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmations",
		jsonResp(models.OBFundsConfirmationResponse1{
			Data: models.OBFundsConfirmationResponseData1{
				FundsConfirmationId: "fc-1",
				ConsentId:           "cbpii-cid-1",
				FundsAvailable:      true,
				Reference:           "purchase-abc",
				InstructedAmount:    models.OBActiveOrHistoricCurrencyAndAmount{Amount: "25.00", Currency: "GBP"},
				CreationDateTime:    time.Now(),
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.ConfirmFundsAvailability(context.Background(), &models.OBFundsConfirmation1{
		Data: models.OBFundsConfirmationData1{
			ConsentId: "cbpii-cid-1",
			Reference: "purchase-abc",
			InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{
				Amount:   "25.00",
				Currency: "GBP",
			},
		},
	})
	if err != nil {
		t.Fatalf("ConfirmFundsAvailability: %v", err)
	}
	if !resp.Data.FundsAvailable {
		t.Error("expected FundsAvailable=true")
	}
	if resp.Data.FundsConfirmationId != "fc-1" {
		t.Errorf("FundsConfirmationId: got %q, want fc-1", resp.Data.FundsConfirmationId)
	}
}

func TestConfirmFundsAvailability_Unavailable(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmations",
		jsonResp(models.OBFundsConfirmationResponse1{
			Data: models.OBFundsConfirmationResponseData1{
				FundsConfirmationId: "fc-2",
				FundsAvailable:      false,
				Reference:           "expensive-purchase",
				InstructedAmount:    models.OBActiveOrHistoricCurrencyAndAmount{Amount: "9999.99", Currency: "GBP"},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.ConfirmFundsAvailability(context.Background(), &models.OBFundsConfirmation1{
		Data: models.OBFundsConfirmationData1{
			ConsentId: "cbpii-cid-1",
			Reference: "expensive-purchase",
			InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{
				Amount:   "9999.99",
				Currency: "GBP",
			},
		},
	})
	if err != nil {
		t.Fatalf("ConfirmFundsAvailability: %v", err)
	}
	if resp.Data.FundsAvailable {
		t.Error("expected FundsAvailable=false for insufficient funds")
	}
}

func TestConfirmFundsAvailability_ConsentExpired(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmations",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(map[string]any{
				"Code":    "UK.OBIE.Resource.ConsentMismatch",
				"Message": "Consent is not Authorised",
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	_, err := svc.ConfirmFundsAvailability(context.Background(), &models.OBFundsConfirmation1{
		Data: models.OBFundsConfirmationData1{
			ConsentId:        "expired-consent",
			Reference:        "ref",
			InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "10.00", Currency: "GBP"},
		},
	})
	if err == nil {
		t.Fatal("expected error for forbidden response, got nil")
	}
	apiErr, ok := err.(*obie.APIError)
	if !ok {
		t.Fatalf("expected *obie.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode: got %d, want 403", apiErr.StatusCode)
	}
}

func TestConfirmFundsAvailability_WithExpirationDateTime(t *testing.T) {
	expiry := time.Now().Add(30 * 24 * time.Hour)
	mux := http.NewServeMux()

	// First: create consent with expiry
	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmation-consents",
		jsonResp(models.OBFundsConfirmationConsentResponse1{
			Data: models.OBFundsConfirmationConsentResponseData1{
				ConsentId:          "cbpii-exp-1",
				Status:             "AwaitingAuthorisation",
				ExpirationDateTime: &expiry,
				DebtorAccount: models.OBCashAccount3{
					SchemeName:     "UK.OBIE.SortCodeAccountNumber",
					Identification: "20000319825731",
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateConsent(context.Background(), &models.OBFundsConfirmationConsent1{
		Data: models.OBFundsConfirmationConsentData1{
			ExpirationDateTime: &expiry,
			DebtorAccount: models.OBCashAccount3{
				SchemeName:     "UK.OBIE.SortCodeAccountNumber",
				Identification: "20000319825731",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateConsent with expiry: %v", err)
	}
	if resp.Data.ExpirationDateTime == nil {
		t.Error("expected ExpirationDateTime to be set")
	}
}

func TestFullLifecycle(t *testing.T) {
	// Full CBPII lifecycle: Create → Get → Confirm → Delete
	mux := http.NewServeMux()

	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmation-consents",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Create: expected POST, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBFundsConfirmationConsentResponse1{
				Data: models.OBFundsConfirmationConsentResponseData1{
					ConsentId: "lifecycle-cid",
					Status:    "AwaitingAuthorisation",
					DebtorAccount: models.OBCashAccount3{
						SchemeName:     "UK.OBIE.SortCodeAccountNumber",
						Identification: "20000319825731",
					},
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmation-consents/lifecycle-cid",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				if err := json.NewEncoder(w).Encode(models.OBFundsConfirmationConsentResponse1{
					Data: models.OBFundsConfirmationConsentResponseData1{
						ConsentId: "lifecycle-cid",
						Status:    "Authorised",
					},
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Errorf("unexpected method %s", r.Method)
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})

	mux.HandleFunc("/open-banking/v3.1/cbpii/funds-confirmations",
		jsonResp(models.OBFundsConfirmationResponse1{
			Data: models.OBFundsConfirmationResponseData1{
				FundsConfirmationId: "fc-lifecycle",
				ConsentId:           "lifecycle-cid",
				FundsAvailable:      true,
				Reference:           "test-ref",
			},
		}))

	svc, _ := newSvc(t, mux)
	ctx := context.Background()

	// Step 1: Create consent
	createResp, err := svc.CreateConsent(ctx, &models.OBFundsConfirmationConsent1{
		Data: models.OBFundsConfirmationConsentData1{
			DebtorAccount: models.OBCashAccount3{
				SchemeName:     "UK.OBIE.SortCodeAccountNumber",
				Identification: "20000319825731",
			},
		},
	})
	if err != nil {
		t.Fatalf("Step 1 CreateConsent: %v", err)
	}
	consentID := createResp.Data.ConsentId

	// Step 2: Get consent (simulating PSU authorisation check)
	getResp, err := svc.GetConsent(ctx, consentID)
	if err != nil {
		t.Fatalf("Step 2 GetConsent: %v", err)
	}
	if getResp.Data.Status != "Authorised" {
		t.Errorf("Step 2 Status: got %q, want Authorised", getResp.Data.Status)
	}

	// Step 3: Confirm funds
	confirmResp, err := svc.ConfirmFundsAvailability(ctx, &models.OBFundsConfirmation1{
		Data: models.OBFundsConfirmationData1{
			ConsentId: consentID,
			Reference: "test-ref",
			InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{
				Amount: "50.00", Currency: "GBP",
			},
		},
	})
	if err != nil {
		t.Fatalf("Step 3 ConfirmFundsAvailability: %v", err)
	}
	if !confirmResp.Data.FundsAvailable {
		t.Error("Step 3: expected FundsAvailable=true")
	}

	// Step 4: Delete consent
	if err := svc.DeleteConsent(ctx, consentID); err != nil {
		t.Fatalf("Step 4 DeleteConsent: %v", err)
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

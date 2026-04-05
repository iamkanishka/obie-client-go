package payments_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/obie"
	"github.com/iamkanishka/obie-client-go/payments"
)

// ─── test infrastructure ─────────────────────────────────────────────────

type testDoer struct {
	client *http.Client
	base   string
}

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

type stubSigner struct{}

func (s *stubSigner) SignJSON(_ any) (string, error) { return "header..sig", nil }

func newSvc(t *testing.T, mux *http.ServeMux) (*payments.Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	doer := &testDoer{client: srv.Client(), base: srv.URL}
	return payments.New(doer, &stubSigner{}, srv.URL), srv
}

func jsonHandler(t *testing.T, code int, v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)

		if v != nil {
			if err := json.NewEncoder(w).Encode(v); err != nil {
				t.Fatalf("encode JSON response: %v", err)
			}
		}
	}
}

func deleteHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ─── Domestic payment consent ─────────────────────────────────────────────

func TestCreateDomesticPaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	want := models.OBWriteDomesticConsentResponse5{
		Data: models.OBWriteDomesticConsentResponseData5{ConsentId: "cid-1", Status: "AwaitingAuthorisation"},
	}
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-payment-consents",
		jsonHandler(t, 201, want))

	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateDomesticPaymentConsent(context.Background(), &models.OBWriteDomesticConsent5{
		Data: models.OBWriteDomesticConsentData5{
			Initiation: models.OBDomesticInitiation{
				InstructionIdentification: "INSTR-001",
				EndToEndIdentification:    "E2E-001",
				InstructedAmount:          models.OBActiveOrHistoricCurrencyAndAmount{Amount: "10.00", Currency: "GBP"},
				CreditorAccount:           models.OBCashAccount3{SchemeName: "UK.OBIE.SortCodeAccountNumber", Identification: "20000319825731"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateDomesticPaymentConsent: %v", err)
	}
	if resp.Data.ConsentId != "cid-1" {
		t.Errorf("ConsentId: got %q", resp.Data.ConsentId)
	}
}

func TestGetDomesticPaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-payment-consents/cid-1",
		jsonHandler(t, 200, models.OBWriteDomesticConsentResponse5{
			Data: models.OBWriteDomesticConsentResponseData5{ConsentId: "cid-1", Status: "Authorised"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetDomesticPaymentConsent(context.Background(), "cid-1")
	if err != nil {
		t.Fatalf("GetDomesticPaymentConsent: %v", err)
	}
	if resp.Data.Status != "Authorised" {
		t.Errorf("Status: %s", resp.Data.Status)
	}
}

func TestGetDomesticPaymentConsentFundsConfirmation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-payment-consents/cid-1/funds-confirmation",
		jsonHandler(t, 200, models.OBFundsConfirmationResponse1{
			Data: models.OBFundsConfirmationResponseData1{FundsAvailable: true},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetDomesticPaymentConsentFundsConfirmation(context.Background(), "cid-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !resp.Data.FundsAvailable {
		t.Error("expected FundsAvailable=true")
	}
}

// ─── Domestic payment submission ──────────────────────────────────────────

func TestSubmitDomesticPayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-payments",
		jsonHandler(t, 201, models.OBWriteDomesticResponse5{
			Data: models.OBWriteDomesticResponseData5{DomesticPaymentId: "pay-1", Status: "Pending"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.SubmitDomesticPayment(context.Background(), &models.OBWriteDomestic2{
		Data: models.OBWriteDomesticData2{ConsentId: "cid-1"},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.DomesticPaymentId != "pay-1" {
		t.Errorf("PaymentId: %s", resp.Data.DomesticPaymentId)
	}
}

func TestGetDomesticPayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-payments/pay-1",
		jsonHandler(t, 200, models.OBWriteDomesticResponse5{
			Data: models.OBWriteDomesticResponseData5{DomesticPaymentId: "pay-1", Status: "AcceptedSettlementCompleted"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetDomesticPayment(context.Background(), "pay-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.Status != "AcceptedSettlementCompleted" {
		t.Errorf("Status: %s", resp.Data.Status)
	}
}

func TestGetDomesticPaymentDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-payments/pay-1/payment-details",
		jsonHandler(t, 200, models.OBWritePaymentDetailsResponse1{
			Data: models.OBWritePaymentDetailsResponseData1{
				PaymentStatus: []models.OBPaymentDetailsStatus1{{Status: "AcceptedSettlementCompleted"}},
			},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetDomesticPaymentDetails(context.Background(), "pay-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(resp.Data.PaymentStatus) == 0 {
		t.Error("expected payment status entries")
	}
}

// ─── Domestic scheduled payment ───────────────────────────────────────────

func TestCreateDomesticScheduledPaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-scheduled-payment-consents",
		jsonHandler(t, 201, models.OBWriteDomesticScheduledConsentResponse4{
			Data: models.OBWriteDomesticScheduledConsentResponseData4{ConsentId: "sched-cid-1", Status: "AwaitingAuthorisation"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateDomesticScheduledPaymentConsent(context.Background(), &models.OBWriteDomesticScheduledConsent4{
		Data: models.OBWriteDomesticScheduledConsentData4{
			Initiation: models.OBDomesticScheduledInitiation{
				InstructionIdentification:  "INSTR-SCHED-001",
				RequestedExecutionDateTime: time.Now().Add(24 * time.Hour),
				InstructedAmount:           models.OBActiveOrHistoricCurrencyAndAmount{Amount: "50.00", Currency: "GBP"},
				CreditorAccount:            models.OBCashAccount3{SchemeName: "UK.OBIE.SortCodeAccountNumber", Identification: "20000319825731"},
			},
		},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.ConsentId != "sched-cid-1" {
		t.Errorf("ConsentId: %s", resp.Data.ConsentId)
	}
}

func TestGetDomesticScheduledPaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-scheduled-payment-consents/sched-cid-1",
		jsonHandler(t, 200, models.OBWriteDomesticScheduledConsentResponse4{
			Data: models.OBWriteDomesticScheduledConsentResponseData4{ConsentId: "sched-cid-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetDomesticScheduledPaymentConsent(context.Background(), "sched-cid-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.ConsentId != "sched-cid-1" {
		t.Errorf("ConsentId: %s", resp.Data.ConsentId)
	}
}

func TestDeleteDomesticScheduledPaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-scheduled-payment-consents/sched-cid-1",
		deleteHandler(t))
	svc, _ := newSvc(t, mux)
	if err := svc.DeleteDomesticScheduledPaymentConsent(context.Background(), "sched-cid-1"); err != nil {
		t.Fatalf("DeleteDomesticScheduledPaymentConsent: %v", err)
	}
}

func TestSubmitDomesticScheduledPayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-scheduled-payments",
		jsonHandler(t, 201, models.OBWriteDomesticScheduledResponse5{
			Data: models.OBWriteDomesticScheduledResponseData5{DomesticScheduledPaymentId: "sched-pay-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.SubmitDomesticScheduledPayment(context.Background(), &models.OBWriteDomesticScheduled3{
		Data: models.OBWriteDomesticScheduledData3{ConsentId: "sched-cid-1"},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.DomesticScheduledPaymentId != "sched-pay-1" {
		t.Errorf("got %s", resp.Data.DomesticScheduledPaymentId)
	}
}

func TestGetDomesticScheduledPayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-scheduled-payments/sched-pay-1",
		jsonHandler(t, 200, models.OBWriteDomesticScheduledResponse5{
			Data: models.OBWriteDomesticScheduledResponseData5{DomesticScheduledPaymentId: "sched-pay-1", Status: "Pending"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetDomesticScheduledPayment(context.Background(), "sched-pay-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.Status != "Pending" {
		t.Errorf("Status: %s", resp.Data.Status)
	}
}

func TestGetDomesticScheduledPaymentDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-scheduled-payments/sched-pay-1/payment-details",
		jsonHandler(t, 200, models.OBWritePaymentDetailsResponse1{}))
	svc, _ := newSvc(t, mux)
	_, err := svc.GetDomesticScheduledPaymentDetails(context.Background(), "sched-pay-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
}

// ─── Domestic standing order ──────────────────────────────────────────────

func TestCreateDomesticStandingOrderConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-standing-order-consents",
		jsonHandler(t, 201, models.OBWriteDomesticStandingOrderConsentResponse6{
			Data: models.OBWriteDomesticStandingOrderConsentResponseData6{ConsentId: "so-cid-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateDomesticStandingOrderConsent(context.Background(), &models.OBWriteDomesticStandingOrderConsent5{
		Data: models.OBWriteDomesticStandingOrderConsentData5{
			Initiation: models.OBDomesticStandingOrderInitiation{
				Frequency:            "EvryWorkgDay",
				FirstPaymentDateTime: time.Now().Add(24 * time.Hour),
				FirstPaymentAmount:   models.OBActiveOrHistoricCurrencyAndAmount{Amount: "100.00", Currency: "GBP"},
				CreditorAccount:      models.OBCashAccount3{SchemeName: "UK.OBIE.SortCodeAccountNumber", Identification: "20000319825731"},
			},
		},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.ConsentId != "so-cid-1" {
		t.Errorf("ConsentId: %s", resp.Data.ConsentId)
	}
}

func TestGetDomesticStandingOrderConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-standing-order-consents/so-cid-1",
		jsonHandler(t, 200, models.OBWriteDomesticStandingOrderConsentResponse6{
			Data: models.OBWriteDomesticStandingOrderConsentResponseData6{ConsentId: "so-cid-1", Status: "Authorised"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetDomesticStandingOrderConsent(context.Background(), "so-cid-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.Status != "Authorised" {
		t.Errorf("Status: %s", resp.Data.Status)
	}
}

func TestSubmitDomesticStandingOrder(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-standing-orders",
		jsonHandler(t, 201, models.OBWriteDomesticStandingOrderResponse6{
			Data: models.OBWriteDomesticStandingOrderResponseData6{DomesticStandingOrderId: "so-1", Status: "Pending"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.SubmitDomesticStandingOrder(context.Background(), &models.OBWriteDomesticStandingOrder4{
		Data: models.OBWriteDomesticStandingOrderData4{ConsentId: "so-cid-1"},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.DomesticStandingOrderId != "so-1" {
		t.Errorf("ID: %s", resp.Data.DomesticStandingOrderId)
	}
}

func TestGetDomesticStandingOrder(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-standing-orders/so-1",
		jsonHandler(t, 200, models.OBWriteDomesticStandingOrderResponse6{
			Data: models.OBWriteDomesticStandingOrderResponseData6{DomesticStandingOrderId: "so-1", Status: "AcceptedSettlementCompleted"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetDomesticStandingOrder(context.Background(), "so-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.Status != "AcceptedSettlementCompleted" {
		t.Errorf("Status: %s", resp.Data.Status)
	}
}

func TestGetDomesticStandingOrderDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-standing-orders/so-1/payment-details",
		jsonHandler(t, 200, models.OBWritePaymentDetailsResponse1{}))
	svc, _ := newSvc(t, mux)
	_, err := svc.GetDomesticStandingOrderDetails(context.Background(), "so-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
}

// ─── International payment ────────────────────────────────────────────────

func TestCreateInternationalPaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-payment-consents",
		jsonHandler(t, 201, models.OBWriteInternationalConsentResponse6{
			Data: models.OBWriteInternationalConsentResponseData6{ConsentId: "intl-cid-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateInternationalPaymentConsent(context.Background(), &models.OBWriteInternationalConsent5{
		Data: models.OBWriteInternationalConsentData5{
			Initiation: models.OBInternationalInitiation{
				InstructionIdentification: "INSTR-001",
				EndToEndIdentification:    "E2E-001",
				CurrencyOfTransfer:        "USD",
				InstructedAmount:          models.OBActiveOrHistoricCurrencyAndAmount{Amount: "100.00", Currency: "GBP"},
				CreditorAccount:           models.OBCashAccount3{SchemeName: "UK.OBIE.IBAN", Identification: "DE89370400440532013000"},
			},
		},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.ConsentId != "intl-cid-1" {
		t.Errorf("ConsentId: %s", resp.Data.ConsentId)
	}
}

func TestGetInternationalPaymentConsentFundsConfirmation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-payment-consents/intl-cid-1/funds-confirmation",
		jsonHandler(t, 200, models.OBFundsConfirmationResponse1{
			Data: models.OBFundsConfirmationResponseData1{FundsAvailable: true},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetInternationalPaymentConsentFundsConfirmation(context.Background(), "intl-cid-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !resp.Data.FundsAvailable {
		t.Error("FundsAvailable should be true")
	}
}

func TestSubmitInternationalPayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-payments",
		jsonHandler(t, 201, models.OBWriteInternationalResponse5{
			Data: models.OBWriteInternationalResponseData5{InternationalPaymentId: "intl-pay-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.SubmitInternationalPayment(context.Background(), &models.OBWriteInternational3{
		Data: models.OBWriteInternationalData3{ConsentId: "intl-cid-1"},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.InternationalPaymentId != "intl-pay-1" {
		t.Errorf("ID: %s", resp.Data.InternationalPaymentId)
	}
}

func TestGetInternationalPaymentDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-payments/intl-pay-1/payment-details",
		jsonHandler(t, 200, models.OBWritePaymentDetailsResponse1{}))
	svc, _ := newSvc(t, mux)
	_, err := svc.GetInternationalPaymentDetails(context.Background(), "intl-pay-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
}

// ─── International scheduled payment ─────────────────────────────────────

func TestCreateInternationalScheduledPaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-scheduled-payment-consents",
		jsonHandler(t, 201, models.OBWriteInternationalScheduledConsentResponse6{
			Data: models.OBWriteInternationalScheduledConsentResponseData6{ConsentId: "intl-sched-cid-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateInternationalScheduledPaymentConsent(context.Background(), &models.OBWriteInternationalScheduledConsent5{
		Data: models.OBWriteInternationalScheduledConsentData5{
			Initiation: models.OBInternationalScheduledInitiation{
				InstructionIdentification:  "INSTR-001",
				RequestedExecutionDateTime: time.Now().Add(48 * time.Hour),
				CurrencyOfTransfer:         "EUR",
				InstructedAmount:           models.OBActiveOrHistoricCurrencyAndAmount{Amount: "200.00", Currency: "GBP"},
				CreditorAccount:            models.OBCashAccount3{SchemeName: "UK.OBIE.IBAN", Identification: "DE89370400440532013000"},
			},
		},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.ConsentId != "intl-sched-cid-1" {
		t.Errorf("ConsentId: %s", resp.Data.ConsentId)
	}
}

func TestDeleteInternationalScheduledPaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-scheduled-payment-consents/intl-sched-cid-1",
		deleteHandler(t))
	svc, _ := newSvc(t, mux)
	if err := svc.DeleteInternationalScheduledPaymentConsent(context.Background(), "intl-sched-cid-1"); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestSubmitInternationalScheduledPayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-scheduled-payments",
		jsonHandler(t, 201, models.OBWriteInternationalScheduledResponse6{
			Data: models.OBWriteInternationalScheduledResponseData6{InternationalScheduledPaymentId: "intl-sched-pay-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.SubmitInternationalScheduledPayment(context.Background(), &models.OBWriteInternationalScheduled3{
		Data: models.OBWriteInternationalScheduledData3{ConsentId: "intl-sched-cid-1"},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.InternationalScheduledPaymentId != "intl-sched-pay-1" {
		t.Errorf("ID: %s", resp.Data.InternationalScheduledPaymentId)
	}
}

// ─── International standing order ─────────────────────────────────────────

func TestCreateInternationalStandingOrderConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-standing-order-consents",
		jsonHandler(t, 201, models.OBWriteInternationalStandingOrderConsentResponse7{
			Data: models.OBWriteInternationalStandingOrderConsentResponseData7{ConsentId: "intl-so-cid-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateInternationalStandingOrderConsent(context.Background(), &models.OBWriteInternationalStandingOrderConsent6{
		Data: models.OBWriteInternationalStandingOrderConsentData6{
			Initiation: models.OBInternationalStandingOrderInitiation6{
				Frequency:            "EvryWorkgDay",
				FirstPaymentDateTime: time.Now().Add(24 * time.Hour),
				CurrencyOfTransfer:   "USD",
				InstructedAmount:     models.OBActiveOrHistoricCurrencyAndAmount{Amount: "500.00", Currency: "GBP"},
				CreditorAccount:      models.OBCashAccount3{SchemeName: "UK.OBIE.IBAN", Identification: "DE89370400440532013000"},
			},
		},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.ConsentId != "intl-so-cid-1" {
		t.Errorf("ConsentId: %s", resp.Data.ConsentId)
	}
}

func TestSubmitInternationalStandingOrder(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-standing-orders",
		jsonHandler(t, 201, models.OBWriteInternationalStandingOrderResponse7{
			Data: models.OBWriteInternationalStandingOrderResponseData7{InternationalStandingOrderId: "intl-so-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.SubmitInternationalStandingOrder(context.Background(), &models.OBWriteInternationalStandingOrder6{
		Data: models.OBWriteInternationalStandingOrderData6{ConsentId: "intl-so-cid-1"},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp.Data.InternationalStandingOrderId != "intl-so-1" {
		t.Errorf("ID: %s", resp.Data.InternationalStandingOrderId)
	}
}

func TestGetInternationalStandingOrderDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/international-standing-orders/intl-so-1/payment-details",
		jsonHandler(t, 200, models.OBWritePaymentDetailsResponse1{}))
	svc, _ := newSvc(t, mux)
	_, err := svc.GetInternationalStandingOrderDetails(context.Background(), "intl-so-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
}

// ─── Status polling ───────────────────────────────────────────────────────

func TestPollDomesticPaymentUntilTerminal(t *testing.T) {
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-payments/pay-poll",
		func(w http.ResponseWriter, r *http.Request) {
			calls++
			status := "Pending"
			if calls >= 3 {
				status = "AcceptedSettlementCompleted"
			}
			if err := json.NewEncoder(w).Encode(models.OBWriteDomesticResponse5{
				Data: models.OBWriteDomesticResponseData5{DomesticPaymentId: "pay-poll", Status: models.PaymentStatus(status)},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	svc, _ := newSvc(t, mux)
	resp, err := svc.PollDomesticPaymentUntilTerminal(context.Background(), "pay-poll", time.Millisecond)
	if err != nil {
		t.Fatalf("PollDomesticPayment: %v", err)
	}
	if resp.Data.Status != "AcceptedSettlementCompleted" {
		t.Errorf("Status: %s", resp.Data.Status)
	}
	if calls != 3 {
		t.Errorf("calls: got %d, want 3", calls)
	}
}

func TestGetPaymentStatus_Generic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-payments/pay-1",
		jsonHandler(t, 200, map[string]any{"Data": map[string]string{"Status": "Pending"}}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetPaymentStatus(context.Background(), payments.PaymentTypeDomestic, "pay-1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if resp == nil {
		t.Error("expected non-nil response")
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

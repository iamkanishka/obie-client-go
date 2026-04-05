package vrp_test

import (
	"fmt"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/obie"
	"github.com/iamkanishka/obie-client-go/vrp"
)

// ─── test infrastructure ─────────────────────────────────────────────────

type testDoer struct{ client *http.Client }

func (d *testDoer) Get(ctx context.Context, url string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := d.client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 400 { return &obie.APIError{StatusCode: resp.StatusCode} }
	return json.NewDecoder(resp.Body).Decode(out)
}

func (d *testDoer) Post(ctx context.Context, url string, body, out any, _ transport.DoOptions) error {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	resp, err := d.client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 400 { return &obie.APIError{StatusCode: resp.StatusCode} }
	if out != nil { return json.NewDecoder(resp.Body).Decode(out) }
	return nil
}

func (d *testDoer) Delete(ctx context.Context, url string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	resp, err := d.client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 400 { return &obie.APIError{StatusCode: resp.StatusCode} }
	return nil
}

type stubSigner struct{}
func (s *stubSigner) SignJSON(_ any) (string, error) { return "hdr..sig", nil }

func newSvc(t *testing.T, mux *http.ServeMux) (*vrp.Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return vrp.New(&testDoer{client: srv.Client()}, &stubSigner{}, srv.URL), srv
}

func jsonResp(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(v); err != nil {
  	http.Error(w, err.Error(), http.StatusInternalServerError)
  }
	}
}

func vrpConsentReq() *models.OBDomesticVRPConsentRequest {
	return &models.OBDomesticVRPConsentRequest{
		Data: models.OBDomesticVRPConsentRequestData{
			ControlParameters: models.OBDomesticVRPControlParameters{
				MaximumIndividualAmount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "100.00", Currency: "GBP"},
				PeriodicLimits: []models.OBDomesticVRPControlParametersPeriodic{{
					PeriodType: "Month", PeriodAlignment: "Calendar",
					Amount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "500.00", Currency: "GBP"},
				}},
				VRPType:                  []string{"UK.OBIE.VRPType.Sweeping"},
				PSUAuthenticationMethods: []string{"UK.OBIE.SCA"},
			},
			Initiation: models.OBDomesticVRPInitiation{
				CreditorAccount: models.OBCashAccount3{SchemeName: "UK.OBIE.SortCodeAccountNumber", Identification: "20000319825731"},
			},
		},
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────

func TestCreateConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-vrp-consents", jsonResp(
		models.OBDomesticVRPConsentResponse{
			Data: models.OBDomesticVRPConsentResponseData{ConsentId: "vrp-cid-1", Status: "AwaitingAuthorisation"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateConsent(context.Background(), vrpConsentReq())
	if err != nil { t.Fatalf("CreateConsent: %v", err) }
	if resp.Data.ConsentId != "vrp-cid-1" { t.Errorf("ConsentId: %s", resp.Data.ConsentId) }
	if resp.Data.Status != "AwaitingAuthorisation" { t.Errorf("Status: %s", resp.Data.Status) }
}

func TestGetConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-vrp-consents/vrp-cid-1", jsonResp(
		models.OBDomesticVRPConsentResponse{
			Data: models.OBDomesticVRPConsentResponseData{ConsentId: "vrp-cid-1", Status: "Authorised"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetConsent(context.Background(), "vrp-cid-1")
	if err != nil { t.Fatalf("GetConsent: %v", err) }
	if resp.Data.Status != "Authorised" { t.Errorf("Status: %s", resp.Data.Status) }
}

func TestDeleteConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-vrp-consents/vrp-cid-1",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		})
	svc, _ := newSvc(t, mux)
	if err := svc.DeleteConsent(context.Background(), "vrp-cid-1"); err != nil {
		t.Fatalf("DeleteConsent: %v", err)
	}
}

func TestGetConsentFundsConfirmation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-vrp-consents/vrp-cid-1/funds-confirmation",
		jsonResp(models.OBVRPFundsConfirmationResponse{
			Data: models.OBVRPFundsConfirmationResponseData{FundsAvailable: true, ConsentId: "vrp-cid-1"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetConsentFundsConfirmation(context.Background(), "vrp-cid-1",
		&models.OBVRPFundsConfirmationRequest{
			Data: models.OBVRPFundsConfirmationRequestData{
				ConsentId:        "vrp-cid-1",
				Reference:        "ref-001",
				InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{Amount: "50.00", Currency: "GBP"},
			},
		})
	if err != nil { t.Fatalf("GetConsentFundsConfirmation: %v", err) }
	if !resp.Data.FundsAvailable { t.Error("FundsAvailable should be true") }
}

func TestSubmitPayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-vrps", jsonResp(
		models.OBDomesticVRPResponse{
			Data: models.OBDomesticVRPResponseData{DomesticVRPId: "vrp-pay-1", Status: "Pending"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.SubmitPayment(context.Background(), &models.OBDomesticVRPRequest{
		Data: models.OBDomesticVRPRequestData{
			ConsentId:               "vrp-cid-1",
			PSUAuthenticationMethod: "UK.OBIE.SCA",
			Initiation: models.OBDomesticVRPInitiation{
				CreditorAccount: models.OBCashAccount3{SchemeName: "UK.OBIE.SortCodeAccountNumber", Identification: "20000319825731"},
			},
			Instruction: models.OBDomesticVRPInstruction{
				InstructionIdentification: "INSTR-VRP-001",
				EndToEndIdentification:    "E2E-VRP-001",
				InstructedAmount:          models.OBActiveOrHistoricCurrencyAndAmount{Amount: "50.00", Currency: "GBP"},
				CreditorAccount:           models.OBCashAccount3{SchemeName: "UK.OBIE.SortCodeAccountNumber", Identification: "20000319825731"},
			},
		},
	})
	if err != nil { t.Fatalf("SubmitPayment: %v", err) }
	if resp.Data.DomesticVRPId != "vrp-pay-1" { t.Errorf("DomesticVRPId: %s", resp.Data.DomesticVRPId) }
}

func TestGetPayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-vrps/vrp-pay-1", jsonResp(
		models.OBDomesticVRPResponse{
			Data: models.OBDomesticVRPResponseData{DomesticVRPId: "vrp-pay-1", Status: "AcceptedSettlementCompleted"},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetPayment(context.Background(), "vrp-pay-1")
	if err != nil { t.Fatalf("GetPayment: %v", err) }
	if resp.Data.Status != "AcceptedSettlementCompleted" { t.Errorf("Status: %s", resp.Data.Status) }
}

func TestGetPaymentDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-vrps/vrp-pay-1/payment-details", jsonResp(
		models.OBWritePaymentDetailsResponse1{
			Data: models.OBWritePaymentDetailsResponseData1{
				PaymentStatus: []models.OBPaymentDetailsStatus1{{Status: "AcceptedSettlementCompleted"}},
			},
		}))
	svc, _ := newSvc(t, mux)
	resp, err := svc.GetPaymentDetails(context.Background(), "vrp-pay-1")
	if err != nil { t.Fatalf("GetPaymentDetails: %v", err) }
	if len(resp.Data.PaymentStatus) == 0 { t.Error("expected payment status entries") }
}

func TestPollPaymentUntilTerminal(t *testing.T) {
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/domestic-vrps/vrp-poll",
		func(w http.ResponseWriter, r *http.Request) {
			calls++
			status := "Pending"
			if calls >= 2 { status = "AcceptedCreditSettlementCompleted" }
			if err := json.NewEncoder(w).Encode(models.OBDomesticVRPResponse{
				Data: models.OBDomesticVRPResponseData{DomesticVRPId: "vrp-poll", Status: models.PaymentStatus(status)},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	svc, _ := newSvc(t, mux)
	resp, err := svc.PollPaymentUntilTerminal(context.Background(), "vrp-poll", time.Millisecond)
	if err != nil { t.Fatalf("PollPaymentUntilTerminal: %v", err) }
	if resp.Data.Status != "AcceptedCreditSettlementCompleted" { t.Errorf("Status: %s", resp.Data.Status) }
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

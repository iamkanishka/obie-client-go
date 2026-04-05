package filepayments_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/filepayments"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/obie"
)

// ─── test infrastructure ─────────────────────────────────────────────────

type testDoer struct{ srv *httptest.Server }

func (d *testDoer) Get(ctx context.Context, url string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := d.srv.Client().Do(req)
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
	resp, err := d.srv.Client().Do(req)
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

type stubSigner struct{}

func (s *stubSigner) SignJSON(_ any) (string, error) { return "hdr..sig", nil }

func newSvc(t *testing.T, mux *http.ServeMux) (*filepayments.Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	doer := &testDoer{srv: srv}
	tokenFn := func(_ context.Context) (string, error) { return "test-token", nil }
	return filepayments.New(doer, srv.Client(), &stubSigner{}, srv.URL, tokenFn), srv
}

func jsonH(code int, v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(v); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// ─── File Payment Consent tests ───────────────────────────────────────────

func TestCreateFilePaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payment-consents",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBWriteFileConsentResponse4{
				Data: models.OBWriteFileConsentResponseData4{
					ConsentId:            "file-cid-1",
					Status:               "AwaitingUpload",
					CreationDateTime:     time.Now(),
					StatusUpdateDateTime: time.Now(),
					Initiation: models.OBFile2{
						FileType: models.FileTypeUK_OBIE_PaymentInitiation_3_1,
						FileHash: "sha256-abc123",
					},
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)

	controlSum := 5000.00
	numTx := "10"
	resp, err := svc.CreateFilePaymentConsent(context.Background(), &models.OBWriteFileConsent3{
		Data: models.OBWriteFileConsentData3{
			Initiation: models.OBFile2{
				FileType:             models.FileTypeUK_OBIE_PaymentInitiation_3_1,
				FileHash:             "sha256-abc123",
				NumberOfTransactions: numTx,
				ControlSum:           &controlSum,
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateFilePaymentConsent: %v", err)
	}
	if resp.Data.ConsentId != "file-cid-1" {
		t.Errorf("ConsentId: got %q", resp.Data.ConsentId)
	}
}

func TestGetFilePaymentConsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payment-consents/file-cid-1",
		jsonH(200, models.OBWriteFileConsentResponse4{
			Data: models.OBWriteFileConsentResponseData4{
				ConsentId: "file-cid-1",
				Status:    "AwaitingAuthorisation",
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetFilePaymentConsent(context.Background(), "file-cid-1")
	if err != nil {
		t.Fatalf("GetFilePaymentConsent: %v", err)
	}
	if resp.Data.ConsentId != "file-cid-1" {
		t.Errorf("ConsentId: got %q", resp.Data.ConsentId)
	}
}

// ─── File Upload / Download tests ────────────────────────────────────────

func TestUploadFile(t *testing.T) {
	var uploadedBody []byte
	var uploadedContentType string

	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payment-consents/file-cid-1/file",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				uploadedBody, _ = io.ReadAll(r.Body)
				uploadedContentType = r.Header.Get("Content-Type")
				w.WriteHeader(http.StatusOK)
			default:
				t.Errorf("unexpected method: %s", r.Method)
			}
		})

	svc, _ := newSvc(t, mux)

	fileContent := []byte(`{"Data":{"DomesticPaymentInstructions":[]}}`)
	if err := svc.UploadFile(context.Background(), "file-cid-1", fileContent, "application/json"); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if string(uploadedBody) != string(fileContent) {
		t.Errorf("uploaded body mismatch: got %q, want %q", uploadedBody, fileContent)
	}
	if uploadedContentType != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", uploadedContentType)
	}
}

func TestDownloadFile(t *testing.T) {
	fileContent := []byte(`{"Data":{"DomesticPaymentInstructions":[]}}`)

	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payment-consents/file-cid-1/file",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")

			if _, err := w.Write(fileContent); err != nil {
				t.Fatalf("write response: %v", err)
			}
		})

	svc, _ := newSvc(t, mux)
	data, contentType, err := svc.DownloadFile(context.Background(), "file-cid-1")
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	if string(data) != string(fileContent) {
		t.Errorf("downloaded content mismatch")
	}
	if contentType != "application/json" {
		t.Errorf("Content-Type: got %q", contentType)
	}
}

// ─── File Payment Submission tests ───────────────────────────────────────

func TestSubmitFilePayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payments",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBWriteFileResponse3{
				Data: models.OBWriteFileResponseData3{
					FilePaymentId:        "file-pay-1",
					ConsentId:            "file-cid-1",
					Status:               "Pending",
					CreationDateTime:     time.Now(),
					StatusUpdateDateTime: time.Now(),
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	resp, err := svc.SubmitFilePayment(context.Background(), &models.OBWriteFile2{
		Data: models.OBWriteFileData2{
			ConsentId: "file-cid-1",
			Initiation: models.OBFile2{
				FileType: models.FileTypeUK_OBIE_PaymentInitiation_3_1,
				FileHash: "sha256-abc123",
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitFilePayment: %v", err)
	}
	if resp.Data.FilePaymentId != "file-pay-1" {
		t.Errorf("FilePaymentId: got %q", resp.Data.FilePaymentId)
	}
}

func TestGetFilePayment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payments/file-pay-1",
		jsonH(200, models.OBWriteFileResponse3{
			Data: models.OBWriteFileResponseData3{
				FilePaymentId: "file-pay-1",
				Status:        "AcceptedSettlementCompleted",
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetFilePayment(context.Background(), "file-pay-1")
	if err != nil {
		t.Fatalf("GetFilePayment: %v", err)
	}
	if resp.Data.Status != "AcceptedSettlementCompleted" {
		t.Errorf("Status: got %q", resp.Data.Status)
	}
}

func TestGetFilePaymentDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payments/file-pay-1/payment-details",
		jsonH(200, models.OBWriteFilePaymentDetailsResponse1{
			Data: models.OBWriteFilePaymentDetailsResponseData1{
				PaymentStatus: []models.OBPaymentDetailsStatus1{
					{Status: "AcceptedSettlementCompleted"},
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetFilePaymentDetails(context.Background(), "file-pay-1")
	if err != nil {
		t.Fatalf("GetFilePaymentDetails: %v", err)
	}
	if len(resp.Data.PaymentStatus) == 0 {
		t.Error("expected payment status entries")
	}
}

func TestGetFilePaymentReport(t *testing.T) {
	reportContent := []byte(`<?xml version="1.0"?><CstmrPmtStsRpt/>`)

	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payments/file-pay-1/report",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/xml")

			if _, err := w.Write(reportContent); err != nil {
				t.Fatalf("write response: %v", err)
			}
		})

	svc, _ := newSvc(t, mux)
	data, contentType, err := svc.GetFilePaymentReport(context.Background(), "file-pay-1")
	if err != nil {
		t.Fatalf("GetFilePaymentReport: %v", err)
	}
	if string(data) != string(reportContent) {
		t.Errorf("report content mismatch")
	}
	if contentType != "application/xml" {
		t.Errorf("Content-Type: got %q", contentType)
	}
}

// ─── Full file payment lifecycle ──────────────────────────────────────────

func TestFullFilePaymentLifecycle(t *testing.T) {
	mux := http.NewServeMux()

	// POST consent
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payment-consents",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBWriteFileConsentResponse4{
				Data: models.OBWriteFileConsentResponseData4{
					ConsentId: "lifecycle-file-cid",
					Status:    "AwaitingUpload",
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	// File upload/download endpoint
	var storedFile []byte
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payment-consents/lifecycle-file-cid/file",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				var err error
				storedFile, err = io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("read request body: %v", err)
				}
				w.WriteHeader(http.StatusOK)

			case http.MethodGet:
				w.Header().Set("Content-Type", "application/json")

				if _, err := w.Write(storedFile); err != nil {
					t.Fatalf("write response: %v", err)
				}
			}
		})

	// GET consent (after upload, simulate status update)
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payment-consents/lifecycle-file-cid",
		jsonH(200, models.OBWriteFileConsentResponse4{
			Data: models.OBWriteFileConsentResponseData4{
				ConsentId: "lifecycle-file-cid",
				Status:    "AwaitingAuthorisation",
			},
		}))

	// POST payment
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payments",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBWriteFileResponse3{
				Data: models.OBWriteFileResponseData3{
					FilePaymentId: "lifecycle-file-pay",
					ConsentId:     "lifecycle-file-cid",
					Status:        "Pending",
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	// GET payment status
	mux.HandleFunc("/open-banking/v3.1/pisp/file-payments/lifecycle-file-pay",
		jsonH(200, models.OBWriteFileResponse3{
			Data: models.OBWriteFileResponseData3{
				FilePaymentId: "lifecycle-file-pay",
				Status:        "AcceptedSettlementCompleted",
			},
		}))

	svc, _ := newSvc(t, mux)
	ctx := context.Background()

	// Step 1: Create consent
	consent, err := svc.CreateFilePaymentConsent(ctx, &models.OBWriteFileConsent3{
		Data: models.OBWriteFileConsentData3{
			Initiation: models.OBFile2{
				FileType: models.FileTypeUK_OBIE_PaymentInitiation_3_1,
				FileHash: "sha256-test",
			},
		},
	})
	if err != nil {
		t.Fatalf("Step 1 CreateConsent: %v", err)
	}

	// Step 2: Upload file
	fileData := []byte(`{"Data":{"DomesticPaymentInstructions":[]}}`)
	if err := svc.UploadFile(ctx, consent.Data.ConsentId, fileData, "application/json"); err != nil {
		t.Fatalf("Step 2 UploadFile: %v", err)
	}

	// Step 3: Verify file was stored
	downloaded, _, err := svc.DownloadFile(ctx, consent.Data.ConsentId)
	if err != nil {
		t.Fatalf("Step 3 DownloadFile: %v", err)
	}
	if string(downloaded) != string(fileData) {
		t.Error("downloaded file does not match uploaded file")
	}

	// Step 4: Get consent (simulating post-authorisation check)
	consentStatus, err := svc.GetFilePaymentConsent(ctx, consent.Data.ConsentId)
	if err != nil {
		t.Fatalf("Step 4 GetConsent: %v", err)
	}
	if consentStatus.Data.Status != "AwaitingAuthorisation" {
		t.Errorf("Step 4 status: got %q", consentStatus.Data.Status)
	}

	// Step 5: Submit payment
	payment, err := svc.SubmitFilePayment(ctx, &models.OBWriteFile2{
		Data: models.OBWriteFileData2{
			ConsentId: consent.Data.ConsentId,
			Initiation: models.OBFile2{
				FileType: models.FileTypeUK_OBIE_PaymentInitiation_3_1,
				FileHash: "sha256-test",
			},
		},
	})
	if err != nil {
		t.Fatalf("Step 5 SubmitPayment: %v", err)
	}

	// Step 6: Poll status
	finalStatus, err := svc.GetFilePayment(ctx, payment.Data.FilePaymentId)
	if err != nil {
		t.Fatalf("Step 6 GetFilePayment: %v", err)
	}
	if finalStatus.Data.Status != "AcceptedSettlementCompleted" {
		t.Errorf("Step 6 final status: got %q", finalStatus.Data.Status)
	}
}

func (d *testDoer) Delete(ctx context.Context, url string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	resp, err := d.srv.Client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	return nil
}

func (d *testDoer) Put(ctx context.Context, url string, body, out any, _ transport.DoOptions) error {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.srv.Client().Do(req)
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

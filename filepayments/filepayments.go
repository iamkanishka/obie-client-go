// Package filepayments implements the OBIE File Payments API (v3.1.2).
//
// File payments allow a TPP to submit a batch of payments as a single file.
// The flow is:
//
//  1. POST /file-payment-consents        → create consent
//  2. POST /file-payment-consents/{id}/file → upload the bulk payment file
//  3. GET  /file-payment-consents/{id}/file → retrieve the uploaded file
//  4. (PSU authorises via redirect)
//  5. POST /file-payments               → submit the payment
//  6. GET  /file-payments/{id}           → poll status
//  7. GET  /file-payments/{id}/report    → download payment report
//  8. GET  /file-payments/{id}/payment-details → detailed status
//
// Ref: https://openbankinguk.github.io/read-write-api-site2/standards/v3.1.3/resources-and-data-models/pisp/file-payment-consents/
package filepayments

import (
	"github.com/iamkanishka/obie-client-go/internal/apierror"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

const pispBase = "/open-banking/v3.1/pisp"

// Service implements OBIE File Payment endpoints.
type Service struct {
	http       transport.HTTPDoer
	rawHTTP    rawHTTPDoer
	signer     jwsSigner
	baseURL    string
	tokenFn    func(ctx context.Context) (string, error)
}

// transport.HTTPDoer provides standard JSON-over-HTTP operations.
// rawHTTPDoer is aliased from transport.RawDoer for file upload/download.
type rawHTTPDoer = transport.RawDoer

// jwsSigner signs request payloads.
type jwsSigner interface {
	SignJSON(v any) (string, error)
}

// New creates a file payments Service.
// rawHTTP is used for multipart file uploads and binary file downloads.
// tokenFn is called to get a Bearer token for raw requests.
func New(h transport.HTTPDoer, rawHTTP rawHTTPDoer, signer jwsSigner, baseURL string, tokenFn func(ctx context.Context) (string, error)) *Service {
	return &Service{http: h, rawHTTP: rawHTTP, signer: signer, baseURL: baseURL, tokenFn: tokenFn}
}

// ── File Payment Consent ──────────────────────────────────────────────────

// CreateFilePaymentConsent creates a new file payment consent.
//
// POST /file-payment-consents
func (s *Service) CreateFilePaymentConsent(
	ctx context.Context,
	req *models.OBWriteFileConsent3,
) (*models.OBWriteFileConsentResponse4, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("filepayments: sign consent: %w", err)
	}
	opts := transport.DoOptions{
		IdempotencyKey: uuid.New().String(),
		JWSSignature:   sig,
	}
	var resp models.OBWriteFileConsentResponse4
	if err := s.http.Post(ctx,
		s.baseURL+pispBase+"/file-payment-consents",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("filepayments: CreateFilePaymentConsent: %w", err)
	}
	return &resp, nil
}

// GetFilePaymentConsent retrieves a file payment consent by ID.
//
// GET /file-payment-consents/{ConsentId}
func (s *Service) GetFilePaymentConsent(
	ctx context.Context, consentID string,
) (*models.OBWriteFileConsentResponse4, error) {
	var resp models.OBWriteFileConsentResponse4
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s%s/file-payment-consents/%s", s.baseURL, pispBase, consentID),
		&resp); err != nil {
		return nil, fmt.Errorf("filepayments: GetFilePaymentConsent(%s): %w", consentID, err)
	}
	return &resp, nil
}

// UploadFile uploads the bulk payment file against an existing consent.
//
// The file content should be a valid payment initiation file matching the
// FileType declared in the consent (e.g. UK.OBIE.PaymentInitiation.3.1 JSON
// or UK.OBIE.pain.001.001.08 XML).
//
// POST /file-payment-consents/{ConsentId}/file
func (s *Service) UploadFile(
	ctx context.Context,
	consentID string,
	fileContent []byte,
	contentType string, // e.g. "application/json" or "application/xml"
) error {
	token, err := s.tokenFn(ctx)
	if err != nil {
		return fmt.Errorf("filepayments: get token for file upload: %w", err)
	}

	url := fmt.Sprintf("%s%s/file-payment-consents/%s/file", s.baseURL, pispBase, consentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("filepayments: build file upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-idempotency-key", uuid.New().String())

	resp, err := s.rawHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("filepayments: upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &apierror.APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	return nil
}

// DownloadFile downloads the file previously uploaded for a consent.
//
// GET /file-payment-consents/{ConsentId}/file
func (s *Service) DownloadFile(ctx context.Context, consentID string) ([]byte, string, error) {
	token, err := s.tokenFn(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("filepayments: get token for file download: %w", err)
	}

	url := fmt.Sprintf("%s%s/file-payment-consents/%s/file", s.baseURL, pispBase, consentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("filepayments: build file download request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.rawHTTP.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("filepayments: download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, "", &apierror.APIError{StatusCode: resp.StatusCode}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("filepayments: read file body: %w", err)
	}
	return data, resp.Header.Get("Content-Type"), nil
}

// ── File Payment Submission ───────────────────────────────────────────────

// SubmitFilePayment submits the file payment against an authorised consent.
//
// POST /file-payments
func (s *Service) SubmitFilePayment(
	ctx context.Context,
	req *models.OBWriteFile2,
) (*models.OBWriteFileResponse3, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("filepayments: sign payment: %w", err)
	}
	opts := transport.DoOptions{
		IdempotencyKey: uuid.New().String(),
		JWSSignature:   sig,
	}
	var resp models.OBWriteFileResponse3
	if err := s.http.Post(ctx,
		s.baseURL+pispBase+"/file-payments",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("filepayments: SubmitFilePayment: %w", err)
	}
	return &resp, nil
}

// GetFilePayment retrieves the status of a submitted file payment.
//
// GET /file-payments/{FilePaymentId}
func (s *Service) GetFilePayment(
	ctx context.Context, filePaymentID string,
) (*models.OBWriteFileResponse3, error) {
	var resp models.OBWriteFileResponse3
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s%s/file-payments/%s", s.baseURL, pispBase, filePaymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("filepayments: GetFilePayment(%s): %w", filePaymentID, err)
	}
	return &resp, nil
}

// GetFilePaymentDetails retrieves detailed per-transaction status for a file payment.
//
// GET /file-payments/{FilePaymentId}/payment-details
func (s *Service) GetFilePaymentDetails(
	ctx context.Context, filePaymentID string,
) (*models.OBWriteFilePaymentDetailsResponse1, error) {
	var resp models.OBWriteFilePaymentDetailsResponse1
	if err := s.http.Get(ctx,
		fmt.Sprintf("%s%s/file-payments/%s/payment-details", s.baseURL, pispBase, filePaymentID),
		&resp); err != nil {
		return nil, fmt.Errorf("filepayments: GetFilePaymentDetails(%s): %w", filePaymentID, err)
	}
	return &resp, nil
}

// GetFilePaymentReport downloads the payment report for a completed file payment.
// Returns raw report bytes and the content-type header.
//
// GET /file-payments/{FilePaymentId}/report
func (s *Service) GetFilePaymentReport(ctx context.Context, filePaymentID string) ([]byte, string, error) {
	token, err := s.tokenFn(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("filepayments: get token for report: %w", err)
	}

	url := fmt.Sprintf("%s%s/file-payments/%s/report", s.baseURL, pispBase, filePaymentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("filepayments: build report request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json, application/xml")

	resp, err := s.rawHTTP.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("filepayments: get report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, "", &apierror.APIError{StatusCode: resp.StatusCode}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("filepayments: read report body: %w", err)
	}
	return data, resp.Header.Get("Content-Type"), nil
}

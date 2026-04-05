// Package obie provides a production-grade Go SDK for the UK Open Banking
// (OBIE) standard v3.1.3, supporting AIS (with consent), PIS (all payment types),
// CBPII, VRP, File Payments, Event Notifications, DCR, OAuth2/mTLS.
package obie

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/iamkanishka/obie-client-go/accounts"
	"github.com/iamkanishka/obie-client-go/aisp"
	"github.com/iamkanishka/obie-client-go/auth"
	"github.com/iamkanishka/obie-client-go/circuitbreaker"
	"github.com/iamkanishka/obie-client-go/eventnotifications"
	"github.com/iamkanishka/obie-client-go/filepayments"
	"github.com/iamkanishka/obie-client-go/funds"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/middleware"
	"github.com/iamkanishka/obie-client-go/observability"
	"github.com/iamkanishka/obie-client-go/payments"
	"github.com/iamkanishka/obie-client-go/ratelimit"
	"github.com/iamkanishka/obie-client-go/signing"
	"github.com/iamkanishka/obie-client-go/vrp"
)

// DoOptions is the set of optional per-request settings passed to service HTTP methods.
// It is an alias for internal/transport.DoOptions.
type DoOptions = transport.DoOptions

// Client is the root OBIE SDK client. Instantiate one per ASPSP connection.
// It is safe to use concurrently from multiple goroutines.
type Client struct {
	cfg     *Config
	hc      *httpClient
	baseURL string

	// ── AIS services ──────────────────────────────────────────────────────
	// AISConsent manages account-access-consent lifecycle (POST/GET/DELETE).
	// This MUST be used before any Accounts calls to establish consent.
	AISConsent *aisp.ConsentService

	// Accounts exposes all AIS resource endpoints (accounts, balances, txns…).
	Accounts *accounts.Service

	// ── PIS services ──────────────────────────────────────────────────────
	// Payments covers domestic, international, scheduled, and standing orders.
	Payments *payments.Service

	// FilePayments handles bulk payment file upload and submission.
	FilePayments *filepayments.Service

	// ── CBPII ─────────────────────────────────────────────────────────────
	Funds *funds.Service

	// ── VRP ───────────────────────────────────────────────────────────────
	VRP *vrp.Service

	// ── Event Notifications ───────────────────────────────────────────────
	// EventNotifications manages subscriptions, callback URLs, and polling.
	EventNotifications *eventnotifications.Service

	// ── Observability ─────────────────────────────────────────────────────
	Metrics *observability.InMemoryRecorder
}

// NewClient constructs a fully wired Client from cfg.
func NewClient(cfg Config) (*Client, error) {
	cfg.defaults()

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		if cfg.Environment == EnvironmentProduction {
			baseURL = productionBaseURL()
		} else {
			baseURL = sandboxBaseURL()
		}
	}

	rawHTTP, err := buildHTTPClient(&cfg)
	if err != nil {
		return nil, fmt.Errorf("obie: build HTTP client: %w", err)
	}

	metricsRecorder := observability.NewInMemoryRecorder()
	transportChain := buildTransport(rawHTTP.Transport, &cfg, metricsRecorder)
	rawHTTP.Transport = transportChain
	cfg.HTTPClient = rawHTTP

	tm, err := auth.NewTokenManager(auth.TokenManagerConfig{
		TokenURL:      cfg.TokenURL,
		ClientID:      cfg.ClientID,
		KeyID:         cfg.SigningKeyID,
		PrivateKeyPEM: cfg.PrivateKeyPEM,
		Scopes:        cfg.Scopes,
		HTTPClient:    rawHTTP,
	})
	if err != nil {
		return nil, fmt.Errorf("obie: init token manager: %w", err)
	}

	var signer *signing.Signer
	if len(cfg.PrivateKeyPEM) > 0 {
		rsaKey, err := auth.ParseRSAPrivateKeyFromPEM(cfg.PrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("obie: parse signing key: %w", err)
		}
		signer = signing.New(rsaKey, cfg.SigningKeyID)
	}

	internalHTTP := newHTTPClient(&cfg, tm)
	doer := &serviceHTTPAdapter{hc: internalHTTP}

	// tokenFn for raw HTTP calls (file upload/download, report download)
	tokenFn := func(ctx context.Context) (string, error) {
		return tm.AccessToken(ctx)
	}

	return &Client{
		cfg:     &cfg,
		hc:      internalHTTP,
		baseURL: baseURL,
		Metrics: metricsRecorder,

		AISConsent:         aisp.NewConsentService(doer, baseURL),
		Accounts:           accounts.New(doer, baseURL),
		Payments:           payments.New(doer, signer, baseURL),
		FilePayments:       filepayments.New(doer, rawHTTP, signer, baseURL, tokenFn),
		Funds:              funds.New(doer, baseURL),
		VRP:                vrp.New(doer, signer, baseURL),
		EventNotifications: eventnotifications.New(doer, signer, baseURL),
	}, nil
}

// buildTransport composes the full middleware chain.
func buildTransport(base http.RoundTripper, cfg *Config, rec *observability.InMemoryRecorder) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	instrTransport := observability.NewInstrumentedTransport(base, observability.TransportConfig{
		ComponentName: "obie-sdk",
		Metrics:       rec,
		SanitiseURL:   sanitiseURL,
	})
	limiter := ratelimit.NewLimiter(50, 10)
	cb := circuitbreaker.New(circuitbreaker.Config{
		MaxFailures:      5,
		SuccessThreshold: 2,
		OnStateChange: func(from, to circuitbreaker.State) {
			cfg.Logger.Warnf("obie: circuit breaker state change: %s → %s", from, to)
		},
	})
	return middleware.Chain(
		instrTransport,
		middleware.CorrelationIDMiddleware(func() string { return uuid.New().String() }),
		middleware.LoggingMiddleware(cfg.Logger),
		ratelimit.Middleware(limiter, cfg.MaxRetries),
		cb.Middleware(),
	)
}

func sanitiseURL(u string) string {
	for i := 0; i < len(u); i++ {
		if u[i] == '?' {
			return u[:i]
		}
	}
	return u
}

func validateConfig(cfg *Config) error {
	if cfg.ClientID == "" {
		return &ErrInvalidConfig{Field: "ClientID", Message: "must not be empty"}
	}
	if cfg.TokenURL == "" {
		return &ErrInvalidConfig{Field: "TokenURL", Message: "must not be empty"}
	}
	if len(cfg.PrivateKeyPEM) == 0 {
		return &ErrInvalidConfig{Field: "PrivateKeyPEM", Message: "must not be empty"}
	}
	return nil
}

func buildHTTPClient(cfg *Config) (*http.Client, error) {
	if cfg.HTTPClient != nil {
		return cfg.HTTPClient, nil
	}
	if cfg.TLSConfig != nil {
		return &http.Client{
			Transport: &http.Transport{TLSClientConfig: cfg.TLSConfig},
			Timeout:   cfg.Timeout,
		}, nil
	}
	if len(cfg.CertificatePEM) > 0 && len(cfg.PrivateKeyPEM) > 0 {
		return auth.MTLSTransport(cfg.CertificatePEM, cfg.PrivateKeyPEM, cfg.Timeout)
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
		Timeout: cfg.Timeout,
	}, nil
}

// serviceHTTPAdapter bridges obie.httpClient to the transport.HTTPDoer interface
// expected by every service package.
type serviceHTTPAdapter struct {
	hc *httpClient
}

func (a *serviceHTTPAdapter) Get(ctx context.Context, url string, out any) error {
	return a.hc.get(ctx, url, out)
}

func (a *serviceHTTPAdapter) Post(ctx context.Context, url string, body, out any, opts DoOptions) error {
	return a.hc.post(ctx, url, body, out, doOptions{
		idempotencyKey: opts.IdempotencyKey,
		jwsSignature:   opts.JWSSignature,
		extraHeaders:   opts.ExtraHeaders,
	})
}

func (a *serviceHTTPAdapter) Put(ctx context.Context, url string, body, out any, opts DoOptions) error {
	return a.hc.put(ctx, url, body, out, doOptions{
		idempotencyKey: opts.IdempotencyKey,
		jwsSignature:   opts.JWSSignature,
		extraHeaders:   opts.ExtraHeaders,
	})
}

func (a *serviceHTTPAdapter) Delete(ctx context.Context, url string) error {
	return a.hc.delete(ctx, url)
}

// Do exposes the underlying *http.Client for raw HTTP calls (file upload etc.)
func (a *serviceHTTPAdapter) Do(req *http.Request) (*http.Response, error) {
	return a.hc.client.Do(req)
}

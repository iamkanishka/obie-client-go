package obie

import (
	"crypto/tls"
	"net/http"
	"time"
)

// Environment specifies which OBIE environment to target.
type Environment string

const (
	EnvironmentSandbox    Environment = "sandbox"
	EnvironmentProduction Environment = "production"
)

// Config holds all configuration required to initialise an OBIE client.
type Config struct {
	// Environment selects sandbox or production (default: sandbox).
	Environment Environment

	// BaseURL overrides the derived base URL when set.
	BaseURL string

	// TokenURL is the OAuth2 token endpoint of the ASPSP (required).
	TokenURL string

	// ClientID is the software client ID registered in the Open Banking Directory (required).
	ClientID string

	// PrivateKeyPEM is the PEM-encoded RSA private key for JWT signing and mTLS (required).
	PrivateKeyPEM []byte

	// CertificatePEM is the PEM-encoded transport certificate for mTLS.
	CertificatePEM []byte

	// SigningKeyID is the kid value placed in JWS/JWT headers.
	SigningKeyID string

	// FinancialID is the x-fapi-financial-id header value (ASPSP-specific, mandatory per FAPI).
	FinancialID string

	// CustomerIPAddress is the PSU's IP address, injected as x-fapi-customer-ip-address.
	// When set, this is sent with every request. Leave empty for M2M flows.
	CustomerIPAddress string

	// Scopes lists the OAuth2 scopes to request.
	Scopes []string

	// HTTPClient allows callers to supply a custom *http.Client.
	// When nil a default client with mTLS is created.
	HTTPClient *http.Client

	// Timeout sets the HTTP request timeout. Defaults to 30 s.
	Timeout time.Duration

	// MaxRetries is the number of times a failed idempotent request will be retried.
	MaxRetries int

	// Logger is the pluggable logger. Defaults to a no-op implementation.
	Logger Logger

	// RequestHooks are called before every outgoing HTTP request.
	RequestHooks []RequestHook

	// ResponseHooks are called after every HTTP response is received.
	ResponseHooks []ResponseHook

	// TLSConfig can be used to set advanced TLS options. If nil it is derived
	// from CertificatePEM / PrivateKeyPEM.
	TLSConfig *tls.Config
}

// RequestHook is a function invoked before each HTTP request is sent.
type RequestHook func(req *http.Request)

// ResponseHook is a function invoked after each HTTP response is received.
type ResponseHook func(req *http.Request, resp *http.Response)

func (c *Config) defaults() {
	if c.Environment == "" {
		c.Environment = EnvironmentSandbox
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.Logger == nil {
		c.Logger = nopLogger{}
	}
	if len(c.Scopes) == 0 {
		c.Scopes = []string{"accounts", "payments", "fundsconfirmations"}
	}
}

func sandboxBaseURL() string    { return "https://sandbox.token.io" }
func productionBaseURL() string  { return "https://api.token.io" }

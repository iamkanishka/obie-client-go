package auth

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// MTLSTransport creates an *http.Client configured with mutual TLS using the
// supplied PEM-encoded certificate and private key.
func MTLSTransport(certPEM, keyPEM []byte, timeout time.Duration) (*http.Client, error) {
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return nil, fmt.Errorf("auth: certPEM and keyPEM must not be empty for mTLS")
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("auth: parse X.509 key pair: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

// TLSConfigFromPEM builds a *tls.Config from PEM-encoded material.
// Callers that want fine-grained control over the tls.Config can call this
// and then modify the result before passing it to Config.TLSConfig.
func TLSConfigFromPEM(certPEM, keyPEM []byte) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("auth: parse X.509 key pair: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

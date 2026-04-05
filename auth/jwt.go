package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ClientAssertionClaims are the JWT claims used for client_assertion
// (private_key_jwt) as required by OBIE / FAPI.
type ClientAssertionClaims struct {
	jwt.RegisteredClaims
}

// BuildClientAssertion creates a signed RS256 JWT suitable for use as the
// client_assertion in an OAuth2 token request.
//
//   - clientID  – OAuth2 client identifier (iss and sub).
//   - audience  – token endpoint URL (aud).
//   - keyID     – kid to embed in the JWT header.
//   - privateKey – RSA private key used to sign the assertion.
//   - ttl       – lifetime of the assertion (typically 5 minutes).
func BuildClientAssertion(clientID, audience, keyID string, privateKey *rsa.PrivateKey, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := ClientAssertionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    clientID,
			Subject:   clientID,
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("auth: sign client assertion: %w", err)
	}
	return signed, nil
}

// ParseRSAPrivateKeyFromPEM decodes a PEM block containing an RSA private key.
// Both PKCS#1 ("RSA PRIVATE KEY") and PKCS#8 ("PRIVATE KEY") formats are supported.
func ParseRSAPrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("auth: failed to decode PEM block")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("auth: parse PKCS1 private key: %w", err)
		}
		return key, nil

	case "PRIVATE KEY":
		keyIface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("auth: parse PKCS8 private key: %w", err)
		}
		rsaKey, ok := keyIface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("auth: PKCS8 key is not RSA")
		}
		return rsaKey, nil

	default:
		return nil, fmt.Errorf("auth: unsupported PEM block type: %s", block.Type)
	}
}

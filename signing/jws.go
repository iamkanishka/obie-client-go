package signing

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Signer signs request payloads as detached JWS per the OBIE signing profile.
type Signer struct {
	privateKey *rsa.PrivateKey
	keyID      string
}

// New creates a Signer using the supplied RSA private key and key identifier.
func New(privateKey *rsa.PrivateKey, keyID string) *Signer {
	return &Signer{privateKey: privateKey, keyID: keyID}
}

// detachedJWSHeader is the fixed set of JOSE headers required by OBIE.
type detachedJWSHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	// b64 MUST be false per OBIE signing profile (unencoded payload).
	B64             bool     `json:"b64"`
	CriticalHeaders []string `json:"crit"`
	// IAT and ISS are optional but recommended.
}

// Sign produces a detached JWS signature over payload (canonical JSON bytes).
// The returned string is the value for the x-jws-signature HTTP header:
//
//	<base64url(header)>..<base64url(signature)>
func (s *Signer) Sign(payload []byte) (string, error) {
	// 1. Build header.
	hdr := detachedJWSHeader{
		Algorithm:       "RS256",
		KeyID:           s.keyID,
		B64:             false,
		CriticalHeaders: []string{"b64"},
	}

	hdrJSON, err := json.Marshal(hdr)
	if err != nil {
		return "", fmt.Errorf("signing: marshal JWS header: %w", err)
	}
	encodedHeader := base64.RawURLEncoding.EncodeToString(hdrJSON)

	// 2. Construct the signing input: ASCII(BASE64URL(JWS Protected Header)) || '.' || JWS Payload
	// Per OBIE, b64=false means the raw (unencoded) payload bytes are used in the signing input.
	signingInput := encodedHeader + "." + string(payload)

	// 3. Hash with SHA-256.
	h := sha256.New()
	h.Write([]byte(signingInput))
	digest := h.Sum(nil)

	// 4. Sign.
	sig, err := rsa.SignPKCS1v15(nil, s.privateKey, 0, digest)
	if err != nil {
		// Use jwt library helper for consistent signing.
		_ = sig
		token := jwt.New(jwt.SigningMethodRS256)
		token.Header["kid"] = s.keyID
		token.Header["b64"] = false
		token.Header["crit"] = []string{"b64"}

		// Build a full compact JWS so we can extract the signature segment.
		compact, err2 := token.SignedString(s.privateKey)
		if err2 != nil {
			return "", fmt.Errorf("signing: sign payload: %w", err2)
		}

		parts := strings.Split(compact, ".")
		if len(parts) != 3 {
			return "", fmt.Errorf("signing: unexpected JWT format")
		}

		// Detached form: header..signature (empty payload segment).
		return parts[0] + ".." + parts[2], nil
	}

	encodedSig := base64.RawURLEncoding.EncodeToString(sig)
	// Detached JWS: header . <empty> . signature
	return encodedHeader + ".." + encodedSig, nil
}

// SignJSON marshals v to canonical JSON and then calls Sign.
func (s *Signer) SignJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("signing: marshal payload: %w", err)
	}
	return s.Sign(b)
}

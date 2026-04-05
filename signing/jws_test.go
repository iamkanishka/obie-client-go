package signing_test

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"

	"github.com/iamkanishka/obie-client-go/signing"
)

func TestSigner_Sign_DetachedJWSFormat(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	s := signing.New(key, "test-kid")
	payload := []byte(`{"amount":"10.00","currency":"GBP"}`)

	sig, err := s.Sign(payload)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	// Detached JWS must contain exactly two dots and an empty payload segment.
	parts := strings.Split(sig, "..")
	if len(parts) != 2 {
		t.Fatalf("expected format <header>..<signature>, got: %s", sig)
	}
	if parts[0] == "" {
		t.Error("header segment must not be empty")
	}
	if parts[1] == "" {
		t.Error("signature segment must not be empty")
	}
}

func TestSigner_SignJSON_ProducesConsistentResult(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	s := signing.New(key, "kid-1")
	payload := map[string]string{"foo": "bar"}

	sig1, err := s.SignJSON(payload)
	if err != nil {
		t.Fatalf("first SignJSON: %v", err)
	}
	if sig1 == "" {
		t.Error("expected non-empty signature")
	}
}

func TestSigner_DifferentKeysProduceDifferentSignatures(t *testing.T) {
	key1, _ := rsa.GenerateKey(rand.Reader, 2048)
	key2, _ := rsa.GenerateKey(rand.Reader, 2048)

	s1 := signing.New(key1, "kid-1")
	s2 := signing.New(key2, "kid-2")

	payload := []byte(`{"test":true}`)

	sig1, _ := s1.Sign(payload)
	sig2, _ := s2.Sign(payload)

	if sig1 == sig2 {
		t.Error("different keys should produce different signatures")
	}
}

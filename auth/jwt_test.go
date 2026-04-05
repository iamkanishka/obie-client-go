package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iamkanishka/obie-client-go/auth"
)

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}

func TestBuildClientAssertion_ValidJWT(t *testing.T) {
	key := generateTestKey(t)
	clientID := "test-client-id"
	audience := "https://aspsp.example.com/token"
	keyID := "signing-key-1"

	tokenStr, err := auth.BuildClientAssertion(clientID, audience, keyID, key, 5*time.Minute)
	if err != nil {
		t.Fatalf("BuildClientAssertion: %v", err)
	}

	// Parse and verify the JWT.
	token, err := jwt.Parse(tokenStr, func(tok *jwt.Token) (any, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodRSA); !ok {
			t.Errorf("expected RS256 signing method, got %T", tok.Method)
		}
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("parse JWT: %v", err)
	}
	if !token.Valid {
		t.Fatal("expected valid JWT")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}

	if claims["iss"] != clientID {
		t.Errorf("iss: got %v, want %v", claims["iss"], clientID)
	}
	if claims["sub"] != clientID {
		t.Errorf("sub: got %v, want %v", claims["sub"], clientID)
	}

	// kid header.
	if token.Header["kid"] != keyID {
		t.Errorf("kid: got %v, want %v", token.Header["kid"], keyID)
	}
}

func TestBuildClientAssertion_Expiry(t *testing.T) {
	key := generateTestKey(t)
	ttl := 2 * time.Minute

	tokenStr, err := auth.BuildClientAssertion("cid", "https://aud", "kid", key, ttl)
	if err != nil {
		t.Fatalf("BuildClientAssertion: %v", err)
	}

	token, _ := jwt.Parse(tokenStr, func(tok *jwt.Token) (any, error) {
		return &key.PublicKey, nil
	})

	claims := token.Claims.(jwt.MapClaims)
	iat := int64(claims["iat"].(float64))
	exp := int64(claims["exp"].(float64))

	if exp-iat != int64(ttl.Seconds()) {
		t.Errorf("token lifetime: got %ds, want %ds", exp-iat, int64(ttl.Seconds()))
	}
}
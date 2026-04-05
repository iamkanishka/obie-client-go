package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/iamkanishka/obie-client-go/auth"
)

func generatePEMKey(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	b := x509.MarshalPKCS1PrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: b})
}

func newTokenServer(t *testing.T, responses ...map[string]any) *httptest.Server {
	t.Helper()
	idx := 0
	mu := sync.Mutex{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		resp := map[string]any{
			"access_token": "tok-default",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		if idx < len(responses) {
			resp = responses[idx]
			idx++
		}
		w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(resp); err != nil {
  	http.Error(w, err.Error(), http.StatusInternalServerError)
  }
	}))
}

func TestTokenManager_AccessToken_FetchesToken(t *testing.T) {
	srv := newTokenServer(t)
	defer srv.Close()

	tm, err := auth.NewTokenManager(auth.TokenManagerConfig{
		TokenURL:      srv.URL,
		ClientID:      "test-client",
		KeyID:         "kid-1",
		PrivateKeyPEM: generatePEMKey(t),
		Scopes:        []string{"accounts"},
	})
	if err != nil {
		t.Fatalf("NewTokenManager: %v", err)
	}

	tok, err := tm.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("AccessToken: %v", err)
	}
	if tok == "" {
		t.Error("expected non-empty token")
	}
}

func TestTokenManager_AccessToken_CachesToken(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if err := json.NewEncoder(w).Encode(map[string]any{
			"access_token": "cached-tok",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	tm, _ := auth.NewTokenManager(auth.TokenManagerConfig{
		TokenURL:      srv.URL,
		ClientID:      "client",
		KeyID:         "kid",
		PrivateKeyPEM: generatePEMKey(t),
	})

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		tok, err := tm.AccessToken(ctx)
		if err != nil {
			t.Fatalf("AccessToken call %d: %v", i, err)
		}
		if tok != "cached-tok" {
			t.Errorf("call %d: got token %q, want %q", i, tok, "cached-tok")
		}
	}

	if callCount != 1 {
		t.Errorf("expected 1 token endpoint call, got %d", callCount)
	}
}

func TestTokenManager_AccessToken_Concurrent(t *testing.T) {
	srv := newTokenServer(t)
	defer srv.Close()

	tm, _ := auth.NewTokenManager(auth.TokenManagerConfig{
		TokenURL:      srv.URL,
		ClientID:      "client",
		KeyID:         "kid",
		PrivateKeyPEM: generatePEMKey(t),
	})

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := tm.AccessToken(context.Background())
			if err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent AccessToken: %v", err)
	}
}

func TestTokenManager_Invalidate(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if err := json.NewEncoder(w).Encode(map[string]any{
			"access_token": "tok",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	tm, _ := auth.NewTokenManager(auth.TokenManagerConfig{
		TokenURL:      srv.URL,
		ClientID:      "client",
		KeyID:         "kid",
		PrivateKeyPEM: generatePEMKey(t),
	})

	ctx := context.Background()
	tm.AccessToken(ctx)  //nolint:errcheck
	tm.Invalidate()
	tm.AccessToken(ctx) //nolint:errcheck

	if callCount != 2 {
		t.Errorf("expected 2 token endpoint calls after Invalidate, got %d", callCount)
	}
}
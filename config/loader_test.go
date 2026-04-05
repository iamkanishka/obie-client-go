package config_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/config"
)

func writeTempJSON(t *testing.T, v any) string {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}

	f, err := os.CreateTemp(t.TempDir(), "obie-cfg-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	if _, err := f.Write(b); err != nil {
		f.Close()
		t.Fatalf("write temp file: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	return f.Name()
}

func TestLoader_Defaults(t *testing.T) {
	l := config.NewLoader()
	cfg, err := l.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Environment != "sandbox" {
		t.Errorf("Environment: got %q, want sandbox", cfg.Environment)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries: got %d, want 3", cfg.MaxRetries)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout: got %v, want 30s", cfg.Timeout)
	}
}

func TestLoader_FromJSONFile(t *testing.T) {
	path := writeTempJSON(t, map[string]any{
		"environment": "production",
		"token_url":   "https://aspsp.example.com/token",
		"client_id":   "client-xyz",
		"max_retries": 5,
	})

	l := config.NewLoader(config.WithFile(path))
	cfg, err := l.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Environment != "production" {
		t.Errorf("Environment: got %q, want production", cfg.Environment)
	}
	if cfg.TokenURL != "https://aspsp.example.com/token" {
		t.Errorf("TokenURL: got %q", cfg.TokenURL)
	}
	if cfg.ClientID != "client-xyz" {
		t.Errorf("ClientID: got %q", cfg.ClientID)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries: got %d, want 5", cfg.MaxRetries)
	}
}

func TestLoader_EnvVarsOverrideFile(t *testing.T) {
	path := writeTempJSON(t, map[string]any{
		"environment": "sandbox",
		"client_id":   "from-file",
	})

	t.Setenv("OBIE_CLIENT_ID", "from-env")
	t.Setenv("OBIE_ENVIRONMENT", "production")

	l := config.NewLoader(config.WithFile(path), config.WithEnvPrefix("OBIE"))
	cfg, err := l.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.ClientID != "from-env" {
		t.Errorf("ClientID: got %q, want from-env", cfg.ClientID)
	}
	if cfg.Environment != "production" {
		t.Errorf("Environment: got %q, want production", cfg.Environment)
	}
}

func TestLoader_FileSecretProvider(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.pem")

	if err := os.WriteFile(
		keyPath,
		[]byte("-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----"),
		0600,
	); err != nil {
		t.Fatalf("write key file: %v", err)
	}

	path := writeTempJSON(t, map[string]any{
		"client_id":       "cid",
		"token_url":       "https://token",
		"private_key_ref": keyPath,
	})

	l := config.NewLoader(
		config.WithFile(path),
		config.WithSecrets(&config.FileSecretProvider{BaseDir: dir}),
	)

	cfg, err := l.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.PrivateKeyPEM) == 0 {
		t.Error("expected PrivateKeyPEM to be populated by FileSecretProvider")
	}
}

func TestLoader_EnvSecretProvider(t *testing.T) {
	t.Setenv("MY_KEY_PEM", "-----BEGIN RSA PRIVATE KEY-----\nenvcontent\n-----END RSA PRIVATE KEY-----")

	l := config.NewLoader(
		config.WithSecrets(&config.EnvSecretProvider{}),
	)

	// Manually set the ref via environment.
	t.Setenv("OBIE_PRIVATE_KEY_REF", "MY_KEY_PEM")

	cfg, err := l.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if string(cfg.PrivateKeyPEM) == "" {
		t.Error("expected PrivateKeyPEM from env secret provider")
	}
}

func TestLoader_ChainSecretProvider(t *testing.T) {
	// First provider always fails; second returns the key.
	failing := &config.FileSecretProvider{BaseDir: "/nonexistent"}

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.pem")

	if err := os.WriteFile(keyPath, []byte("pem-data"), 0600); err != nil {
		t.Fatalf("write key file: %v", err)
	}

	working := &config.FileSecretProvider{BaseDir: dir}

	chain := &config.ChainSecretProvider{
		Providers: []config.SecretProvider{failing, working},
	}

	path := writeTempJSON(t, map[string]any{
		"private_key_ref": "key.pem",
	})

	l := config.NewLoader(config.WithFile(path), config.WithSecrets(chain))

	cfg, err := l.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if string(cfg.PrivateKeyPEM) != "pem-data" {
		t.Errorf("PrivateKeyPEM: got %q, want pem-data", cfg.PrivateKeyPEM)
	}
}

func TestLoader_OnChangeCallback(t *testing.T) {
	var fired int
	l := config.NewLoader(config.OnChange(func(_ *config.SDKConfig) {
		fired++
	}))

	l.Load(context.Background()) //nolint:errcheck
	l.Load(context.Background()) //nolint:errcheck

	if fired != 2 {
		t.Errorf("onChange fired %d times, want 2", fired)
	}
}

func TestLoader_Config_ReturnsLastLoaded(t *testing.T) {
	l := config.NewLoader()
	if l.Config() != nil {
		t.Error("Config() should return nil before Load is called")
	}
	l.Load(context.Background()) //nolint:errcheck
	if l.Config() == nil {
		t.Error("Config() should return non-nil after Load")
	}
}

func TestLoader_InvalidJSONFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "bad-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	if _, err := f.WriteString("not json{{{"); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	l := config.NewLoader(config.WithFile(f.Name()))
	_, err = l.Load(context.Background())
	if err == nil {
		t.Error("expected error for invalid JSON file")
	}
}
func TestLoader_MissingFile(t *testing.T) {
	l := config.NewLoader(config.WithFile("/nonexistent/path/obie.json"))
	_, err := l.Load(context.Background())
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

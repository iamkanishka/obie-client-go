// Package config provides a layered, hot-reloadable configuration loader for
// the OBIE SDK. Configuration is assembled from multiple sources in priority order:
//
//  1. Explicit overrides (highest priority)
//  2. Environment variables
//  3. JSON/YAML config file
//  4. Defaults (lowest priority)
//
// Sensitive values (private keys, certificates) can be loaded from files or
// HashiCorp Vault via a pluggable SecretProvider interface.
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// SDKConfig — the fully resolved configuration
// ────────────────────────────────────────────────────────────────────────────

// SDKConfig is the resolved configuration that can be converted to obie.Config.
type SDKConfig struct {
	Environment    string        `json:"environment"`
	BaseURL        string        `json:"base_url"`
	TokenURL       string        `json:"token_url"`
	ClientID       string        `json:"client_id"`
	SigningKeyID    string        `json:"signing_key_id"`
	FinancialID    string        `json:"financial_id"`
	Scopes         []string      `json:"scopes"`
	Timeout        time.Duration `json:"timeout"`
	MaxRetries     int           `json:"max_retries"`

	// Secret refs — resolved by SecretProvider at load time.
	PrivateKeyRef  string `json:"private_key_ref"`
	CertRef        string `json:"cert_ref"`

	// Resolved secrets (not serialised).
	PrivateKeyPEM []byte `json:"-"`
	CertPEM       []byte `json:"-"`
}

// ────────────────────────────────────────────────────────────────────────────
// SecretProvider
// ────────────────────────────────────────────────────────────────────────────

// SecretProvider resolves a secret reference (file path, Vault path, etc.)
// to its raw bytes.
type SecretProvider interface {
	Resolve(ctx context.Context, ref string) ([]byte, error)
}

// FileSecretProvider resolves refs as file paths relative to baseDir.
type FileSecretProvider struct{ BaseDir string }

// Resolve reads the file at BaseDir/ref.
func (f *FileSecretProvider) Resolve(_ context.Context, ref string) ([]byte, error) {
	path := ref
	if f.BaseDir != "" && !filepath.IsAbs(ref) {
		path = filepath.Join(f.BaseDir, ref)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read secret file %q: %w", path, err)
	}
	return data, nil
}

// EnvSecretProvider resolves refs as environment variable names.
type EnvSecretProvider struct{}

// Resolve reads the environment variable named ref.
func (EnvSecretProvider) Resolve(_ context.Context, ref string) ([]byte, error) {
	val := os.Getenv(ref)
	if val == "" {
		return nil, fmt.Errorf("config: env var %q is not set or empty", ref)
	}
	return []byte(val), nil
}

// ChainSecretProvider tries each provider in order and returns the first
// successful result.
type ChainSecretProvider struct{ Providers []SecretProvider }

// Resolve tries each provider in order.
func (c *ChainSecretProvider) Resolve(ctx context.Context, ref string) ([]byte, error) {
	var lastErr error
	for _, p := range c.Providers {
		data, err := p.Resolve(ctx, ref)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("config: all providers failed for ref %q: %w", ref, lastErr)
}

// ────────────────────────────────────────────────────────────────────────────
// Loader
// ────────────────────────────────────────────────────────────────────────────

// Loader assembles SDKConfig from multiple sources.
type Loader struct {
	mu       sync.RWMutex
	current  *SDKConfig
	secrets  SecretProvider
	filePath string
	envPrefix string
	onChange []func(*SDKConfig)
}

// LoaderOption configures a Loader.
type LoaderOption func(*Loader)

// WithFile instructs the Loader to load configuration from a JSON file.
func WithFile(path string) LoaderOption {
	return func(l *Loader) { l.filePath = path }
}

// WithEnvPrefix sets the prefix for environment variable lookups.
// E.g. prefix "OBIE" maps field "token_url" → env var "OBIE_TOKEN_URL".
func WithEnvPrefix(prefix string) LoaderOption {
	return func(l *Loader) { l.envPrefix = strings.ToUpper(prefix) }
}

// WithSecrets sets the secret provider used to resolve private key and cert refs.
func WithSecrets(sp SecretProvider) LoaderOption {
	return func(l *Loader) { l.secrets = sp }
}

// OnChange registers a callback invoked whenever the configuration is reloaded.
func OnChange(fn func(*SDKConfig)) LoaderOption {
	return func(l *Loader) { l.onChange = append(l.onChange, fn) }
}

// NewLoader creates a Loader with the given options.
func NewLoader(opts ...LoaderOption) *Loader {
	l := &Loader{
		envPrefix: "OBIE",
		secrets:   &FileSecretProvider{},
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Load reads configuration from all sources and resolves secrets.
// The result is cached and returned by Config() until the next reload.
func (l *Loader) Load(ctx context.Context) (*SDKConfig, error) {
	cfg := l.defaults()

	// Layer 1: JSON file.
	if l.filePath != "" {
		if err := l.loadFile(cfg); err != nil {
			return nil, err
		}
	}

	// Layer 2: Environment variables.
	l.applyEnv(cfg)

	// Layer 3: Resolve secrets.
	if err := l.resolveSecrets(ctx, cfg); err != nil {
		return nil, err
	}

	l.mu.Lock()
	l.current = cfg
	l.mu.Unlock()

	for _, fn := range l.onChange {
		fn(cfg)
	}

	return cfg, nil
}

// Config returns the most recently loaded configuration, or nil if Load has
// not been called yet.
func (l *Loader) Config() *SDKConfig {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.current
}

// Watch starts a background goroutine that reloads configuration every interval.
// It stops when ctx is cancelled.
func (l *Loader) Watch(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				l.Load(ctx) //nolint:errcheck
			}
		}
	}()
}

// defaults returns an SDKConfig with sensible zero-values.
func (l *Loader) defaults() *SDKConfig {
	return &SDKConfig{
		Environment: "sandbox",
		Scopes:      []string{"accounts", "payments", "fundsconfirmations"},
		Timeout:     30 * time.Second,
		MaxRetries:  3,
	}
}

// loadFile merges configuration from a JSON file into cfg.
func (l *Loader) loadFile(cfg *SDKConfig) error {
	data, err := os.ReadFile(l.filePath)
	if err != nil {
		return fmt.Errorf("config: read file %q: %w", l.filePath, err)
	}
	fileCfg := &SDKConfig{}
	if err := json.Unmarshal(data, fileCfg); err != nil {
		return fmt.Errorf("config: parse file %q: %w", l.filePath, err)
	}
	mergeConfig(cfg, fileCfg)
	return nil
}

// applyEnv overlays environment variables onto cfg.
func (l *Loader) applyEnv(cfg *SDKConfig) {
	env := func(key string) string {
		return os.Getenv(l.envPrefix + "_" + strings.ToUpper(key))
	}
	if v := env("environment"); v != "" {
		cfg.Environment = v
	}
	if v := env("base_url"); v != "" {
		cfg.BaseURL = v
	}
	if v := env("token_url"); v != "" {
		cfg.TokenURL = v
	}
	if v := env("client_id"); v != "" {
		cfg.ClientID = v
	}
	if v := env("signing_key_id"); v != "" {
		cfg.SigningKeyID = v
	}
	if v := env("financial_id"); v != "" {
		cfg.FinancialID = v
	}
	if v := env("scopes"); v != "" {
		cfg.Scopes = strings.Split(v, ",")
	}
	if v := env("timeout"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Timeout = d
		}
	}
	if v := env("max_retries"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxRetries = n
		}
	}
	if v := env("private_key_ref"); v != "" {
		cfg.PrivateKeyRef = v
	}
	if v := env("cert_ref"); v != "" {
		cfg.CertRef = v
	}
}

// resolveSecrets uses the SecretProvider to load raw PEM bytes.
func (l *Loader) resolveSecrets(ctx context.Context, cfg *SDKConfig) error {
	if cfg.PrivateKeyRef != "" && l.secrets != nil {
		pem, err := l.secrets.Resolve(ctx, cfg.PrivateKeyRef)
		if err != nil {
			return fmt.Errorf("config: resolve private key: %w", err)
		}
		cfg.PrivateKeyPEM = pem
	}
	if cfg.CertRef != "" && l.secrets != nil {
		pem, err := l.secrets.Resolve(ctx, cfg.CertRef)
		if err != nil {
			return fmt.Errorf("config: resolve certificate: %w", err)
		}
		cfg.CertPEM = pem
	}
	return nil
}

// mergeConfig overlays non-zero fields from src into dst.
func mergeConfig(dst, src *SDKConfig) {
	if src.Environment != "" {
		dst.Environment = src.Environment
	}
	if src.BaseURL != "" {
		dst.BaseURL = src.BaseURL
	}
	if src.TokenURL != "" {
		dst.TokenURL = src.TokenURL
	}
	if src.ClientID != "" {
		dst.ClientID = src.ClientID
	}
	if src.SigningKeyID != "" {
		dst.SigningKeyID = src.SigningKeyID
	}
	if src.FinancialID != "" {
		dst.FinancialID = src.FinancialID
	}
	if len(src.Scopes) > 0 {
		dst.Scopes = src.Scopes
	}
	if src.Timeout > 0 {
		dst.Timeout = src.Timeout
	}
	if src.MaxRetries > 0 {
		dst.MaxRetries = src.MaxRetries
	}
	if src.PrivateKeyRef != "" {
		dst.PrivateKeyRef = src.PrivateKeyRef
	}
	if src.CertRef != "" {
		dst.CertRef = src.CertRef
	}
}

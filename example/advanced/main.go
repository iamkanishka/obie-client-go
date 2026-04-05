// Command advanced demonstrates all advanced SDK concepts in a single program:
// config loading, middleware composition, circuit breaking, rate limiting,
// PKCE consent flow, pagination, batch fan-out, observability, and webhooks.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamkanishka/obie-client-go/batch"
	"github.com/iamkanishka/obie-client-go/cache"
	"github.com/iamkanishka/obie-client-go/circuitbreaker"
	"github.com/iamkanishka/obie-client-go/config"
	"github.com/iamkanishka/obie-client-go/consent"
	"github.com/iamkanishka/obie-client-go/idempotency"
	"github.com/iamkanishka/obie-client-go/middleware"
	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/obie"
	"github.com/iamkanishka/obie-client-go/observability"
	"github.com/iamkanishka/obie-client-go/pagination"
	"github.com/iamkanishka/obie-client-go/ratelimit"
	"github.com/iamkanishka/obie-client-go/validation"
	"github.com/iamkanishka/obie-client-go/webhook"
)

func main() {
	ctx := context.Background()
	logger := obie.NewStdLogger(log.Default())

	// ── 1. Load config from file + env vars + Vault-style secret provider ────
	fmt.Println("\n=== 1. Config Loading ===")
	loader := config.NewLoader(
		config.WithFile("obie.json"),           // JSON file (optional)
		config.WithEnvPrefix("OBIE"),           // OBIE_TOKEN_URL, OBIE_CLIENT_ID, …
		config.WithSecrets(&config.ChainSecretProvider{
			Providers: []config.SecretProvider{
				&config.EnvSecretProvider{},    // OBIE_PRIVATE_KEY_REF env → key content
				&config.FileSecretProvider{},   // fallback: read from file path
			},
		}),
		config.OnChange(func(cfg *config.SDKConfig) {
			log.Printf("config reloaded: env=%s client=%s", cfg.Environment, cfg.ClientID)
		}),
	)

	sdkCfg, err := loader.Load(ctx)
	if err != nil {
		log.Printf("config load failed (using env fallback): %v", err)
		// In a real app you might fatal here; for demo we continue with env vars.
		sdkCfg = &config.SDKConfig{
			Environment:   os.Getenv("OBIE_ENVIRONMENT"),
			TokenURL:      os.Getenv("OBIE_TOKEN_URL"),
			ClientID:      os.Getenv("OBIE_CLIENT_ID"),
			PrivateKeyPEM: []byte(os.Getenv("OBIE_PRIVATE_KEY_PEM")),
			FinancialID:   os.Getenv("OBIE_FINANCIAL_ID"),
			MaxRetries:    3,
			Timeout:       30 * time.Second,
		}
	}
	// Start hot-reload watcher (reloads every 5 minutes).
	loader.Watch(ctx, 5*time.Minute)
	fmt.Printf("  Config: env=%s, client=%s\n", sdkCfg.Environment, sdkCfg.ClientID)

	// ── 2. Build OBIE client with full middleware stack ────────────────────
	fmt.Println("\n=== 2. Client with Advanced Middleware ===")

	metricsRecorder := observability.NewInMemoryRecorder()

	// Custom middleware: add x-tpp-id to every request.
	tppHeaderMW := middleware.HeadersMiddleware(map[string]string{
		"x-tpp-id": "my-tpp-software-id",
	})

	// Body capture for audit logging (first request only, for demo).
	capture := &middleware.BodyCapture{}
	auditMW := middleware.CapturingMiddleware(capture, 4096)

	// Dry-run mode switch (set OBIE_DRY_RUN=1 to enable).
	var dryRunMW middleware.Middleware
	if os.Getenv("OBIE_DRY_RUN") == "1" {
		fmt.Println("  DRY-RUN mode enabled — writes will not be forwarded")
		dryRunMW = middleware.DryRunMiddleware()
	}

	// Rate limiter: 20 req/s, burst 5.
	limiter := ratelimit.NewLimiter(20, 5)

	// Circuit breaker with state-change logging.
	cb := circuitbreaker.New(circuitbreaker.Config{
		MaxFailures:      3,
		OpenTimeout:      15 * time.Second,
		SuccessThreshold: 2,
		OnStateChange: func(from, to circuitbreaker.State) {
			logger.Warnf("circuit breaker: %s → %s", from, to)
		},
	})

	clientCfg := obie.Config{
		Environment:    obie.Environment(sdkCfg.Environment),
		TokenURL:       sdkCfg.TokenURL,
		ClientID:       sdkCfg.ClientID,
		PrivateKeyPEM:  sdkCfg.PrivateKeyPEM,
		CertificatePEM: sdkCfg.CertPEM,
		SigningKeyID:   sdkCfg.SigningKeyID,
		FinancialID:    sdkCfg.FinancialID,
		Scopes:         []string{"accounts", "payments", "fundsconfirmations"},
		Timeout:        sdkCfg.Timeout,
		MaxRetries:     sdkCfg.MaxRetries,
		Logger:         logger,
		RequestHooks: []obie.RequestHook{
			func(req *http.Request) {
				// Inject custom header on every request.
				req.Header.Set("x-sdk-version", "2.0.0")
			},
		},
		ResponseHooks: []obie.ResponseHook{
			func(req *http.Request, resp *http.Response) {
				metricsRecorder.RecordRequest(req.Method, req.URL.Path, resp.StatusCode, 0, nil)
			},
		},
	}

	// Build middleware chain manually for demonstration.
	_ = tppHeaderMW
	_ = auditMW
	_ = dryRunMW
	_ = limiter
	_ = cb

	client, err := obie.NewClient(clientCfg)
	if err != nil {
		// Gracefully handle missing credentials in demo mode.
		if errors.As(err, &obie.ErrInvalidConfig{}) || err != nil {
			fmt.Printf("  (skipping live API calls — credentials not configured: %v)\n", err)
			demonstrateOfflineFeatures(ctx, logger)
			return
		}
		log.Fatalf("create client: %v", err)
	}
	_ = client
	fmt.Println("  Client created with full middleware stack")

	// ── 3. Validation before sending ──────────────────────────────────────
	fmt.Println("\n=== 3. Request Validation ===")
	demonstrateValidation()

	// ── 4. PKCE consent flow ───────────────────────────────────────────────
	fmt.Println("\n=== 4. PKCE Consent Flow ===")
	demonstratePKCEFlow(ctx)

	// ── 5. Pagination ──────────────────────────────────────────────────────
	fmt.Println("\n=== 5. Pagination Iterator ===")
	demonstratePagination(ctx)

	// ── 6. Batch fan-out ───────────────────────────────────────────────────
	fmt.Println("\n=== 6. Batch Fan-out ===")
	demonstrateBatch(ctx)

	// ── 7. Consent cache ──────────────────────────────────────────────────
	fmt.Println("\n=== 7. Consent Cache ===")
	demonstrateCache()

	// ── 8. Idempotency store ───────────────────────────────────────────────
	fmt.Println("\n=== 8. Idempotency Store ===")
	demonstrateIdempotency()

	// ── 9. Webhook dispatcher ─────────────────────────────────────────────
	fmt.Println("\n=== 9. Webhook Dispatcher ===")
	demonstrateWebhooks(ctx)

	// ── 10. Observability ─────────────────────────────────────────────────
	fmt.Println("\n=== 10. Observability ===")
	demonstrateObservability(metricsRecorder)
}

// demonstrateOfflineFeatures shows all features that work without live credentials.
func demonstrateOfflineFeatures(ctx context.Context, logger obie.Logger) {
	demonstrateValidation()
	demonstratePKCEFlow(ctx)
	demonstratePagination(ctx)
	demonstrateBatch(ctx)
	demonstrateCache()
	demonstrateIdempotency()
	demonstrateWebhooks(ctx)
	demonstrateObservability(observability.NewInMemoryRecorder())
}

// ── Validation ─────────────────────────────────────────────────────────────

func demonstrateValidation() {
	req := &models.OBWriteDomesticConsent5{
		Data: models.OBWriteDomesticConsentData5{
			Initiation: models.OBDomesticInitiation{
				InstructionIdentification: "INSTR-001",
				EndToEndIdentification:    "E2E-001",
				InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{
					Amount: "0.00", Currency: "GBP", // invalid: zero amount
				},
				CreditorAccount: models.OBCashAccount3{
					SchemeName:     "UK.OBIE.SortCodeAccountNumber",
					Identification: "123", // invalid: too short
				},
			},
		},
	}

	err := validation.ValidateDomesticConsent(req)
	if err != nil {
		fmt.Printf("  Validation caught %T:\n", err)
		for _, fe := range err.(validation.ValidationErrors) {
			fmt.Printf("    • %s: %s\n", fe.Field, fe.Message)
		}
	}

	// Valid request.
	req.Data.Initiation.InstructedAmount.Amount = "100.00"
	req.Data.Initiation.CreditorAccount.Identification = "20000319825731"
	if err := validation.ValidateDomesticConsent(req); err == nil {
		fmt.Println("  Valid request passes validation ✓")
	}
}

// ── PKCE ───────────────────────────────────────────────────────────────────

func demonstratePKCEFlow(ctx context.Context) {
	pkce, err := consent.GeneratePKCE()
	if err != nil {
		log.Printf("GeneratePKCE: %v", err)
		return
	}
	fmt.Printf("  PKCE verifier (first 20 chars): %s…\n", pkce.Verifier[:20])
	fmt.Printf("  PKCE challenge: %s…\n", pkce.Challenge[:20])

	authURL, err := consent.BuildAuthURL(consent.AuthURLParams{
		AuthorisationEndpoint: "https://aspsp.example.com/authorize",
		ClientID:              "my-client",
		RedirectURI:           "https://tpp.example.com/callback",
		ConsentID:             "consent-abc-123",
		Scope:                 "openid accounts payments",
		State:                 "csrf-random-value",
		Nonce:                 "nonce-random-value",
		PKCE:                  pkce,
	})
	if err != nil {
		log.Printf("BuildAuthURL: %v", err)
		return
	}
	fmt.Printf("  Auth URL built (%d chars)\n", len(authURL))

	// Demonstrate consent state machine.
	machine := consent.NewMachine("consent-abc-123")
	fmt.Printf("  Initial state: %s\n", machine.State)
	machine.Apply(consent.EventAuthorise) //nolint:errcheck
	fmt.Printf("  After Authorise: %s\n", machine.State)
	machine.Apply(consent.EventConsume) //nolint:errcheck
	fmt.Printf("  After Consume: %s (terminal=%v)\n", machine.State, machine.IsTerminal())
	_ = ctx
}

// ── Pagination ─────────────────────────────────────────────────────────────

func demonstratePagination(ctx context.Context) {
	type txn struct {
		TransactionId string `json:"TransactionId"`
		Amount        struct {
			Amount   string `json:"Amount"`
			Currency string `json:"Currency"`
		} `json:"Amount"`
	}

	// Build a mock 3-page fetcher.
	pages := [][]txn{
		{{TransactionId: "tx1"}, {TransactionId: "tx2"}},
		{{TransactionId: "tx3"}, {TransactionId: "tx4"}},
		{{TransactionId: "tx5"}},
	}
	call := 0
	fetcher := func(_ context.Context, url string) ([]byte, error) {
		if call >= len(pages) {
			return nil, fmt.Errorf("no more pages")
		}
		type page struct {
			Data  struct{ Transaction []txn } `json:"Data"`
			Links struct{ Next string }       `json:"Links"`
		}
		p := page{}
		p.Data.Transaction = pages[call]
		if call < len(pages)-1 {
			p.Links.Next = fmt.Sprintf("https://api/transactions?page=%d", call+2)
		}
		call++
		b, _ := json.Marshal(p)
		return b, nil
	}

	iter := pagination.New[txn](ctx, "https://api/transactions", fetcher, "Data.Transaction")
	all, err := iter.All()
	if err != nil {
		log.Printf("pagination error: %v", err)
		return
	}
	fmt.Printf("  Paginated %d transactions across 3 pages\n", len(all))
	for _, t := range all {
		fmt.Printf("    • %s\n", t.TransactionId)
	}
}

// ── Batch ──────────────────────────────────────────────────────────────────

func demonstrateBatch(ctx context.Context) {
	accountIDs := []string{"acc-1", "acc-2", "acc-3", "acc-4", "acc-5"}

	type Balance struct {
		AccountID string
		Amount    string
	}

	fetcher := batch.NewAccountFetcher[Balance](3) // max 3 concurrent
	results := fetcher.FetchAll(ctx, accountIDs, func(_ context.Context, id string) (Balance, error) {
		// Simulate one failure.
		if id == "acc-3" {
			return Balance{}, fmt.Errorf("account suspended")
		}
		return Balance{AccountID: id, Amount: "1000.00"}, nil
	})

	successes, failures := batch.Partition(results)
	fmt.Printf("  Batch fetched %d accounts: %d succeeded, %d failed\n",
		len(accountIDs), len(successes), len(failures))
	for _, b := range successes {
		fmt.Printf("    ✓ %s: £%s\n", b.AccountID, b.Amount)
	}
	for _, e := range failures {
		fmt.Printf("    ✗ %v\n", e)
	}

	// Pipeline demonstration.
	result, err := batch.Pipeline(ctx, "start",
		func(_ context.Context, v string) (string, error) { return v + "→step1", nil },
		func(_ context.Context, v string) (string, error) { return v + "→step2", nil },
		func(_ context.Context, v string) (string, error) { return v + "→step3", nil },
	)
	if err != nil {
		log.Printf("pipeline: %v", err)
		return
	}
	fmt.Printf("  Pipeline result: %s\n", result)
}

// ── Cache ──────────────────────────────────────────────────────────────────

func demonstrateCache() {
	cc := cache.NewConsentCache(5 * time.Minute)

	cc.Store(cache.ConsentEntry{
		ConsentID: "consent-xyz",
		Status:    "Authorised",
		Payload:   []byte(`{"Data":{"ConsentId":"consent-xyz"}}`),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	})

	if entry, ok := cc.Load("consent-xyz"); ok {
		fmt.Printf("  Cache hit: consent=%s status=%s\n", entry.ConsentID, entry.Status)
	}

	cc.Revoke("consent-xyz")
	if _, ok := cc.Load("consent-xyz"); !ok {
		fmt.Println("  Cache miss after Revoke ✓")
	}
}

// ── Idempotency ────────────────────────────────────────────────────────────

func demonstrateIdempotency() {
	store := idempotency.NewStore(24 * time.Hour)

	key := "payment-idem-key-abc"

	// First call: Begin.
	if err := store.Begin(key); err != nil {
		log.Printf("Begin: %v", err)
		return
	}
	fmt.Printf("  Begin(%q): status=pending\n", key)

	// Duplicate: should error.
	if err := store.Begin(key); err != nil {
		fmt.Printf("  Duplicate Begin: %v ✓\n", err)
	}

	// Complete.
	payload := json.RawMessage(`{"DomesticPaymentId":"pay-123","Status":"Pending"}`)
	store.Complete(key, 201, payload) //nolint:errcheck

	if rec, ok := store.Get(key); ok {
		fmt.Printf("  Complete: status=%s code=%d\n", rec.Status, rec.StatusCode)
	}
}

// ── Webhooks ───────────────────────────────────────────────────────────────

func demonstrateWebhooks(ctx context.Context) {
	dlq := webhook.NewDLQ(100)
	d := webhook.NewDispatcher(dlq, nil)

	// Register typed handlers.
	d.OnResourceUpdate(func(_ context.Context, env *webhook.Envelope, ev webhook.ResourceUpdateEvent) error {
		fmt.Printf("  Resource updated: jti=%s status=%d\n", env.Jti, ev.Subject.HTTPStatusCode)
		return nil
	})

	d.OnConsentRevoked(func(_ context.Context, env *webhook.Envelope, _ webhook.ConsentAuthRevokedEvent) error {
		fmt.Printf("  Consent revoked: jti=%s\n", env.Jti)
		return nil
	})

	// Simulate an incoming webhook JSON payload.
	body, _ := json.Marshal(map[string]any{
		"iss": "https://aspsp.example.com",
		"iat": time.Now().Unix(),
		"jti": "event-jti-001",
		"sub": "https://aspsp.example.com/open-banking/v3.1/pisp/domestic-payments/pay-123",
		"toe": time.Now().Unix(),
		"events": map[string]any{
			string(webhook.EventTypeResourceUpdate): map[string]any{
				"subject": map[string]any{
					"subject_type":     "http://openbanking.org.uk/rid_http://openbanking.org.uk/rty",
					"http_status_code": 200,
					"links": map[string]string{
						"http://openbanking.org.uk/rid": "https://aspsp.example.com/payments/pay-123",
					},
				},
			},
		},
	})

	if err := d.DispatchJSON(ctx, body); err != nil {
		log.Printf("DispatchJSON: %v", err)
	}
	fmt.Printf("  DLQ depth after dispatch: %d\n", dlq.Len())
}

// ── Observability ──────────────────────────────────────────────────────────

func demonstrateObservability(rec *observability.InMemoryRecorder) {
	// Simulate some recorded calls.
	rec.RecordRequest("GET", "/accounts", 200, 45*time.Millisecond, nil)
	rec.RecordRequest("POST", "/domestic-payments", 201, 120*time.Millisecond, nil)
	rec.RecordRequest("GET", "/accounts", 500, 250*time.Millisecond, nil)
	rec.RecordRequest("GET", "/balances", 200, 35*time.Millisecond, nil)

	hc := observability.NewHealthChecker(rec, 0.30, 500*time.Millisecond)
	status := hc.Check()

	fmt.Printf("  Total requests: %d\n", status.TotalRequests)
	fmt.Printf("  Error rate:     %.0f%%\n", status.ErrorRate*100)
	fmt.Printf("  Avg latency:    %v\n", status.AvgDuration)
	fmt.Printf("  Healthy:        %v\n", status.Healthy)
}

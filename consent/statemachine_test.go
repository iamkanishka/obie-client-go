package consent_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/consent"
)

// ── Machine ───────────────────────────────────────────────────────────────

func TestMachine_InitialState(t *testing.T) {
	m := consent.NewMachine("cid-1")
	if m.State != consent.StateAwaitingAuthorisation {
		t.Errorf("initial state: got %s, want AwaitingAuthorisation", m.State)
	}
}

func TestMachine_Authorise(t *testing.T) {
	m := consent.NewMachine("cid-1")
	if err := m.Apply(consent.EventAuthorise); err != nil {
		t.Fatalf("Apply(Authorise): %v", err)
	}
	if m.State != consent.StateAuthorised {
		t.Errorf("state after Authorise: got %s, want Authorised", m.State)
	}
}

func TestMachine_Consume(t *testing.T) {
	m := consent.NewMachine("cid-1")
	m.Apply(consent.EventAuthorise) //nolint:errcheck
	if err := m.Apply(consent.EventConsume); err != nil {
		t.Fatalf("Apply(Consume): %v", err)
	}
	if m.State != consent.StateConsumed {
		t.Errorf("state after Consume: got %s, want Consumed", m.State)
	}
}

func TestMachine_Reject(t *testing.T) {
	m := consent.NewMachine("cid-1")
	if err := m.Apply(consent.EventReject); err != nil {
		t.Fatalf("Apply(Reject): %v", err)
	}
	if m.State != consent.StateRejected {
		t.Errorf("state: got %s, want Rejected", m.State)
	}
}

func TestMachine_InvalidTransition(t *testing.T) {
	m := consent.NewMachine("cid-1")
	err := m.Apply(consent.EventConsume) // cannot consume before authorise
	if err == nil {
		t.Fatal("expected ErrInvalidTransition, got nil")
	}
	var te *consent.ErrInvalidTransition
	if !errors.As(err, &te) {
		t.Errorf("expected *ErrInvalidTransition, got %T", err)
	}
}

func TestMachine_IsTerminal(t *testing.T) {
	m := consent.NewMachine("cid-1")
	if m.IsTerminal() {
		t.Error("AwaitingAuthorisation should not be terminal")
	}
	m.Apply(consent.EventReject) //nolint:errcheck
	if !m.IsTerminal() {
		t.Error("Rejected should be terminal")
	}
}

func TestMachine_History(t *testing.T) {
	m := consent.NewMachine("cid-1")
	m.Apply(consent.EventAuthorise) //nolint:errcheck
	m.Apply(consent.EventRevoke)    //nolint:errcheck

	if len(m.History) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(m.History))
	}
	if m.History[0].Event != consent.EventAuthorise {
		t.Errorf("history[0].Event: got %s, want Authorise", m.History[0].Event)
	}
	if m.History[1].Event != consent.EventRevoke {
		t.Errorf("history[1].Event: got %s, want Revoke", m.History[1].Event)
	}
}

func TestMachine_SyncFromASPSP(t *testing.T) {
	m := consent.NewMachine("cid-1")
	if err := m.SyncFromASPSP("Authorised"); err != nil {
		t.Fatalf("SyncFromASPSP: %v", err)
	}
	if m.State != consent.StateAuthorised {
		t.Errorf("state: got %s, want Authorised", m.State)
	}
}

func TestMachine_SyncFromASPSP_NoChange(t *testing.T) {
	m := consent.NewMachine("cid-1")
	before := len(m.History)
	m.SyncFromASPSP("AwaitingAuthorisation") //nolint:errcheck
	if len(m.History) != before {
		t.Error("SyncFromASPSP should not record history entry when state is unchanged")
	}
}

func TestMachine_SyncFromASPSP_UnknownStatus(t *testing.T) {
	m := consent.NewMachine("cid-1")
	err := m.SyncFromASPSP("BogusStatus")
	if err == nil {
		t.Error("expected error for unknown ASPSP status")
	}
}

// ── PKCE ─────────────────────────────────────────────────────────────────

func TestGeneratePKCE_Uniqueness(t *testing.T) {
	p1, err := consent.GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE: %v", err)
	}
	p2, err := consent.GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE: %v", err)
	}
	if p1.Verifier == p2.Verifier {
		t.Error("PKCE verifiers should be unique")
	}
	if p1.Challenge == p2.Challenge {
		t.Error("PKCE challenges should be unique")
	}
}

func TestGeneratePKCE_MethodIsS256(t *testing.T) {
	p, _ := consent.GeneratePKCE()
	if p.Method != "S256" {
		t.Errorf("Method: got %q, want S256", p.Method)
	}
}

func TestGeneratePKCE_ChallengeIsBase64URL(t *testing.T) {
	p, _ := consent.GeneratePKCE()
	// Base64URL characters: A-Z a-z 0-9 - _  (no = padding in raw encoding)
	for _, c := range p.Challenge {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			t.Errorf("challenge contains non-base64url character: %q", c)
		}
	}
}

// ── BuildAuthURL ──────────────────────────────────────────────────────────

func TestBuildAuthURL_ContainsRequiredParams(t *testing.T) {
	pkce, _ := consent.GeneratePKCE()
	authURL, err := consent.BuildAuthURL(consent.AuthURLParams{
		AuthorisationEndpoint: "https://aspsp.example.com/authorize",
		ClientID:              "client-123",
		RedirectURI:           "https://tpp.example.com/callback",
		ConsentID:             "consent-abc",
		Scope:                 "openid accounts",
		State:                 "random-state",
		Nonce:                 "random-nonce",
		PKCE:                  pkce,
	})
	if err != nil {
		t.Fatalf("BuildAuthURL: %v", err)
	}

	checks := []string{
		"response_type=",
		"client_id=client-123",
		"redirect_uri=",
		"state=random-state",
		"code_challenge=",
		"code_challenge_method=S256",
		"claims=",
	}
	for _, check := range checks {
		if !strings.Contains(authURL, check) {
			t.Errorf("auth URL missing %q: %s", check, authURL)
		}
	}
}

func TestBuildAuthURL_MissingEndpoint(t *testing.T) {
	_, err := consent.BuildAuthURL(consent.AuthURLParams{
		ClientID:  "client-123",
		ConsentID: "consent-abc",
	})
	if err == nil {
		t.Error("expected error when AuthorisationEndpoint is empty")
	}
}

// ── PollUntilAuthorised ───────────────────────────────────────────────────

func TestPollUntilAuthorised_SucceedsWhenAuthorised(t *testing.T) {
	calls := 0
	fetcher := func(_ context.Context, _ string) (string, error) {
		calls++
		if calls < 3 {
			return "AwaitingAuthorisation", nil
		}
		return "Authorised", nil
	}

	m := consent.NewMachine("cid-poll")
	err := consent.PollUntilAuthorised(context.Background(), m, fetcher, time.Millisecond)
	if err != nil {
		t.Fatalf("PollUntilAuthorised: %v", err)
	}
	if m.State != consent.StateAuthorised {
		t.Errorf("final state: got %s, want Authorised", m.State)
	}
}

func TestPollUntilAuthorised_AbortOnRejected(t *testing.T) {
	fetcher := func(_ context.Context, _ string) (string, error) {
		return "Rejected", nil
	}
	m := consent.NewMachine("cid-poll")
	err := consent.PollUntilAuthorised(context.Background(), m, fetcher, time.Millisecond)
	if err == nil {
		t.Error("expected error when consent is rejected")
	}
}

func TestPollUntilAuthorised_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	fetcher := func(c context.Context, _ string) (string, error) {
		calls++
		if calls == 2 {
			cancel()
		}
		return "AwaitingAuthorisation", nil
	}
	m := consent.NewMachine("cid-poll")
	err := consent.PollUntilAuthorised(ctx, m, fetcher, time.Millisecond)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}

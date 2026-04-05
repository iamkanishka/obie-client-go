// Command example demonstrates end-to-end usage of the OBIE SDK:
// authentication → account listing → domestic payment.
//
// Usage:
//
//	go run ./example \
//	  -token-url  https://aspsp.example.com/token \
//	  -client-id  your-client-id \
//	  -key-file   /path/to/private.pem \
//	  -cert-file  /path/to/transport.pem \
//	  -base-url   https://aspsp.example.com
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/obie"
)

func main() {
	tokenURL := flag.String("token-url", "", "OAuth2 token endpoint URL")
	clientID := flag.String("client-id", "", "OBIE client ID")
	keyFile  := flag.String("key-file", "", "Path to PEM-encoded private key")
	certFile := flag.String("cert-file", "", "Path to PEM-encoded mTLS certificate (optional)")
	baseURL  := flag.String("base-url", "", "ASPSP base URL (overrides environment default)")
	sandbox  := flag.Bool("sandbox", true, "Target sandbox environment")
	flag.Parse()

	if *tokenURL == "" || *clientID == "" || *keyFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	keyPEM, err := os.ReadFile(*keyFile)
	if err != nil {
		log.Fatalf("read key file: %v", err)
	}

	var certPEM []byte
	if *certFile != "" {
		certPEM, err = os.ReadFile(*certFile)
		if err != nil {
			log.Fatalf("read cert file: %v", err)
		}
	}

	env := obie.EnvironmentSandbox
	if !*sandbox {
		env = obie.EnvironmentProduction
	}

	cfg := obie.Config{
		Environment:    env,
		BaseURL:        *baseURL,
		TokenURL:       *tokenURL,
		ClientID:       *clientID,
		PrivateKeyPEM:  keyPEM,
		CertificatePEM: certPEM,
		SigningKeyID:   "my-signing-key",
		FinancialID:    "0015800001041RHAAY",
		Scopes:         []string{"accounts", "payments"},
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		Logger:         obie.NewStdLogger(log.Default()),
	}

	client, err := obie.NewClient(cfg)
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	ctx := context.Background()

	// ── 1. List accounts ──────────────────────────────────────────────────────
	fmt.Println("\n=== Accounts ===")
	accs, err := client.Accounts.GetAccounts(ctx)
	if err != nil {
		log.Fatalf("GetAccounts: %v", err)
	}
	for _, acc := range accs.Data.Account {
		fmt.Printf("  [%s] %s (%s %s)\n", acc.AccountId, acc.Nickname, acc.Currency, acc.AccountSubType)
	}

	if len(accs.Data.Account) == 0 {
		log.Println("No accounts returned – check consent authorisation.")
		return
	}

	// ── 2. Fetch balances for the first account ───────────────────────────────
	firstID := accs.Data.Account[0].AccountId
	fmt.Printf("\n=== Balances for %s ===\n", firstID)
	bals, err := client.Accounts.GetAccountBalances(ctx, firstID)
	if err != nil {
		log.Fatalf("GetAccountBalances: %v", err)
	}
	for _, b := range bals.Data.Balance {
		fmt.Printf("  %s %s %s\n", b.Type, b.Amount.Amount, b.Amount.Currency)
	}

	// ── 3. Create a domestic payment consent ─────────────────────────────────
	fmt.Println("\n=== Domestic Payment Consent ===")
	consentReq := &models.OBWriteDomesticConsent5{
		Data: models.OBWriteDomesticConsentData5{
			Initiation: models.OBDomesticInitiation{
				InstructionIdentification: "INSTR-001",
				EndToEndIdentification:    "E2E-001",
				InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{
					Amount:   "10.00",
					Currency: "GBP",
				},
				CreditorAccount: models.OBCashAccount3{
					SchemeName:     "UK.OBIE.SortCodeAccountNumber",
					Identification: "20000319825731",
					Name:           "Acme Ltd",
				},
				RemittanceInformation: &models.OBRemittanceInformation1{
					Reference:    "OBIE-SDK-TEST",
					Unstructured: "Test payment from OBIE SDK",
				},
			},
		},
		Risk: models.OBRisk1{
			PaymentContextCode: "EcommerceGoods",
		},
	}

	consent, err := client.Payments.CreateDomesticPaymentConsent(ctx, consentReq)
	if err != nil {
		log.Fatalf("CreateDomesticPaymentConsent: %v", err)
	}

	fmt.Printf("  Consent ID: %s\n", consent.Data.ConsentId)
	fmt.Printf("  Status:     %s\n", consent.Data.Status)
	fmt.Printf("  (Redirect the PSU to authorise this consent before submitting the payment)\n")

	// ── 4. Submit payment (requires prior PSU authorisation in a real flow) ───
	fmt.Println("\n=== Domestic Payment Submission (demo – skipped in example) ===")
	paymentReq := &models.OBWriteDomestic2{
		Data: models.OBWriteDomesticData2{
			ConsentId:  consent.Data.ConsentId,
			Initiation: consentReq.Data.Initiation,
		},
		Risk: consentReq.Risk,
	}
	_ = paymentReq
	fmt.Println("  Payment request built. Submit after PSU authorises the consent.")

	// ── 5. Funds confirmation ─────────────────────────────────────────────────
	fmt.Println("\n=== Funds Confirmation ===")
	fmt.Println("  (Requires a CBPII consent — skipped in example)")

	// ── 6. Pretty-print full consent response ─────────────────────────────────
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	fmt.Println("\n=== Full Consent Response ===")
	enc.Encode(consent) //nolint:errcheck
}

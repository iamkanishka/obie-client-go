# obie-client-go

[![Go Reference](https://pkg.go.dev/badge/github.com/iamkanishka/obie-client-go/obie.svg)](https://pkg.go.dev/github.com/iamkanishka/obie-client-go/obie)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Production-grade Go client for the [UK Open Banking (OBIE) Read/Write API v3.1.3](https://openbankinguk.github.io/read-write-api-site2/).**

Complete implementation of every OBIE endpoint — AIS, PIS, CBPII, VRP, File Payments,
Event Notifications, DCR — with full type safety, FAPI-compliant headers, mTLS,
exponential-backoff retry, circuit breaker, token-bucket rate limiter, and a
generic LRU cache.

---

## Installation

```bash
go get github.com/iamkanishka/obie-client-go
go mod tidy
```

Requires **Go 1.23** or later. The two external dependencies are fetched by `go mod tidy`:

| Dependency | Purpose |
|---|---|
| `github.com/golang-jwt/jwt/v5` | RS256 JWT client assertion |
| `github.com/google/uuid` | idempotency key generation |

---

## Quick start

```go
import (
    "context"
    "log"

    "github.com/iamkanishka/obie-client-go/models"
    "github.com/iamkanishka/obie-client-go/obie"
)

func main() {
    keyPEM, _ := os.ReadFile("private.pem")

    client, err := obie.NewClient(obie.Config{
        Environment:   obie.EnvironmentSandbox,
        TokenURL:      "https://aspsp.example.com/token",
        ClientID:      "your-client-id",
        PrivateKeyPEM: keyPEM,
        SigningKeyID:  "your-signing-kid",
        FinancialID:   "0015800001041RHAAY",
        Scopes:        []string{"accounts", "payments", "fundsconfirmations"},
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // 1. Create AIS consent (required before any account reads)
    consent, err := client.AISConsent.CreateAccountAccessConsent(ctx,
        &models.OBReadConsent1{
            Data: models.OBReadData1{
                Permissions: models.AllPermissions(),
            },
        })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Redirect PSU to authorise consent: %s", consent.Data.ConsentId)

    // 2. After PSU authorises, read accounts
    accounts, err := client.Accounts.GetAccounts(ctx)
    for _, acc := range accounts.Data.Account {
        log.Printf("%s — %s %s", acc.AccountId, acc.Currency, acc.AccountSubType)
    }
}
```

---

## Configuration

```go
client, err := obie.NewClient(obie.Config{
    // Required
    TokenURL:      "https://aspsp.example.com/token",
    ClientID:      "your-client-id",
    PrivateKeyPEM: keyPEM,  // PEM-encoded RSA private key

    // Recommended
    CertificatePEM:    certPEM,      // mTLS transport certificate
    SigningKeyID:       "kid-value",
    FinancialID:        "0015800001041RHAAY",
    CustomerIPAddress:  r.RemoteAddr, // omit for scheduled/M2M flows

    // Optional
    Environment: obie.EnvironmentSandbox, // or EnvironmentProduction
    BaseURL:     "https://ob.bank.example.com",
    Scopes:      []string{"accounts", "payments", "fundsconfirmations"},
    Timeout:     30 * time.Second, // default
    MaxRetries:  3,                // default
    Logger:      myLogger,         // implements obie.Logger
})
```

| Field | Type | Default | Description |
|---|---|---|---|
| `Environment` | `Environment` | `EnvironmentSandbox` | `EnvironmentSandbox` or `EnvironmentProduction` |
| `BaseURL` | `string` | derived | Override ASPSP base URL |
| `TokenURL` | `string` | **required** | OAuth2 token endpoint |
| `ClientID` | `string` | **required** | Software client ID |
| `PrivateKeyPEM` | `[]byte` | **required** | RSA private key (PKCS#1 or PKCS#8) |
| `CertificatePEM` | `[]byte` | — | mTLS transport certificate |
| `SigningKeyID` | `string` | — | `kid` for JWS/JWT headers |
| `FinancialID` | `string` | — | `x-fapi-financial-id` header |
| `CustomerIPAddress` | `string` | — | `x-fapi-customer-ip-address` |
| `Scopes` | `[]string` | `accounts payments fundsconfirmations` | OAuth2 scopes |
| `Timeout` | `time.Duration` | `30s` | Per-request timeout |
| `MaxRetries` | `int` | `3` | Retry attempts on idempotent failures |
| `Logger` | `Logger` | no-op | Pluggable structured logger |
| `RequestHooks` | `[]RequestHook` | — | Pre-request interceptors |
| `ResponseHooks` | `[]ResponseHook` | — | Post-response interceptors |
| `TLSConfig` | `*tls.Config` | — | Advanced TLS configuration |

---

## Services

All services are fields on `*obie.Client`:

| Field | Package | Description |
|---|---|---|
| `client.AISConsent` | `aisp` | Account-access-consent + offers + AIS standing orders |
| `client.Accounts` | `accounts` | Accounts, balances, transactions, beneficiaries, direct debits, scheduled payments, statements, parties, products |
| `client.Payments` | `payments` | Domestic, international, scheduled, and standing-order payments |
| `client.FilePayments` | `filepayments` | File payment consent, file upload/download, submission, report |
| `client.Funds` | `funds` | CBPII funds-confirmation consent and check |
| `client.VRP` | `vrp` | Variable recurring payments |
| `client.EventNotifications` | `eventnotifications` | Event subscriptions, callback URLs, aggregated polling |

---

## AIS — Account Access Consents

Account access consent is required before any AIS resource read.

```go
// Create — returns ConsentId for PSU redirect
expiry := time.Now().Add(90 * 24 * time.Hour)
consent, err := client.AISConsent.CreateAccountAccessConsent(ctx, &models.OBReadConsent1{
    Data: models.OBReadData1{
        Permissions: []models.Permission{
            models.PermissionReadAccountsDetail,
            models.PermissionReadBalances,
            models.PermissionReadTransactionsDetail,
            models.PermissionReadBeneficiariesDetail,
            models.PermissionReadDirectDebits,
            models.PermissionReadStandingOrdersDetail,
            models.PermissionReadScheduledPaymentsDetail,
            models.PermissionReadStatementsDetail,
            models.PermissionReadParty,
            models.PermissionReadOffers,
            models.PermissionReadProducts,
        },
        ExpirationDateTime: &expiry,
    },
})

// Poll status
status, err := client.AISConsent.GetAccountAccessConsent(ctx, consent.Data.ConsentId)
// status.Data.Status: "AwaitingAuthorisation" → "Authorised"

// Delete when PSU revokes
err = client.AISConsent.DeleteAccountAccessConsent(ctx, consent.Data.ConsentId)
```

---

## AIS — Resource reads

```go
// Accounts
accounts, _    := client.Accounts.GetAccounts(ctx)
account, _     := client.Accounts.GetAccount(ctx, "acc-id")

// Balances
balances, _    := client.Accounts.GetBalances(ctx)
accBal, _      := client.Accounts.GetAccountBalances(ctx, "acc-id")

// Transactions (with optional date filter)
txns, _ := client.Accounts.GetAccountTransactions(ctx, "acc-id",
    accounts.TransactionFilter{
        FromBookingDateTime: &from,
        ToBookingDateTime:   &to,
    })

// Beneficiaries, Direct Debits, Scheduled Payments
bens, _  := client.Accounts.GetBeneficiaries(ctx)
dds, _   := client.Accounts.GetDirectDebits(ctx)
sps, _   := client.Accounts.GetScheduledPayments(ctx)

// Statements
stmts, _ := client.Accounts.GetStatements(ctx)
stmt, _  := client.Accounts.GetStatement(ctx, "acc-id", "stmt-id")
stmtTx, _:= client.Accounts.GetStatementTransactions(ctx, "acc-id", "stmt-id")

// Parties, Products, Offers
party, _ := client.Accounts.GetParty(ctx)
prods, _ := client.Accounts.GetProducts(ctx)
offers, _:= client.AISConsent.GetOffers(ctx)
accOffers,_:= client.AISConsent.GetAccountOffers(ctx, "acc-id")

// AIS Standing Orders (read-only view)
sos, _   := client.AISConsent.GetStandingOrders(ctx)
accSOs, _:= client.AISConsent.GetAccountStandingOrders(ctx, "acc-id")
```

---

## PIS — Domestic Payments

```go
// 1. Create consent
consent, err := client.Payments.CreateDomesticPaymentConsent(ctx,
    &models.OBWriteDomesticConsent5{
        Data: models.OBWriteDomesticConsentData5{
            Initiation: models.OBDomesticInitiation{
                InstructionIdentification: "INSTR-001",
                EndToEndIdentification:    "E2E-001",
                InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{
                    Amount: "10.50", Currency: "GBP",
                },
                CreditorAccount: models.OBCashAccount3{
                    SchemeName:     "UK.OBIE.SortCodeAccountNumber",
                    Identification: "20000319825731",
                    Name:           "Receiver Name",
                },
            },
        },
        Risk: models.OBRisk1{
            PaymentContextCode: models.PaymentContextPartyToParty,
        },
    })

// 2. Redirect PSU to authorise consent.Data.ConsentId

// 3. Submit payment
payment, err := client.Payments.SubmitDomesticPayment(ctx,
    &models.OBWriteDomestic2{
        Data: models.OBWriteDomesticData2{
            ConsentId:  consent.Data.ConsentId,
            Initiation: consent.Data.Initiation,
        },
        Risk: models.OBRisk1{PaymentContextCode: models.PaymentContextPartyToParty},
    })

// 4. Poll until terminal status
final, err := client.Payments.PollDomesticPaymentUntilTerminal(
    ctx, payment.Data.DomesticPaymentId, 5*time.Second)
// final.Data.Status: "AcceptedSettlementCompleted" | "Rejected"
```

## PIS — All payment types

```go
// International
client.Payments.CreateInternationalPaymentConsent(ctx, req)
client.Payments.SubmitInternationalPayment(ctx, req)
client.Payments.PollInternationalPaymentUntilTerminal(ctx, id, interval)

// Domestic Scheduled
client.Payments.CreateDomesticScheduledPaymentConsent(ctx, req)
client.Payments.DeleteDomesticScheduledPaymentConsent(ctx, consentId)
client.Payments.SubmitDomesticScheduledPayment(ctx, req)

// Domestic Standing Orders
client.Payments.CreateDomesticStandingOrderConsent(ctx, req)
client.Payments.SubmitDomesticStandingOrder(ctx, req)

// International Scheduled + International Standing Orders
client.Payments.CreateInternationalScheduledPaymentConsent(ctx, req)
client.Payments.SubmitInternationalScheduledPayment(ctx, req)
client.Payments.CreateInternationalStandingOrderConsent(ctx, req)
client.Payments.SubmitInternationalStandingOrder(ctx, req)

// Payment details (all types)
client.Payments.GetDomesticPaymentDetails(ctx, paymentId)
client.Payments.GetPaymentStatus(ctx, payments.PaymentTypeDomestic, paymentId)
```

---

## PIS — File Payments

```go
// 1. Create consent
sum := 1500.00
consent, _ := client.FilePayments.CreateFilePaymentConsent(ctx,
    &models.OBWriteFileConsent3{
        Data: models.OBWriteFileConsentData3{
            Initiation: models.OBFile2{
                FileType: models.FileTypeUK_OBIE_PaymentInitiation_3_1,
                FileHash: "sha256-base64-hash",
                NumberOfTransactions: "10",
                ControlSum: &sum,
            },
        },
    })

// 2. Upload file (status moves to AwaitingAuthorisation)
fileJSON, _ := os.ReadFile("payments.json")
err = client.FilePayments.UploadFile(ctx, consent.Data.ConsentId, fileJSON, "application/json")

// 3. Redirect PSU to authorise

// 4. Submit and poll
payment, _ := client.FilePayments.SubmitFilePayment(ctx, &models.OBWriteFile2{
    Data: models.OBWriteFileData2{ConsentId: consent.Data.ConsentId, Initiation: consent.Data.Initiation},
})
status, _ := client.FilePayments.GetFilePayment(ctx, payment.Data.FilePaymentId)
report, contentType, _ := client.FilePayments.GetFilePaymentReport(ctx, payment.Data.FilePaymentId)
```

---

## CBPII — Funds Confirmation

```go
consent, _ := client.Funds.CreateConsent(ctx, &models.OBFundsConfirmationConsent1{
    Data: models.OBFundsConfirmationConsentData1{
        DebtorAccount: models.OBCashAccount3{
            SchemeName:     "UK.OBIE.SortCodeAccountNumber",
            Identification: "20000319825731",
        },
    },
})
// After PSU authorises:
result, _ := client.Funds.ConfirmFundsAvailability(ctx, &models.OBFundsConfirmation1{
    Data: models.OBFundsConfirmationData1{
        ConsentId: consent.Data.ConsentId,
        Reference: "purchase-ref",
        InstructedAmount: models.OBActiveOrHistoricCurrencyAndAmount{
            Amount: "150.00", Currency: "GBP",
        },
    },
})
log.Printf("Funds available: %v", result.Data.FundsAvailable)

// Revoke when done
_ = client.Funds.DeleteConsent(ctx, consent.Data.ConsentId)
```

---

## VRP — Variable Recurring Payments

```go
// Create consent
consent, _ := client.VRP.CreateConsent(ctx, &models.OBDomesticVRPConsentRequest{
    Data: models.OBDomesticVRPConsentRequestData{
        ControlParameters: models.OBDomesticVRPControlParameters{
            VRPType:                  []string{"UK.OBIE.VRPType.Sweeping"},
            PSUAuthenticationMethods: []string{"UK.OBIE.SCA"},
            MaximumIndividualAmount: models.OBActiveOrHistoricCurrencyAndAmount{
                Amount: "500.00", Currency: "GBP",
            },
            PeriodicLimits: []models.OBDomesticVRPControlParametersPeriodic{{
                PeriodType:      "Month",
                PeriodAlignment: "Calendar",
                Amount: models.OBActiveOrHistoricCurrencyAndAmount{
                    Amount: "2000.00", Currency: "GBP",
                },
            }},
        },
    },
})

// Delete consent (revoke)
_ = client.VRP.DeleteConsent(ctx, consent.Data.Data.ConsentId)

// Submit payment and poll
payment, _ := client.VRP.SubmitPayment(ctx, req)
final, _   := client.VRP.PollPaymentUntilTerminal(ctx, payment.Data.DomesticVRPId, 3*time.Second)
```

---

## Event Notifications

```go
// Subscribe (ASPSP pushes events to your URL)
sub, _ := client.EventNotifications.CreateEventSubscription(ctx,
    &models.OBEventSubscription1{
        Data: models.OBEventSubscriptionData1{
            CallbackUrl: "https://tpp.example.com/events",
            Version:     "3.1",
            EventTypes: []models.EventNotificationType{
                models.EventNotificationResourceUpdate,
                models.EventNotificationConsentAuthorizationRevoked,
            },
        },
    })

// Update
_, _ = client.EventNotifications.UpdateEventSubscription(ctx, sub.Data.EventSubscriptionId, ...)

// Delete
_ = client.EventNotifications.DeleteEventSubscription(ctx, sub.Data.EventSubscriptionId)

// Aggregated polling (pull model)
maxEvts := 10
resp, _ := client.EventNotifications.PollEvents(ctx, &models.OBEventPolling1{
    MaxEvents:         &maxEvts,
    ReturnImmediately: boolPtr(true),
    Ack:               []string{"jti-001", "jti-002"},
})
for jti, jwt := range resp.Sets {
    // verify and process each event JWT
    _ = jti; _ = jwt
}

// Callback URLs (legacy)
cb, _ := client.EventNotifications.CreateCallbackUrl(ctx, &models.OBCallbackUrl1{
    Data: models.OBCallbackUrlData1{Url: "https://tpp.example.com/cb", Version: "3.1"},
})
_ = client.EventNotifications.DeleteCallbackUrl(ctx, cb.Data.CallbackUrlId)
```

---

## Error handling

```go
import "github.com/iamkanishka/obie-client-go/obie"

_, err := client.Accounts.GetAccounts(ctx)
if err != nil {
    var apiErr *obie.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d — interaction: %s\n", apiErr.StatusCode, apiErr.InteractionID)
        if apiErr.OBError != nil {
            fmt.Println(apiErr.OBError.Code, apiErr.OBError.Message)
        }
        // Check specific OBIE error code
        if apiErr.IsErrorCode(models.OBIEErrorFieldMissing) {
            // handle validation error
        }
    }
}
```

---

## Permission codes

```go
// All 21 OBIE permission codes:
models.PermissionReadAccountsBasic
models.PermissionReadAccountsDetail
models.PermissionReadBalances
models.PermissionReadBeneficiariesBasic
models.PermissionReadBeneficiariesDetail
models.PermissionReadDirectDebits
models.PermissionReadOffers
models.PermissionReadPAN
models.PermissionReadParty
models.PermissionReadPartyPSU
models.PermissionReadProducts
models.PermissionReadScheduledPaymentsBasic
models.PermissionReadScheduledPaymentsDetail
models.PermissionReadStandingOrdersBasic
models.PermissionReadStandingOrdersDetail
models.PermissionReadStatementsBasic
models.PermissionReadStatementsDetail
models.PermissionReadTransactionsBasic
models.PermissionReadTransactionsCredits
models.PermissionReadTransactionsDebits
models.PermissionReadTransactionsDetail

// Helper that returns all 15 Detail-level permissions:
models.AllPermissions()
```

---

## Advanced packages

| Package | Import | Purpose |
|---|---|---|
| `middleware` | `.../middleware` | Composable `http.RoundTripper` chain (logging, correlation ID, dry-run, audit capture) |
| `ratelimit` | `.../ratelimit` | Token-bucket rate limiter honouring `Retry-After` |
| `circuitbreaker` | `.../circuitbreaker` | Closed/Open/HalfOpen circuit breaker with state-change hooks |
| `cache` | `.../cache` | Generic TTL LRU cache, `ConsentCache`, `ResponseCache` |
| `pagination` | `.../pagination` | Lazy HATEOAS `next`-link iterator |
| `observability` | `.../observability` | `Tracer`/`Span` interfaces + `InMemoryRecorder` + `HealthChecker` |
| `idempotency` | `.../idempotency` | Server-side idempotency key store + middleware |
| `validation` | `.../validation` | Deep request validation (amounts, IBAN, sort codes, VRP limits) |
| `batch` | `.../batch` | Bounded-concurrency fan-out + sequential `Pipeline` |
| `consent` | `.../consent` | PKCE helper, `BuildAuthURL`, consent state machine, `PollUntilAuthorised` |
| `webhook` | `.../webhook` | Typed event dispatcher with dead-letter queue |
| `config` | `.../config` | Layered config loader (file + env + secret provider) with hot-reload |
| `signing` | `.../signing` | Detached JWS (OBIE b64=false profile) |
| `auth` | `.../auth` | OAuth2 token manager, JWT assertion, mTLS transport |

---

## Running tests

```bash
# Fetch dependencies first
go mod tidy

# Run all tests with race detector
make test

# Short subset (no network)
make test-short

# HTML coverage report
make coverage

# Vet + build
make vet
make build
```

---

## Project layout

```
github.com/iamkanishka/obie-client-go/
├── obie/               Core client, config, HTTP layer, errors, logger, doc
├── models/             All OBIE v3.1.3 request/response structs + typed enums
├── aisp/               AIS consent + offers + standing orders (read)
├── accounts/           AIS resource reads (all 9 resource types)
├── payments/           PIS — all 6 payment types + status polling
├── filepayments/       File payment consent, upload, submission, report
├── funds/              CBPII funds confirmation
├── vrp/                Variable recurring payments
├── eventnotifications/ Event subscriptions, callback URLs, aggregated polling
├── events/             Webhook JWS parsing + signature verification
├── webhook/            Typed event dispatcher + dead-letter queue
├── auth/               OAuth2, JWT RS256 client assertion, mTLS
├── signing/            Detached JWS (OBIE signing profile)
├── dcr/                Dynamic Client Registration
├── directory/          Open Banking Directory client
├── middleware/         Composable HTTP transport middleware
├── ratelimit/          Token-bucket rate limiter
├── circuitbreaker/     Circuit breaker
├── cache/              Generic TTL LRU cache
├── pagination/         HATEOAS next-link iterator
├── observability/      Tracing + metrics interfaces + InMemoryRecorder
├── idempotency/        Idempotency key store
├── validation/         Request validation
├── batch/              Parallel fan-out + pipeline
├── consent/            PKCE, auth URL builder, consent state machine
├── config/             Layered config with hot-reload
└── example/            Usage examples
```

---

## Licence

MIT © Kanishka Naik

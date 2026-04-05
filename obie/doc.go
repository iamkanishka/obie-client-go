// Package obie is the core package of the OBIE (Open Banking Implementation
// Entity) Go client library.
//
// # Overview
//
// This library provides a complete, production-grade Go client for the UK Open
// Banking Read/Write API v3.1.3. It covers every resource defined in the official
// specification:
//
//   - AIS (Account Information Service) — account access consents, accounts,
//     balances, transactions, beneficiaries, direct debits, standing orders,
//     scheduled payments, statements, parties, products, and offers.
//   - PIS (Payment Initiation Service) — domestic, international, scheduled,
//     standing-order, and file payments.
//   - CBPII (Confirmation of Funds) — funds confirmation consents and checks.
//   - VRP (Variable Recurring Payments) — consent lifecycle, payment submission,
//     and polling.
//   - Event Notifications — push subscriptions, callback URLs, and aggregated polling.
//   - DCR (Dynamic Client Registration) — RFC 7591 / FAPI compliant.
//
// # Quick start
//
//	client, err := obie.NewClient(obie.Config{
//	    Environment:   obie.EnvironmentSandbox,
//	    TokenURL:      "https://aspsp.example.com/token",
//	    ClientID:      "your-client-id",
//	    PrivateKeyPEM: keyPEM,
//	    SigningKeyID:  "your-kid",
//	    FinancialID:   "0015800001041RHAAY",
//	    Scopes:        []string{"accounts", "payments", "fundsconfirmations"},
//	})
//
// # Services
//
// All services are accessed as fields on [Client]:
//
//	client.AISConsent         — account-access-consent lifecycle (POST/GET/DELETE)
//	client.Accounts           — AIS resource reads (accounts, balances, transactions…)
//	client.Payments           — PIS payment initiation (domestic, international, scheduled, SO)
//	client.FilePayments       — bulk file payment flow (consent → upload → submit → report)
//	client.Funds              — CBPII funds confirmation
//	client.VRP                — variable recurring payments
//	client.EventNotifications — event subscriptions, callback URLs, aggregated polling
//	client.Metrics            — in-process RED metrics and health checker
package obie

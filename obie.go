// Package obie is the root convenience package for the
// github.com/iamkanishka/obie-client-go module.
//
// It re-exports the most commonly used types from the
// [github.com/iamkanishka/obie-client-go/obie] sub-package so callers can
// import a single path for common use-cases.
//
// # Primary import path
//
// Most users will import the core client directly:
//
//	import "github.com/iamkanishka/obie-client-go/obie"
//
//	client, err := obie.NewClient(obie.Config{ ... })
//
// # Module overview
//
// The github.com/iamkanishka/obie-client-go module provides a
// production-grade Go client for the UK Open Banking (OBIE) Read/Write
// API v3.1.3. It covers:
//
//   - AIS  — Account Information Service (consents, accounts, transactions…)
//   - PIS  — Payment Initiation Service (domestic, international, file…)
//   - CBPII — Confirmation of Funds
//   - VRP  — Variable Recurring Payments
//   - Event Notifications — subscriptions, callback URLs, polling
//   - DCR  — Dynamic Client Registration
//
// # Package map
//
//	obie/               — Client, Config, HTTP layer, errors, logger
//	models/             — All OBIE v3.1.3 typed structs and enums
//	aisp/               — AIS consent service + offers + standing orders
//	accounts/           — AIS resource reads
//	payments/           — PIS all payment types
//	filepayments/       — File payment flow
//	funds/              — CBPII funds confirmation
//	vrp/                — Variable recurring payments
//	eventnotifications/ — Event subscriptions + polling
//	auth/               — OAuth2 token manager + mTLS
//	signing/            — Detached JWS (OBIE signing profile)
//	validation/         — Request validation
//	internal/transport/ — Shared HTTP option types (breaks import cycle)
//
// # Getting started
//
//	go get github.com/iamkanishka/obie-client-go
//	go mod tidy
//
// See the [obie] sub-package for full documentation and examples.
package obie

// Version is the current module version.
const Version = "1.0.2"

// SpecVersion is the OBIE Read/Write API specification version targeted by
// this module.
const SpecVersion = "v3.1.3"

package models

import "time"

// ────────────────────────────────────────────────────────────────────────────
// VRP Consent Request/Response (already in models/payments.go but extended here
// with typed enums and missing v3.1.3 fields)
// ────────────────────────────────────────────────────────────────────────────

// OBDomesticVRPConsentRequestTyped uses typed enums — use this for new code.
// The existing OBDomesticVRPConsentRequest in payments.go is retained for compat.

// OBDomesticVRPControlParametersTyped is a spec-accurate control parameters block.
type OBDomesticVRPControlParametersTyped struct {
	ValidFromDateTime       *time.Time                                    `json:"ValidFromDateTime,omitempty"`
	ValidToDateTime         *time.Time                                    `json:"ValidToDateTime,omitempty"`
	MaximumIndividualAmount OBActiveOrHistoricCurrencyAndAmount           `json:"MaximumIndividualAmount"`
	PeriodicLimits          []OBDomesticVRPControlParametersPeriodicTyped `json:"PeriodicLimits"`
	VRPType                 []OBVRPType                                   `json:"VRPType"`
	PSUAuthenticationMethods []OBVRPAuthenticationMethods                 `json:"PSUAuthenticationMethods"`
	// SupplementaryData allows ASPSPs to add bank-specific parameters.
	SupplementaryData       map[string]any                        `json:"SupplementaryData,omitempty"`
}

// OBDomesticVRPControlParametersPeriodicTyped uses typed PeriodType and alignment.
type OBDomesticVRPControlParametersPeriodicTyped struct {
	// PeriodAlignment: Calendar or Consent.
	PeriodAlignment OBPeriodAlignment                   `json:"PeriodAlignment"`
	// PeriodType: Day, Week, Fortnight, Month, Half-year, Year.
	PeriodType      OBPeriodType                        `json:"PeriodType"`
	Amount          OBActiveOrHistoricCurrencyAndAmount  `json:"Amount"`
}

// OBDomesticVRPResponseData5Typed includes all v3.1.3 response fields.
type OBDomesticVRPResponseData5Typed struct {
	DomesticVRPId                string                   `json:"DomesticVRPId"`
	ConsentId                    string                   `json:"ConsentId"`
	CreationDateTime             time.Time                `json:"CreationDateTime"`
	Status                       PaymentStatus            `json:"Status"`
	StatusUpdateDateTime         time.Time                `json:"StatusUpdateDateTime"`
	// ExpectedExecutionDateTime is the date/time the payment is expected to execute.
	ExpectedExecutionDateTime    *time.Time               `json:"ExpectedExecutionDateTime,omitempty"`
	// ExpectedSettlementDateTime is the date/time the payment is expected to settle.
	ExpectedSettlementDateTime   *time.Time               `json:"ExpectedSettlementDateTime,omitempty"`
	Refund                       *OBCashAccount3          `json:"Refund,omitempty"`
	Charges                      []OBCharge2              `json:"Charges,omitempty"`
	Initiation                   OBDomesticVRPInitiation  `json:"Initiation"`
	Instruction                  OBDomesticVRPInstruction `json:"Instruction"`
	DebtorAccount                *OBCashAccount3          `json:"DebtorAccount,omitempty"`
}

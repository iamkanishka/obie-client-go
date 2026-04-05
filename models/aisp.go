package models

import "time"

// ────────────────────────────────────────────────────────────────────────────
// Account Access Consent (OBReadConsent1 / OBReadConsentResponse1)
// Ref: /resources-and-data-models/aisp/account-access-consents/
// ────────────────────────────────────────────────────────────────────────────

// OBReadConsent1 is the request body for POST /account-access-consents.
type OBReadConsent1 struct {
	Data OBReadData1 `json:"Data"`
	Risk OBRisk2     `json:"Risk"`
}

// OBReadData1 carries the permissions and optional datetime bounds for AIS consent.
type OBReadData1 struct {
	// Permissions lists the data clusters being consented.
	// Must contain at least one value. Use AllPermissions() for full access.
	Permissions             []Permission `json:"Permissions"`
	// ExpirationDateTime: when nil, permissions are open-ended.
	ExpirationDateTime      *time.Time   `json:"ExpirationDateTime,omitempty"`
	// TransactionFromDateTime: when nil, returns data from earliest available.
	TransactionFromDateTime *time.Time   `json:"TransactionFromDateTime,omitempty"`
	// TransactionToDateTime: when nil, returns data up to latest available.
	TransactionToDateTime   *time.Time   `json:"TransactionToDateTime,omitempty"`
}

// OBRisk2 is the (currently empty) risk block for AIS requests.
// The spec reserves this for future use; always send as an empty object {}.
type OBRisk2 struct{}

// OBReadConsentResponse1 is the response for POST/GET /account-access-consents.
type OBReadConsentResponse1 struct {
	Data  OBReadDataConsentResponse1 `json:"Data"`
	Risk  OBRisk2                    `json:"Risk"`
	Links Links                      `json:"Links"`
	Meta  Meta                       `json:"Meta"`
}

// OBReadDataConsentResponse1 carries the full consent response data.
type OBReadDataConsentResponse1 struct {
	ConsentId               string        `json:"ConsentId"`
	CreationDateTime        time.Time     `json:"CreationDateTime"`
	Status                  ConsentStatus `json:"Status"`
	StatusUpdateDateTime    time.Time     `json:"StatusUpdateDateTime"`
	Permissions             []Permission  `json:"Permissions"`
	ExpirationDateTime      *time.Time    `json:"ExpirationDateTime,omitempty"`
	TransactionFromDateTime *time.Time    `json:"TransactionFromDateTime,omitempty"`
	TransactionToDateTime   *time.Time    `json:"TransactionToDateTime,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Offers (OBReadOffer1)
// Ref: /resources-and-data-models/aisp/Offers/
// ────────────────────────────────────────────────────────────────────────────

// OBOffer1 represents a single account offer.
// Rate format: ^(-?\d{1,3}){1}(.\d{1,4}){0,1}$ (e.g. "1.5000")
type OBOffer1 struct {
	AccountId   string                               `json:"AccountId"`
	OfferId     string                               `json:"OfferId,omitempty"`
	OfferType   OfferType                            `json:"OfferType,omitempty"`
	Description string                               `json:"Description,omitempty"`
	StartDateTime *time.Time                         `json:"StartDateTime,omitempty"`
	EndDateTime   *time.Time                         `json:"EndDateTime,omitempty"`
	// Rate is the interest/promotional rate, e.g. "1.5000"
	Rate        string                               `json:"Rate,omitempty"`
	// Value is a numeric value (e.g. number of months, points)
	Value       *int                                 `json:"Value,omitempty"`
	// Term describes the terms of the offer in free text
	Term        string                               `json:"Term,omitempty"`
	// URL links to documentation for the offer
	URL         string                               `json:"URL,omitempty"`
	Amount      *OBActiveOrHistoricCurrencyAndAmount  `json:"Amount,omitempty"`
	Fee         *OBActiveOrHistoricCurrencyAndAmount  `json:"Fee,omitempty"`
}

// OBReadOffer1 is the response for GET /offers and GET /accounts/{id}/offers.
type OBReadOffer1 struct {
	Data  OBReadDataOffer1 `json:"Data"`
	Links Links            `json:"Links"`
	Meta  Meta             `json:"Meta"`
}

// OBReadDataOffer1 wraps the offers array.
type OBReadDataOffer1 struct {
	Offer []OBOffer1 `json:"Offer"`
}

// ────────────────────────────────────────────────────────────────────────────
// AIS Standing Orders (OBReadStandingOrder6)
// Ref: /resources-and-data-models/aisp/standing-orders/
// ────────────────────────────────────────────────────────────────────────────

// OBStandingOrder6 represents a standing order mandate as seen via AIS (read-only).
// This is distinct from PIS standing orders which are for initiation.
type OBStandingOrder6 struct {
	AccountId               string                               `json:"AccountId"`
	StandingOrderId         string                               `json:"StandingOrderId,omitempty"`
	// Frequency follows the OB frequency pattern (e.g. "IntrvlMnthDay:01:03")
	Frequency               string                               `json:"Frequency"`
	Reference               string                               `json:"Reference,omitempty"`
	FirstPaymentDateTime    *time.Time                           `json:"FirstPaymentDateTime,omitempty"`
	NextPaymentDateTime     *time.Time                           `json:"NextPaymentDateTime,omitempty"`
	LastPaymentDateTime     *time.Time                           `json:"LastPaymentDateTime,omitempty"`
	FinalPaymentDateTime    *time.Time                           `json:"FinalPaymentDateTime,omitempty"`
	// NumberOfPayments: how many payments have been made
	NumberOfPayments        string                               `json:"NumberOfPayments,omitempty"`
	// StandingOrderStatusCode: Active, Inactive
	StandingOrderStatusCode string                               `json:"StandingOrderStatusCode,omitempty"`
	FirstPaymentAmount      *OBActiveOrHistoricCurrencyAndAmount `json:"FirstPaymentAmount,omitempty"`
	NextPaymentAmount       *OBActiveOrHistoricCurrencyAndAmount `json:"NextPaymentAmount,omitempty"`
	LastPaymentAmount       *OBActiveOrHistoricCurrencyAndAmount `json:"LastPaymentAmount,omitempty"`
	FinalPaymentAmount      *OBActiveOrHistoricCurrencyAndAmount `json:"FinalPaymentAmount,omitempty"`
	CreditorAgent           *OBBranchAndFinancialInstitutionIdentification6 `json:"CreditorAgent,omitempty"`
	CreditorAccount         *OBCashAccount3                      `json:"CreditorAccount,omitempty"`
	SupplementaryData       map[string]any               `json:"SupplementaryData,omitempty"`
}

// OBReadStandingOrder6 is the response for GET /standing-orders and GET /accounts/{id}/standing-orders.
type OBReadStandingOrder6 struct {
	Data  OBReadDataStandingOrder6 `json:"Data"`
	Links Links                    `json:"Links"`
	Meta  Meta                     `json:"Meta"`
}

// OBReadDataStandingOrder6 wraps the standing orders array.
type OBReadDataStandingOrder6 struct {
	StandingOrder []OBStandingOrder6 `json:"StandingOrder"`
}

// ────────────────────────────────────────────────────────────────────────────
// Products (OBReadProduct2 v3.1.3)
// Ref: /resources-and-data-models/aisp/Products/
// ────────────────────────────────────────────────────────────────────────────

// OBProduct2 represents a product associated with an account.
type OBProduct2 struct {
	AccountId        string               `json:"AccountId"`
	ProductId        string               `json:"ProductId,omitempty"`
	// ProductType: PersonalCurrentAccount, BusinessCurrentAccount, etc.
	ProductType      string               `json:"ProductType"`
	MarketingStateId string               `json:"MarketingStateId,omitempty"`
	ProductName      string               `json:"ProductName,omitempty"`
	OtherProductType *OBOtherProductType1 `json:"OtherProductType,omitempty"`
	// BCA/PCA data models are complex and bank-specific — raw JSON for flexibility.
	BCA any `json:"BCA,omitempty"`
	PCA any `json:"PCA,omitempty"`
}

// OBOtherProductType1 holds custom product-type description.
type OBOtherProductType1 struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

// OBReadProduct2 is the response for GET /products and GET /accounts/{id}/product.
type OBReadProduct2 struct {
	Data  OBReadDataProduct2 `json:"Data"`
	Links Links              `json:"Links"`
	Meta  Meta               `json:"Meta"`
}

// OBReadDataProduct2 wraps the products array.
type OBReadDataProduct2 struct {
	Product []OBProduct2 `json:"Product"`
}

// ────────────────────────────────────────────────────────────────────────────
// Typed response wrappers for AISP resources
// ────────────────────────────────────────────────────────────────────────────

// OBReadAccount6 is the v3.1.3 response for GET /accounts.
type OBReadAccount6 struct {
	Data  OBReadDataAccount6 `json:"Data"`
	Links Links              `json:"Links"`
	Meta  Meta               `json:"Meta"`
}

type OBReadDataAccount6 struct {
	Account []OBAccount6 `json:"Account"`
}

// OBReadScheduledPayment3 is the response for GET /scheduled-payments.
type OBReadScheduledPayment3 struct {
	Data  OBReadDataScheduledPayment3 `json:"Data"`
	Links Links                       `json:"Links"`
	Meta  Meta                        `json:"Meta"`
}

type OBReadDataScheduledPayment3 struct {
	ScheduledPayment []OBScheduledPayment3 `json:"ScheduledPayment"`
}

// OBReadParty3 wraps multiple parties (bulk parties endpoint).
type OBReadParty3 struct {
	Data  OBReadDataParty3 `json:"Data"`
	Links Links            `json:"Links"`
	Meta  Meta             `json:"Meta"`
}

type OBReadDataParty3 struct {
	Party []OBParty2 `json:"Party"`
}

// OBReadParty2 wraps the single party returned by GET /party.
type OBReadParty2 struct {
	Data  OBReadDataParty2 `json:"Data"`
	Links Links            `json:"Links"`
	Meta  Meta             `json:"Meta"`
}

type OBReadDataParty2 struct {
	Party OBParty2 `json:"Party"`
}

// OBReadBeneficiary5 is the response for GET /beneficiaries.
type OBReadBeneficiary5 struct {
	Data  OBReadDataBeneficiary5 `json:"Data"`
	Links Links                  `json:"Links"`
	Meta  Meta                   `json:"Meta"`
}

type OBReadDataBeneficiary5 struct {
	Beneficiary []OBBeneficiary5 `json:"Beneficiary"`
}

// OBReadDirectDebit2 is the response for GET /direct-debits.
type OBReadDirectDebit2 struct {
	Data  OBReadDataDirectDebit2 `json:"Data"`
	Links Links                  `json:"Links"`
	Meta  Meta                   `json:"Meta"`
}

type OBReadDataDirectDebit2 struct {
	DirectDebit []OBDirectDebit2 `json:"DirectDebit"`
}

// OBReadTransaction6 is the response for GET /transactions.
type OBReadTransaction6 struct {
	Data  OBReadDataTransaction6 `json:"Data"`
	Links Links                  `json:"Links"`
	Meta  Meta                   `json:"Meta"`
}

type OBReadDataTransaction6 struct {
	Transaction []OBTransaction6 `json:"Transaction"`
}

// OBReadBalance1 is the response for GET /balances.
type OBReadBalance1 struct {
	Data  OBReadDataBalance1 `json:"Data"`
	Links Links              `json:"Links"`
	Meta  Meta               `json:"Meta"`
}

type OBReadDataBalance1 struct {
	Balance []OBBalance1 `json:"Balance"`
}

// OBReadStatement2 is the response for GET /statements.
type OBReadStatement2 struct {
	Data  OBReadDataStatement2 `json:"Data"`
	Links Links                `json:"Links"`
	Meta  Meta                 `json:"Meta"`
}

type OBReadDataStatement2 struct {
	Statement []OBStatement2 `json:"Statement"`
}

package models

import "time"

// ── Account Status ────────────────────────────────────────────────────────

// AccountStatus reflects the operational status of an account per the spec.
type AccountStatus string

const (
	AccountStatusEnabled  AccountStatus = "Enabled"
	AccountStatusDisabled AccountStatus = "Disabled"
	AccountStatusDeleted  AccountStatus = "Deleted"
	AccountStatusProForma AccountStatus = "ProForma"
	AccountStatusPending  AccountStatus = "Pending"
)

// SwitchStatus values defined by CASS (Current Account Switch Service).
type SwitchStatus string

const (
	SwitchStatusNotSwitched    SwitchStatus = "UK.CASS.NotSwitched"
	SwitchStatusSwitchCompleted SwitchStatus = "UK.CASS.SwitchCompleted"
)

// OBAccount6 represents a v3.1.3 account object returned by GET /accounts.
type OBAccount6 struct {
	AccountId            string          `json:"AccountId"`
	Status               AccountStatus   `json:"Status,omitempty"`
	StatusUpdateDateTime *time.Time      `json:"StatusUpdateDateTime,omitempty"`
	Currency             string          `json:"Currency"`
	AccountType          string          `json:"AccountType"`
	AccountSubType       string          `json:"AccountSubType"`
	Description          string          `json:"Description,omitempty"`
	Nickname             string          `json:"Nickname,omitempty"`
	OpeningDate          *time.Time      `json:"OpeningDate,omitempty"`
	MaturityDate         *time.Time      `json:"MaturityDate,omitempty"`
	// SwitchStatus indicates the CASS switch status (UK.CASS.NotSwitched | UK.CASS.SwitchCompleted).
	SwitchStatus SwitchStatus                                    `json:"SwitchStatus,omitempty"`
	Account      []OBCashAccount3                                `json:"Account,omitempty"`
	Servicer     *OBBranchAndFinancialInstitutionIdentification6 `json:"Servicer,omitempty"`
}

// GetAccountsResponse is the top-level response for GET /accounts.
type GetAccountsResponse struct {
	Data  GetAccountsData `json:"Data"`
	Links Links           `json:"Links"`
	Meta  Meta            `json:"Meta"`
}

type GetAccountsData struct {
	Account []OBAccount6 `json:"Account"`
}

// GetAccountResponse is the top-level response for GET /accounts/{AccountId}.
type GetAccountResponse struct {
	Data  GetAccountData `json:"Data"`
	Links Links          `json:"Links"`
	Meta  Meta           `json:"Meta"`
}

type GetAccountData struct {
	Account []OBAccount6 `json:"Account"`
}

// ── Balance ───────────────────────────────────────────────────────────────

// OBBalance1 represents a balance on an account.
type OBBalance1 struct {
	AccountId            string                              `json:"AccountId"`
	Amount               OBActiveOrHistoricCurrencyAndAmount `json:"Amount"`
	CreditDebitIndicator string                              `json:"CreditDebitIndicator"`
	Type                 string                              `json:"Type"`
	DateTime             time.Time                           `json:"DateTime"`
	CreditLine           []OBCreditLine1                     `json:"CreditLine,omitempty"`
}

type OBCreditLine1 struct {
	Included bool                                `json:"Included"`
	Amount   *OBActiveOrHistoricCurrencyAndAmount `json:"Amount,omitempty"`
	Type     string                              `json:"Type,omitempty"`
}

// GetBalancesResponse is the top-level response for GET /balances.
type GetBalancesResponse struct {
	Data  GetBalancesData `json:"Data"`
	Links Links           `json:"Links"`
	Meta  Meta            `json:"Meta"`
}

type GetBalancesData struct {
	Balance []OBBalance1 `json:"Balance"`
}

// ── Transaction ───────────────────────────────────────────────────────────

// TransactionMutability indicates whether a transaction is mutable.
type TransactionMutability string

const (
	TransactionMutable   TransactionMutability = "Mutable"
	TransactionImmutable TransactionMutability = "Immutable"
)

// OBTransaction6 represents a single transaction per v3.1.3 spec.
type OBTransaction6 struct {
	AccountId                      string                                       `json:"AccountId"`
	TransactionId                  string                                       `json:"TransactionId,omitempty"`
	TransactionReference           string                                       `json:"TransactionReference,omitempty"`
	StatementReference             []string                                     `json:"StatementReference,omitempty"`
	CreditDebitIndicator           string                                       `json:"CreditDebitIndicator"`
	Status                         string                                       `json:"Status"`
	// TransactionMutability indicates whether the transaction can be amended (v3.1.3).
	TransactionMutability          TransactionMutability                        `json:"TransactionMutability,omitempty"`
	BookingDateTime                time.Time                                    `json:"BookingDateTime"`
	ValueDateTime                  *time.Time                                   `json:"ValueDateTime,omitempty"`
	TransactionInformation         string                                       `json:"TransactionInformation,omitempty"`
	AddressLine                    string                                       `json:"AddressLine,omitempty"`
	Amount                         OBActiveOrHistoricCurrencyAndAmount          `json:"Amount"`
	ChargeAmount                   *OBActiveOrHistoricCurrencyAndAmount         `json:"ChargeAmount,omitempty"`
	CurrencyExchange               *OBCurrencyExchange5                         `json:"CurrencyExchange,omitempty"`
	BankTransactionCode            *OBBankTransactionCodeStructure1             `json:"BankTransactionCode,omitempty"`
	ProprietaryBankTransactionCode *OBProprietaryBankTransactionCodeStructure1  `json:"ProprietaryBankTransactionCode,omitempty"`
	Balance                        *OBTransactionCashBalance                    `json:"Balance,omitempty"`
	MerchantDetails                *OBMerchantDetails1                          `json:"MerchantDetails,omitempty"`
	CreditorAgent                  *OBBranchAndFinancialInstitutionIdentification6 `json:"CreditorAgent,omitempty"`
	CreditorAccount                *OBCashAccount3                              `json:"CreditorAccount,omitempty"`
	DebtorAgent                    *OBBranchAndFinancialInstitutionIdentification6 `json:"DebtorAgent,omitempty"`
	DebtorAccount                  *OBCashAccount3                              `json:"DebtorAccount,omitempty"`
	// CardInstrument holds payment card details used for the transaction (v3.1.3).
	CardInstrument                 *OBTransactionCardInstrument1                `json:"CardInstrument,omitempty"`
	SupplementaryData              map[string]any                       `json:"SupplementaryData,omitempty"`
}

// OBTransactionCardInstrument1 describes the payment card used in a transaction.
type OBTransactionCardInstrument1 struct {
	CardSchemeName    string `json:"CardSchemeName"`
	AuthorisationType string `json:"AuthorisationType,omitempty"`
	Name              string `json:"Name,omitempty"`
	// Identification holds a masked PAN, e.g. "****1234".
	Identification    string `json:"Identification,omitempty"`
}

// OBCurrencyExchange5 captures currency exchange data for a cross-currency transaction.
type OBCurrencyExchange5 struct {
	SourceCurrency   string  `json:"SourceCurrency"`
	TargetCurrency   string  `json:"TargetCurrency,omitempty"`
	UnitCurrency     string  `json:"UnitCurrency,omitempty"`
	ExchangeRate     float64 `json:"ExchangeRate"`
	ContractIdentification string `json:"ContractIdentification,omitempty"`
	QuotationDate    *time.Time `json:"QuotationDate,omitempty"`
	InstructedAmount *OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount,omitempty"`
}

type OBBankTransactionCodeStructure1 struct {
	Code    string `json:"Code"`
	SubCode string `json:"SubCode"`
}

type OBProprietaryBankTransactionCodeStructure1 struct {
	Code   string `json:"Code"`
	Issuer string `json:"Issuer,omitempty"`
}

type OBTransactionCashBalance struct {
	Amount               OBActiveOrHistoricCurrencyAndAmount `json:"Amount"`
	CreditDebitIndicator string                              `json:"CreditDebitIndicator"`
	Type                 string                              `json:"Type"`
}

type OBMerchantDetails1 struct {
	MerchantName         string `json:"MerchantName,omitempty"`
	MerchantCategoryCode string `json:"MerchantCategoryCode,omitempty"`
}

// GetTransactionsResponse is the top-level response for GET /transactions.
type GetTransactionsResponse struct {
	Data  GetTransactionsData `json:"Data"`
	Links Links               `json:"Links"`
	Meta  Meta                `json:"Meta"`
}

type GetTransactionsData struct {
	Transaction []OBTransaction6 `json:"Transaction"`
}

// ── Beneficiary ───────────────────────────────────────────────────────────

// BeneficiaryType classifies the beneficiary relationship (v3.1.3).
type BeneficiaryType string

const (
	BeneficiaryTypeUnidentified BeneficiaryType = "Unidentified"
	BeneficiaryTypePersonal     BeneficiaryType = "Personal"
	BeneficiaryTypeBusiness     BeneficiaryType = "Business"
)

// OBBeneficiary5 represents a beneficiary linked to an account (v3.1.3).
type OBBeneficiary5 struct {
	AccountId       string                                       `json:"AccountId,omitempty"`
	BeneficiaryId   string                                       `json:"BeneficiaryId,omitempty"`
	// BeneficiaryType classifies the beneficiary (v3.1.3 field).
	BeneficiaryType BeneficiaryType                              `json:"BeneficiaryType,omitempty"`
	Reference       string                                       `json:"Reference,omitempty"`
	CreditorAgent   *OBBranchAndFinancialInstitutionIdentification6 `json:"CreditorAgent,omitempty"`
	CreditorAccount *OBCashAccount3                              `json:"CreditorAccount,omitempty"`
}

// GetBeneficiariesResponse is the top-level response for GET /beneficiaries.
type GetBeneficiariesResponse struct {
	Data  GetBeneficiariesData `json:"Data"`
	Links Links                `json:"Links"`
	Meta  Meta                 `json:"Meta"`
}

type GetBeneficiariesData struct {
	Beneficiary []OBBeneficiary5 `json:"Beneficiary"`
}

// ── Direct Debit ──────────────────────────────────────────────────────────

// OBDirectDebit2 represents a direct debit mandate.
type OBDirectDebit2 struct {
	AccountId              string     `json:"AccountId"`
	DirectDebitId          string     `json:"DirectDebitId,omitempty"`
	MandateIdentification  string     `json:"MandateIdentification"`
	DirectDebitStatusCode  string     `json:"DirectDebitStatusCode,omitempty"`
	Name                   string     `json:"Name"`
	PreviousPaymentDateTime *time.Time `json:"PreviousPaymentDateTime,omitempty"`
	Frequency              string     `json:"Frequency,omitempty"`
	PreviousPaymentAmount  *OBActiveOrHistoricCurrencyAndAmount `json:"PreviousPaymentAmount,omitempty"`
}

// GetDirectDebitsResponse is the top-level response for GET /direct-debits.
type GetDirectDebitsResponse struct {
	Data  GetDirectDebitsData `json:"Data"`
	Links Links               `json:"Links"`
	Meta  Meta                `json:"Meta"`
}

type GetDirectDebitsData struct {
	DirectDebit []OBDirectDebit2 `json:"DirectDebit"`
}

// ── Scheduled Payment ─────────────────────────────────────────────────────

// OBScheduledPayment3 represents a scheduled payment per v3.1.3 spec.
type OBScheduledPayment3 struct {
	AccountId                string     `json:"AccountId"`
	ScheduledPaymentId       string     `json:"ScheduledPaymentId,omitempty"`
	ScheduledPaymentDateTime time.Time  `json:"ScheduledPaymentDateTime"`
	ScheduledType            string     `json:"ScheduledType"`
	Reference                string     `json:"Reference,omitempty"`
	// DebtorReference is the reference to the originating account (v3.1.3).
	DebtorReference          string     `json:"DebtorReference,omitempty"`
	InstructedAmount         OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
	CreditorAgent            *OBBranchAndFinancialInstitutionIdentification6 `json:"CreditorAgent,omitempty"`
	CreditorAccount          *OBCashAccount3 `json:"CreditorAccount,omitempty"`
}

// GetScheduledPaymentsResponse is the top-level response for GET /scheduled-payments.
type GetScheduledPaymentsResponse struct {
	Data  GetScheduledPaymentsData `json:"Data"`
	Links Links                    `json:"Links"`
	Meta  Meta                     `json:"Meta"`
}

type GetScheduledPaymentsData struct {
	ScheduledPayment []OBScheduledPayment3 `json:"ScheduledPayment"`
}

// ── Statement ─────────────────────────────────────────────────────────────

// OBStatement2 represents a statement.
type OBStatement2 struct {
	AccountId           string     `json:"AccountId"`
	StatementId         string     `json:"StatementId,omitempty"`
	StatementReference  string     `json:"StatementReference,omitempty"`
	Type                string     `json:"Type"`
	StartDateTime       time.Time  `json:"StartDateTime"`
	EndDateTime         time.Time  `json:"EndDateTime"`
	CreationDateTime    time.Time  `json:"CreationDateTime"`
	StatementDescription []string  `json:"StatementDescription,omitempty"`
	StatementBenefit    []OBStatementBenefit1   `json:"StatementBenefit,omitempty"`
	StatementFee        []OBStatementFee2       `json:"StatementFee,omitempty"`
	StatementInterest   []OBStatementInterest2  `json:"StatementInterest,omitempty"`
	StatementAmount     []OBStatementAmount1    `json:"StatementAmount,omitempty"`
	StatementDateTime   []OBStatementDateTime1  `json:"StatementDateTime,omitempty"`
	StatementRate       []OBStatementRate1      `json:"StatementRate,omitempty"`
	StatementValue      []OBStatementValue1     `json:"StatementValue,omitempty"`
}

type OBStatementBenefit1 struct {
	Amount OBActiveOrHistoricCurrencyAndAmount `json:"Amount"`
	Type   string                             `json:"Type"`
}

type OBStatementFee2 struct {
	Amount            OBActiveOrHistoricCurrencyAndAmount `json:"Amount"`
	CreditDebitIndicator string                          `json:"CreditDebitIndicator"`
	Type              string                             `json:"Type"`
	Rate              float64                            `json:"Rate,omitempty"`
	RateType          string                             `json:"RateType,omitempty"`
	Frequency         string                             `json:"Frequency,omitempty"`
}

type OBStatementInterest2 struct {
	Amount            OBActiveOrHistoricCurrencyAndAmount `json:"Amount"`
	CreditDebitIndicator string                          `json:"CreditDebitIndicator"`
	Type              string                             `json:"Type"`
	Rate              float64                            `json:"Rate,omitempty"`
	RateType          string                             `json:"RateType,omitempty"`
	Frequency         string                             `json:"Frequency,omitempty"`
}

type OBStatementAmount1 struct {
	Amount               OBActiveOrHistoricCurrencyAndAmount `json:"Amount"`
	CreditDebitIndicator string                             `json:"CreditDebitIndicator"`
	Type                 string                             `json:"Type"`
}

type OBStatementDateTime1 struct {
	DateTime time.Time `json:"DateTime"`
	Type     string    `json:"Type"`
}

type OBStatementRate1 struct {
	Rate     float64 `json:"Rate"`
	Type     string  `json:"Type"`
	RateType string  `json:"RateType,omitempty"`
}

type OBStatementValue1 struct {
	Value int    `json:"Value"`
	Type  string `json:"Type"`
}

// GetStatementsResponse is the top-level response for GET /statements.
type GetStatementsResponse struct {
	Data  GetStatementsData `json:"Data"`
	Links Links             `json:"Links"`
	Meta  Meta              `json:"Meta"`
}

type GetStatementsData struct {
	Statement []OBStatement2 `json:"Statement"`
}

// ── Party ─────────────────────────────────────────────────────────────────

// OBParty2 represents a party associated with an account.
type OBParty2 struct {
	PartyId             string              `json:"PartyId"`
	PartyNumber         string              `json:"PartyNumber,omitempty"`
	PartyType           string              `json:"PartyType,omitempty"`
	Name                string              `json:"Name,omitempty"`
	FullLegalName       string              `json:"FullLegalName,omitempty"`
	LegalStructure      string              `json:"LegalStructure,omitempty"`
	BeneficialOwnership bool                `json:"BeneficialOwnership,omitempty"`
	AccountRole         string              `json:"AccountRole,omitempty"`
	EmailAddress        string              `json:"EmailAddress,omitempty"`
	Phone               string              `json:"Phone,omitempty"`
	Mobile              string              `json:"Mobile,omitempty"`
	Relationships       *OBPartyRelationships1 `json:"Relationships,omitempty"`
	Address             []OBPostalAddress6  `json:"Address,omitempty"`
}

// OBPartyRelationships1 holds relationships between parties (v3.1.3 field).
type OBPartyRelationships1 struct {
	// Account references the account this party relationship is linked to.
	Account *OBPartyRelationship1 `json:"Account,omitempty"`
}

// OBPartyRelationship1 is a single party-to-resource relationship.
type OBPartyRelationship1 struct {
	Related string `json:"Related"`
	Id      string `json:"Id"`
}

// GetPartiesResponse is the top-level response for GET /parties.
type GetPartiesResponse struct {
	Data  GetPartiesData `json:"Data"`
	Links Links          `json:"Links"`
	Meta  Meta           `json:"Meta"`
}

type GetPartiesData struct {
	Party []OBParty2 `json:"Party"`
}

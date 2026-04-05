// Package models contains all OBIE v3.1.3 typed enumerations and constants.
// All values match the Namespaced Enumerations specification:
// https://openbankinguk.github.io/read-write-api-site2/standards/v3.1.3/references/namespaced-enumerations/
package models

// ── AIS Permission codes (OBExternalPermissions1Code) ─────────────────────

// Permission is a single AIS data-cluster permission code.
type Permission string

const (
	PermissionReadAccountsBasic           Permission = "ReadAccountsBasic"
	PermissionReadAccountsDetail          Permission = "ReadAccountsDetail"
	PermissionReadBalances                Permission = "ReadBalances"
	PermissionReadBeneficiariesBasic      Permission = "ReadBeneficiariesBasic"
	PermissionReadBeneficiariesDetail     Permission = "ReadBeneficiariesDetail"
	PermissionReadDirectDebits            Permission = "ReadDirectDebits"
	PermissionReadOffers                  Permission = "ReadOffers"
	PermissionReadPAN                     Permission = "ReadPAN"
	PermissionReadParty                   Permission = "ReadParty"
	PermissionReadPartyPSU                Permission = "ReadPartyPSU"
	PermissionReadProducts                Permission = "ReadProducts"
	PermissionReadScheduledPaymentsBasic  Permission = "ReadScheduledPaymentsBasic"
	PermissionReadScheduledPaymentsDetail Permission = "ReadScheduledPaymentsDetail"
	PermissionReadStandingOrdersBasic     Permission = "ReadStandingOrdersBasic"
	PermissionReadStandingOrdersDetail    Permission = "ReadStandingOrdersDetail"
	PermissionReadStatementsBasic         Permission = "ReadStatementsBasic"
	PermissionReadStatementsDetail        Permission = "ReadStatementsDetail"
	PermissionReadTransactionsBasic       Permission = "ReadTransactionsBasic"
	PermissionReadTransactionsCredits     Permission = "ReadTransactionsCredits"
	PermissionReadTransactionsDebits      Permission = "ReadTransactionsDebits"
	PermissionReadTransactionsDetail      Permission = "ReadTransactionsDetail"
)

// AllPermissions returns all 21 permission codes defined in the spec.
// Use this when creating a consent that requests full data access.
func AllPermissions() []Permission {
	return []Permission{
		PermissionReadAccountsDetail,
		PermissionReadBalances,
		PermissionReadBeneficiariesDetail,
		PermissionReadDirectDebits,
		PermissionReadOffers,
		PermissionReadPAN,
		PermissionReadParty,
		PermissionReadPartyPSU,
		PermissionReadProducts,
		PermissionReadScheduledPaymentsDetail,
		PermissionReadStandingOrdersDetail,
		PermissionReadStatementsDetail,
		PermissionReadTransactionsCredits,
		PermissionReadTransactionsDebits,
		PermissionReadTransactionsDetail,
	}
}

// ── Offer type codes (OBExternalOfferType1Code) ───────────────────────────

// OfferType classifies an account offer.
type OfferType string

const (
	OfferTypeBalanceTransfer OfferType = "BalanceTransfer"
	OfferTypeLimitIncrease   OfferType = "LimitIncrease"
	OfferTypeMoneyTransfer   OfferType = "MoneyTransfer"
	OfferTypeOther           OfferType = "Other"
	OfferTypePromotionalRate OfferType = "PromotionalRate"
)

// ── Account type / sub-type codes ────────────────────────────────────────

// OBExternalAccountType classifies the account.
type OBExternalAccountType string

const (
	AccountTypeBusiness OBExternalAccountType = "Business"
	AccountTypePersonal OBExternalAccountType = "Personal"
)

// OBExternalAccountSubType further classifies the account.
type OBExternalAccountSubType string

const (
	AccountSubTypeChargeCard     OBExternalAccountSubType = "ChargeCard"
	AccountSubTypeCreditCard     OBExternalAccountSubType = "CreditCard"
	AccountSubTypeCurrentAccount OBExternalAccountSubType = "CurrentAccount"
	AccountSubTypeEMoney         OBExternalAccountSubType = "EMoney"
	AccountSubTypeLoan           OBExternalAccountSubType = "Loan"
	AccountSubTypeMortgage       OBExternalAccountSubType = "Mortgage"
	AccountSubTypePrePaymentCard OBExternalAccountSubType = "PrePaymentCard"
	AccountSubTypeSavings        OBExternalAccountSubType = "Savings"
)

// ── Balance type codes ────────────────────────────────────────────────────

// OBBalanceType classifies a balance entry.
type OBBalanceType string

const (
	BalanceTypeClosingAvailable       OBBalanceType = "ClosingAvailable"
	BalanceTypeClosingBooked          OBBalanceType = "ClosingBooked"
	BalanceTypeClosingCleared         OBBalanceType = "ClosingCleared"
	BalanceTypeExpected               OBBalanceType = "Expected"
	BalanceTypeForwardAvailable       OBBalanceType = "ForwardAvailable"
	BalanceTypeInformation            OBBalanceType = "Information"
	BalanceTypeInterimAvailable       OBBalanceType = "InterimAvailable"
	BalanceTypeInterimBooked          OBBalanceType = "InterimBooked"
	BalanceTypeInterimCleared         OBBalanceType = "InterimCleared"
	BalanceTypeOpeningAvailable       OBBalanceType = "OpeningAvailable"
	BalanceTypeOpeningBooked          OBBalanceType = "OpeningBooked"
	BalanceTypeOpeningCleared         OBBalanceType = "OpeningCleared"
	BalanceTypePreviouslyClosedBooked OBBalanceType = "PreviouslyClosedBooked"
)

// ── Credit/Debit indicator ────────────────────────────────────────────────

// OBCreditDebitCode indicates whether an entry is credit or debit.
type OBCreditDebitCode string

const (
	CreditDebitCredit OBCreditDebitCode = "Credit"
	CreditDebitDebit  OBCreditDebitCode = "Debit"
)

// ── Scheduled payment type codes ─────────────────────────────────────────

// OBExternalScheduleType classifies a scheduled payment.
type OBExternalScheduleType string

const (
	ScheduledTypeArrival   OBExternalScheduleType = "Arrival"
	ScheduledTypeExecution OBExternalScheduleType = "Execution"
)

// ── Standing order frequency codes ───────────────────────────────────────

// OBExternalFrequency1Code defines how often a standing order repeats.
// Format: IntrvlMnthDay:xx:yy (every xx months on day yy)
type OBExternalFrequency1Code string

const (
	FrequencyEveryDay        OBExternalFrequency1Code = "EvryDay"
	FrequencyEveryWorkingDay OBExternalFrequency1Code = "EvryWorkgDay"
	FrequencyIntrvlDay       OBExternalFrequency1Code = "IntrvlDay"
	FrequencyIntrvlWkDay     OBExternalFrequency1Code = "IntrvlWkDay"
	FrequencyWkInMnthDay     OBExternalFrequency1Code = "WkInMnthDay"
	FrequencyIntrvlMnthDay   OBExternalFrequency1Code = "IntrvlMnthDay"
	FrequencyQtrDay          OBExternalFrequency1Code = "QtrDay"
)

// ── Payment instruction priority ─────────────────────────────────────────

// OBPriority2Code defines payment instruction priority.
type OBPriority2Code string

const (
	PriorityNormal OBPriority2Code = "Normal"
	PriorityUrgent OBPriority2Code = "Urgent"
)

// ── Charge bearer codes ───────────────────────────────────────────────────

// OBChargeBearerType1Code defines who bears the payment charges.
type OBChargeBearerType1Code string

const (
	ChargeBearerBorneByCreditor       OBChargeBearerType1Code = "BorneByCreditor"
	ChargeBearerBorneByDebtor         OBChargeBearerType1Code = "BorneByDebtor"
	ChargeBearerFollowingServiceLevel OBChargeBearerType1Code = "FollowingServiceLevel"
	ChargeBearerShared                OBChargeBearerType1Code = "Shared"
	// ChargeBearerSLEV follows the Service Level Agreement.
	ChargeBearerSLEV                  OBChargeBearerType1Code = "SLEV"
)

// ── Exchange rate type codes ──────────────────────────────────────────────

// OBExchangeRateType2Code classifies an exchange rate quote.
type OBExchangeRateType2Code string

const (
	ExchangeRateAgreed     OBExchangeRateType2Code = "Agreed"
	ExchangeRateActual     OBExchangeRateType2Code = "Actual"
	ExchangeRateIndicative OBExchangeRateType2Code = "Indicative"
)

// ── Event notification type codes (OBEventType1Code) ─────────────────────

// EventNotificationType is an OBIE event type URN.
type EventNotificationType string

const (
	EventNotificationResourceUpdate                           EventNotificationType = "urn:uk:org:openbanking:events:resource-update"
	EventNotificationConsentAuthorizationRevoked             EventNotificationType = "urn:uk:org:openbanking:events:consent-authorization-revoked"
	EventNotificationAccountAccessConsentLinkedAccountUpdate EventNotificationType = "urn:uk:org:openbanking:events:account-access-consent-linked-account-update"
)

// ── Account Consent Linked Account Update reason codes ───────────────────

// OBExternalLinkedAccountUpdateReason is the reason for a linked account update.
type OBExternalLinkedAccountUpdateReason string

const (
	// LinkedAccountUpdateAccountClosure indicates an account was closed.
	LinkedAccountUpdateAccountClosure OBExternalLinkedAccountUpdateReason = "UK.OBIE.AccountClosure"
	// LinkedAccountUpdateCASS indicates an account was switched via CASS.
	LinkedAccountUpdateCASS OBExternalLinkedAccountUpdateReason = "UK.OBIE.CASS"
)

// ── File payment type codes ───────────────────────────────────────────────

// FileType identifies the format of a bulk payment file.
type FileType string

const (
	FileTypeUK_OBIE_PaymentInitiation_2_1 FileType = "UK.OBIE.PaymentInitiation.2.1"
	FileTypeUK_OBIE_PaymentInitiation_3_1 FileType = "UK.OBIE.PaymentInitiation.3.1"
	FileTypeUK_OBIE_pain_001_001_08       FileType = "UK.OBIE.pain.001.001.08"
)

// ── Scheme name codes ─────────────────────────────────────────────────────

// OBExternalAccountIdentification4Code identifies the account numbering scheme.
type OBExternalAccountIdentification4Code string

const (
	SchemeNameUKOBIESortCodeAccountNumber OBExternalAccountIdentification4Code = "UK.OBIE.SortCodeAccountNumber"
	SchemeNameUKOBIEIBAN                  OBExternalAccountIdentification4Code = "UK.OBIE.IBAN"
	SchemeNameUKOBIEBBAN                  OBExternalAccountIdentification4Code = "UK.OBIE.BBAN"
	SchemeNameUKOBIEPAN                   OBExternalAccountIdentification4Code = "UK.OBIE.PAN"
	SchemeNameUKOBIEGetBranchCode         OBExternalAccountIdentification4Code = "UK.OBIE.GetBranchCode"
	SchemeNameUKOBIESWIFT                 OBExternalAccountIdentification4Code = "UK.OBIE.SWIFT"
	SchemeNameUKOBIEBICFI                 OBExternalAccountIdentification4Code = "UK.OBIE.BICFI"
)

// ── Financial institution identification scheme codes ─────────────────────

// OBExternalFinancialInstitutionIdentification4Code identifies a financial institution.
type OBExternalFinancialInstitutionIdentification4Code string

const (
	FinancialInstitutionSchemeBICFI OBExternalFinancialInstitutionIdentification4Code = "UK.OBIE.BICFI"
)

// ── External payment context codes ───────────────────────────────────────

// OBExternalPaymentContext1Code classifies the context of a payment.
type OBExternalPaymentContext1Code string

const (
	PaymentContextBillPayment       OBExternalPaymentContext1Code = "BillPayment"
	PaymentContextEcommerceGoods    OBExternalPaymentContext1Code = "EcommerceGoods"
	PaymentContextEcommerceServices OBExternalPaymentContext1Code = "EcommerceServices"
	PaymentContextOther             OBExternalPaymentContext1Code = "Other"
	PaymentContextPartyToParty      OBExternalPaymentContext1Code = "PartyToParty"
)

// ── External SCA exemption codes ─────────────────────────────────────────

// OBExternalExtendedProprietaryBankTransactionCode classifies SCA exemption types.
type OBExternalSCAExemptionType1Code string

const (
	SCAExemptionBillPayment      OBExternalSCAExemptionType1Code = "BillPayment"
	SCAExemptionContactlessTravel OBExternalSCAExemptionType1Code = "ContactlessTravel"
	SCAExemptionEcommerceGoods   OBExternalSCAExemptionType1Code = "EcommerceGoods"
	SCAExemptionEcommerceServices OBExternalSCAExemptionType1Code = "EcommerceServices"
	SCAExemptionKioskPayment     OBExternalSCAExemptionType1Code = "Kiosk"
	SCAExemptionPartyToParty     OBExternalSCAExemptionType1Code = "PartyToParty"
)

// ── Payment status codes (OBTransactionIndividualStatus1Code) ─────────────

// PaymentStatus represents the full set of payment status codes per spec.
type PaymentStatus string

const (
	// Terminal success statuses
	PaymentStatusAcceptedCreditSettlementCompleted PaymentStatus = "AcceptedCreditSettlementCompleted"
	PaymentStatusAcceptedSettlementCompleted       PaymentStatus = "AcceptedSettlementCompleted"

	// In-progress statuses
	PaymentStatusAcceptedSettlementInProcess PaymentStatus = "AcceptedSettlementInProcess"
	PaymentStatusAcceptedWithoutPosting      PaymentStatus = "AcceptedWithoutPosting"

	// Pending / awaiting
	PaymentStatusPending PaymentStatus = "Pending"

	// File payment specific statuses
	PaymentStatusInitiationCompleted PaymentStatus = "InitiationCompleted"
	PaymentStatusInitiationPending   PaymentStatus = "InitiationPending"
	PaymentStatusInitiationFailed    PaymentStatus = "InitiationFailed"

	// Terminal failure
	PaymentStatusRejected PaymentStatus = "Rejected"
)

// ── Consent status codes ──────────────────────────────────────────────────

// ConsentStatus represents the lifecycle status of any consent resource.
type ConsentStatus string

const (
	ConsentStatusAuthorised            ConsentStatus = "Authorised"
	ConsentStatusAwaitingAuthorisation ConsentStatus = "AwaitingAuthorisation"
	ConsentStatusConsumed              ConsentStatus = "Consumed"
	ConsentStatusRejected              ConsentStatus = "Rejected"
	ConsentStatusRevoked               ConsentStatus = "Revoked"
)

// ── File payment consent status (extends consent status) ─────────────────

// FileConsentStatus includes AwaitingUpload which is specific to file payments.
type FileConsentStatus string

const (
	FileConsentStatusAwaitingUpload        FileConsentStatus = "AwaitingUpload"
	FileConsentStatusAwaitingAuthorisation FileConsentStatus = "AwaitingAuthorisation"
	FileConsentStatusAuthorised            FileConsentStatus = "Authorised"
	FileConsentStatusConsumed              FileConsentStatus = "Consumed"
	FileConsentStatusRejected              FileConsentStatus = "Rejected"
)

// ── Card scheme codes ─────────────────────────────────────────────────────

// OBExternalCardSchemeType1Code identifies the payment card scheme.
type OBExternalCardSchemeType1Code string

const (
	CardSchemeAmericanExpress OBExternalCardSchemeType1Code = "AmericanExpress"
	CardSchemeDiners          OBExternalCardSchemeType1Code = "Diners"
	CardSchemeDiscover        OBExternalCardSchemeType1Code = "Discover"
	CardSchemeMasterCard      OBExternalCardSchemeType1Code = "MasterCard"
	CardSchemeVISA            OBExternalCardSchemeType1Code = "VISA"
)

// ── Card authorisation type codes ────────────────────────────────────────

// OBExternalCardAuthorisationType1Code identifies how a card transaction was authorised.
type OBExternalCardAuthorisationType1Code string

const (
	CardAuthTypeConsumerDevice      OBExternalCardAuthorisationType1Code = "ConsumerDevice"
	CardAuthTypeContactless         OBExternalCardAuthorisationType1Code = "Contactless"
	CardAuthTypeNone                OBExternalCardAuthorisationType1Code = "None"
	CardAuthTypePIN                 OBExternalCardAuthorisationType1Code = "PIN"
)

// ── OBIE error codes (OBErrorResponseError1Code) ──────────────────────────

// OBIEErrorCode is a typed OBIE error code for structured error handling.
type OBIEErrorCode string

const (
	// 400 errors
	OBIEErrorFieldExpected             OBIEErrorCode = "UK.OBIE.Field.Expected"
	OBIEErrorFieldInvalid              OBIEErrorCode = "UK.OBIE.Field.Invalid"
	OBIEErrorFieldInvalidDate          OBIEErrorCode = "UK.OBIE.Field.InvalidDate"
	OBIEErrorFieldMissing              OBIEErrorCode = "UK.OBIE.Field.Missing"
	OBIEErrorFieldUnexpected           OBIEErrorCode = "UK.OBIE.Field.Unexpected"
	OBIEErrorHeaderInvalid             OBIEErrorCode = "UK.OBIE.Header.Invalid"
	OBIEErrorHeaderMissing             OBIEErrorCode = "UK.OBIE.Header.Missing"
	OBIEErrorParameterInvalid          OBIEErrorCode = "UK.OBIE.Param.Invalid"
	OBIEErrorParameterMissing          OBIEErrorCode = "UK.OBIE.Param.Missing"
	OBIEErrorParameterUnexpected       OBIEErrorCode = "UK.OBIE.Param.Unexpected"
	OBIEErrorResourceConsentMismatch   OBIEErrorCode = "UK.OBIE.Resource.ConsentMismatch"
	OBIEErrorResourceInvalidConsentStatus OBIEErrorCode = "UK.OBIE.Resource.InvalidConsentStatus"
	OBIEErrorResourceInvalidFormat     OBIEErrorCode = "UK.OBIE.Resource.InvalidFormat"
	OBIEErrorSignatureInvalid          OBIEErrorCode = "UK.OBIE.Signature.Invalid"
	OBIEErrorSignatureMalformed        OBIEErrorCode = "UK.OBIE.Signature.Malformed"
	OBIEErrorSignatureMissing          OBIEErrorCode = "UK.OBIE.Signature.Missing"
	OBIEErrorSignatureInvalidClaim     OBIEErrorCode = "UK.OBIE.Signature.InvalidClaim"
	OBIEErrorUnexpectedError           OBIEErrorCode = "UK.OBIE.Unexpected.Error"

	// 401 errors
	OBIEErrorNotAuthorised OBIEErrorCode = "UK.OBIE.NotAuthorised"

	// 403 errors
	OBIEErrorResourceNotFound          OBIEErrorCode = "UK.OBIE.Resource.NotFound"
	OBIEErrorRuleAfterCutOffDateTime   OBIEErrorCode = "UK.OBIE.Rules.AfterCutOffDateTime"
	OBIEErrorRuleDuplicateReference    OBIEErrorCode = "UK.OBIE.Rules.DuplicateReference"
	OBIEErrorRuleFailsControlParameters OBIEErrorCode = "UK.OBIE.Rules.FailsControlParameters"

	// 409 errors
	OBIEErrorResourceDuplicate OBIEErrorCode = "UK.OBIE.Resource.Duplicate"

	// 429 errors
	OBIEErrorRulesMissingFields OBIEErrorCode = "UK.OBIE.Rules.MissingFields"
)

// ── VRP type codes ────────────────────────────────────────────────────────

// OBVRPType classifies the type of variable recurring payment.
type OBVRPType string

const (
	VRPTypeSweeping          OBVRPType = "UK.OBIE.VRPType.Sweeping"
	VRPTypeOther             OBVRPType = "UK.OBIE.VRPType.Other"
)

// ── PSU authentication method codes ──────────────────────────────────────

// OBVRPAuthenticationMethods classifies how a PSU authenticated for a VRP.
type OBVRPAuthenticationMethods string

const (
	PSUAuthMethodSCA     OBVRPAuthenticationMethods = "UK.OBIE.SCA"
	PSUAuthMethodSCANotRequired OBVRPAuthenticationMethods = "UK.OBIE.SCANotRequired"
)

// ── VRP periodic limit codes ──────────────────────────────────────────────

// OBVRPInteractionTypes classifies user-interaction types during VRP.
type OBVRPConsentType string

// OBPeriodType defines the period for VRP control parameters.
type OBPeriodType string

const (
	PeriodTypeDay       OBPeriodType = "Day"
	PeriodTypeWeek      OBPeriodType = "Week"
	PeriodTypeFortnight OBPeriodType = "Fortnight"
	PeriodTypeMonth     OBPeriodType = "Month"
	PeriodTypeHalfYear  OBPeriodType = "Half-year"
	PeriodTypeYear      OBPeriodType = "Year"
)

// OBPeriodAlignment defines whether the period aligns to calendar or consent start.
type OBPeriodAlignment string

const (
	PeriodAlignmentCalendar OBPeriodAlignment = "Calendar"
	PeriodAlignmentConsent  OBPeriodAlignment = "Consent"
)

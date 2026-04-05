package models

import "time"

// OBWriteDomesticConsent5 is the request body for creating a domestic payment consent.
type OBWriteDomesticConsent5 struct {
	Data OBWriteDomesticConsentData5 `json:"Data"`
	Risk OBRisk1                     `json:"Risk"`
}

type OBWriteDomesticConsentData5 struct {
	Permission              string                         `json:"Permission,omitempty"`
	ReadRefundAccount       string                         `json:"ReadRefundAccount,omitempty"`
	ExpirationDateTime      *time.Time                     `json:"ExpirationDateTime,omitempty"`
	Initiation              OBDomesticInitiation           `json:"Initiation"`
	Authorisation           *OBAuthorisation1              `json:"Authorisation,omitempty"`
	SCASupportData          *OBSCASupportData1             `json:"SCASupportData,omitempty"`
}

// OBDomesticInitiation represents the initiation details of a domestic payment.
type OBDomesticInitiation struct {
	InstructionIdentification    string                              `json:"InstructionIdentification"`
	EndToEndIdentification       string                              `json:"EndToEndIdentification"`
	LocalInstrument              string                              `json:"LocalInstrument,omitempty"`
	InstructedAmount             OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
	DebtorAccount                *OBCashAccount3                     `json:"DebtorAccount,omitempty"`
	CreditorAccount              OBCashAccount3                      `json:"CreditorAccount"`
	CreditorPostalAddress        *OBPostalAddress6                   `json:"CreditorPostalAddress,omitempty"`
	RemittanceInformation        *OBRemittanceInformation1           `json:"RemittanceInformation,omitempty"`
	SupplementaryData            map[string]any              `json:"SupplementaryData,omitempty"`
}

// OBRemittanceInformation1 carries remittance information.
type OBRemittanceInformation1 struct {
	Unstructured string `json:"Unstructured,omitempty"`
	Reference    string `json:"Reference,omitempty"`
}

// OBAuthorisation1 specifies the authorisation type for a consent.
type OBAuthorisation1 struct {
	// AuthorisationType: Any | Single
	AuthorisationType  OBExternalAuthorisation1Code `json:"AuthorisationType"`
	CompletionDateTime *time.Time `json:"CompletionDateTime,omitempty"`
}

// OBSCASupportData1 carries SCA support data.
type OBSCASupportData1 struct {
	RequestedSCAExemptionType  OBExternalSCAExemptionType1Code `json:"RequestedSCAExemptionType,omitempty"`
	AppliedAuthenticationApproach string `json:"AppliedAuthenticationApproach,omitempty"`
	ReferencePaymentOrderId    string `json:"ReferencePaymentOrderId,omitempty"`
}

// OBWriteDomesticConsentResponse5 is the response for a domestic payment consent.
type OBWriteDomesticConsentResponse5 struct {
	Data  OBWriteDomesticConsentResponseData5 `json:"Data"`
	Risk  OBRisk1                             `json:"Risk"`
	Links Links                               `json:"Links"`
	Meta  Meta                                `json:"Meta"`
}

type OBWriteDomesticConsentResponseData5 struct {
	ConsentId               string                         `json:"ConsentId"`
	CreationDateTime        time.Time                      `json:"CreationDateTime"`
	Status                  ConsentStatus                  `json:"Status"`
	StatusUpdateDateTime    time.Time                      `json:"StatusUpdateDateTime"`
	Permission              string                         `json:"Permission,omitempty"`
	ReadRefundAccount       string                         `json:"ReadRefundAccount,omitempty"`
	ExpirationDateTime      *time.Time                     `json:"ExpirationDateTime,omitempty"`
	Initiation              OBDomesticInitiation           `json:"Initiation"`
	Authorisation           *OBAuthorisation1              `json:"Authorisation,omitempty"`
	SCASupportData          *OBSCASupportData1             `json:"SCASupportData,omitempty"`
	Debtor                  *OBCashAccount3                `json:"Debtor,omitempty"`
	Charges                 []OBCharge2                    `json:"Charges,omitempty"`
}

// OBCharge2 represents a charge associated with a payment.
type OBCharge2 struct {
	ChargeBearer OBChargeBearerType1Code `json:"ChargeBearer"`
	Type         string                              `json:"Type"`
	Amount       OBActiveOrHistoricCurrencyAndAmount `json:"Amount"`
}

// OBWriteDomestic2 is the request body for submitting a domestic payment.
type OBWriteDomestic2 struct {
	Data OBWriteDomesticData2 `json:"Data"`
	Risk OBRisk1              `json:"Risk"`
}

type OBWriteDomesticData2 struct {
	ConsentId    string               `json:"ConsentId"`
	Initiation   OBDomesticInitiation `json:"Initiation"`
}

// OBWriteDomesticResponse5 is the response for a submitted domestic payment.
type OBWriteDomesticResponse5 struct {
	Data  OBWriteDomesticResponseData5 `json:"Data"`
	Links Links                        `json:"Links"`
	Meta  Meta                         `json:"Meta"`
}

type OBWriteDomesticResponseData5 struct {
	DomesticPaymentId       string               `json:"DomesticPaymentId"`
	ConsentId               string               `json:"ConsentId"`
	CreationDateTime        time.Time            `json:"CreationDateTime"`
	Status                  PaymentStatus        `json:"Status"`
	StatusUpdateDateTime    time.Time            `json:"StatusUpdateDateTime"`
	Initiation              OBDomesticInitiation `json:"Initiation"`
	MultiAuthorisation      *OBMultiAuthorisation1 `json:"MultiAuthorisation,omitempty"`
	Debtor                  *OBCashAccount3      `json:"Debtor,omitempty"`
	Charges                 []OBCharge2          `json:"Charges,omitempty"`
}

// OBMultiAuthorisation1 captures multi-auth status.
type OBMultiAuthorisation1 struct {
	// Status: Authorised | AwaitingFurtherAuthorisation | Rejected
	Status OBExternalMultiAuthorisation1Code `json:"Status"`
	NumberRequired          int        `json:"NumberRequired,omitempty"`
	NumberReceived          int        `json:"NumberReceived,omitempty"`
	LastUpdateDateTime      *time.Time `json:"LastUpdateDateTime,omitempty"`
	ExpirationDateTime      *time.Time `json:"ExpirationDateTime,omitempty"`
}

// ---- International Payment Models ----

// OBWriteInternationalConsent5 is the request body for an international payment consent.
type OBWriteInternationalConsent5 struct {
	Data OBWriteInternationalConsentData5 `json:"Data"`
	Risk OBRisk1                          `json:"Risk"`
}

type OBWriteInternationalConsentData5 struct {
	Permission         string                      `json:"Permission,omitempty"`
	ReadRefundAccount  string                      `json:"ReadRefundAccount,omitempty"`
	ExpirationDateTime *time.Time                  `json:"ExpirationDateTime,omitempty"`
	Initiation         OBInternationalInitiation   `json:"Initiation"`
	Authorisation      *OBAuthorisation1           `json:"Authorisation,omitempty"`
	SCASupportData     *OBSCASupportData1          `json:"SCASupportData,omitempty"`
}

// OBInternationalInitiation represents the initiation details of an international payment.
type OBInternationalInitiation struct {
	InstructionIdentification    string                              `json:"InstructionIdentification"`
	EndToEndIdentification       string                              `json:"EndToEndIdentification"`
	LocalInstrument              string                              `json:"LocalInstrument,omitempty"`
	InstructionPriority OBPriority2Code `json:"InstructionPriority,omitempty"`
	Purpose                      string                              `json:"Purpose,omitempty"`
	ExtendedPurpose              string                              `json:"ExtendedPurpose,omitempty"`
	ChargeBearer OBChargeBearerType1Code `json:"ChargeBearer,omitempty"`
	CurrencyOfTransfer           string                              `json:"CurrencyOfTransfer"`
	DestinationCountryCode       string                              `json:"DestinationCountryCode,omitempty"`
	InstructedAmount             OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
	ExchangeRateInformation      *OBExchangeRate1                    `json:"ExchangeRateInformation,omitempty"`
	DebtorAccount                *OBCashAccount3                     `json:"DebtorAccount,omitempty"`
	Creditor                     *OBPartyIdentification43            `json:"Creditor,omitempty"`
	CreditorAgent                *OBBranchAndFinancialInstitutionIdentification6 `json:"CreditorAgent,omitempty"`
	CreditorAccount              OBCashAccount3                      `json:"CreditorAccount"`
	RemittanceInformation        *OBRemittanceInformation1           `json:"RemittanceInformation,omitempty"`
}

type OBExchangeRate1 struct {
	// RateType: Agreed, Actual, or Indicative (OBExchangeRateType2Code)
	RateType      OBExchangeRateType2Code `json:"RateType"`
	UnitCurrency  string  `json:"UnitCurrency"`
	ExchangeRate  float64 `json:"ExchangeRate,omitempty"`
	ContractIdentification string `json:"ContractIdentification,omitempty"`
}

type OBPartyIdentification43 struct {
	Name          string            `json:"Name,omitempty"`
	PostalAddress *OBPostalAddress6 `json:"PostalAddress,omitempty"`
}

// OBWriteInternationalConsentResponse6 is the response for international payment consent.
type OBWriteInternationalConsentResponse6 struct {
	Data  OBWriteInternationalConsentResponseData6 `json:"Data"`
	Risk  OBRisk1                                  `json:"Risk"`
	Links Links                                    `json:"Links"`
	Meta  Meta                                     `json:"Meta"`
}

type OBWriteInternationalConsentResponseData6 struct {
	ConsentId            string                    `json:"ConsentId"`
	CreationDateTime     time.Time                 `json:"CreationDateTime"`
	Status               ConsentStatus             `json:"Status"`
	StatusUpdateDateTime time.Time                 `json:"StatusUpdateDateTime"`
	Permission           string                    `json:"Permission,omitempty"`
	ReadRefundAccount    string                    `json:"ReadRefundAccount,omitempty"`
	ExpirationDateTime   *time.Time                `json:"ExpirationDateTime,omitempty"`
	Initiation           OBInternationalInitiation `json:"Initiation"`
	Authorisation        *OBAuthorisation1         `json:"Authorisation,omitempty"`
	SCASupportData       *OBSCASupportData1        `json:"SCASupportData,omitempty"`
	Charges              []OBCharge2               `json:"Charges,omitempty"`
	ExchangeRateInformation *OBExchangeRate1        `json:"ExchangeRateInformation,omitempty"`
}

// OBWriteInternational3 is the request body for submitting an international payment.
type OBWriteInternational3 struct {
	Data OBWriteInternationalData3 `json:"Data"`
	Risk OBRisk1                   `json:"Risk"`
}

type OBWriteInternationalData3 struct {
	ConsentId  string                    `json:"ConsentId"`
	Initiation OBInternationalInitiation `json:"Initiation"`
}

// OBWriteInternationalResponse5 is the response for a submitted international payment.
type OBWriteInternationalResponse5 struct {
	Data  OBWriteInternationalResponseData5 `json:"Data"`
	Links Links                             `json:"Links"`
	Meta  Meta                              `json:"Meta"`
}

type OBWriteInternationalResponseData5 struct {
	InternationalPaymentId  string                    `json:"InternationalPaymentId"`
	ConsentId               string                    `json:"ConsentId"`
	CreationDateTime        time.Time                 `json:"CreationDateTime"`
	Status                  PaymentStatus             `json:"Status"`
	StatusUpdateDateTime    time.Time                 `json:"StatusUpdateDateTime"`
	Initiation              OBInternationalInitiation `json:"Initiation"`
	MultiAuthorisation      *OBMultiAuthorisation1    `json:"MultiAuthorisation,omitempty"`
	Debtor                  *OBCashAccount3           `json:"Debtor,omitempty"`
	Charges                 []OBCharge2               `json:"Charges,omitempty"`
	ExchangeRateInformation *OBExchangeRate1          `json:"ExchangeRateInformation,omitempty"`
}

// ---- Standing Order Models ----

// OBWriteDomesticStandingOrderConsent5 is the request body for a standing order consent.
type OBWriteDomesticStandingOrderConsent5 struct {
	Data OBWriteDomesticStandingOrderConsentData5 `json:"Data"`
	Risk OBRisk1                                  `json:"Risk"`
}

type OBWriteDomesticStandingOrderConsentData5 struct {
	Permission     string                        `json:"Permission,omitempty"`
	ExpirationDateTime *time.Time                `json:"ExpirationDateTime,omitempty"`
	Initiation     OBDomesticStandingOrderInitiation `json:"Initiation"`
	Authorisation  *OBAuthorisation1             `json:"Authorisation,omitempty"`
	SCASupportData *OBSCASupportData1            `json:"SCASupportData,omitempty"`
}

type OBDomesticStandingOrderInitiation struct {
	Frequency                    string                              `json:"Frequency"`
	Reference                    string                              `json:"Reference,omitempty"`
	NumberOfPayments             string                              `json:"NumberOfPayments,omitempty"`
	FirstPaymentDateTime         time.Time                           `json:"FirstPaymentDateTime"`
	RecurringPaymentDateTime     *time.Time                          `json:"RecurringPaymentDateTime,omitempty"`
	FinalPaymentDateTime         *time.Time                          `json:"FinalPaymentDateTime,omitempty"`
	FirstPaymentAmount           OBActiveOrHistoricCurrencyAndAmount `json:"FirstPaymentAmount"`
	RecurringPaymentAmount       *OBActiveOrHistoricCurrencyAndAmount `json:"RecurringPaymentAmount,omitempty"`
	FinalPaymentAmount           *OBActiveOrHistoricCurrencyAndAmount `json:"FinalPaymentAmount,omitempty"`
	DebtorAccount                *OBCashAccount3                     `json:"DebtorAccount,omitempty"`
	CreditorAccount              OBCashAccount3                      `json:"CreditorAccount"`
	SupplementaryData            map[string]any              `json:"SupplementaryData,omitempty"`
}

// VRP Models

// OBDomesticVRPConsentRequest is the request body for a VRP consent.
type OBDomesticVRPConsentRequest struct {
	Data OBDomesticVRPConsentRequestData `json:"Data"`
	Risk OBRisk1                         `json:"Risk"`
}

type OBDomesticVRPConsentRequestData struct {
	ControlParameters OBDomesticVRPControlParameters `json:"ControlParameters"`
	Initiation        OBDomesticVRPInitiation        `json:"Initiation"`
}

type OBDomesticVRPControlParameters struct {
	ValidFromDateTime     *time.Time                            `json:"ValidFromDateTime,omitempty"`
	ValidToDateTime       *time.Time                            `json:"ValidToDateTime,omitempty"`
	MaximumIndividualAmount OBActiveOrHistoricCurrencyAndAmount `json:"MaximumIndividualAmount"`
	PeriodicLimits        []OBDomesticVRPControlParametersPeriodic `json:"PeriodicLimits"`
	VRPType               []string                              `json:"VRPType"`
	PSUAuthenticationMethods []string                           `json:"PSUAuthenticationMethods"`
}

type OBDomesticVRPControlParametersPeriodic struct {
	PeriodAlignment string                              `json:"PeriodAlignment"`
	PeriodType      string                              `json:"PeriodType"`
	Amount          OBActiveOrHistoricCurrencyAndAmount `json:"Amount"`
}

type OBDomesticVRPInitiation struct {
	DebtorAccount         *OBCashAccount3           `json:"DebtorAccount,omitempty"`
	CreditorAccount       OBCashAccount3            `json:"CreditorAccount"`
	RemittanceInformation *OBRemittanceInformation1 `json:"RemittanceInformation,omitempty"`
}

// OBDomesticVRPConsentResponse is the response for a VRP consent.
type OBDomesticVRPConsentResponse struct {
	Data  OBDomesticVRPConsentResponseData `json:"Data"`
	Risk  OBRisk1                          `json:"Risk"`
	Links Links                            `json:"Links"`
	Meta  Meta                             `json:"Meta"`
}

type OBDomesticVRPConsentResponseData struct {
	ConsentId            string                          `json:"ConsentId"`
	CreationDateTime     time.Time                       `json:"CreationDateTime"`
	Status               ConsentStatus                   `json:"Status"`
	StatusUpdateDateTime time.Time                       `json:"StatusUpdateDateTime"`
	ControlParameters    OBDomesticVRPControlParameters  `json:"ControlParameters"`
	Initiation           OBDomesticVRPInitiation         `json:"Initiation"`
	DebtorAccount        *OBCashAccount3                 `json:"DebtorAccount,omitempty"`
}

// OBDomesticVRPRequest is the request body for submitting a VRP payment.
type OBDomesticVRPRequest struct {
	Data OBDomesticVRPRequestData `json:"Data"`
	Risk OBRisk1                  `json:"Risk"`
}

type OBDomesticVRPRequestData struct {
	ConsentId    string                  `json:"ConsentId"`
	PSUAuthenticationMethod string       `json:"PSUAuthenticationMethod"`
	Initiation   OBDomesticVRPInitiation `json:"Initiation"`
	Instruction  OBDomesticVRPInstruction `json:"Instruction"`
}

type OBDomesticVRPInstruction struct {
	InstructionIdentification string                              `json:"InstructionIdentification"`
	EndToEndIdentification    string                              `json:"EndToEndIdentification"`
	LocalInstrument           string                              `json:"LocalInstrument,omitempty"`
	InstructedAmount          OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
	CreditorAccount           OBCashAccount3                      `json:"CreditorAccount"`
	RemittanceInformation     *OBRemittanceInformation1           `json:"RemittanceInformation,omitempty"`
}

// OBDomesticVRPResponse is the response for a submitted VRP payment.
type OBDomesticVRPResponse struct {
	Data  OBDomesticVRPResponseData `json:"Data"`
	Links Links                     `json:"Links"`
	Meta  Meta                      `json:"Meta"`
}

type OBDomesticVRPResponseData struct {
	DomesticVRPId        string                   `json:"DomesticVRPId"`
	ConsentId            string                   `json:"ConsentId"`
	CreationDateTime     time.Time                `json:"CreationDateTime"`
	Status               PaymentStatus            `json:"Status"`
	StatusUpdateDateTime time.Time                `json:"StatusUpdateDateTime"`
	Initiation           OBDomesticVRPInitiation  `json:"Initiation"`
	Instruction          OBDomesticVRPInstruction `json:"Instruction"`
	DebtorAccount        *OBCashAccount3          `json:"DebtorAccount,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Domestic Standing Order — response models
// ────────────────────────────────────────────────────────────────────────────

// OBWriteDomesticStandingOrderConsentResponse6 is the response for a standing order consent.
type OBWriteDomesticStandingOrderConsentResponse6 struct {
	Data  OBWriteDomesticStandingOrderConsentResponseData6 `json:"Data"`
	Risk  OBRisk1                                          `json:"Risk"`
	Links Links                                            `json:"Links"`
	Meta  Meta                                             `json:"Meta"`
}

type OBWriteDomesticStandingOrderConsentResponseData6 struct {
	ConsentId               string                            `json:"ConsentId"`
	CreationDateTime        time.Time                         `json:"CreationDateTime"`
	Status                  ConsentStatus                     `json:"Status"`
	StatusUpdateDateTime    time.Time                         `json:"StatusUpdateDateTime"`
	Permission              string                            `json:"Permission,omitempty"`
	ExpirationDateTime      *time.Time                        `json:"ExpirationDateTime,omitempty"`
	Initiation              OBDomesticStandingOrderInitiation `json:"Initiation"`
	Authorisation           *OBAuthorisation1                 `json:"Authorisation,omitempty"`
	SCASupportData          *OBSCASupportData1                `json:"SCASupportData,omitempty"`
	Charges                 []OBCharge2                       `json:"Charges,omitempty"`
}

// OBWriteDomesticStandingOrder4 is the request body for submitting a standing order.
type OBWriteDomesticStandingOrder4 struct {
	Data OBWriteDomesticStandingOrderData4 `json:"Data"`
	Risk OBRisk1                           `json:"Risk"`
}

type OBWriteDomesticStandingOrderData4 struct {
	ConsentId  string                            `json:"ConsentId"`
	Initiation OBDomesticStandingOrderInitiation `json:"Initiation"`
}

// OBWriteDomesticStandingOrderResponse6 is the response for a submitted standing order.
type OBWriteDomesticStandingOrderResponse6 struct {
	Data  OBWriteDomesticStandingOrderResponseData6 `json:"Data"`
	Links Links                                     `json:"Links"`
	Meta  Meta                                      `json:"Meta"`
}

type OBWriteDomesticStandingOrderResponseData6 struct {
	DomesticStandingOrderId string                            `json:"DomesticStandingOrderId"`
	ConsentId               string                            `json:"ConsentId"`
	CreationDateTime        time.Time                         `json:"CreationDateTime"`
	Status                  PaymentStatus                     `json:"Status"`
	StatusUpdateDateTime    time.Time                         `json:"StatusUpdateDateTime"`
	Initiation              OBDomesticStandingOrderInitiation `json:"Initiation"`
	MultiAuthorisation      *OBMultiAuthorisation1            `json:"MultiAuthorisation,omitempty"`
	Debtor                  *OBCashAccount3                   `json:"Debtor,omitempty"`
	Charges                 []OBCharge2                       `json:"Charges,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Domestic Scheduled Payment — full typed models
// ────────────────────────────────────────────────────────────────────────────

// OBWriteDomesticScheduledConsent4 is the request body for creating a scheduled payment consent.
type OBWriteDomesticScheduledConsent4 struct {
	Data OBWriteDomesticScheduledConsentData4 `json:"Data"`
	Risk OBRisk1                              `json:"Risk"`
}

type OBWriteDomesticScheduledConsentData4 struct {
	Permission         string                             `json:"Permission,omitempty"`
	ExpirationDateTime *time.Time                         `json:"ExpirationDateTime,omitempty"`
	Initiation         OBDomesticScheduledInitiation      `json:"Initiation"`
	Authorisation      *OBAuthorisation1                  `json:"Authorisation,omitempty"`
	SCASupportData     *OBSCASupportData1                 `json:"SCASupportData,omitempty"`
}

// OBDomesticScheduledInitiation represents initiation details of a scheduled payment.
type OBDomesticScheduledInitiation struct {
	InstructionIdentification string                              `json:"InstructionIdentification"`
	EndToEndIdentification    string                              `json:"EndToEndIdentification,omitempty"`
	LocalInstrument           string                              `json:"LocalInstrument,omitempty"`
	RequestedExecutionDateTime time.Time                          `json:"RequestedExecutionDateTime"`
	InstructedAmount          OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
	DebtorAccount             *OBCashAccount3                     `json:"DebtorAccount,omitempty"`
	CreditorAccount           OBCashAccount3                      `json:"CreditorAccount"`
	CreditorPostalAddress     *OBPostalAddress6                   `json:"CreditorPostalAddress,omitempty"`
	RemittanceInformation     *OBRemittanceInformation1           `json:"RemittanceInformation,omitempty"`
	SupplementaryData         map[string]any              `json:"SupplementaryData,omitempty"`
}

// OBWriteDomesticScheduledConsentResponse4 is the response for a scheduled payment consent.
type OBWriteDomesticScheduledConsentResponse4 struct {
	Data  OBWriteDomesticScheduledConsentResponseData4 `json:"Data"`
	Risk  OBRisk1                                      `json:"Risk"`
	Links Links                                        `json:"Links"`
	Meta  Meta                                         `json:"Meta"`
}

type OBWriteDomesticScheduledConsentResponseData4 struct {
	ConsentId            string                        `json:"ConsentId"`
	CreationDateTime     time.Time                     `json:"CreationDateTime"`
	Status               ConsentStatus                 `json:"Status"`
	StatusUpdateDateTime time.Time                     `json:"StatusUpdateDateTime"`
	Permission           string                        `json:"Permission,omitempty"`
	ExpirationDateTime   *time.Time                    `json:"ExpirationDateTime,omitempty"`
	Initiation           OBDomesticScheduledInitiation `json:"Initiation"`
	Authorisation        *OBAuthorisation1             `json:"Authorisation,omitempty"`
	SCASupportData       *OBSCASupportData1            `json:"SCASupportData,omitempty"`
	Charges              []OBCharge2                   `json:"Charges,omitempty"`
}

// OBWriteDomesticScheduled3 is the request body for submitting a scheduled payment.
type OBWriteDomesticScheduled3 struct {
	Data OBWriteDomesticScheduledData3 `json:"Data"`
	Risk OBRisk1                       `json:"Risk"`
}

type OBWriteDomesticScheduledData3 struct {
	ConsentId  string                        `json:"ConsentId"`
	Initiation OBDomesticScheduledInitiation `json:"Initiation"`
}

// OBWriteDomesticScheduledResponse5 is the response for a submitted scheduled payment.
type OBWriteDomesticScheduledResponse5 struct {
	Data  OBWriteDomesticScheduledResponseData5 `json:"Data"`
	Links Links                                 `json:"Links"`
	Meta  Meta                                  `json:"Meta"`
}

type OBWriteDomesticScheduledResponseData5 struct {
	DomesticScheduledPaymentId string                        `json:"DomesticScheduledPaymentId"`
	ConsentId                  string                        `json:"ConsentId"`
	CreationDateTime           time.Time                     `json:"CreationDateTime"`
	Status                     PaymentStatus                 `json:"Status"`
	StatusUpdateDateTime       time.Time                     `json:"StatusUpdateDateTime"`
	Initiation                 OBDomesticScheduledInitiation `json:"Initiation"`
	MultiAuthorisation         *OBMultiAuthorisation1        `json:"MultiAuthorisation,omitempty"`
	Debtor                     *OBCashAccount3               `json:"Debtor,omitempty"`
	Charges                    []OBCharge2                   `json:"Charges,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// International Scheduled Payment
// ────────────────────────────────────────────────────────────────────────────

// OBWriteInternationalScheduledConsent5 is the request body for an international scheduled consent.
type OBWriteInternationalScheduledConsent5 struct {
	Data OBWriteInternationalScheduledConsentData5 `json:"Data"`
	Risk OBRisk1                                   `json:"Risk"`
}

type OBWriteInternationalScheduledConsentData5 struct {
	Permission         string                                `json:"Permission,omitempty"`
	ExpirationDateTime *time.Time                            `json:"ExpirationDateTime,omitempty"`
	Initiation         OBInternationalScheduledInitiation    `json:"Initiation"`
	Authorisation      *OBAuthorisation1                     `json:"Authorisation,omitempty"`
	SCASupportData     *OBSCASupportData1                    `json:"SCASupportData,omitempty"`
}

// OBInternationalScheduledInitiation represents initiation details of an international scheduled payment.
type OBInternationalScheduledInitiation struct {
	InstructionIdentification   string                              `json:"InstructionIdentification"`
	EndToEndIdentification      string                              `json:"EndToEndIdentification,omitempty"`
	LocalInstrument             string                              `json:"LocalInstrument,omitempty"`
	InstructionPriority OBPriority2Code `json:"InstructionPriority,omitempty"`
	Purpose                     string                              `json:"Purpose,omitempty"`
	ChargeBearer OBChargeBearerType1Code `json:"ChargeBearer,omitempty"`
	RequestedExecutionDateTime  time.Time                           `json:"RequestedExecutionDateTime"`
	CurrencyOfTransfer          string                              `json:"CurrencyOfTransfer"`
	DestinationCountryCode      string                              `json:"DestinationCountryCode,omitempty"`
	InstructedAmount            OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
	ExchangeRateInformation     *OBExchangeRate1                    `json:"ExchangeRateInformation,omitempty"`
	DebtorAccount               *OBCashAccount3                     `json:"DebtorAccount,omitempty"`
	Creditor                    *OBPartyIdentification43            `json:"Creditor,omitempty"`
	CreditorAgent               *OBBranchAndFinancialInstitutionIdentification6 `json:"CreditorAgent,omitempty"`
	CreditorAccount             OBCashAccount3                      `json:"CreditorAccount"`
	RemittanceInformation       *OBRemittanceInformation1           `json:"RemittanceInformation,omitempty"`
}

// OBWriteInternationalScheduledConsentResponse6 is the consent response.
type OBWriteInternationalScheduledConsentResponse6 struct {
	Data  OBWriteInternationalScheduledConsentResponseData6 `json:"Data"`
	Risk  OBRisk1                                           `json:"Risk"`
	Links Links                                             `json:"Links"`
	Meta  Meta                                              `json:"Meta"`
}

type OBWriteInternationalScheduledConsentResponseData6 struct {
	ConsentId            string                             `json:"ConsentId"`
	CreationDateTime     time.Time                          `json:"CreationDateTime"`
	Status               ConsentStatus                      `json:"Status"`
	StatusUpdateDateTime time.Time                          `json:"StatusUpdateDateTime"`
	Initiation           OBInternationalScheduledInitiation `json:"Initiation"`
	Charges              []OBCharge2                        `json:"Charges,omitempty"`
	ExchangeRateInformation *OBExchangeRate1                `json:"ExchangeRateInformation,omitempty"`
}

// OBWriteInternationalScheduled3 is the request body for submitting.
type OBWriteInternationalScheduled3 struct {
	Data OBWriteInternationalScheduledData3 `json:"Data"`
	Risk OBRisk1                            `json:"Risk"`
}

type OBWriteInternationalScheduledData3 struct {
	ConsentId  string                             `json:"ConsentId"`
	Initiation OBInternationalScheduledInitiation `json:"Initiation"`
}

// OBWriteInternationalScheduledResponse6 is the response for a submitted international scheduled payment.
type OBWriteInternationalScheduledResponse6 struct {
	Data  OBWriteInternationalScheduledResponseData6 `json:"Data"`
	Links Links                                      `json:"Links"`
	Meta  Meta                                       `json:"Meta"`
}

type OBWriteInternationalScheduledResponseData6 struct {
	InternationalScheduledPaymentId string                             `json:"InternationalScheduledPaymentId"`
	ConsentId                       string                             `json:"ConsentId"`
	CreationDateTime                time.Time                          `json:"CreationDateTime"`
	Status                          PaymentStatus                      `json:"Status"`
	StatusUpdateDateTime            time.Time                          `json:"StatusUpdateDateTime"`
	Initiation                      OBInternationalScheduledInitiation `json:"Initiation"`
	MultiAuthorisation              *OBMultiAuthorisation1             `json:"MultiAuthorisation,omitempty"`
	ExchangeRateInformation         *OBExchangeRate1                   `json:"ExchangeRateInformation,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// International Standing Order
// ────────────────────────────────────────────────────────────────────────────

// OBWriteInternationalStandingOrderConsent6 is the request body for an international standing order consent.
type OBWriteInternationalStandingOrderConsent6 struct {
	Data OBWriteInternationalStandingOrderConsentData6 `json:"Data"`
	Risk OBRisk1                                       `json:"Risk"`
}

type OBWriteInternationalStandingOrderConsentData6 struct {
	Permission     string                                   `json:"Permission,omitempty"`
	ExpirationDateTime *time.Time                           `json:"ExpirationDateTime,omitempty"`
	Initiation     OBInternationalStandingOrderInitiation6  `json:"Initiation"`
	Authorisation  *OBAuthorisation1                        `json:"Authorisation,omitempty"`
	SCASupportData *OBSCASupportData1                       `json:"SCASupportData,omitempty"`
}

// OBInternationalStandingOrderInitiation6 is the initiation block.
type OBInternationalStandingOrderInitiation6 struct {
	Frequency                    string                              `json:"Frequency"`
	Reference                    string                              `json:"Reference,omitempty"`
	NumberOfPayments             string                              `json:"NumberOfPayments,omitempty"`
	FirstPaymentDateTime         time.Time                           `json:"FirstPaymentDateTime"`
	FinalPaymentDateTime         *time.Time                          `json:"FinalPaymentDateTime,omitempty"`
	Purpose                      string                              `json:"Purpose,omitempty"`
	ChargeBearer OBChargeBearerType1Code `json:"ChargeBearer,omitempty"`
	CurrencyOfTransfer           string                              `json:"CurrencyOfTransfer"`
	InstructedAmount             OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
	DebtorAccount                *OBCashAccount3                     `json:"DebtorAccount,omitempty"`
	Creditor                     *OBPartyIdentification43            `json:"Creditor,omitempty"`
	CreditorAgent                *OBBranchAndFinancialInstitutionIdentification6 `json:"CreditorAgent,omitempty"`
	CreditorAccount              OBCashAccount3                      `json:"CreditorAccount"`
}

// OBWriteInternationalStandingOrderConsentResponse7 is the consent response.
type OBWriteInternationalStandingOrderConsentResponse7 struct {
	Data  OBWriteInternationalStandingOrderConsentResponseData7 `json:"Data"`
	Risk  OBRisk1                                               `json:"Risk"`
	Links Links                                                 `json:"Links"`
	Meta  Meta                                                  `json:"Meta"`
}

type OBWriteInternationalStandingOrderConsentResponseData7 struct {
	ConsentId            string                                  `json:"ConsentId"`
	CreationDateTime     time.Time                               `json:"CreationDateTime"`
	Status               ConsentStatus                           `json:"Status"`
	StatusUpdateDateTime time.Time                               `json:"StatusUpdateDateTime"`
	Initiation           OBInternationalStandingOrderInitiation6 `json:"Initiation"`
}

// OBWriteInternationalStandingOrder6 is the request for submitting an international standing order.
type OBWriteInternationalStandingOrder6 struct {
	Data OBWriteInternationalStandingOrderData6 `json:"Data"`
	Risk OBRisk1                                `json:"Risk"`
}

type OBWriteInternationalStandingOrderData6 struct {
	ConsentId  string                                  `json:"ConsentId"`
	Initiation OBInternationalStandingOrderInitiation6 `json:"Initiation"`
}

// OBWriteInternationalStandingOrderResponse7 is the response for a submitted international standing order.
type OBWriteInternationalStandingOrderResponse7 struct {
	Data  OBWriteInternationalStandingOrderResponseData7 `json:"Data"`
	Links Links                                          `json:"Links"`
	Meta  Meta                                           `json:"Meta"`
}

type OBWriteInternationalStandingOrderResponseData7 struct {
	InternationalStandingOrderId string                                  `json:"InternationalStandingOrderId"`
	ConsentId                    string                                  `json:"ConsentId"`
	CreationDateTime             time.Time                               `json:"CreationDateTime"`
	Status                       PaymentStatus                           `json:"Status"`
	StatusUpdateDateTime         time.Time                               `json:"StatusUpdateDateTime"`
	Initiation                   OBInternationalStandingOrderInitiation6 `json:"Initiation"`
	MultiAuthorisation           *OBMultiAuthorisation1                  `json:"MultiAuthorisation,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Payment Details (generic status detail for any payment type)
// ────────────────────────────────────────────────────────────────────────────

// OBPaymentDetailsStatus1 holds detailed multi-step payment status.
type OBPaymentDetailsStatus1 struct {
	LocalInstrument   string                                `json:"LocalInstrument,omitempty"`
	Status            PaymentStatus                         `json:"Status"`
	StatusDetail      []OBPaymentDetailsStatusDetail1       `json:"StatusDetail,omitempty"`
}

// OBPaymentDetailsStatusDetail1 provides a reason code for a payment status step.
type OBPaymentDetailsStatusDetail1 struct {
	LocalInstrument     string `json:"LocalInstrument,omitempty"`
	// Status: Authorised | AwaitingFurtherAuthorisation | Rejected
	Status OBExternalMultiAuthorisation1Code `json:"Status"`
	StatusReason        string `json:"StatusReason,omitempty"`
	StatusReasonDescription string `json:"StatusReasonDescription,omitempty"`
}

// OBWritePaymentDetailsResponse1 wraps a payment details response.
type OBWritePaymentDetailsResponse1 struct {
	Data  OBWritePaymentDetailsResponseData1 `json:"Data"`
	Links Links                              `json:"Links"`
	Meta  Meta                               `json:"Meta"`
}

type OBWritePaymentDetailsResponseData1 struct {
	PaymentStatus []OBPaymentDetailsStatus1 `json:"PaymentStatus"`
}

// ────────────────────────────────────────────────────────────────────────────
// VRP Funds Confirmation typed models
// ────────────────────────────────────────────────────────────────────────────

// OBVRPFundsConfirmationRequest is the request body for a VRP funds confirmation check.
type OBVRPFundsConfirmationRequest struct {
	Data OBVRPFundsConfirmationRequestData `json:"Data"`
}

type OBVRPFundsConfirmationRequestData struct {
	ConsentId        string                              `json:"ConsentId"`
	Reference        string                              `json:"Reference"`
	InstructedAmount OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
}

// OBVRPFundsConfirmationResponse is the response for a VRP funds confirmation check.
type OBVRPFundsConfirmationResponse struct {
	Data  OBVRPFundsConfirmationResponseData `json:"Data"`
	Links Links                              `json:"Links"`
	Meta  Meta                               `json:"Meta"`
}

type OBVRPFundsConfirmationResponseData struct {
	FundsConfirmationId string    `json:"FundsConfirmationId"`
	ConsentId           string    `json:"ConsentId"`
	CreationDateTime    time.Time `json:"CreationDateTime"`
	FundsAvailable      bool      `json:"FundsAvailable"`
	Reference           string    `json:"Reference"`
	InstructedAmount    OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
}

// ────────────────────────────────────────────────────────────────────────────
// Typed enums for payment-specific fields
// ────────────────────────────────────────────────────────────────────────────

// OBExternalPaymentConsentPermission1Code defines the permission scope for a payment consent.
type OBExternalPaymentConsentPermission1Code string

const (
	PaymentConsentPermissionCreate      OBExternalPaymentConsentPermission1Code = "Create"
	PaymentConsentPermissionCreateAndRefund OBExternalPaymentConsentPermission1Code = "CreateAndRefund"
)

// OBReadRefundAccount1Code specifies whether to return debtor account details.
type OBReadRefundAccount1Code string

const (
	ReadRefundAccountNo  OBReadRefundAccount1Code = "No"
	ReadRefundAccountYes OBReadRefundAccount1Code = "Yes"
)

// OBExternalAuthorisation1Code defines the multi-authorisation type.
type OBExternalAuthorisation1Code string

const (
	AuthorisationTypeAny  OBExternalAuthorisation1Code = "Any"
	AuthorisationTypeSingle OBExternalAuthorisation1Code = "Single"
)

// OBExternalMultiAuthorisation1Code defines multi-authorisation status values.
type OBExternalMultiAuthorisation1Code string

const (
	MultiAuthStatusAuthorised        OBExternalMultiAuthorisation1Code = "Authorised"
	MultiAuthStatusAwaitingFurtherAuth OBExternalMultiAuthorisation1Code = "AwaitingFurtherAuthorisation"
	MultiAuthStatusRejected          OBExternalMultiAuthorisation1Code = "Rejected"
)

// OBExternalLocalInstrument1Code identifies local instrument codes.
// These are ASPSP-specific and defined in each bank's documentation.
// Common UK examples:
const (
	LocalInstrumentCHAPS    = "UK.OBIE.CHAPS"
	LocalInstrumentFPS      = "UK.OBIE.FPS"
	LocalInstrumentBACS     = "UK.OBIE.Bacs"
	LocalInstrumentLink     = "UK.OBIE.Link"
	LocalInstrumentPaySpark = "UK.OBIE.PaySpark"
)

// OBExternalRequestStatus1Code defines consent request status values.
// (Alias of ConsentStatus defined in enums.go — kept for explicit payment usage)
type OBExternalConsentStatus1Code = ConsentStatus

// OBExternalPaymentTransactionStatus1Code status values returned in payment responses.
// (Alias of PaymentStatus defined in enums.go)
type OBExternalPaymentTransactionStatus1Code = PaymentStatus

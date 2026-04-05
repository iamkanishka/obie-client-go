package models

import "time"

// OBFundsConfirmationConsent1 is the request body for creating a funds confirmation consent.
type OBFundsConfirmationConsent1 struct {
	Data OBFundsConfirmationConsentData1 `json:"Data"`
}

type OBFundsConfirmationConsentData1 struct {
	ExpirationDateTime *time.Time     `json:"ExpirationDateTime,omitempty"`
	DebtorAccount      OBCashAccount3 `json:"DebtorAccount"`
}

// OBFundsConfirmationConsentResponse1 is the response for a funds confirmation consent.
type OBFundsConfirmationConsentResponse1 struct {
	Data  OBFundsConfirmationConsentResponseData1 `json:"Data"`
	Links Links                                   `json:"Links"`
	Meta  Meta                                    `json:"Meta"`
}

type OBFundsConfirmationConsentResponseData1 struct {
	ConsentId            string         `json:"ConsentId"`
	CreationDateTime     time.Time      `json:"CreationDateTime"`
	Status               ConsentStatus  `json:"Status"`
	StatusUpdateDateTime time.Time      `json:"StatusUpdateDateTime"`
	ExpirationDateTime   *time.Time     `json:"ExpirationDateTime,omitempty"`
	DebtorAccount        OBCashAccount3 `json:"DebtorAccount"`
}

// OBFundsConfirmation1 is the request body for confirming funds availability.
type OBFundsConfirmation1 struct {
	Data OBFundsConfirmationData1 `json:"Data"`
}

type OBFundsConfirmationData1 struct {
	ConsentId      string                              `json:"ConsentId"`
	Reference      string                              `json:"Reference"`
	InstructedAmount OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
}

// OBFundsConfirmationResponse1 is the response for a funds confirmation check.
type OBFundsConfirmationResponse1 struct {
	Data  OBFundsConfirmationResponseData1 `json:"Data"`
	Links Links                            `json:"Links"`
	Meta  Meta                             `json:"Meta"`
}

type OBFundsConfirmationResponseData1 struct {
	FundsConfirmationId string    `json:"FundsConfirmationId"`
	ConsentId           string    `json:"ConsentId"`
	CreationDateTime    time.Time `json:"CreationDateTime"`
	FundsAvailable      bool      `json:"FundsAvailable"`
	Reference           string    `json:"Reference"`
	InstructedAmount    OBActiveOrHistoricCurrencyAndAmount `json:"InstructedAmount"`
}

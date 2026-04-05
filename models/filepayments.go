package models

import "time"

// ────────────────────────────────────────────────────────────────────────────
// File Payment Consent
// Ref: /resources-and-data-models/pisp/file-payment-consents/
// ────────────────────────────────────────────────────────────────────────────

// OBWriteFileConsent3 is the request body for POST /file-payment-consents.
type OBWriteFileConsent3 struct {
	Data OBWriteFileConsentData3 `json:"Data"`
	Risk OBRisk1                 `json:"Risk"`
}

// OBWriteFileConsentData3 carries the file consent initiation.
type OBWriteFileConsentData3 struct {
	Initiation     OBFile2            `json:"Initiation"`
	Authorisation  *OBAuthorisation1  `json:"Authorisation,omitempty"`
	SCASupportData *OBSCASupportData1 `json:"SCASupportData,omitempty"`
}

// OBFile2 describes the bulk payment file metadata.
// FileHash must be a SHA256 hash of the file content, Base64 encoded.
// NumberOfTransactions must match the actual count in the file.
// ControlSum must match the sum of all transaction amounts in the file.
type OBFile2 struct {
	// FileType: UK.OBIE.PaymentInitiation.3.1, UK.OBIE.pain.001.001.08, etc.
	FileType             FileType   `json:"FileType"`
	// FileHash: SHA256 hash of the file, Base64-encoded.
	FileHash             string     `json:"FileHash"`
	FileReference        string     `json:"FileReference,omitempty"`
	// NumberOfTransactions must exactly match the count in the uploaded file.
	NumberOfTransactions string     `json:"NumberOfTransactions,omitempty"`
	// ControlSum must exactly match the sum of all instructed amounts in the file.
	ControlSum           *float64   `json:"ControlSum,omitempty"`
	RequestedExecutionDateTime *time.Time `json:"RequestedExecutionDateTime,omitempty"`
	LocalInstrument      string     `json:"LocalInstrument,omitempty"`
	DebtorAccount        *OBCashAccount3 `json:"DebtorAccount,omitempty"`
	RemittanceInformation *OBRemittanceInformation1 `json:"RemittanceInformation,omitempty"`
	SupplementaryData    map[string]any `json:"SupplementaryData,omitempty"`
}

// OBWriteFileConsentResponse4 is the response for POST/GET /file-payment-consents/{id}.
type OBWriteFileConsentResponse4 struct {
	Data  OBWriteFileConsentResponseData4 `json:"Data"`
	Risk  OBRisk1                         `json:"Risk"`
	Links Links                           `json:"Links"`
	Meta  Meta                            `json:"Meta"`
}

// OBWriteFileConsentResponseData4 extends the request with ASPSP-assigned fields.
// Status transitions: AwaitingUpload → AwaitingAuthorisation → Authorised → Consumed/Rejected
type OBWriteFileConsentResponseData4 struct {
	ConsentId            string            `json:"ConsentId"`
	CreationDateTime     time.Time         `json:"CreationDateTime"`
	// Status: AwaitingUpload | AwaitingAuthorisation | Authorised | Consumed | Rejected
	Status               FileConsentStatus `json:"Status"`
	StatusUpdateDateTime time.Time         `json:"StatusUpdateDateTime"`
	Charges              []OBCharge2       `json:"Charges,omitempty"`
	Initiation           OBFile2           `json:"Initiation"`
	Authorisation        *OBAuthorisation1 `json:"Authorisation,omitempty"`
	SCASupportData       *OBSCASupportData1 `json:"SCASupportData,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// File Payment Submission
// Ref: /resources-and-data-models/pisp/file-payments/
// ────────────────────────────────────────────────────────────────────────────

// OBWriteFile2 is the request body for POST /file-payments.
// The file itself is uploaded separately via POST /file-payment-consents/{id}/file.
type OBWriteFile2 struct {
	Data OBWriteFileData2 `json:"Data"`
	Risk OBRisk1          `json:"Risk"`
}

// OBWriteFileData2 references the consent and repeats the initiation for verification.
type OBWriteFileData2 struct {
	ConsentId  string  `json:"ConsentId"`
	Initiation OBFile2 `json:"Initiation"`
}

// OBWriteFileResponse3 is the response for POST/GET /file-payments.
type OBWriteFileResponse3 struct {
	Data  OBWriteFileResponseData3 `json:"Data"`
	Links Links                    `json:"Links"`
	Meta  Meta                     `json:"Meta"`
}

// OBWriteFileResponseData3 carries the submitted file payment details.
// Status: Pending | InitiationPending | InitiationFailed | InitiationCompleted
type OBWriteFileResponseData3 struct {
	FilePaymentId        string         `json:"FilePaymentId"`
	ConsentId            string         `json:"ConsentId"`
	CreationDateTime     time.Time      `json:"CreationDateTime"`
	Status               PaymentStatus  `json:"Status"`
	StatusUpdateDateTime time.Time      `json:"StatusUpdateDateTime"`
	Charges              []OBCharge2    `json:"Charges,omitempty"`
	Initiation           OBFile2        `json:"Initiation"`
	Debtor               *OBCashAccount3 `json:"Debtor,omitempty"`
	MultiAuthorisation   *OBMultiAuthorisation1 `json:"MultiAuthorisation,omitempty"`
}

// OBWriteFilePaymentDetailsResponse1 is the response for payment details.
type OBWriteFilePaymentDetailsResponse1 struct {
	Data  OBWriteFilePaymentDetailsResponseData1 `json:"Data"`
	Links Links                                  `json:"Links"`
	Meta  Meta                                   `json:"Meta"`
}

// OBWriteFilePaymentDetailsResponseData1 wraps file payment status details.
type OBWriteFilePaymentDetailsResponseData1 struct {
	PaymentStatus []OBPaymentDetailsStatus1 `json:"PaymentStatus"`
}

package models

import "time"

// OBError represents the top-level OBIE API error response.
type OBError struct {
	Code    string        `json:"Code"`
	Message string        `json:"Message"`
	Errors  []OBErrorDetail `json:"Errors,omitempty"`
}

// OBErrorDetail provides detail on a single validation or processing error.
type OBErrorDetail struct {
	ErrorCode OBIEErrorCode `json:"ErrorCode"`
	Message   string        `json:"Message"`
	Path      string        `json:"Path,omitempty"`
	URL       string        `json:"Url,omitempty"`
}

// Links represents HATEOAS links returned in OBIE responses.
type Links struct {
	Self  string `json:"Self"`
	First string `json:"First,omitempty"`
	Prev  string `json:"Prev,omitempty"`
	Next  string `json:"Next,omitempty"`
	Last  string `json:"Last,omitempty"`
}

// Meta contains pagination and availability metadata.
type Meta struct {
	TotalPages             int        `json:"TotalPages,omitempty"`
	FirstAvailableDateTime *time.Time `json:"FirstAvailableDateTime,omitempty"`
	LastAvailableDateTime  *time.Time `json:"LastAvailableDateTime,omitempty"`
}

// OBActiveOrHistoricCurrencyAndAmount represents a monetary amount with ISO 4217 currency.
// Amount format: up to 13 digits before decimal, up to 5 after: ^\d{1,13}\.\d{1,5}$
type OBActiveOrHistoricCurrencyAndAmount struct {
	Amount   string `json:"Amount"`
	Currency string `json:"Currency"`
}

// OBPostalAddress6 represents a postal address per ISO 20022.
type OBPostalAddress6 struct {
	AddressType        string   `json:"AddressType,omitempty"`
	Department         string   `json:"Department,omitempty"`
	SubDepartment      string   `json:"SubDepartment,omitempty"`
	StreetName         string   `json:"StreetName,omitempty"`
	BuildingNumber     string   `json:"BuildingNumber,omitempty"`
	PostCode           string   `json:"PostCode,omitempty"`
	TownName           string   `json:"TownName,omitempty"`
	CountrySubDivision string   `json:"CountrySubDivision,omitempty"`
	// Country is an ISO 3166-1 alpha-2 code (e.g. "GB", "US").
	Country            string   `json:"Country,omitempty"`
	AddressLine        []string `json:"AddressLine,omitempty"`
}

// OBCashAccount3 represents an account identification block.
// SchemeName values: UK.OBIE.SortCodeAccountNumber, UK.OBIE.IBAN, UK.OBIE.BBAN, UK.OBIE.PAN
type OBCashAccount3 struct {
	SchemeName              string `json:"SchemeName"`
	Identification          string `json:"Identification"`
	Name                    string `json:"Name,omitempty"`
	SecondaryIdentification string `json:"SecondaryIdentification,omitempty"`
}

// OBBranchAndFinancialInstitutionIdentification6 represents a financial institution.
type OBBranchAndFinancialInstitutionIdentification6 struct {
	// SchemeName: UK.OBIE.BICFI
	SchemeName     string            `json:"SchemeName,omitempty"`
	Identification string            `json:"Identification,omitempty"`
	Name           string            `json:"Name,omitempty"`
	PostalAddress  *OBPostalAddress6 `json:"PostalAddress,omitempty"`
}

// OBRisk1 carries risk information for payment consent requests.
type OBRisk1 struct {
	// PaymentContextCode classifies the payment context.
	PaymentContextCode              OBExternalPaymentContext1Code `json:"PaymentContextCode,omitempty"`
	MerchantCategoryCode            string                        `json:"MerchantCategoryCode,omitempty"`
	MerchantCustomerIdentification  string                        `json:"MerchantCustomerIdentification,omitempty"`
	// DeliveryAddress is used for card-present / ecommerce flows.
	DeliveryAddress                 *OBPostalAddress6             `json:"DeliveryAddress,omitempty"`
}

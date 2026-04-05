// Package validation provides deep validation of OBIE request structs before
// they are sent to an ASPSP. Catching malformed requests client-side avoids
// unnecessary round-trips and gives callers richer error messages than a raw
// HTTP 400 would provide.
package validation

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/iamkanishka/obie-client-go/models"
)

// ────────────────────────────────────────────────────────────────────────────
// ValidationError
// ────────────────────────────────────────────────────────────────────────────

// FieldError describes a single validation failure.
type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a slice of FieldErrors that implements the error interface.
type ValidationErrors []FieldError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = e.Error()
	}
	return "validation: " + strings.Join(msgs, "; ")
}

// HasErrors returns true if the slice is non-empty.
func (ve ValidationErrors) HasErrors() bool { return len(ve) > 0 }

// ────────────────────────────────────────────────────────────────────────────
// Validator
// ────────────────────────────────────────────────────────────────────────────

// Validator collects validation errors while walking a request struct.
type Validator struct {
	errs ValidationErrors
}

// New creates a fresh Validator.
func New() *Validator { return &Validator{} }

// add records a field error.
func (v *Validator) add(field, message string) {
	v.errs = append(v.errs, FieldError{Field: field, Message: message})
}

// Errors returns all accumulated errors, or nil if none exist.
func (v *Validator) Errors() error {
	if len(v.errs) == 0 {
		return nil
	}
	return v.errs
}

// ────────────────────────────────────────────────────────────────────────────
// Shared field validators
// ────────────────────────────────────────────────────────────────────────────

// ISO 4217 currency regex (3 uppercase letters).
var reISO4217 = regexp.MustCompile(`^[A-Z]{3}$`)

// Amount regex: up to 15 digits, optional decimal with up to 5 decimal places.
var reAmount = regexp.MustCompile(`^\d{1,15}(\.\d{1,5})?$`)

// Sort-code account-number (6+8 = 14 digits).
var reSortCodeAccNum = regexp.MustCompile(`^\d{14}$`)

// IBAN: 2 letters + 2 digits + up to 30 alphanumeric.
var reIBAN = regexp.MustCompile(`^[A-Z]{2}\d{2}[A-Z0-9]{1,30}$`)

// BIC/SWIFT: 8 or 11 characters.
//var reBIC = regexp.MustCompile(`^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$`)

// validateAmount checks that s is a well-formed monetary amount string.
func (v *Validator) validateAmount(field, s string) {
	if s == "" {
		v.add(field, "amount must not be empty")
		return
	}
	if !reAmount.MatchString(s) {
		v.add(field, fmt.Sprintf("amount %q does not match expected format (e.g. \"10.50\")", s))
		return
	}
	// Must be positive.
	n, _, err := new(big.Float).Parse(s, 10)
	if err != nil || n.Sign() <= 0 {
		v.add(field, "amount must be a positive number")
	}
}

// validateCurrency checks that s is a valid ISO 4217 currency code.
func (v *Validator) validateCurrency(field, s string) {
	if !reISO4217.MatchString(s) {
		v.add(field, fmt.Sprintf("currency %q is not a valid ISO 4217 code", s))
	}
}

// validateStringLen checks that s is within [min, max] runes.
func (v *Validator) validateStringLen(field, s string, min, max int) {
	l := utf8.RuneCountInString(s)
	if l < min {
		v.add(field, fmt.Sprintf("must be at least %d characters (got %d)", min, l))
	}
	if max > 0 && l > max {
		v.add(field, fmt.Sprintf("must be at most %d characters (got %d)", max, l))
	}
}

// validateRequired checks that s is non-empty.
func (v *Validator) validateRequired(field, s string) {
	if strings.TrimSpace(s) == "" {
		v.add(field, "is required")
	}
}

// validateAccountIdentification validates a sort-code+account-number or IBAN.
func (v *Validator) validateAccountIdentification(field, scheme, identification string) {
	switch scheme {
	case "UK.OBIE.SortCodeAccountNumber":
		// Strip spaces/dashes for flexible input.
		stripped := strings.Map(func(r rune) rune {
			if r == '-' || r == ' ' {
				return -1
			}
			return r
		}, identification)
		if !reSortCodeAccNum.MatchString(stripped) {
			v.add(field, fmt.Sprintf("SortCodeAccountNumber %q must be 14 digits (6 sort code + 8 account)", identification))
		}
	case "UK.OBIE.IBAN":
		upper := strings.ToUpper(strings.ReplaceAll(identification, " ", ""))
		if !reIBAN.MatchString(upper) {
			v.add(field, fmt.Sprintf("IBAN %q is malformed", identification))
		}
	case "UK.OBIE.BBAN":
		if len(identification) == 0 {
			v.add(field, "BBAN identification must not be empty")
		}
	default:
		// Unknown scheme — just require non-empty.
		if identification == "" {
			v.add(field, "identification must not be empty when scheme is set")
		}
	}
}

// ────────────────────────────────────────────────────────────────────────────
// OBCashAccount3
// ────────────────────────────────────────────────────────────────────────────

// ValidateCashAccount validates an OBCashAccount3 at the given field path.
func (v *Validator) ValidateCashAccount(field string, acc models.OBCashAccount3) {
	v.validateRequired(field+".SchemeName", acc.SchemeName)
	v.validateAccountIdentification(field+".Identification", acc.SchemeName, acc.Identification)
	if acc.Name != "" {
		v.validateStringLen(field+".Name", acc.Name, 1, 70)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// OBActiveOrHistoricCurrencyAndAmount
// ────────────────────────────────────────────────────────────────────────────

// ValidateAmount validates a currency+amount pair.
func (v *Validator) ValidateAmount(field string, a models.OBActiveOrHistoricCurrencyAndAmount) {
	v.validateAmount(field+".Amount", a.Amount)
	v.validateCurrency(field+".Currency", a.Currency)
}

// ────────────────────────────────────────────────────────────────────────────
// Domestic payment consent
// ────────────────────────────────────────────────────────────────────────────

// ValidateDomesticConsent validates an OBWriteDomesticConsent5 and returns
// any validation errors, or nil.
func ValidateDomesticConsent(req *models.OBWriteDomesticConsent5) error {
	v := New()
	if req == nil {
		v.add("request", "must not be nil")
		return v.Errors()
	}

	init := req.Data.Initiation
	v.validateRequired("Data.Initiation.InstructionIdentification", init.InstructionIdentification)
	v.validateStringLen("Data.Initiation.InstructionIdentification", init.InstructionIdentification, 1, 35)
	v.validateRequired("Data.Initiation.EndToEndIdentification", init.EndToEndIdentification)
	v.validateStringLen("Data.Initiation.EndToEndIdentification", init.EndToEndIdentification, 1, 35)

	v.ValidateAmount("Data.Initiation.InstructedAmount", init.InstructedAmount)
	v.ValidateCashAccount("Data.Initiation.CreditorAccount", init.CreditorAccount)

	if init.DebtorAccount != nil {
		v.ValidateCashAccount("Data.Initiation.DebtorAccount", *init.DebtorAccount)
	}
	if ri := init.RemittanceInformation; ri != nil {
		if ri.Reference != "" {
			v.validateStringLen("Data.Initiation.RemittanceInformation.Reference", ri.Reference, 1, 35)
		}
		if ri.Unstructured != "" {
			v.validateStringLen("Data.Initiation.RemittanceInformation.Unstructured", ri.Unstructured, 1, 140)
		}
	}

	return v.Errors()
}

// ────────────────────────────────────────────────────────────────────────────
// Domestic payment submission
// ────────────────────────────────────────────────────────────────────────────

// ValidateDomesticPayment validates an OBWriteDomestic2 submission request.
func ValidateDomesticPayment(req *models.OBWriteDomestic2) error {
	v := New()
	if req == nil {
		v.add("request", "must not be nil")
		return v.Errors()
	}
	v.validateRequired("Data.ConsentId", req.Data.ConsentId)
	v.validateStringLen("Data.ConsentId", req.Data.ConsentId, 1, 128)

	init := req.Data.Initiation
	v.validateRequired("Data.Initiation.InstructionIdentification", init.InstructionIdentification)
	v.validateRequired("Data.Initiation.EndToEndIdentification", init.EndToEndIdentification)
	v.ValidateAmount("Data.Initiation.InstructedAmount", init.InstructedAmount)
	v.ValidateCashAccount("Data.Initiation.CreditorAccount", init.CreditorAccount)
	return v.Errors()
}

// ────────────────────────────────────────────────────────────────────────────
// International payment consent
// ────────────────────────────────────────────────────────────────────────────

// ValidateInternationalConsent validates an OBWriteInternationalConsent5.
func ValidateInternationalConsent(req *models.OBWriteInternationalConsent5) error {
	v := New()
	if req == nil {
		v.add("request", "must not be nil")
		return v.Errors()
	}
	init := req.Data.Initiation
	v.validateRequired("Data.Initiation.InstructionIdentification", init.InstructionIdentification)
	v.validateRequired("Data.Initiation.EndToEndIdentification", init.EndToEndIdentification)
	v.validateCurrency("Data.Initiation.CurrencyOfTransfer", init.CurrencyOfTransfer)
	v.ValidateAmount("Data.Initiation.InstructedAmount", init.InstructedAmount)
	v.ValidateCashAccount("Data.Initiation.CreditorAccount", init.CreditorAccount)

	if init.DestinationCountryCode != "" {
		if matched, _ := regexp.MatchString(`^[A-Z]{2}$`, init.DestinationCountryCode); !matched {
			v.add("Data.Initiation.DestinationCountryCode",
				fmt.Sprintf("%q is not a valid ISO 3166-1 alpha-2 country code", init.DestinationCountryCode))
		}
	}
	if init.ChargeBearer != "" {
		validBearers := map[models.OBChargeBearerType1Code]bool{
			models.ChargeBearerBorneByCreditor:       true,
			models.ChargeBearerBorneByDebtor:         true,
			models.ChargeBearerFollowingServiceLevel: true,
			models.ChargeBearerShared:                true,
			models.ChargeBearerSLEV:                  true,
		}
		if !validBearers[init.ChargeBearer] {
			v.add("Data.Initiation.ChargeBearer", fmt.Sprintf("%q is not a valid charge bearer value", init.ChargeBearer))
		}
	}
	return v.Errors()
}

// ────────────────────────────────────────────────────────────────────────────
// Funds confirmation
// ────────────────────────────────────────────────────────────────────────────

// ValidateFundsConfirmation validates an OBFundsConfirmation1.
func ValidateFundsConfirmation(req *models.OBFundsConfirmation1) error {
	v := New()
	if req == nil {
		v.add("request", "must not be nil")
		return v.Errors()
	}
	v.validateRequired("Data.ConsentId", req.Data.ConsentId)
	v.validateRequired("Data.Reference", req.Data.Reference)
	v.validateStringLen("Data.Reference", req.Data.Reference, 1, 35)
	v.ValidateAmount("Data.InstructedAmount", req.Data.InstructedAmount)
	return v.Errors()
}

// ────────────────────────────────────────────────────────────────────────────
// VRP consent
// ────────────────────────────────────────────────────────────────────────────

// ValidateVRPConsent validates an OBDomesticVRPConsentRequest.
func ValidateVRPConsent(req *models.OBDomesticVRPConsentRequest) error {
	v := New()
	if req == nil {
		v.add("request", "must not be nil")
		return v.Errors()
	}
	cp := req.Data.ControlParameters
	v.ValidateAmount("Data.ControlParameters.MaximumIndividualAmount", cp.MaximumIndividualAmount)

	if len(cp.PeriodicLimits) == 0 {
		v.add("Data.ControlParameters.PeriodicLimits", "at least one periodic limit is required")
	}
	for i, pl := range cp.PeriodicLimits {
		field := fmt.Sprintf("Data.ControlParameters.PeriodicLimits[%d]", i)
		validPeriods := map[string]bool{"Day": true, "Week": true, "Fortnight": true, "Month": true, "Half-year": true, "Year": true}
		if !validPeriods[pl.PeriodType] {
			v.add(field+".PeriodType", fmt.Sprintf("%q is not a valid period type", pl.PeriodType))
		}
		v.ValidateAmount(field+".Amount", pl.Amount)
	}
	if len(cp.VRPType) == 0 {
		v.add("Data.ControlParameters.VRPType", "at least one VRP type is required")
	}

	v.ValidateCashAccount("Data.Initiation.CreditorAccount", req.Data.Initiation.CreditorAccount)
	return v.Errors()
}

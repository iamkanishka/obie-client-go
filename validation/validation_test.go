package validation_test

import (
	"errors"
	"testing"

	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/validation"
)

// helpers
func gbpAmount(a string) models.OBActiveOrHistoricCurrencyAndAmount {
	return models.OBActiveOrHistoricCurrencyAndAmount{Amount: a, Currency: "GBP"}
}

func ukAccount(id string) models.OBCashAccount3 {
	return models.OBCashAccount3{
		SchemeName:     "UK.OBIE.SortCodeAccountNumber",
		Identification: id,
		Name:           "Test Account",
	}
}

func validDomesticConsent() *models.OBWriteDomesticConsent5 {
	return &models.OBWriteDomesticConsent5{
		Data: models.OBWriteDomesticConsentData5{
			Initiation: models.OBDomesticInitiation{
				InstructionIdentification: "INSTR-001",
				EndToEndIdentification:    "E2E-001",
				InstructedAmount:          gbpAmount("100.00"),
				CreditorAccount:           ukAccount("20000319825731"),
			},
		},
	}
}

// ── DomesticConsent ───────────────────────────────────────────────────────

func TestValidateDomesticConsent_Valid(t *testing.T) {
	if err := validation.ValidateDomesticConsent(validDomesticConsent()); err != nil {
		t.Errorf("expected no error for valid consent, got: %v", err)
	}
}

func TestValidateDomesticConsent_NilRequest(t *testing.T) {
	if err := validation.ValidateDomesticConsent(nil); err == nil {
		t.Error("expected error for nil request")
	}
}

func TestValidateDomesticConsent_MissingInstructionID(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.InstructionIdentification = ""
	assertFieldError(t, validation.ValidateDomesticConsent(req), "InstructionIdentification")
}

func TestValidateDomesticConsent_InstructionIDTooLong(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.InstructionIdentification = string(make([]byte, 36))
	assertFieldError(t, validation.ValidateDomesticConsent(req), "InstructionIdentification")
}

func TestValidateDomesticConsent_InvalidAmount(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.InstructedAmount.Amount = "-5.00"
	assertFieldError(t, validation.ValidateDomesticConsent(req), "Amount")
}

func TestValidateDomesticConsent_ZeroAmount(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.InstructedAmount.Amount = "0.00"
	assertFieldError(t, validation.ValidateDomesticConsent(req), "Amount")
}

func TestValidateDomesticConsent_BadCurrency(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.InstructedAmount.Currency = "gbp" // lowercase
	assertFieldError(t, validation.ValidateDomesticConsent(req), "Currency")
}

func TestValidateDomesticConsent_InvalidSortCodeAccountNumber(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.CreditorAccount.Identification = "12345" // too short
	assertFieldError(t, validation.ValidateDomesticConsent(req), "Identification")
}

func TestValidateDomesticConsent_ValidIBAN(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.CreditorAccount.SchemeName = "UK.OBIE.IBAN"
	req.Data.Initiation.CreditorAccount.Identification = "GB29NWBK60161331926819"
	if err := validation.ValidateDomesticConsent(req); err != nil {
		t.Errorf("expected valid IBAN to pass, got: %v", err)
	}
}

func TestValidateDomesticConsent_InvalidIBAN(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.CreditorAccount.SchemeName = "UK.OBIE.IBAN"
	req.Data.Initiation.CreditorAccount.Identification = "NOTANIBAN"
	assertFieldError(t, validation.ValidateDomesticConsent(req), "IBAN")
}

func TestValidateDomesticConsent_ReferenceTooLong(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.RemittanceInformation = &models.OBRemittanceInformation1{
		Reference: string(make([]byte, 36)), // > 35
	}
	assertFieldError(t, validation.ValidateDomesticConsent(req), "Reference")
}

func TestValidateDomesticConsent_MultipleErrors(t *testing.T) {
	req := validDomesticConsent()
	req.Data.Initiation.InstructionIdentification = ""
	req.Data.Initiation.EndToEndIdentification = ""
	req.Data.Initiation.InstructedAmount.Amount = ""
	err := validation.ValidateDomesticConsent(req)
	if err == nil {
		t.Fatal("expected multiple errors, got nil")
	}
	var ve validation.ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
	if len(ve) < 3 {
		t.Errorf("expected >= 3 errors, got %d: %v", len(ve), ve)
	}
}

// ── InternationalConsent ──────────────────────────────────────────────────

func TestValidateInternationalConsent_Valid(t *testing.T) {
	req := &models.OBWriteInternationalConsent5{
		Data: models.OBWriteInternationalConsentData5{
			Initiation: models.OBInternationalInitiation{
				InstructionIdentification: "INSTR-001",
				EndToEndIdentification:    "E2E-001",
				CurrencyOfTransfer:        "USD",
				InstructedAmount:          gbpAmount("500.00"),
				CreditorAccount:           ukAccount("20000319825731"),
				DestinationCountryCode:    "US",
				ChargeBearer:              "Shared",
			},
		},
	}
	if err := validation.ValidateInternationalConsent(req); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateInternationalConsent_BadCountryCode(t *testing.T) {
	req := &models.OBWriteInternationalConsent5{
		Data: models.OBWriteInternationalConsentData5{
			Initiation: models.OBInternationalInitiation{
				InstructionIdentification: "INSTR-001",
				EndToEndIdentification:    "E2E-001",
				CurrencyOfTransfer:        "USD",
				InstructedAmount:          gbpAmount("100.00"),
				CreditorAccount:           ukAccount("20000319825731"),
				DestinationCountryCode:    "USA", // 3 letters — invalid
			},
		},
	}
	assertFieldError(t, validation.ValidateInternationalConsent(req), "DestinationCountryCode")
}

func TestValidateInternationalConsent_BadChargeBearer(t *testing.T) {
	req := &models.OBWriteInternationalConsent5{
		Data: models.OBWriteInternationalConsentData5{
			Initiation: models.OBInternationalInitiation{
				InstructionIdentification: "INSTR-001",
				EndToEndIdentification:    "E2E-001",
				CurrencyOfTransfer:        "EUR",
				InstructedAmount:          gbpAmount("100.00"),
				CreditorAccount:           ukAccount("20000319825731"),
				ChargeBearer:              "InvalidBearer",
			},
		},
	}
	assertFieldError(t, validation.ValidateInternationalConsent(req), "ChargeBearer")
}

// ── FundsConfirmation ─────────────────────────────────────────────────────

func TestValidateFundsConfirmation_Valid(t *testing.T) {
	req := &models.OBFundsConfirmation1{
		Data: models.OBFundsConfirmationData1{
			ConsentId:        "consent-123",
			Reference:        "purchase-ref",
			InstructedAmount: gbpAmount("25.00"),
		},
	}
	if err := validation.ValidateFundsConfirmation(req); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateFundsConfirmation_MissingFields(t *testing.T) {
	err := validation.ValidateFundsConfirmation(&models.OBFundsConfirmation1{})
	if err == nil {
		t.Fatal("expected errors for empty request")
	}
}

// ── VRPConsent ────────────────────────────────────────────────────────────

func TestValidateVRPConsent_Valid(t *testing.T) {
	req := &models.OBDomesticVRPConsentRequest{
		Data: models.OBDomesticVRPConsentRequestData{
			ControlParameters: models.OBDomesticVRPControlParameters{
				MaximumIndividualAmount: gbpAmount("100.00"),
				PeriodicLimits: []models.OBDomesticVRPControlParametersPeriodic{
					{PeriodType: "Month", PeriodAlignment: "Calendar", Amount: gbpAmount("500.00")},
				},
				VRPType:                  []string{"UK.OBIE.VRPType.Sweeping"},
				PSUAuthenticationMethods: []string{"UK.OBIE.SCA"},
			},
			Initiation: models.OBDomesticVRPInitiation{
				CreditorAccount: ukAccount("20000319825731"),
			},
		},
	}
	if err := validation.ValidateVRPConsent(req); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateVRPConsent_InvalidPeriodType(t *testing.T) {
	req := &models.OBDomesticVRPConsentRequest{
		Data: models.OBDomesticVRPConsentRequestData{
			ControlParameters: models.OBDomesticVRPControlParameters{
				MaximumIndividualAmount: gbpAmount("100.00"),
				PeriodicLimits: []models.OBDomesticVRPControlParametersPeriodic{
					{PeriodType: "Quarter", Amount: gbpAmount("300.00")}, // invalid
				},
				VRPType: []string{"UK.OBIE.VRPType.Sweeping"},
			},
			Initiation: models.OBDomesticVRPInitiation{
				CreditorAccount: ukAccount("20000319825731"),
			},
		},
	}
	assertFieldError(t, validation.ValidateVRPConsent(req), "PeriodType")
}

// ── helpers ───────────────────────────────────────────────────────────────

func assertFieldError(t *testing.T, err error, fieldSubstring string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected validation error containing %q, got nil", fieldSubstring)
		return
	}
	if !containsField(err, fieldSubstring) {
		t.Errorf("expected error mentioning %q, got: %v", fieldSubstring, err)
	}
}

func containsField(err error, sub string) bool {
	if err == nil {
		return false
	}
	return len(sub) == 0 || stringContains(err.Error(), sub)
}

func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

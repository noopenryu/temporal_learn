package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{7,14}$`)
)

// ValidatePhoneNumber checks international phone format
func ValidatePhoneNumber(phone string) error {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return fmt.Errorf("phone number is required")
	}
	// Strip spaces and dashes
	cleaned := strings.ReplaceAll(strings.ReplaceAll(trimmed, " ", ""), "-", "")
	if !phoneRegex.MatchString(cleaned) {
		return fmt.Errorf("invalid phone number format; expected E.164 or standard international format with 8-15 digits")
	}
	return nil
}

// CleanPhoneNumber normalizes phone number
func CleanPhoneNumber(phone string) string {
	trimmed := strings.TrimSpace(phone)
	cleaned := strings.ReplaceAll(strings.ReplaceAll(trimmed, " ", ""), "-", "")
	if !strings.HasPrefix(cleaned, "+") {
		// Default to + if not present
		return "+" + cleaned
	}
	return cleaned
}

// ValidateEmail checks email address format
func ValidateEmail(email string) error {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return fmt.Errorf("email address is required")
	}
	if len(trimmed) > 254 || !emailRegex.MatchString(trimmed) {
		return fmt.Errorf("invalid email address format")
	}
	return nil
}

// ValidatePassword validates password complexity and length
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	if len(password) > 128 {
		return fmt.Errorf("password cannot exceed 128 characters")
	}
	return nil
}

// ValidateDateOfBirth validates that DOB is valid, not year 0001, not in the future, and applicant is 18+
func ValidateDateOfBirth(rawDob string) (time.Time, error) {
	trimmed := strings.TrimSpace(rawDob)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("date of birth is required")
	}

	// Reject Go zero-time and year 0001
	if strings.HasPrefix(trimmed, "0001") || strings.HasPrefix(trimmed, "0000") {
		return time.Time{}, fmt.Errorf("date of birth is invalid; cannot be year 0001")
	}

	// Support YYYY-MM-DD or RFC3339 formats
	var parsed time.Time
	var err error

	if strings.Contains(trimmed, "T") {
		parsed, err = time.Parse(time.RFC3339, trimmed)
		if err != nil {
			parsed, err = time.Parse("2006-01-02T15:04:05", trimmed)
		}
	} else {
		parsed, err = time.Parse("2006-01-02", trimmed)
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date of birth format, expected YYYY-MM-DD: %w", err)
	}

	now := time.Now().UTC()
	if parsed.After(now) {
		return time.Time{}, fmt.Errorf("date of birth cannot be in the future")
	}

	if parsed.Year() < 1900 {
		return time.Time{}, fmt.Errorf("date of birth year must be 1900 or later")
	}

	// Calculate age accurately
	age := now.Year() - parsed.Year()
	if now.YearDay() < parsed.YearDay() {
		age--
	}

	if age < 18 {
		return time.Time{}, fmt.Errorf("applicant must be at least 18 years old to open a wallet account (current age: %d)", age)
	}

	if age > 120 {
		return time.Time{}, fmt.Errorf("applicant age exceeds maximum limit of 120 years (current age: %d)", age)
	}

	return parsed, nil
}

// ValidateNames checks first and last name
func ValidateNames(firstName, lastName string) error {
	fn := strings.TrimSpace(firstName)
	ln := strings.TrimSpace(lastName)

	if len(fn) < 2 {
		return fmt.Errorf("first name must be at least 2 characters")
	}
	if len(ln) < 2 {
		return fmt.Errorf("last name must be at least 2 characters")
	}
	if len(fn) > 50 || len(ln) > 50 {
		return fmt.Errorf("names cannot exceed 50 characters")
	}
	return nil
}

// ValidateAddress checks address fields
func ValidateAddress(country, city, street, postalCode string) error {
	if strings.TrimSpace(country) == "" {
		return fmt.Errorf("country is required")
	}
	if len(strings.TrimSpace(city)) < 2 {
		return fmt.Errorf("city must be at least 2 characters")
	}
	if len(strings.TrimSpace(street)) < 5 {
		return fmt.Errorf("street address must be at least 5 characters")
	}
	if len(strings.TrimSpace(postalCode)) < 3 {
		return fmt.Errorf("postal code must be at least 3 characters")
	}
	return nil
}

// ValidateIdentityDocs checks identity document attributes
func ValidateIdentityDocs(docType, docNumber, issuingCountry, frontUrl, selfieUrl string) error {
	dt := strings.ToUpper(strings.TrimSpace(docType))
	if dt != "NATIONAL_ID" && dt != "PASSPORT" && dt != "DRIVING_LICENSE" {
		return fmt.Errorf("invalid document type; must be NATIONAL_ID, PASSPORT, or DRIVING_LICENSE")
	}
	if len(strings.TrimSpace(docNumber)) < 5 {
		return fmt.Errorf("document number must be at least 5 characters")
	}
	if strings.TrimSpace(issuingCountry) == "" {
		return fmt.Errorf("issuing country is required")
	}
	if strings.TrimSpace(frontUrl) == "" {
		return fmt.Errorf("front document photo or file URL is required")
	}
	if strings.TrimSpace(selfieUrl) == "" {
		return fmt.Errorf("selfie verification photo URL is required")
	}
	return nil
}

// ValidateFinancial checks financial profile
func ValidateFinancial(occupation string, annualIncome float64, sourceOfFunds string) error {
	if strings.TrimSpace(occupation) == "" {
		return fmt.Errorf("occupation is required")
	}
	if annualIncome < 0 {
		return fmt.Errorf("annual income cannot be negative")
	}
	if strings.TrimSpace(sourceOfFunds) == "" {
		return fmt.Errorf("source of funds is required")
	}
	return nil
}

// ValidateWalletTier checks tier selection
func ValidateWalletTier(tier string) error {
	t := strings.ToUpper(strings.TrimSpace(tier))
	if t != "TIER_1_BASIC" && t != "TIER_2_PREMIUM" {
		return fmt.Errorf("invalid wallet tier; must be TIER_1_BASIC or TIER_2_PREMIUM")
	}
	return nil
}

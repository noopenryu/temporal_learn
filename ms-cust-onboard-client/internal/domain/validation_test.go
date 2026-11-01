package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ms-cust-onboard-client/internal/domain"
)

func TestValidatePhoneNumber(t *testing.T) {
	assert.NoError(t, domain.ValidatePhoneNumber("+15551234567"))
	assert.NoError(t, domain.ValidatePhoneNumber("+84901234567"))
	assert.NoError(t, domain.ValidatePhoneNumber("84901234567"))
	assert.NoError(t, domain.ValidatePhoneNumber("+1 555-123-4567"))

	assert.Error(t, domain.ValidatePhoneNumber(""))
	assert.Error(t, domain.ValidatePhoneNumber("123"))
	assert.Error(t, domain.ValidatePhoneNumber("abcde"))
}

func TestValidateEmail(t *testing.T) {
	assert.NoError(t, domain.ValidateEmail("test@example.com"))
	assert.NoError(t, domain.ValidateEmail("alice.smith+wallet@domain.co.uk"))

	assert.Error(t, domain.ValidateEmail(""))
	assert.Error(t, domain.ValidateEmail("invalid-email"))
	assert.Error(t, domain.ValidateEmail("no@domain"))
}

func TestValidatePassword(t *testing.T) {
	assert.NoError(t, domain.ValidatePassword("secret123"))
	assert.NoError(t, domain.ValidatePassword("123456"))

	assert.Error(t, domain.ValidatePassword(""))
	assert.Error(t, domain.ValidatePassword("12345")) // Too short
}

func TestValidateDateOfBirth(t *testing.T) {
	// Valid adult DOB
	validAdult := time.Now().AddDate(-25, 0, 0).Format("2006-01-02")
	parsed, err := domain.ValidateDateOfBirth(validAdult)
	assert.NoError(t, err)
	assert.Equal(t, time.Now().AddDate(-25, 0, 0).Year(), parsed.Year())

	// Year 0001 (Go zero time) must be rejected
	_, err = domain.ValidateDateOfBirth("0001-01-01")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be year 0001")

	_, err = domain.ValidateDateOfBirth("0001-01-01T00:00:00Z")
	assert.Error(t, err)

	// Underage (< 18) must be rejected
	underage := time.Now().AddDate(-16, 0, 0).Format("2006-01-02")
	_, err = domain.ValidateDateOfBirth(underage)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 18 years old")

	// Future date must be rejected
	future := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	_, err = domain.ValidateDateOfBirth(future)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "future")

	// Before 1900 must be rejected
	_, err = domain.ValidateDateOfBirth("1899-12-31")
	assert.Error(t, err)
}

func TestValidateNames(t *testing.T) {
	assert.NoError(t, domain.ValidateNames("John", "Doe"))
	assert.Error(t, domain.ValidateNames("J", "Doe"))
	assert.Error(t, domain.ValidateNames("John", ""))
}

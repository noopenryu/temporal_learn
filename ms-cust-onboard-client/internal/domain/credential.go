package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials     = errors.New("invalid phone number or password")
	ErrCredentialNotFound     = errors.New("no active application found for this phone number, or invalid credentials")
	ErrCredentialAlreadyExists = errors.New("an application credential already exists for this phone number")
)

// Credential represents stored customer authentication for resuming onboarding
type Credential struct {
	PhoneNumber  string    `json:"phone_number"`
	ApplicationID uuid.UUID `json:"application_id"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

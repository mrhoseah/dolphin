package auth

import (
	"github.com/mrhoseah/dolphin/internal/auth"
	"gorm.io/gorm"
)

// MFAManager represents the MFA manager
type MFAManager = auth.MFAManager

// NewMFAManager creates a new MFA manager
func NewMFAManager(db *gorm.DB, issuer string) *MFAManager {
	return auth.NewMFAManager(db, issuer)
}

// TOTPProvider represents a TOTP provider
type TOTPProvider = auth.TOTPProvider

// NewTOTPProvider creates a new TOTP provider
func NewTOTPProvider(issuer string) *TOTPProvider {
	return auth.NewTOTPProvider(issuer)
}

// SMSProvider represents an SMS provider
type SMSProvider = auth.SMSProvider

// NewSMSProvider creates a new SMS provider
func NewSMSProvider(sender func(phone, code string) error) *SMSProvider {
	return auth.NewSMSProvider(sender)
}

// EmailProvider represents an email provider
type EmailProvider = auth.EmailProvider

// NewEmailProvider creates a new email provider
func NewEmailProvider(sender func(email, code string) error) *EmailProvider {
	return auth.NewEmailProvider(sender)
}

// MFAMethod represents an MFA method
type MFAMethod = auth.MFAMethod

// MFAProvider represents an MFA provider interface
type MFAProvider = auth.MFAProvider


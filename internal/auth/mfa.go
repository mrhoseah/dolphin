package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"time"

	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

// MFAProvider represents a multi-factor authentication provider
type MFAProvider interface {
	GenerateSecret(userID uint) (string, error)
	ValidateCode(secret, code string) (bool, error)
	GenerateQRCode(secret, accountName, issuer string) (string, error)
	GenerateRecoveryCodes(count int) ([]string, error)
}

// TOTPProvider implements MFA using TOTP (Time-based One-Time Password)
type TOTPProvider struct {
	issuer string
}

// NewTOTPProvider creates a new TOTP provider
func NewTOTPProvider(issuer string) *TOTPProvider {
	return &TOTPProvider{
		issuer: issuer,
	}
}

// GenerateSecret generates a TOTP secret for a user
func (t *TOTPProvider) GenerateSecret(userID uint) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.issuer,
		AccountName: fmt.Sprintf("%d", userID),
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

// ValidateCode validates a TOTP code
func (t *TOTPProvider) ValidateCode(secret, code string) (bool, error) {
	return totp.Validate(code, secret), nil
}

// GenerateQRCode generates a QR code data URL for TOTP setup
func (t *TOTPProvider) GenerateQRCode(secret, accountName, issuer string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
		Secret:      []byte(secret),
	})
	if err != nil {
		return "", err
	}

	// Generate QR code image
	img, err := key.Image(200, 200)
	if err != nil {
		return "", err
	}

	// Convert image to PNG bytes
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	// Convert to base64 data URL
	return fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes())), nil
}

// GenerateRecoveryCodes generates recovery codes for MFA
func (t *TOTPProvider) GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := generateSecureCode(8)
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	return codes, nil
}

// SMSProvider implements MFA using SMS codes
type SMSProvider struct {
	sender func(phone, code string) error
}

// NewSMSProvider creates a new SMS provider
func NewSMSProvider(sender func(phone, code string) error) *SMSProvider {
	return &SMSProvider{
		sender: sender,
	}
}

// GenerateSecret generates a random code and sends it via SMS
func (s *SMSProvider) GenerateSecret(userID uint) (string, error) {
	// This would typically look up the user's phone number
	// For now, return a placeholder
	return "", errors.New("SMS provider requires phone number")
}

// SendSMSCode sends a code via SMS
func (s *SMSProvider) SendSMSCode(phone string) (string, error) {
	code, err := generateSecureCode(6)
	if err != nil {
		return "", err
	}

	if s.sender != nil {
		if err := s.sender(phone, code); err != nil {
			return "", err
		}
	}

	return code, nil
}

// ValidateCode validates an SMS code (this would typically check against a stored code)
func (s *SMSProvider) ValidateCode(secret, code string) (bool, error) {
	return secret == code, nil
}

// GenerateQRCode is not applicable for SMS
func (s *SMSProvider) GenerateQRCode(secret, accountName, issuer string) (string, error) {
	return "", errors.New("QR code not applicable for SMS MFA")
}

// GenerateRecoveryCodes generates recovery codes for MFA
func (s *SMSProvider) GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := generateSecureCode(8)
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	return codes, nil
}

// EmailProvider implements MFA using email codes
type EmailProvider struct {
	sender func(email, code string) error
}

// NewEmailProvider creates a new email provider
func NewEmailProvider(sender func(email, code string) error) *EmailProvider {
	return &EmailProvider{
		sender: sender,
	}
}

// GenerateSecret generates a random code and sends it via email
func (e *EmailProvider) GenerateSecret(userID uint) (string, error) {
	return "", errors.New("Email provider requires email address")
}

// SendEmailCode sends a code via email
func (e *EmailProvider) SendEmailCode(email string) (string, error) {
	code, err := generateSecureCode(6)
	if err != nil {
		return "", err
	}

	if e.sender != nil {
		if err := e.sender(email, code); err != nil {
			return "", err
		}
	}

	return code, nil
}

// ValidateCode validates an email code
func (e *EmailProvider) ValidateCode(secret, code string) (bool, error) {
	return secret == code, nil
}

// GenerateQRCode is not applicable for email
func (e *EmailProvider) GenerateQRCode(secret, accountName, issuer string) (string, error) {
	return "", errors.New("QR code not applicable for email MFA")
}

// GenerateRecoveryCodes generates recovery codes for MFA
func (e *EmailProvider) GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := generateSecureCode(8)
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	return codes, nil
}

// MFAMethod represents an MFA method for a user
type MFAMethod struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Type          string    `gorm:"not null" json:"type"` // "totp", "sms", "email"
	Secret        string    `gorm:"not null" json:"-"`
	Enabled       bool      `gorm:"default:false" json:"enabled"`
	RecoveryCodes []string  `gorm:"type:json" json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// MFAManager manages MFA for users
type MFAManager struct {
	db       *gorm.DB
	providers map[string]MFAProvider
	totp     *TOTPProvider
}

// NewMFAManager creates a new MFA manager
func NewMFAManager(db *gorm.DB, issuer string) *MFAManager {
	totpProvider := NewTOTPProvider(issuer)
	return &MFAManager{
		db:       db,
		providers: make(map[string]MFAProvider),
		totp:     totpProvider,
	}
}

// RegisterProvider registers an MFA provider
func (m *MFAManager) RegisterProvider(name string, provider MFAProvider) {
	m.providers[name] = provider
}

// EnableTOTP enables TOTP for a user
func (m *MFAManager) EnableTOTP(userID uint, accountName string) (string, string, error) {
	secret, err := m.totp.GenerateSecret(userID)
	if err != nil {
		return "", "", err
	}

	qrCode, err := m.totp.GenerateQRCode(secret, accountName, m.totp.issuer)
	if err != nil {
		return "", "", err
	}

	// Generate recovery codes
	recoveryCodes, err := m.totp.GenerateRecoveryCodes(10)
	if err != nil {
		return "", "", err
	}

	// Store MFA method (not enabled yet until verified)
	mfaMethod := &MFAMethod{
		UserID:        userID,
		Type:          "totp",
		Secret:        secret,
		Enabled:       false,
		RecoveryCodes: recoveryCodes,
	}

	if err := m.db.Create(mfaMethod).Error; err != nil {
		return "", "", err
	}

	return secret, qrCode, nil
}

// VerifyAndEnableTOTP verifies a TOTP code and enables MFA
func (m *MFAManager) VerifyAndEnableTOTP(userID uint, code string) error {
	var mfaMethod MFAMethod
	if err := m.db.Where("user_id = ? AND type = ?", userID, "totp").First(&mfaMethod).Error; err != nil {
		return errors.New("TOTP not set up for this user")
	}

	valid, err := m.totp.ValidateCode(mfaMethod.Secret, code)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("invalid TOTP code")
	}

	// Enable MFA
	mfaMethod.Enabled = true
	return m.db.Save(&mfaMethod).Error
}

// VerifyTOTP verifies a TOTP code during login
func (m *MFAManager) VerifyTOTP(userID uint, code string) (bool, error) {
	var mfaMethod MFAMethod
	if err := m.db.Where("user_id = ? AND type = ? AND enabled = ?", userID, "totp", true).First(&mfaMethod).Error; err != nil {
		return false, errors.New("TOTP not enabled for this user")
	}

	// Check recovery codes first
	for i, recoveryCode := range mfaMethod.RecoveryCodes {
		if recoveryCode == code {
			// Remove used recovery code
			mfaMethod.RecoveryCodes = append(mfaMethod.RecoveryCodes[:i], mfaMethod.RecoveryCodes[i+1:]...)
			m.db.Save(&mfaMethod)
			return true, nil
		}
	}

	// Validate TOTP code
	return m.totp.ValidateCode(mfaMethod.Secret, code)
}

// IsMFAEnabled checks if MFA is enabled for a user
func (m *MFAManager) IsMFAEnabled(userID uint) (bool, error) {
	var count int64
	err := m.db.Model(&MFAMethod{}).Where("user_id = ? AND enabled = ?", userID, true).Count(&count).Error
	return count > 0, err
}

// DisableMFA disables MFA for a user
func (m *MFAManager) DisableMFA(userID uint) error {
	return m.db.Where("user_id = ?", userID).Delete(&MFAMethod{}).Error
}

// GetRecoveryCodes returns recovery codes for a user
func (m *MFAManager) GetRecoveryCodes(userID uint) ([]string, error) {
	var mfaMethod MFAMethod
	if err := m.db.Where("user_id = ? AND type = ?", userID, "totp").First(&mfaMethod).Error; err != nil {
		return nil, err
	}
	return mfaMethod.RecoveryCodes, nil
}

// GenerateNewRecoveryCodes generates new recovery codes
func (m *MFAManager) GenerateNewRecoveryCodes(userID uint) ([]string, error) {
	var mfaMethod MFAMethod
	if err := m.db.Where("user_id = ? AND type = ?", userID, "totp").First(&mfaMethod).Error; err != nil {
		return nil, err
	}

	codes, err := m.totp.GenerateRecoveryCodes(10)
	if err != nil {
		return nil, err
	}

	mfaMethod.RecoveryCodes = codes
	if err := m.db.Save(&mfaMethod).Error; err != nil {
		return nil, err
	}

	return codes, nil
}

// generateSecureCode generates a secure random code
func generateSecureCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Use base32 encoding for readability
	code := base32.StdEncoding.EncodeToString(bytes)
	return code[:length], nil
}


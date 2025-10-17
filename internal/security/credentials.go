package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// CredentialManager manages encrypted credentials
type CredentialManager struct {
	masterKey []byte
	keyFile   string
	encrypted map[string]string
}

// CredentialEntry represents an encrypted credential
type CredentialEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Salt  string `json:"salt"`
	IV    string `json:"iv"`
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager(keyFile string) (*CredentialManager, error) {
	cm := &CredentialManager{
		keyFile:   keyFile,
		encrypted: make(map[string]string),
	}

	// Load or generate master key
	if err := cm.loadOrGenerateMasterKey(); err != nil {
		return nil, fmt.Errorf("failed to load master key: %w", err)
	}

	// Load existing encrypted credentials
	if err := cm.loadEncryptedCredentials(); err != nil {
		return nil, fmt.Errorf("failed to load encrypted credentials: %w", err)
	}

	return cm, nil
}

// SetCredential encrypts and stores a credential
func (cm *CredentialManager) SetCredential(key, value string) error {
	encrypted, err := cm.encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt credential: %w", err)
	}

	cm.encrypted[key] = encrypted

	// Save to file
	if err := cm.saveEncryptedCredentials(); err != nil {
		return fmt.Errorf("failed to save encrypted credentials: %w", err)
	}

	return nil
}

// GetCredential decrypts and returns a credential
func (cm *CredentialManager) GetCredential(key string) (string, error) {
	encrypted, exists := cm.encrypted[key]
	if !exists {
		return "", fmt.Errorf("credential not found: %s", key)
	}

	decrypted, err := cm.decrypt(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	return decrypted, nil
}

// DeleteCredential removes a credential
func (cm *CredentialManager) DeleteCredential(key string) error {
	delete(cm.encrypted, key)

	// Save to file
	if err := cm.saveEncryptedCredentials(); err != nil {
		return fmt.Errorf("failed to save encrypted credentials: %w", err)
	}

	return nil
}

// ListCredentials returns all credential keys
func (cm *CredentialManager) ListCredentials() []string {
	var keys []string
	for key := range cm.encrypted {
		keys = append(keys, key)
	}
	return keys
}

// EncryptFile encrypts a file containing credentials
func (cm *CredentialManager) EncryptFile(filePath string) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse as key=value pairs
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		if err := cm.SetCredential(key, value); err != nil {
			return fmt.Errorf("failed to encrypt credential %s: %w", key, err)
		}
	}

	return nil
}

// DecryptToFile decrypts credentials and writes them to a file
func (cm *CredentialManager) DecryptToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for key := range cm.encrypted {
		value, err := cm.GetCredential(key)
		if err != nil {
			return fmt.Errorf("failed to decrypt credential %s: %w", key, err)
		}

		_, err = fmt.Fprintf(file, "%s=%s\n", key, value)
		if err != nil {
			return fmt.Errorf("failed to write credential: %w", err)
		}
	}

	return nil
}

// loadOrGenerateMasterKey loads or generates the master key
func (cm *CredentialManager) loadOrGenerateMasterKey() error {
	// Try to load existing key
	if data, err := os.ReadFile(cm.keyFile); err == nil {
		cm.masterKey = data
		return nil
	}

	// Generate new key
	cm.masterKey = make([]byte, 32)
	if _, err := rand.Read(cm.masterKey); err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	// Save key to file
	if err := os.MkdirAll(filepath.Dir(cm.keyFile), 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	if err := os.WriteFile(cm.keyFile, cm.masterKey, 0600); err != nil {
		return fmt.Errorf("failed to save master key: %w", err)
	}

	return nil
}

// loadEncryptedCredentials loads encrypted credentials from file
func (cm *CredentialManager) loadEncryptedCredentials() error {
	credentialsFile := cm.keyFile + ".credentials"

	if data, err := os.ReadFile(credentialsFile); err == nil {
		var entries []CredentialEntry
		if err := json.Unmarshal(data, &entries); err != nil {
			return fmt.Errorf("failed to unmarshal credentials: %w", err)
		}

		for _, entry := range entries {
			// Reconstruct the encrypted string
			encrypted := fmt.Sprintf("%s:%s:%s", entry.Salt, entry.IV, entry.Value)
			cm.encrypted[entry.Key] = encrypted
		}
	}

	return nil
}

// saveEncryptedCredentials saves encrypted credentials to file
func (cm *CredentialManager) saveEncryptedCredentials() error {
	credentialsFile := cm.keyFile + ".credentials"

	var entries []CredentialEntry
	for key, encrypted := range cm.encrypted {
		parts := strings.Split(encrypted, ":")
		if len(parts) != 3 {
			continue
		}

		entries = append(entries, CredentialEntry{
			Key:   key,
			Value: parts[2],
			Salt:  parts[0],
			IV:    parts[1],
		})
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(credentialsFile, data, 0600); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}

// encrypt encrypts a value using AES-GCM
func (cm *CredentialManager) encrypt(plaintext string) (string, error) {
	// Generate random salt and IV
	salt := make([]byte, 16)
	iv := make([]byte, 12)

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	// Derive key from master key and salt
	key := pbkdf2.Key(cm.masterKey, salt, 10000, 32, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nil, iv, []byte(plaintext), nil)

	// Encode as base64
	encrypted := base64.StdEncoding.EncodeToString(ciphertext)
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	ivB64 := base64.StdEncoding.EncodeToString(iv)

	return fmt.Sprintf("%s:%s:%s", saltB64, ivB64, encrypted), nil
}

// decrypt decrypts a value using AES-GCM
func (cm *CredentialManager) decrypt(encrypted string) (string, error) {
	parts := strings.Split(encrypted, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid encrypted format")
	}

	saltB64, ivB64, ciphertextB64 := parts[0], parts[1], parts[2]

	// Decode from base64
	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return "", err
	}

	iv, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}

	// Derive key from master key and salt
	key := pbkdf2.Key(cm.masterKey, salt, 10000, 32, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Decrypt
	plaintext, err := aesGCM.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// KMSDriver defines the interface for external key management systems
type KMSDriver interface {
	Encrypt(keyID string, plaintext []byte) ([]byte, error)
	Decrypt(keyID string, ciphertext []byte) ([]byte, error)
}

// AWSKMSDriver implements KMS using AWS KMS
type AWSKMSDriver struct {
	region string
	keyID  string
}

// NewAWSKMSDriver creates a new AWS KMS driver
func NewAWSKMSDriver(region, keyID string) *AWSKMSDriver {
	return &AWSKMSDriver{
		region: region,
		keyID:  keyID,
	}
}

// Encrypt encrypts data using AWS KMS
func (kms *AWSKMSDriver) Encrypt(keyID string, plaintext []byte) ([]byte, error) {
	// This would use the AWS SDK for Go
	// For now, return an error indicating it needs implementation
	return nil, fmt.Errorf("AWS KMS driver not implemented - requires AWS SDK")
}

// Decrypt decrypts data using AWS KMS
func (kms *AWSKMSDriver) Decrypt(keyID string, ciphertext []byte) ([]byte, error) {
	// This would use the AWS SDK for Go
	// For now, return an error indicating it needs implementation
	return nil, fmt.Errorf("AWS KMS driver not implemented - requires AWS SDK")
}

// VaultDriver implements KMS using HashiCorp Vault
type VaultDriver struct {
	address string
	token   string
}

// NewVaultDriver creates a new Vault driver
func NewVaultDriver(address, token string) *VaultDriver {
	return &VaultDriver{
		address: address,
		token:   token,
	}
}

// Encrypt encrypts data using Vault
func (vault *VaultDriver) Encrypt(keyID string, plaintext []byte) ([]byte, error) {
	// This would use the Vault API
	// For now, return an error indicating it needs implementation
	return nil, fmt.Errorf("Vault driver not implemented - requires Vault client")
}

// Decrypt decrypts data using Vault
func (vault *VaultDriver) Decrypt(keyID string, ciphertext []byte) ([]byte, error) {
	// This would use the Vault API
	// For now, return an error indicating it needs implementation
	return nil, fmt.Errorf("Vault driver not implemented - requires Vault client")
}

// EnvironmentCredentialManager manages credentials from environment variables
type EnvironmentCredentialManager struct {
	prefix string
}

// NewEnvironmentCredentialManager creates a new environment credential manager
func NewEnvironmentCredentialManager(prefix string) *EnvironmentCredentialManager {
	return &EnvironmentCredentialManager{
		prefix: prefix,
	}
}

// GetCredential gets a credential from environment variables
func (ecm *EnvironmentCredentialManager) GetCredential(key string) (string, error) {
	envKey := ecm.prefix + strings.ToUpper(key)
	value := os.Getenv(envKey)
	if value == "" {
		return "", fmt.Errorf("environment variable not found: %s", envKey)
	}
	return value, nil
}

// SetCredential sets a credential in environment variables (not supported)
func (ecm *EnvironmentCredentialManager) SetCredential(key, value string) error {
	return fmt.Errorf("setting environment variables not supported")
}

// ListCredentials lists available credentials
func (ecm *EnvironmentCredentialManager) ListCredentials() []string {
	var keys []string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, ecm.prefix) {
			key := strings.Split(env, "=")[0]
			key = strings.TrimPrefix(key, ecm.prefix)
			keys = append(keys, strings.ToLower(key))
		}
	}
	return keys
}

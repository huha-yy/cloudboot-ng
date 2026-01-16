package crypto

import (
	"crypto/ecdsa"
	"fmt"
)

// DRMManager handles the complete DRM workflow for CloudBoot Providers
type DRMManager struct {
	masterKey     []byte // Global product master key (32 bytes)
	officialPubKey *ecdsa.PublicKey // Official CloudBoot public key for signature verification
}

// NewDRMManager creates a new DRM manager
func NewDRMManager(masterKey []byte, officialPubKey *ecdsa.PublicKey) (*DRMManager, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes")
	}
	if officialPubKey == nil {
		return nil, fmt.Errorf("official public key is required")
	}

	return &DRMManager{
		masterKey:      masterKey,
		officialPubKey: officialPubKey,
	}, nil
}

// VerifyPackageSignature verifies the signature of a CBP package
func (d *DRMManager) VerifyPackageSignature(packageData []byte, signature string) (bool, error) {
	return VerifySignature(packageData, signature, d.officialPubKey)
}

// DecryptProviderWithMasterKey decrypts a provider binary using the master key
func (d *DRMManager) DecryptProviderWithMasterKey(encryptedBinary []byte) ([]byte, error) {
	return DecryptFile(encryptedBinary, d.masterKey)
}

// EncryptProviderWithMasterKey encrypts a provider binary using the master key
func (d *DRMManager) EncryptProviderWithMasterKey(binary []byte) ([]byte, error) {
	return EncryptFile(binary, d.masterKey)
}

// GenerateSessionKey generates a random session key for re-encryption
func (d *DRMManager) GenerateSessionKey() ([]byte, error) {
	return GenerateAES256Key()
}

// ReEncryptWithSessionKey re-encrypts a provider with a session key
func (d *DRMManager) ReEncryptWithSessionKey(plainProvider []byte, sessionKey []byte) ([]byte, error) {
	if len(sessionKey) != 32 {
		return nil, fmt.Errorf("session key must be 32 bytes")
	}
	return EncryptFile(plainProvider, sessionKey)
}

// DecryptWithSessionKey decrypts a provider using a session key
func (d *DRMManager) DecryptWithSessionKey(encryptedProvider []byte, sessionKey []byte) ([]byte, error) {
	if len(sessionKey) != 32 {
		return nil, fmt.Errorf("session key must be 32 bytes")
	}
	return DecryptFile(encryptedProvider, sessionKey)
}

// CompleteDecryptionFlow performs the complete DRM decryption workflow:
// 1. Decrypt with master key (from .cbp package)
// 2. Generate session key
// 3. Re-encrypt with session key (for transmission to BootOS)
func (d *DRMManager) CompleteDecryptionFlow(encryptedBinary []byte) (plainProvider []byte, sessionKey []byte, reEncrypted []byte, err error) {
	// Step 1: Decrypt with master key
	plainProvider, err = d.DecryptProviderWithMasterKey(encryptedBinary)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("master key decryption failed: %w", err)
	}

	// Step 2: Generate session key
	sessionKey, err = d.GenerateSessionKey()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("session key generation failed: %w", err)
	}

	// Step 3: Re-encrypt with session key
	reEncrypted, err = d.ReEncryptWithSessionKey(plainProvider, sessionKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("session key encryption failed: %w", err)
	}

	return plainProvider, sessionKey, reEncrypted, nil
}

// LicenseInfo represents a customer license
type LicenseInfo struct {
	CustomerID     string
	CustomerName   string
	LicenseKey     []byte // This is the master key encrypted for this customer
	Features       []string
	ExpiresAt      string
}

// DecryptLicenseKey decrypts the license key to get the master key
// In a real implementation, this would use the customer's private key
func (d *DRMManager) DecryptLicenseKey(encryptedLicenseKey string, customerPrivateKey []byte) ([]byte, error) {
	// For now, we assume the license key IS the master key
	// In production, you'd use envelope encryption here
	return DecryptAES256(encryptedLicenseKey, customerPrivateKey)
}

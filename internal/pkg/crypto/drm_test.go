package crypto

import (
	"bytes"
	"testing"
)

func TestDRMManagerCreation(t *testing.T) {
	masterKey, _ := GenerateAES256Key()
	privateKey, _ := GenerateECDSAKeyPair()

	drm, err := NewDRMManager(masterKey, &privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to create DRM manager: %v", err)
	}

	if drm == nil {
		t.Error("DRM manager is nil")
	}
}

func TestCompleteDecryptionFlow(t *testing.T) {
	// 准备测试数据
	masterKey, _ := GenerateAES256Key()
	privateKey, _ := GenerateECDSAKeyPair()
	drm, _ := NewDRMManager(masterKey, &privateKey.PublicKey)

	originalProvider := []byte("#!/bin/bash\necho 'Provider binary content'")

	// 加密Provider（模拟.cbp包中的加密二进制）
	encryptedWithMaster, err := drm.EncryptProviderWithMasterKey(originalProvider)
	if err != nil {
		t.Fatalf("Failed to encrypt with master key: %v", err)
	}

	// 执行完整的解密流程
	plainProvider, sessionKey, reEncrypted, err := drm.CompleteDecryptionFlow(encryptedWithMaster)
	if err != nil {
		t.Fatalf("Complete decryption flow failed: %v", err)
	}

	// 验证明文Provider
	if !bytes.Equal(plainProvider, originalProvider) {
		t.Error("Decrypted provider doesn't match original")
	}

	// 验证session key长度
	if len(sessionKey) != 32 {
		t.Errorf("Expected session key length 32, got %d", len(sessionKey))
	}

	// 验证重加密后的数据可以用session key解密
	decrypted, err := drm.DecryptWithSessionKey(reEncrypted, sessionKey)
	if err != nil {
		t.Fatalf("Failed to decrypt with session key: %v", err)
	}

	if !bytes.Equal(decrypted, originalProvider) {
		t.Error("Session key decryption doesn't match original")
	}
}

func TestPackageSignatureVerification(t *testing.T) {
	masterKey, _ := GenerateAES256Key()
	privateKey, _ := GenerateECDSAKeyPair()
	drm, _ := NewDRMManager(masterKey, &privateKey.PublicKey)

	packageData := []byte("CBP package content")

	// 签名
	signature, err := SignData(packageData, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign package: %v", err)
	}

	// 验证
	valid, err := drm.VerifyPackageSignature(packageData, signature)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	if !valid {
		t.Error("Valid signature was rejected")
	}

	// 篡改数据后验证应该失败
	tamperedData := []byte("Tampered package content")
	valid, _ = drm.VerifyPackageSignature(tamperedData, signature)
	if valid {
		t.Error("Tampered package was accepted")
	}
}

func TestInvalidMasterKeySize(t *testing.T) {
	invalidKey := []byte("short")
	privateKey, _ := GenerateECDSAKeyPair()

	_, err := NewDRMManager(invalidKey, &privateKey.PublicKey)
	if err == nil {
		t.Error("Expected error for invalid master key size")
	}
}

func TestMissingPublicKey(t *testing.T) {
	masterKey, _ := GenerateAES256Key()

	_, err := NewDRMManager(masterKey, nil)
	if err == nil {
		t.Error("Expected error for missing public key")
	}
}

func TestSessionKeyReEncryption(t *testing.T) {
	masterKey, _ := GenerateAES256Key()
	privateKey, _ := GenerateECDSAKeyPair()
	drm, _ := NewDRMManager(masterKey, &privateKey.PublicKey)

	providerData := []byte("Provider binary")
	sessionKey, _ := drm.GenerateSessionKey()

	// 加密
	encrypted, err := drm.ReEncryptWithSessionKey(providerData, sessionKey)
	if err != nil {
		t.Fatalf("Re-encryption failed: %v", err)
	}

	// 解密
	decrypted, err := drm.DecryptWithSessionKey(encrypted, sessionKey)
	if err != nil {
		t.Fatalf("Session key decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, providerData) {
		t.Error("Session key round-trip failed")
	}
}

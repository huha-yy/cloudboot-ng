package crypto

import (
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	// 生成密钥对
	privateKey, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	data := []byte("CloudBoot NG Provider Package v1.0.0")

	// 签名
	signature, err := SignData(data, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	if signature == "" {
		t.Error("Signature is empty")
	}

	// 验证签名（正确的公钥）
	valid, err := VerifySignature(data, signature, &privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	if !valid {
		t.Error("Signature verification failed with correct key")
	}
}

func TestVerifyWithWrongData(t *testing.T) {
	privateKey, _ := GenerateECDSAKeyPair()
	originalData := []byte("Original data")
	tamperedData := []byte("Tampered data")

	signature, _ := SignData(originalData, privateKey)

	// 验证被篡改的数据应该失败
	valid, _ := VerifySignature(tamperedData, signature, &privateKey.PublicKey)
	if valid {
		t.Error("Signature verification should fail with tampered data")
	}
}

func TestVerifyWithWrongKey(t *testing.T) {
	privateKey1, _ := GenerateECDSAKeyPair()
	privateKey2, _ := GenerateECDSAKeyPair()

	data := []byte("Test data")
	signature, _ := SignData(data, privateKey1)

	// 使用错误的公钥验证应该失败
	valid, _ := VerifySignature(data, signature, &privateKey2.PublicKey)
	if valid {
		t.Error("Signature verification should fail with wrong public key")
	}
}

func TestPrivateKeyPEMRoundTrip(t *testing.T) {
	originalKey, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// 导出为PEM
	pemStr, err := PrivateKeyToPEM(originalKey)
	if err != nil {
		t.Fatalf("Failed to export to PEM: %v", err)
	}

	// 从PEM加载
	loadedKey, err := PrivateKeyFromPEM(pemStr)
	if err != nil {
		t.Fatalf("Failed to load from PEM: %v", err)
	}

	// 验证密钥是否相同（通过签名验证）
	data := []byte("test data")
	signature, _ := SignData(data, originalKey)
	valid, _ := VerifySignature(data, signature, &loadedKey.PublicKey)

	if !valid {
		t.Error("Loaded key doesn't match original key")
	}
}

func TestPublicKeyPEMRoundTrip(t *testing.T) {
	privateKey, _ := GenerateECDSAKeyPair()
	publicKey := &privateKey.PublicKey

	// 导出为PEM
	pemStr, err := PublicKeyToPEM(publicKey)
	if err != nil {
		t.Fatalf("Failed to export to PEM: %v", err)
	}

	// 从PEM加载
	loadedKey, err := PublicKeyFromPEM(pemStr)
	if err != nil {
		t.Fatalf("Failed to load from PEM: %v", err)
	}

	// 验证公钥是否相同
	data := []byte("test data")
	signature, _ := SignData(data, privateKey)
	valid, _ := VerifySignature(data, signature, loadedKey)

	if !valid {
		t.Error("Loaded public key doesn't match original")
	}
}

func TestInvalidSignature(t *testing.T) {
	privateKey, _ := GenerateECDSAKeyPair()
	data := []byte("test")

	// 无效的Base64
	_, err := VerifySignature(data, "invalid!!!base64", &privateKey.PublicKey)
	if err == nil {
		t.Error("Expected error for invalid base64")
	}

	// 长度错误的签名
	invalidSig := "dGVzdA==" // 有效的Base64但长度错误
	valid, _ := VerifySignature(data, invalidSig, &privateKey.PublicKey)
	if valid {
		t.Error("Should reject signature with invalid length")
	}
}

package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptAES256(t *testing.T) {
	key, err := GenerateAES256Key()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	plaintext := []byte("Hello, CloudBoot NG! This is a secret message.")

	// 加密
	ciphertext, err := EncryptAES256(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if ciphertext == "" {
		t.Error("Ciphertext is empty")
	}

	// 解密
	decrypted, err := DecryptAES256(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text doesn't match.\nExpected: %s\nGot: %s", plaintext, decrypted)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1, _ := GenerateAES256Key()
	key2, _ := GenerateAES256Key()

	plaintext := []byte("Secret data")
	ciphertext, _ := EncryptAES256(plaintext, key1)

	// 使用错误的密钥解密应该失败
	_, err := DecryptAES256(ciphertext, key2)
	if err == nil {
		t.Error("Expected decryption to fail with wrong key")
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	key, err := GenerateAES256Key()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	originalData := []byte("Binary file content: \x00\x01\x02\xFF")

	// 加密文件
	encrypted, err := EncryptFile(originalData, key)
	if err != nil {
		t.Fatalf("File encryption failed: %v", err)
	}

	// 解密文件
	decrypted, err := DecryptFile(encrypted, key)
	if err != nil {
		t.Fatalf("File decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, originalData) {
		t.Errorf("Decrypted file doesn't match original")
	}
}

func TestInvalidKeySize(t *testing.T) {
	invalidKey := []byte("short") // 5 bytes instead of 32

	_, err := EncryptAES256([]byte("test"), invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key size")
	}

	_, err = DecryptAES256("dGVzdA==", invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key size")
	}
}

func TestGenerateKeyLength(t *testing.T) {
	key, err := GenerateAES256Key()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}
}

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
)

// GenerateECDSAKeyPair 生成ECDSA密钥对（P-256曲线）
func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// SignData 使用私钥对数据进行签名
func SignData(data []byte, privateKey *ecdsa.PrivateKey) (string, error) {
	// 计算数据哈希
	hash := sha256.Sum256(data)

	// 签名
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// 序列化签名（r和s拼接）
	signature := append(r.Bytes(), s.Bytes()...)

	// Base64编码
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature 使用公钥验证签名
func VerifySignature(data []byte, signatureB64 string, publicKey *ecdsa.PublicKey) (bool, error) {
	// Base64解码签名
	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	// 签名应该是r和s的拼接，长度为64字节（P-256）
	if len(signature) != 64 {
		return false, fmt.Errorf("invalid signature length: %d", len(signature))
	}

	// 分离r和s
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])

	// 计算数据哈希
	hash := sha256.Sum256(data)

	// 验证签名
	valid := ecdsa.Verify(publicKey, hash[:], r, s)
	return valid, nil
}

// PrivateKeyToPEM 将私钥导出为PEM格式
func PrivateKeyToPEM(key *ecdsa.PrivateKey) (string, error) {
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}

	return string(pem.EncodeToMemory(block)), nil
}

// PrivateKeyFromPEM 从PEM格式加载私钥
func PrivateKeyFromPEM(pemStr string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

// PublicKeyToPEM 将公钥导出为PEM格式
func PublicKeyToPEM(key *ecdsa.PublicKey) (string, error) {
	keyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keyBytes,
	}

	return string(pem.EncodeToMemory(block)), nil
}

// PublicKeyFromPEM 从PEM格式加载公钥
func PublicKeyFromPEM(pemStr string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	return ecdsaPub, nil
}

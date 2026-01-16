package cspm

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloudboot/cloudboot-ng/internal/core/audit"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/crypto"
)

// PluginManager Provider插件管理器
type PluginManager struct {
	storeDir           string // Private Store目录
	mu                 sync.RWMutex
	plugins            map[string]*ProviderInfo
	drmManager         *crypto.DRMManager
	watermarkValidator *audit.WatermarkValidator
}

// ProviderInfo Provider信息
type ProviderInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Vendor   string `json:"vendor"`
	Model    string `json:"model"`
	FilePath string `json:"file_path"`
	Checksum string `json:"checksum"` // SHA256
	Manifest Manifest `json:"manifest"`
	Watermark audit.Watermark `json:"watermark"`
	WatermarkViolation *audit.WatermarkViolation `json:"watermark_violation,omitempty"`
}

// NewPluginManager 创建Plugin Manager
func NewPluginManager(storeDir string, masterKey []byte, officialPubKey *ecdsa.PublicKey, currentLicenseID string) (*PluginManager, error) {
	// 确保Store目录存在
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	// 创建DRM管理器
	drmManager, err := crypto.NewDRMManager(masterKey, officialPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create DRM manager: %w", err)
	}

	// 创建水印验证器
	auditLogPath := filepath.Join(storeDir, "../audit/watermark_violations.log")
	watermarkValidator, err := audit.NewWatermarkValidator(currentLicenseID, auditLogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create watermark validator: %w", err)
	}

	pm := &PluginManager{
		storeDir:           storeDir,
		plugins:            make(map[string]*ProviderInfo),
		drmManager:         drmManager,
		watermarkValidator: watermarkValidator,
	}

	// 扫描已存在的Provider
	if err := pm.scanProviders(); err != nil {
		return nil, fmt.Errorf("failed to scan providers: %w", err)
	}

	return pm, nil
}

// ImportProvider 导入Provider包（.cbp文件）
// 完整的DRM解密和水印验证逻辑
func (pm *PluginManager) ImportProvider(cbpPath string) (*ProviderInfo, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 步骤1: 解析.cbp包
	pkg, err := ParseCBP(cbpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cbp package: %w", err)
	}

	// 步骤2: 验证签名（防止篡改）
	// 注意：实际签名应该是对整个.cbp文件的签名，这里简化为对manifest的签名
	packageData := []byte(pkg.Manifest.ID + pkg.Manifest.Version)
	valid, err := pm.drmManager.VerifyPackageSignature(packageData, pkg.Signature)
	if err != nil || !valid {
		return nil, fmt.Errorf("signature verification failed: invalid or tampered package")
	}

	// 步骤3: 验证水印
	watermarkViolation, err := pm.watermarkValidator.ValidateWatermark(
		pkg.Manifest.ID,
		pkg.Manifest.Name,
		pkg.Watermark,
	)
	if err != nil {
		return nil, fmt.Errorf("watermark validation failed: %w", err)
	}

	// 步骤4: 解密Provider二进制
	plainProvider, err := pm.drmManager.DecryptProviderWithMasterKey(pkg.ProviderBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt provider: %w", err)
	}

	// 步骤5: 保存明文Provider到临时文件（用于执行）
	providerID := pkg.Manifest.ID
	providerPath := filepath.Join(pm.storeDir, providerID)

	if err := os.WriteFile(providerPath, plainProvider, 0755); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}

	// 步骤6: 计算校验和
	hash := sha256.Sum256(plainProvider)
	checksum := hex.EncodeToString(hash[:])

	// 步骤7: 创建Provider信息
	info := &ProviderInfo{
		ID:                 providerID,
		Name:               pkg.Manifest.Name,
		Version:            pkg.Manifest.Version,
		Vendor:             pkg.Manifest.Vendor,
		Model:              pkg.Manifest.Model,
		FilePath:           providerPath,
		Checksum:           checksum,
		Manifest:           pkg.Manifest,
		Watermark:          pkg.Watermark,
		WatermarkViolation: watermarkViolation, // 如果有违规，记录下来
	}

	pm.plugins[providerID] = info

	return info, nil
}

// GetProvider 获取Provider信息
func (pm *PluginManager) GetProvider(id string) (*ProviderInfo, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info, ok := pm.plugins[id]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", id)
	}

	return info, nil
}

// ListProviders 列出所有Provider
func (pm *PluginManager) ListProviders() []*ProviderInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	providers := make([]*ProviderInfo, 0, len(pm.plugins))
	for _, info := range pm.plugins {
		providers = append(providers, info)
	}

	return providers
}

// DeleteProvider 删除Provider
func (pm *PluginManager) DeleteProvider(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	info, ok := pm.plugins[id]
	if !ok {
		return fmt.Errorf("provider not found: %s", id)
	}

	// 删除文件
	if err := os.Remove(info.FilePath); err != nil {
		return fmt.Errorf("failed to delete provider file: %w", err)
	}

	// 从内存中移除
	delete(pm.plugins, id)

	return nil
}

// CreateExecutor 为指定Provider创建Executor
func (pm *PluginManager) CreateExecutor(id string) (*Executor, error) {
	info, err := pm.GetProvider(id)
	if err != nil {
		return nil, err
	}

	// TODO: 实现内存解密逻辑
	// 当前直接使用文件路径（开发阶段）

	return NewExecutor(info.FilePath), nil
}

// scanProviders 扫描Store目录中已存在的Provider
func (pm *PluginManager) scanProviders() error {
	entries, err := os.ReadDir(pm.storeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 目录不存在，忽略
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(pm.storeDir, entry.Name())

		// 计算校验和
		checksum, err := calculateFileChecksum(filePath)
		if err != nil {
			continue // 跳过无法计算的文件
		}

		info := &ProviderInfo{
			ID:       entry.Name(),
			Name:     entry.Name(),
			Version:  "unknown",
			FilePath: filePath,
			Checksum: checksum,
		}

		pm.plugins[entry.Name()] = info
	}

	return nil
}

// 辅助函数

func generateProviderID(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])[:16]
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func calculateFileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

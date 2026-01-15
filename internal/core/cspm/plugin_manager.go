package cspm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// PluginManager Provider插件管理器
type PluginManager struct {
	storeDir string // Private Store目录
	mu       sync.RWMutex
	plugins  map[string]*ProviderInfo
}

// ProviderInfo Provider信息
type ProviderInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	FilePath string `json:"file_path"`
	Checksum string `json:"checksum"` // SHA256
}

// NewPluginManager 创建Plugin Manager
func NewPluginManager(storeDir string) (*PluginManager, error) {
	// 确保Store目录存在
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	pm := &PluginManager{
		storeDir: storeDir,
		plugins:  make(map[string]*ProviderInfo),
	}

	// 扫描已存在的Provider
	if err := pm.scanProviders(); err != nil {
		return nil, fmt.Errorf("failed to scan providers: %w", err)
	}

	return pm, nil
}

// ImportProvider 导入Provider包（.cbp文件）
// TODO: 实现完整的DRM解密和水印验证逻辑
func (pm *PluginManager) ImportProvider(cbpPath string) (*ProviderInfo, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 读取.cbp文件
	file, err := os.Open(cbpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cbp file: %w", err)
	}
	defer file.Close()

	// 计算校验和
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	checksum := hex.EncodeToString(hash.Sum(nil))

	// 生成Provider ID
	providerID := generateProviderID(cbpPath)

	// 复制到Store目录
	destPath := filepath.Join(pm.storeDir, providerID)
	if err := copyFile(cbpPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy provider: %w", err)
	}

	// 设置可执行权限
	if err := os.Chmod(destPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to set executable permission: %w", err)
	}

	// 创建Provider信息
	info := &ProviderInfo{
		ID:       providerID,
		Name:     filepath.Base(cbpPath),
		Version:  "1.0.0", // TODO: 从manifest.json读取
		FilePath: destPath,
		Checksum: checksum,
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

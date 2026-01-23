package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupManager 数据库备份管理器
type BackupManager struct {
	dbPath     string
	backupDir  string
	maxBackups int
}

// NewBackupManager 创建备份管理器
func NewBackupManager(dbPath, backupDir string) *BackupManager {
	return &BackupManager{
		dbPath:     dbPath,
		backupDir:  backupDir,
		maxBackups: 7, // 保留最近7天的备份
	}
}

// Backup 执行数据库备份（使用 VACUUM INTO）
func (bm *BackupManager) Backup() (string, error) {
	// 确保备份目录存在
	if err := os.MkdirAll(bm.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 生成备份文件名（带时间戳，包含纳秒以避免冲突）
	timestamp := time.Now().Format("20060102-150405.000000")
	backupFile := filepath.Join(bm.backupDir, fmt.Sprintf("cloudboot-%s.db", timestamp))

	// 如果文件已存在，删除它（VACUUM INTO 不允许覆盖）
	if _, err := os.Stat(backupFile); err == nil {
		if err := os.Remove(backupFile); err != nil {
			return "", fmt.Errorf("failed to remove existing backup file: %w", err)
		}
	}

	// 使用 VACUUM INTO 进行在线备份
	sql := fmt.Sprintf("VACUUM INTO '%s'", backupFile)
	if err := DB.Exec(sql).Error; err != nil {
		return "", fmt.Errorf("backup failed: %w", err)
	}

	// 清理旧备份
	if err := bm.cleanOldBackups(); err != nil {
		return backupFile, fmt.Errorf("backup succeeded but cleanup failed: %w", err)
	}

	return backupFile, nil
}

// Restore 从备份文件恢复数据库
func (bm *BackupManager) Restore(backupFile string) error {
	// 验证备份文件存在
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	// 关闭当前数据库连接
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	// 备份当前数据库文件（以防恢复失败）
	currentBackup := bm.dbPath + ".before-restore"
	if err := copyFile(bm.dbPath, currentBackup); err != nil {
		return fmt.Errorf("failed to backup current database: %w", err)
	}

	// 用备份文件替换当前数据库
	if err := copyFile(backupFile, bm.dbPath); err != nil {
		// 恢复失败，回滚
		_ = copyFile(currentBackup, bm.dbPath)
		return fmt.Errorf("failed to restore database: %w", err)
	}

	// 删除临时备份
	_ = os.Remove(currentBackup)

	return nil
}

// ListBackups 列出所有备份文件
func (bm *BackupManager) ListBackups() ([]BackupInfo, error) {
	files, err := filepath.Glob(filepath.Join(bm.backupDir, "cloudboot-*.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	backups := make([]BackupInfo, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Path:      file,
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	return backups, nil
}

// cleanOldBackups 清理超过保留期限的备份
func (bm *BackupManager) cleanOldBackups() error {
	backups, err := bm.ListBackups()
	if err != nil {
		return err
	}

	if len(backups) <= bm.maxBackups {
		return nil
	}

	// 按时间排序，删除最旧的备份
	for i := 0; i < len(backups)-bm.maxBackups; i++ {
		if err := os.Remove(backups[i].Path); err != nil {
			return fmt.Errorf("failed to remove old backup: %w", err)
		}
	}

	return nil
}

// BackupInfo 备份文件信息
type BackupInfo struct {
	Path      string
	Size      int64
	CreatedAt time.Time
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

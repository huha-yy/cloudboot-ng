package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupManager_Backup(t *testing.T) {
	// 创建临时数据库
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	backupDir := filepath.Join(tempDir, "backups")

	// 初始化测试数据库
	err := Init(Config{DSN: dbPath})
	require.NoError(t, err)

	// 创建备份管理器
	bm := NewBackupManager(dbPath, backupDir)

	// 执行备份
	backupFile, err := bm.Backup()
	require.NoError(t, err)
	assert.NotEmpty(t, backupFile)

	// 验证备份文件存在
	_, err = os.Stat(backupFile)
	assert.NoError(t, err)

	// 验证备份文件大小 > 0
	info, _ := os.Stat(backupFile)
	assert.Greater(t, info.Size(), int64(0))
}

func TestBackupManager_ListBackups(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	backupDir := filepath.Join(tempDir, "backups")

	err := Init(Config{DSN: dbPath})
	require.NoError(t, err)

	bm := NewBackupManager(dbPath, backupDir)

	// 创建多个备份
	for i := 0; i < 3; i++ {
		_, err := bm.Backup()
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // 确保时间戳不同
	}

	// 列出备份
	backups, err := bm.ListBackups()
	require.NoError(t, err)
	assert.Len(t, backups, 3)

	// 验证备份按时间排序
	for i := 0; i < len(backups)-1; i++ {
		assert.True(t, backups[i].CreatedAt.Before(backups[i+1].CreatedAt) ||
			backups[i].CreatedAt.Equal(backups[i+1].CreatedAt))
	}
}

func TestBackupManager_CleanOldBackups(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	backupDir := filepath.Join(tempDir, "backups")

	err := Init(Config{DSN: dbPath})
	require.NoError(t, err)

	bm := NewBackupManager(dbPath, backupDir)
	bm.maxBackups = 3 // 只保留3个备份

	// 创建5个备份
	for i := 0; i < 5; i++ {
		_, err := bm.Backup()
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// 验证只保留了3个备份
	backups, err := bm.ListBackups()
	require.NoError(t, err)
	assert.Len(t, backups, 3)
}

func TestBackupManager_Restore(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	backupDir := filepath.Join(tempDir, "backups")

	// 初始化数据库并插入测试数据
	err := Init(Config{DSN: dbPath})
	require.NoError(t, err)

	// 插入测试数据
	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"type:varchar(100)"`
	}
	err = DB.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	err = DB.Create(&TestModel{Name: "test1"}).Error
	require.NoError(t, err)

	// 创建备份
	bm := NewBackupManager(dbPath, backupDir)
	backupFile, err := bm.Backup()
	require.NoError(t, err)

	// 修改数据
	err = DB.Create(&TestModel{Name: "test2"}).Error
	require.NoError(t, err)

	// 验证有2条记录
	var count int64
	DB.Model(&TestModel{}).Count(&count)
	assert.Equal(t, int64(2), count)

	// 恢复备份
	err = bm.Restore(backupFile)
	require.NoError(t, err)

	// 重新连接数据库
	err = Init(Config{DSN: dbPath})
	require.NoError(t, err)

	// 验证只有1条记录（恢复到备份时的状态）
	DB.Model(&TestModel{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

package database

import (
	"log"
	"time"
)

// BackupScheduler 定时备份调度器
type BackupScheduler struct {
	manager  *BackupManager
	interval time.Duration
	stopCh   chan struct{}
}

// NewBackupScheduler 创建定时备份调度器
func NewBackupScheduler(manager *BackupManager, interval time.Duration) *BackupScheduler {
	return &BackupScheduler{
		manager:  manager,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动定时备份
func (bs *BackupScheduler) Start() {
	log.Printf("📦 启动数据库定时备份 (间隔: %v)", bs.interval)

	// 立即执行一次备份
	go bs.runBackup()

	// 启动定时器
	ticker := time.NewTicker(bs.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				bs.runBackup()
			case <-bs.stopCh:
				ticker.Stop()
				log.Println("📦 数据库定时备份已停止")
				return
			}
		}
	}()
}

// Stop 停止定时备份
func (bs *BackupScheduler) Stop() {
	close(bs.stopCh)
}

// runBackup 执行备份
func (bs *BackupScheduler) runBackup() {
	log.Println("📦 开始数据库备份...")
	start := time.Now()

	backupFile, err := bs.manager.Backup()
	if err != nil {
		log.Printf("❌ 数据库备份失败: %v", err)
		return
	}

	duration := time.Since(start)
	log.Printf("✅ 数据库备份成功: %s (耗时: %v)", backupFile, duration)
}

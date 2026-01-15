package database

import (
	"fmt"
	"log"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// Config 数据库配置
type Config struct {
	DSN      string
	LogLevel logger.LogLevel
}

// Init 初始化数据库连接
func Init(config Config) error {
	var err error

	// 默认配置
	if config.DSN == "" {
		config.DSN = "cloudboot.db?_journal_mode=WAL"
	}

	// 打开数据库连接
	DB, err = gorm.Open(sqlite.Open(config.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	log.Printf("✅ 数据库连接成功: %s", config.DSN)

	// 自动迁移
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	log.Println("🔄 开始数据库迁移...")

	err := DB.AutoMigrate(
		&models.Machine{},
		&models.Job{},
		&models.OSProfile{},
		&models.License{},
	)

	if err != nil {
		return err
	}

	log.Println("✅ 数据库迁移完成")
	return nil
}

// Close 关闭数据库连接
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// SetDB 设置数据库实例（用于测试）
func SetDB(db *gorm.DB) {
	DB = db
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

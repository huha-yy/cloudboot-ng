package models

import (
	"time"
)

// OSProfile 操作系统安装模板
type OSProfile struct {
	ID        string        `gorm:"primaryKey" json:"id"`
	Name      string        `gorm:"uniqueIndex;type:varchar(100)" json:"name"`
	Distro    string        `gorm:"type:varchar(50)" json:"distro"`
	Config    ProfileConfig `gorm:"serializer:json;type:text" json:"config"`
	CreatedAt time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// ProfileConfig 安装配置详情
type ProfileConfig struct {
	RootPasswordHash string          `json:"root_password_hash"`
	Timezone         string          `json:"timezone"`
	RepoURL          string          `json:"repo_url"`      // Mirror URL for package installation
	Partitions       []Partition     `json:"partitions"`
	Network          NetworkConfig   `json:"network"`
	Packages         []string        `json:"packages"`
	PostScript       string          `json:"post_script,omitempty"` // Post-installation script
}

// Partition 分区配置
type Partition struct {
	MountPoint string `json:"mount_point"` // /, /boot, /home, etc.
	Size       string `json:"size"`        // 100GB, 20%
	FSType     string `json:"fstype"`      // ext4, xfs, swap
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	DHCP     bool     `json:"dhcp"`
	Hostname string   `json:"hostname"`
	IP       string   `json:"ip,omitempty"`
	Netmask  string   `json:"netmask,omitempty"`
	Gateway  string   `json:"gateway,omitempty"`
	DNS      []string `json:"dns,omitempty"`
}

// TableName 指定表名
func (OSProfile) TableName() string {
	return "os_profiles"
}

// Validate 验证配置合法性
func (p *OSProfile) Validate() error {
	// TODO: 实现验证逻辑
	// - 检查分区配置（必须有 / 分区）
	// - 检查网络配置
	// - 检查密码哈希格式
	return nil
}

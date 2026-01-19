package models

import (
	"time"
)

// OSProfile 操作系统安装模板
type OSProfile struct {
	ID        string        `gorm:"primaryKey" json:"id"`
	Name      string        `gorm:"uniqueIndex;type:varchar(100)" json:"name"`
	Distro    string        `gorm:"type:varchar(50)" json:"distro"`    // centos7, ubuntu22, rocky8, suse15
	Version   string        `gorm:"type:varchar(20)" json:"version"`   // 7.9, 22.04, 8.8, 15.5
	Config    ProfileConfig `gorm:"serializer:json;type:text" json:"config"`
	CreatedAt time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// ProfileConfig 安装配置详情
type ProfileConfig struct {
	RootPasswordHash string               `json:"root_password_hash"`
	Timezone         string               `json:"timezone"`             // America/New_York, Asia/Shanghai
	RepoURL          string               `json:"repo_url"`             // Mirror URL for package installation
	Partitions       []PartitionConfig    `json:"partitions"`
	NetworkConfig    *NetworkConfigDetail `json:"network_config"`
	Packages         []string             `json:"packages"`             // Additional packages to install
	PostScript       string               `json:"post_script"`          // Post-installation script
	InstallAgent     bool                 `json:"install_agent"`        // Install CloudBoot Agent
	KernelURL        string               `json:"kernel_url,omitempty"` // Custom kernel URL
	InitrdURL        string               `json:"initrd_url,omitempty"` // Custom initrd URL
}

// PartitionConfig 分区配置
type PartitionConfig struct {
	MountPoint string `json:"mount_point"` // /, /boot, /boot/efi, /home, swap
	SizeMB     int    `json:"size_mb"`     // Size in MB (0 = use remaining)
	FileSystem string `json:"file_system"` // ext4, xfs, swap, vfat
	Grow       bool   `json:"grow"`        // Grow to fill available space
}

// NetworkConfigDetail 网络配置详情
type NetworkConfigDetail struct {
	BootProto string `json:"boot_proto"` // dhcp, static
	Device    string `json:"device"`     // eth0, ens192
	IPAddress string `json:"ip_address,omitempty"`
	Netmask   string `json:"netmask,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
	DNS       string `json:"dns,omitempty"` // Single DNS server
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

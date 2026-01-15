package models

import (
	"time"
)

// Machine 表示一台物理服务器资产
type Machine struct {
	ID            string         `gorm:"primaryKey" json:"id"`
	Hostname      string         `gorm:"uniqueIndex" json:"hostname"`
	MacAddress    string         `gorm:"uniqueIndex;column:mac_address" json:"mac_address"`
	IPAddress     string         `gorm:"column:ip_address" json:"ip_address"`
	Status        MachineStatus  `gorm:"type:varchar(20);index" json:"status"`
	HardwareSpec  HardwareInfo   `gorm:"serializer:json;type:text" json:"hardware_spec"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// MachineStatus 机器状态枚举
type MachineStatus string

const (
	// MachineStatusDiscovered PXE启动发现，未就绪
	MachineStatusDiscovered MachineStatus = "discovered"
	// MachineStatusReady 已就绪，可开始安装
	MachineStatusReady MachineStatus = "ready"
	// MachineStatusInstalling 安装中
	MachineStatusInstalling MachineStatus = "installing"
	// MachineStatusActive 安装完成，激活
	MachineStatusActive MachineStatus = "active"
	// MachineStatusError 发生错误
	MachineStatusError MachineStatus = "error"
)

// HardwareInfo 标准化硬件指纹 (Schema v1.0)
type HardwareInfo struct {
	SchemaVersion       string              `json:"schema_version"`
	System              SystemInfo          `json:"system"`
	CPU                 CPUInfo             `json:"cpu"`
	Memory              MemoryInfo          `json:"memory"`
	StorageControllers  []ControllerInfo    `json:"storage_controllers"`
	NetworkInterfaces   []NICInfo           `json:"network_interfaces"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Manufacturer string `json:"manufacturer"`
	ProductName  string `json:"product_name"`
	SerialNumber string `json:"serial_number"`
}

// CPUInfo CPU信息
type CPUInfo struct {
	Arch    string `json:"arch"`     // x86_64, arm64
	Model   string `json:"model"`    // Intel Xeon E5-2680 v4
	Cores   int    `json:"cores"`    // 总核心数
	Sockets int    `json:"sockets"`  // CPU插槽数
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	TotalBytes int64      `json:"total_bytes"`
	DIMMs      []DimmInfo `json:"dimms"`
}

// DimmInfo 内存条信息
type DimmInfo struct {
	Slot      string `json:"slot"`       // DIMM_A1
	SizeBytes int64  `json:"size_bytes"` // 34359738368 (32GB)
	Speed     int    `json:"speed"`      // 3200 MHz
}

// ControllerInfo 存储控制器信息
type ControllerInfo struct {
	PCIID  string `json:"pci_id"`  // 1000:005f
	Vendor string `json:"vendor"`  // LSI Logic
	Model  string `json:"model"`   // MegaRAID SAS 3108
	Driver string `json:"driver"`  // megaraid_sas
}

// NICInfo 网卡信息
type NICInfo struct {
	Name  string `json:"name"`  // eth0
	MAC   string `json:"mac"`   // aa:bb:cc:dd:ee:ff
	Speed int    `json:"speed"` // 10000 (Mbps)
	Link  bool   `json:"link"`  // true/false
}

// TableName 指定表名
func (Machine) TableName() string {
	return "machines"
}

// IsOnline 检查机器是否在线（基于UpdatedAt时间戳）
func (m *Machine) IsOnline() bool {
	// 如果最后更新时间在5分钟以内，认为在线
	return time.Since(m.UpdatedAt) < 5*time.Minute
}

// IsReady 检查机器是否就绪可部署
func (m *Machine) IsReady() bool {
	return m.Status == MachineStatusReady || m.Status == MachineStatusActive
}

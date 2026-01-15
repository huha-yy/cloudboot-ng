package configgen

import (
	"fmt"
	"net"
	"strings"

	"github.com/cloudboot/cloudboot-ng/internal/models"
)

// Validate 验证OSProfile配置
func (g *Generator) Validate(profile *models.OSProfile) error {
	if profile == nil {
		return fmt.Errorf("profile is nil")
	}

	// 验证OS类型
	if err := validateOSType(profile.Distro); err != nil {
		return err
	}

	// 验证分区配置
	if err := validatePartitions(profile.Config.Partitions); err != nil {
		return err
	}

	// 验证网络配置
	if err := validateNetwork(&profile.Config.Network); err != nil {
		return err
	}

	return nil
}

// validateOSType 验证操作系统类型
func validateOSType(osType string) error {
	validTypes := map[string]bool{
		"centos7":  true,
		"centos8":  true,
		"ubuntu20": true,
		"ubuntu22": true,
		"sles12":   true,
		"sles15":   true,
	}

	if !validTypes[osType] {
		return fmt.Errorf("unsupported OS type: %s", osType)
	}

	return nil
}

// validatePartitions 验证分区配置
func validatePartitions(partitions []models.Partition) error {
	if len(partitions) == 0 {
		return fmt.Errorf("no partitions defined")
	}

	hasRoot := false

	for i, part := range partitions {
		// 检查挂载点
		if part.MountPoint == "" {
			return fmt.Errorf("partition %d: mount point is empty", i)
		}

		// 检查文件系统类型
		if part.FSType == "" {
			return fmt.Errorf("partition %d: filesystem type is empty", i)
		}

		// 检查大小 (简化版本：只检查非空)
		if part.Size == "" {
			return fmt.Errorf("partition %d: size is empty", i)
		}

		// 标记关键分区
		if part.MountPoint == "/" {
			hasRoot = true
		}

		if part.MountPoint == "swap" && part.FSType != "swap" {
			return fmt.Errorf("swap partition must have fstype=swap, got %s", part.FSType)
		}

		// 验证文件系统类型
		if err := validateFilesystem(part.FSType, part.MountPoint); err != nil {
			return fmt.Errorf("partition %d: %w", i, err)
		}
	}

	// 检查必需分区
	if !hasRoot {
		return fmt.Errorf("root (/) partition is required")
	}

	return nil
}

// validateFilesystem 验证文件系统类型
func validateFilesystem(fstype, mount string) error {
	validFS := map[string]bool{
		"ext4":  true,
		"xfs":   true,
		"swap":  true,
		"vfat":  true,
		"btrfs": true,
	}

	if !validFS[fstype] {
		return fmt.Errorf("unsupported filesystem type: %s", fstype)
	}

	// /boot 特殊检查 (UEFI系统需要vfat)
	if mount == "/boot/efi" && fstype != "vfat" {
		return fmt.Errorf("/boot/efi must use vfat filesystem")
	}

	return nil
}

// validateNetwork 验证网络配置
func validateNetwork(network *models.NetworkConfig) error {
	if network == nil {
		return fmt.Errorf("network configuration is required")
	}

	// 验证主机名
	if network.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	if len(network.Hostname) > 63 {
		return fmt.Errorf("hostname too long (max 63 characters)")
	}

	// 验证IP地址
	if network.IP == "" {
		return fmt.Errorf("IP address is required")
	}

	ip := net.ParseIP(network.IP)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", network.IP)
	}

	// 验证子网掩码
	if network.Netmask == "" {
		return fmt.Errorf("netmask is required")
	}

	if !isValidNetmask(network.Netmask) {
		return fmt.Errorf("invalid netmask: %s", network.Netmask)
	}

	// 验证网关
	if network.Gateway != "" {
		gw := net.ParseIP(network.Gateway)
		if gw == nil {
			return fmt.Errorf("invalid gateway address: %s", network.Gateway)
		}
	}

	// 验证DNS
	if len(network.DNS) > 0 {
		for _, dns := range network.DNS {
			dns = strings.TrimSpace(dns)
			if net.ParseIP(dns) == nil {
				return fmt.Errorf("invalid DNS server address: %s", dns)
			}
		}
	}

	return nil
}

// isValidNetmask 验证子网掩码格式
func isValidNetmask(netmask string) bool {
	validMasks := []string{
		"255.255.255.255", // /32
		"255.255.255.254", // /31
		"255.255.255.252", // /30
		"255.255.255.248", // /29
		"255.255.255.240", // /28
		"255.255.255.224", // /27
		"255.255.255.192", // /26
		"255.255.255.128", // /25
		"255.255.255.0",   // /24
		"255.255.254.0",   // /23
		"255.255.252.0",   // /22
		"255.255.248.0",   // /21
		"255.255.240.0",   // /20
		"255.255.224.0",   // /19
		"255.255.192.0",   // /18
		"255.255.128.0",   // /17
		"255.255.0.0",     // /16
		"255.254.0.0",     // /15
		"255.252.0.0",     // /14
		"255.248.0.0",     // /13
		"255.240.0.0",     // /12
		"255.224.0.0",     // /11
		"255.192.0.0",     // /10
		"255.128.0.0",     // /9
		"255.0.0.0",       // /8
	}

	for _, mask := range validMasks {
		if netmask == mask {
			return true
		}
	}

	return false
}

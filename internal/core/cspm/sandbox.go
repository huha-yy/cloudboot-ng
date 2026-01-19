package cspm

import (
	"context"
	"os/exec"
)

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	// AllowedPaths Provider允许访问的路径列表
	AllowedPaths []string

	// WorkDir Provider工作目录（默认: /opt/cloudboot/runtime）
	WorkDir string

	// MaxMemoryMB 最大内存限制（MB）
	MaxMemoryMB int64

	// MaxCPUPercent CPU使用率限制（百分比）
	MaxCPUPercent int

	// MaxProcesses 最大进程数限制
	MaxProcesses int

	// NetworkIsolation 是否隔离网络
	NetworkIsolation bool

	// ReadOnlyPaths 只读路径列表（Provider不能写入）
	ReadOnlyPaths []string
}

// DefaultSandboxConfig 默认沙箱配置
func DefaultSandboxConfig() *SandboxConfig {
	return &SandboxConfig{
		AllowedPaths: []string{
			"/opt/cloudboot/runtime", // Provider工作目录
		},
		WorkDir:          "/opt/cloudboot/runtime",
		MaxMemoryMB:      512,  // 512MB内存限制
		MaxCPUPercent:    50,   // 50% CPU限制
		MaxProcesses:     10,   // 最多10个进程
		NetworkIsolation: true, // 默认隔离网络
		ReadOnlyPaths: []string{
			"/usr",     // 系统二进制只读
			"/lib",     // 系统库只读
			"/lib64",   // 系统库只读
			"/bin",     // 系统命令只读
			"/sbin",    // 系统命令只读
			"/etc/ld.so.cache", // 动态链接器缓存只读
		},
	}
}

// Sandbox Provider沙箱接口
type Sandbox interface {
	// Apply 将沙箱配置应用到命令
	Apply(ctx context.Context, cmd *exec.Cmd, config *SandboxConfig) error

	// Cleanup 清理沙箱资源
	Cleanup() error
}

// NewSandbox 创建平台特定的沙箱实例
// 具体实现在各平台的sandbox_*.go文件中

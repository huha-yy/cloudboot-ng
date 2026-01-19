// +build darwin

package cspm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// NewSandbox 创建macOS沙箱实例
func NewSandbox() Sandbox {
	return &DarwinSandbox{}
}

// DarwinSandbox macOS平台的沙箱实现
// macOS不支持Linux namespace，使用基础隔离机制
type DarwinSandbox struct {
	// macOS特定的沙箱状态
}

// Apply 应用沙箱配置到命令
func (s *DarwinSandbox) Apply(ctx context.Context, cmd *exec.Cmd, config *SandboxConfig) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	// 1. 设置工作目录
	if config.WorkDir != "" {
		cmd.Dir = config.WorkDir

		// 确保工作目录存在
		if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
			return fmt.Errorf("failed to create workdir: %w", err)
		}
	}

	// 2. 创建新的进程组（便于管理子进程）
	cmd.SysProcAttr.Setpgid = true

	// 3. macOS资源限制（通过setrlimit）
	// 注意：macOS的rlimit支持有限，某些限制可能不可用

	// 内存限制 - macOS不直接支持RLIMIT_RSS，使用RLIMIT_DATA代替
	if config.MaxMemoryMB > 0 {
		memLimitBytes := uint64(config.MaxMemoryMB * 1024 * 1024)

		// 使用RLIMIT_DATA (数据段大小限制)
		if err := syscall.Setrlimit(syscall.RLIMIT_DATA, &syscall.Rlimit{
			Cur: memLimitBytes,
			Max: memLimitBytes,
		}); err != nil {
			// macOS可能不支持某些限制，记录警告但继续
			fmt.Fprintf(os.Stderr, "Warning: failed to set memory limit: %v\n", err)
		}
	}

	// 进程数限制 - macOS不支持RLIMIT_NPROC（每用户进程数限制）
	// 可以使用其他机制或跳过此限制
	if config.MaxProcesses > 0 {
		// macOS不支持，仅记录警告
		fmt.Fprintf(os.Stderr, "Warning: process limit not supported on macOS\n")
	}

	// 4. 文件描述符限制（防止资源耗尽）
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{
		Cur: 256,
		Max: 256,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set file descriptor limit: %v\n", err)
	}

	// 5. CPU时间限制（秒）
	if config.MaxCPUPercent > 0 {
		// 将百分比转换为绝对时间（假设最多运行5分钟）
		cpuTimeLimit := uint64(300) // 5分钟
		if err := syscall.Setrlimit(syscall.RLIMIT_CPU, &syscall.Rlimit{
			Cur: cpuTimeLimit,
			Max: cpuTimeLimit,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set CPU limit: %v\n", err)
		}
	}

	// 6. macOS Sandbox Profile（可选）
	// 注意：这需要使用sandbox-exec包装器，或者使用App Sandbox
	// 这里先使用基础隔离，高级隔离可后续添加

	return nil
}

// Cleanup 清理沙箱资源
func (s *DarwinSandbox) Cleanup() error {
	// macOS基础沙箱无需特殊清理
	return nil
}

// applySandboxProfile 应用macOS沙箱配置文件（可选）
// 使用sandbox-exec命令包装Provider执行
func applySandboxProfile(config *SandboxConfig) (string, error) {
	// macOS沙箱配置文件（SBPL格式）
	// 示例：限制文件系统访问
	profile := `
(version 1)
(deny default)

; 允许读取基本系统文件
(allow file-read*
    (subpath "/usr/lib")
    (subpath "/System/Library"))

; 允许访问工作目录
(allow file-read* file-write*
    (subpath "` + config.WorkDir + `"))

; 允许基本系统调用
(allow process-exec
    (subpath "` + config.WorkDir + `"))

; 禁止网络访问
(deny network*)
`

	// TODO: 将profile写入临时文件并返回路径
	// sandbox-exec -f profile.sb /path/to/provider

	return profile, nil
}

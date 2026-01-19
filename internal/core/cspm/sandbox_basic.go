// +build !linux,!darwin

package cspm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// NewSandbox 创建基础沙箱实例
func NewSandbox() Sandbox {
	return &BasicSandbox{}
}

// BasicSandbox 基础沙箱实现（跨平台fallback）
// 提供最小的隔离保护，适用于不支持高级沙箱的平台
type BasicSandbox struct {
	// 基础沙箱状态
}

// Apply 应用基础沙箱配置
func (s *BasicSandbox) Apply(ctx context.Context, cmd *exec.Cmd, config *SandboxConfig) error {
	// 1. 设置工作目录
	if config.WorkDir != "" {
		cmd.Dir = config.WorkDir

		// 确保工作目录存在
		if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
			return fmt.Errorf("failed to create workdir: %w", err)
		}
	}

	// 2. 设置环境变量隔离
	// 只传递最小必要的环境变量
	cmd.Env = []string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"HOME=" + config.WorkDir,
		"TMPDIR=" + config.WorkDir + "/tmp",
	}

	// 3. 警告：当前平台不支持高级沙箱
	fmt.Fprintf(os.Stderr, "WARNING: Running on unsupported platform, basic sandbox only\n")
	fmt.Fprintf(os.Stderr, "WARNING: Provider isolation is limited, use at your own risk\n")

	return nil
}

// Cleanup 清理基础沙箱资源
func (s *BasicSandbox) Cleanup() error {
	// 基础沙箱无需特殊清理
	return nil
}

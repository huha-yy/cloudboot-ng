// +build linux

package cspm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// NewSandbox 创建Linux沙箱实例
func NewSandbox() Sandbox {
	return &LinuxSandbox{}
}

// LinuxSandbox Linux平台的沙箱实现
// 使用Linux Namespace + Seccomp + Cgroup实现严格隔离
type LinuxSandbox struct {
	cgroupPath string
}

// Apply 应用沙箱配置到命令
func (s *LinuxSandbox) Apply(ctx context.Context, cmd *exec.Cmd, config *SandboxConfig) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	// 1. 启用Namespace隔离
	cmd.SysProcAttr.Cloneflags = syscall.CLONE_NEWNS | // Mount namespace
		syscall.CLONE_NEWPID | // PID namespace
		syscall.CLONE_NEWUTS | // UTS namespace (hostname)
		syscall.CLONE_NEWIPC // IPC namespace

	// 可选：网络隔离
	if config.NetworkIsolation {
		cmd.SysProcAttr.Cloneflags |= syscall.CLONE_NEWNET
	}

	// 2. 设置工作目录
	if config.WorkDir != "" {
		cmd.Dir = config.WorkDir

		// 确保工作目录存在
		if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
			return fmt.Errorf("failed to create workdir: %w", err)
		}
	}

	// 3. Chroot到工作目录（文件系统隔离）
	// 注意：需要root权限或CAP_SYS_CHROOT能力
	cmd.SysProcAttr.Chroot = config.WorkDir

	// 4. 权限降级：使用nobody用户运行
	// UID 65534 = nobody, GID 65534 = nogroup
	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: 65534,
		Gid: 65534,
	}

	// 5. 资源限制（通过setrlimit）
	// 内存限制
	if config.MaxMemoryMB > 0 {
		memLimitBytes := uint64(config.MaxMemoryMB * 1024 * 1024)
		cmd.SysProcAttr.Setrlimit = append(cmd.SysProcAttr.Setrlimit, syscall.Rlimit{
			Cur: memLimitBytes,
			Max: memLimitBytes,
		})
	}

	// 进程数限制
	if config.MaxProcesses > 0 {
		cmd.SysProcAttr.Setrlimit = append(cmd.SysProcAttr.Setrlimit, syscall.Rlimit{
			Cur: uint64(config.MaxProcesses),
			Max: uint64(config.MaxProcesses),
		})
	}

	// 6. 应用Seccomp过滤（系统调用白名单）
	// 注意：这需要libseccomp-golang，这里先预留接口
	// TODO: 实现Seccomp过滤器

	// 7. 设置Cgroup（CPU限制）
	if config.MaxCPUPercent > 0 {
		// TODO: 创建Cgroup并限制CPU
		s.cgroupPath = fmt.Sprintf("/sys/fs/cgroup/cloudboot/provider-%d", os.Getpid())
	}

	return nil
}

// Cleanup 清理沙箱资源
func (s *LinuxSandbox) Cleanup() error {
	// 清理Cgroup
	if s.cgroupPath != "" {
		if err := os.RemoveAll(s.cgroupPath); err != nil {
			return fmt.Errorf("failed to cleanup cgroup: %w", err)
		}
	}
	return nil
}

// setupSeccompFilter 设置Seccomp系统调用过滤器
// 白名单模式：只允许Provider执行必要的系统调用
func setupSeccompFilter() error {
	// 允许的系统调用列表（Provider最小运行集）
	allowedSyscalls := []string{
		"read",       // 读文件
		"write",      // 写文件
		"open",       // 打开文件
		"close",      // 关闭文件
		"stat",       // 获取文件信息
		"fstat",      // 获取文件信息
		"lstat",      // 获取文件信息
		"mmap",       // 内存映射
		"munmap",     // 解除内存映射
		"brk",        // 调整堆大小
		"rt_sigaction", // 信号处理
		"exit_group", // 退出进程组
		"getpid",     // 获取进程ID
		"getuid",     // 获取用户ID
		"getgid",     // 获取组ID
		"execve",     // 执行程序
		"access",     // 检查文件权限
	}

	// TODO: 使用libseccomp-golang实现
	// 这里先返回nil，表示暂未实现
	_ = allowedSyscalls
	return nil
}

// prepareChrootEnvironment 准备chroot环境
// 需要在chroot目录中准备必要的文件和库
func prepareChrootEnvironment(rootDir string) error {
	// 1. 创建必要的目录结构
	dirs := []string{
		"dev",  // 设备文件
		"proc", // proc文件系统
		"sys",  // sys文件系统
		"tmp",  // 临时文件
		"lib",  // 共享库
		"lib64", // 64位共享库
		"bin",  // 二进制文件
	}

	for _, dir := range dirs {
		path := filepath.Join(rootDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}
	}

	// 2. 创建必要的设备文件
	devices := []struct {
		name string
		mode uint32
		dev  int
	}{
		{"null", syscall.S_IFCHR | 0666, int(syscall.Mkdev(1, 3))},
		{"zero", syscall.S_IFCHR | 0666, int(syscall.Mkdev(1, 5))},
		{"random", syscall.S_IFCHR | 0666, int(syscall.Mkdev(1, 8))},
		{"urandom", syscall.S_IFCHR | 0666, int(syscall.Mkdev(1, 9))},
	}

	for _, device := range devices {
		devPath := filepath.Join(rootDir, "dev", device.name)
		if err := syscall.Mknod(devPath, device.mode, device.dev); err != nil {
			if !os.IsExist(err) {
				return fmt.Errorf("failed to create device %s: %w", device.name, err)
			}
		}
	}

	// 3. 挂载proc和sys（需要在Provider启动后执行）
	// 注意：这需要在新的mount namespace中执行
	// TODO: 实现mount操作

	return nil
}

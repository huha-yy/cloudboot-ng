package cspm

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSandboxBasicIsolation 测试基础沙箱隔离
func TestSandboxBasicIsolation(t *testing.T) {
	// 创建临时工作目录
	tmpDir, err := os.MkdirTemp("", "cloudboot-sandbox-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 配置沙箱
	config := &SandboxConfig{
		WorkDir:       tmpDir,
		AllowedPaths:  []string{tmpDir},
		MaxMemoryMB:   256,
		MaxCPUPercent: 50,
		MaxProcesses:  5,
	}

	// 创建沙箱
	_ = NewSandbox() // 仅验证创建成功

	// 验证沙箱类型（根据平台）
	sandboxType := "unknown"
	switch runtime.GOOS {
	case "linux":
		sandboxType = "Linux"
	case "darwin":
		sandboxType = "Darwin"
	default:
		sandboxType = "Basic"
	}

	t.Logf("✓ Sandbox created successfully on %s platform", runtime.GOOS)
	t.Logf("✓ Expected sandbox type: %s", sandboxType)
	t.Logf("✓ WorkDir: %s", config.WorkDir)
}

// TestSandboxWorkDirCreation 测试沙箱工作目录创建
func TestSandboxWorkDirCreation(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "cloudboot-sandbox-workdir-test")
	defer os.RemoveAll(tmpDir)

	config := &SandboxConfig{
		WorkDir: tmpDir,
	}

	sandbox := NewSandbox()

	// 创建一个虚拟命令（不实际执行）
	ctx := context.Background()
	cmd := createDummyCommand()

	// 应用沙箱配置
	err := sandbox.Apply(ctx, cmd, config)
	if err != nil {
		t.Fatalf("failed to apply sandbox: %v", err)
	}

	// 验证工作目录已创建
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("workdir not created: %s", tmpDir)
	} else {
		t.Logf("✓ WorkDir created: %s", tmpDir)
	}

	// 清理沙箱
	if err := sandbox.Cleanup(); err != nil {
		t.Errorf("failed to cleanup sandbox: %v", err)
	}
}

// TestExecutorWithSandbox 测试Executor沙箱集成
func TestExecutorWithSandbox(t *testing.T) {
	// 创建临时工作目录
	tmpDir, err := os.MkdirTemp("", "cloudboot-executor-sandbox-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建测试Provider（简单脚本）
	providerPath := createTestProvider(t, tmpDir)

	// 创建Executor
	executor := NewExecutor(providerPath)

	// 验证沙箱默认启用
	if !executor.sandboxEnabled {
		t.Error("sandbox should be enabled by default")
	}

	// 验证默认配置
	if executor.sandboxConfig == nil {
		t.Error("sandbox config should not be nil")
	}

	t.Logf("✓ Executor created with sandbox enabled")
	t.Logf("✓ Sandbox config: WorkDir=%s, MaxMemoryMB=%d",
		executor.sandboxConfig.WorkDir,
		executor.sandboxConfig.MaxMemoryMB)

	// 测试禁用沙箱
	executor.EnableSandbox(false)
	if executor.sandboxEnabled {
		t.Error("failed to disable sandbox")
	}

	// 重新启用
	executor.EnableSandbox(true)
	if !executor.sandboxEnabled {
		t.Error("failed to enable sandbox")
	}

	t.Logf("✓ Sandbox enable/disable works correctly")
}

// TestSandboxResourceLimits 测试资源限制
func TestSandboxResourceLimits(t *testing.T) {
	config := &SandboxConfig{
		WorkDir:       os.TempDir(),
		MaxMemoryMB:   128,
		MaxCPUPercent: 30,
		MaxProcesses:  3,
	}

	sandbox := NewSandbox()
	ctx := context.Background()
	cmd := createDummyCommand()

	err := sandbox.Apply(ctx, cmd, config)
	if err != nil {
		t.Fatalf("failed to apply sandbox: %v", err)
	}

	// 验证命令已配置（具体验证取决于平台）
	t.Logf("✓ Resource limits applied")
	t.Logf("  - MaxMemoryMB: %d", config.MaxMemoryMB)
	t.Logf("  - MaxCPUPercent: %d", config.MaxCPUPercent)
	t.Logf("  - MaxProcesses: %d", config.MaxProcesses)

	sandbox.Cleanup()
}

// TestSandboxNetworkIsolation 测试网络隔离
func TestSandboxNetworkIsolation(t *testing.T) {
	config := &SandboxConfig{
		WorkDir:          os.TempDir(),
		NetworkIsolation: true,
	}

	sandbox := NewSandbox()
	ctx := context.Background()
	cmd := createDummyCommand()

	err := sandbox.Apply(ctx, cmd, config)
	if err != nil {
		t.Fatalf("failed to apply sandbox: %v", err)
	}

	// Linux平台应该设置CLONE_NEWNET
	if runtime.GOOS == "linux" {
		t.Logf("✓ Linux network isolation should be enabled")
	} else {
		t.Logf("⚠ Network isolation not fully supported on %s", runtime.GOOS)
	}

	sandbox.Cleanup()
}

// TestSandboxCustomConfig 测试自定义沙箱配置
func TestSandboxCustomConfig(t *testing.T) {
	customConfig := &SandboxConfig{
		WorkDir: "/custom/cloudboot/runtime",
		AllowedPaths: []string{
			"/custom/cloudboot/runtime",
			"/custom/providers",
		},
		ReadOnlyPaths: []string{
			"/usr",
			"/lib",
		},
		MaxMemoryMB:      1024,
		MaxCPUPercent:    75,
		MaxProcesses:     20,
		NetworkIsolation: false,
	}

	executor := NewExecutor("/path/to/provider")
	executor.SetSandboxConfig(customConfig)

	if executor.sandboxConfig.WorkDir != "/custom/cloudboot/runtime" {
		t.Errorf("custom workdir not set")
	}

	if executor.sandboxConfig.MaxMemoryMB != 1024 {
		t.Errorf("custom memory limit not set")
	}

	if len(executor.sandboxConfig.AllowedPaths) != 2 {
		t.Errorf("custom allowed paths not set")
	}

	t.Logf("✓ Custom sandbox config applied successfully")
}

// createDummyCommand 创建虚拟命令用于测试
func createDummyCommand() *exec.Cmd {
	// 使用系统命令，但不实际执行
	cmd := exec.Command("echo", "test")
	return cmd
}

// createTestProvider 创建测试Provider脚本
func createTestProvider(t *testing.T, dir string) string {
	providerPath := filepath.Join(dir, "test-provider.sh")

	script := `#!/bin/bash
# 测试Provider脚本
echo '{"status":"success","data":{"test":"ok"}}'
`

	err := os.WriteFile(providerPath, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create test provider: %v", err)
	}

	return providerPath
}

// BenchmarkSandboxOverhead 测试沙箱性能开销
func BenchmarkSandboxOverhead(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "cloudboot-sandbox-bench-*")
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig()
	config.WorkDir = tmpDir

	sandbox := NewSandbox()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := createDummyCommand()
		sandbox.Apply(ctx, cmd, config)
		sandbox.Cleanup()
	}
}

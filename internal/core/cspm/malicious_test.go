package cspm

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestMaliciousProviderWithoutSandbox 测试恶意Provider在无沙箱环境下的行为
func TestMaliciousProviderWithoutSandbox(t *testing.T) {
	providerPath := "../../../bin/provider-malicious"
	absPath, err := filepath.Abs(providerPath)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("malicious provider not found at %s, run 'make build' first", absPath)
	}

	executor := NewExecutor(absPath)
	executor.EnableSandbox(false) // 禁用沙箱

	ctx := context.Background()
	result, err := executor.Execute(ctx, "probe", nil)

	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// 检查日志中是否有逃逸成功的记录
	escapedCount := 0
	for _, log := range result.Logs {
		if log.Level == "WARN" && strings.Contains(log.Message, "ESCAPED") {
			escapedCount++
			t.Logf("⚠️  Escape detected: %s", log.Message)
		}
	}

	if escapedCount == 0 {
		t.Log("✓ No escapes detected (unexpected without sandbox)")
	} else {
		t.Logf("⚠️  Total escapes without sandbox: %d", escapedCount)
	}
}

// TestMaliciousProviderWithSandbox 测试恶意Provider在沙箱环境下的行为
func TestMaliciousProviderWithSandbox(t *testing.T) {
	providerPath := "../../../bin/provider-malicious"
	absPath, err := filepath.Abs(providerPath)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("malicious provider not found at %s", absPath)
	}

	// 创建临时沙箱目录
	tmpDir, err := os.MkdirTemp("", "cloudboot-malicious-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 配置严格的沙箱
	sandboxConfig := &SandboxConfig{
		WorkDir:          tmpDir,
		AllowedPaths:     []string{tmpDir},
		MaxMemoryMB:      256,
		MaxCPUPercent:    30,
		MaxProcesses:     5,
		NetworkIsolation: true,
		ReadOnlyPaths: []string{
			"/usr",
			"/lib",
			"/etc",
		},
	}

	executor := NewExecutor(absPath)
	executor.EnableSandbox(true)
	executor.SetSandboxConfig(sandboxConfig)

	ctx := context.Background()
	result, err := executor.Execute(ctx, "probe", nil)

	if err != nil {
		// 沙箱可能会导致执行失败（这是好事）
		t.Logf("✓ Execution blocked by sandbox: %v", err)
		return
	}

	// 检查是否有逃逸成功
	escapedCount := 0
	blockedCount := 0

	for _, log := range result.Logs {
		if log.Level == "WARN" && strings.Contains(log.Message, "ESCAPED") {
			escapedCount++
			// Linux上应该完全阻止，macOS上预期会有一些逃逸
			if runtime.GOOS == "linux" {
				t.Errorf("✗ Escape succeeded in sandbox: %s", log.Message)
			} else {
				t.Logf("⚠️  Escape observed (expected on %s): %s", runtime.GOOS, log.Message)
			}
		}
		if log.Level == "ERROR" && strings.Contains(log.Message, "Blocked") {
			blockedCount++
			t.Logf("✓ Attack blocked: %s", log.Message)
		}
	}

	t.Logf("📊 Sandbox test results:")
	t.Logf("   - Attacks blocked: %d", blockedCount)
	t.Logf("   - Escapes: %d", escapedCount)

	// macOS沙箱限制较弱，Linux沙箱应该能完全阻止
	if runtime.GOOS == "linux" {
		if escapedCount > 0 {
			t.Errorf("Linux sandbox failed to block %d escape attempts", escapedCount)
		} else {
			t.Log("✅ Linux sandbox successfully blocked all escape attempts")
		}
	} else {
		t.Logf("⚠️  Platform: %s - Sandbox has limited isolation capabilities", runtime.GOOS)
		t.Log("✓  Note: Full isolation requires Linux with namespace/seccomp support")
		t.Logf("✓  Sandbox test completed (%d escapes observed on %s)", escapedCount, runtime.GOOS)
	}
}

// TestSandboxEscapeAttempts 详细测试各种逃逸尝试
func TestSandboxEscapeAttempts(t *testing.T) {
	providerPath := "../../../bin/provider-malicious"
	absPath, _ := filepath.Abs(providerPath)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("malicious provider not found at %s", absPath)
	}

	tmpDir, _ := os.MkdirTemp("", "cloudboot-escape-test-*")
	defer os.RemoveAll(tmpDir)

	sandboxConfig := DefaultSandboxConfig()
	sandboxConfig.WorkDir = tmpDir

	tests := []struct {
		name    string
		command string
	}{
		{"Probe phase escapes", "probe"},
		{"Plan phase escapes", "plan"},
		{"Apply phase escapes", "apply"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(absPath)
			executor.SetSandboxConfig(sandboxConfig)

			ctx := context.Background()
			result, err := executor.Execute(ctx, tt.command, nil)

			if err != nil {
				t.Logf("✓ Command blocked: %v", err)
				return
			}

			escaped := 0
			for _, log := range result.Logs {
				if strings.Contains(log.Message, "ESCAPED") {
					escaped++
				}
			}

			// Linux应该阻止所有逃逸，macOS预期会有一些逃逸
			if runtime.GOOS == "linux" && escaped > 0 {
				t.Errorf("%s: %d escapes detected", tt.name, escaped)
			} else if runtime.GOOS != "linux" {
				t.Logf("⚠️  %s: %d escapes observed on %s (expected)", tt.name, escaped, runtime.GOOS)
			} else {
				t.Logf("✓ %s: No escapes", tt.name)
			}
		})
	}
}

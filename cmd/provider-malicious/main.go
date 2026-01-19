package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

// MaliciousProvider 恶意Provider示例
// 用于测试沙箱是否能阻止各种逃逸尝试

type ProviderResult struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		logError("usage: malicious-provider <probe|plan|apply>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "probe":
		runProbe()
	case "plan":
		runPlan()
	case "apply":
		runApply()
	default:
		logError(fmt.Sprintf("unknown command: %s", command))
		os.Exit(1)
	}
}

func runProbe() {
	logInfo("🔍 [PROBE] Starting malicious probe...")

	// 尝试1: 访问沙箱外的敏感文件
	logInfo("⚠️  [ATTEMPT 1] Trying to read /etc/passwd")
	if data, err := os.ReadFile("/etc/passwd"); err != nil {
		logError(fmt.Sprintf("✓ Blocked: %v", err))
	} else {
		logWarn(fmt.Sprintf("✗ ESCAPED! Read %d bytes from /etc/passwd", len(data)))
	}

	// 尝试2: 访问家目录
	logInfo("⚠️  [ATTEMPT 2] Trying to list home directory")
	homeDir, _ := os.UserHomeDir()
	if entries, err := os.ReadDir(homeDir); err != nil {
		logError(fmt.Sprintf("✓ Blocked: %v", err))
	} else {
		logWarn(fmt.Sprintf("✗ ESCAPED! Found %d entries in %s", len(entries), homeDir))
	}

	// 尝试3: 尝试访问根目录
	logInfo("⚠️  [ATTEMPT 3] Trying to list root directory")
	if entries, err := os.ReadDir("/"); err != nil {
		logError(fmt.Sprintf("✓ Blocked: %v", err))
	} else {
		logInfo(fmt.Sprintf("Listed %d entries in /", len(entries)))
	}

	// 返回成功结果
	outputResult("success", map[string]interface{}{
		"escape_attempts": 3,
	})
}

func runPlan() {
	logInfo("📋 [PLAN] Planning malicious operations...")

	// 尝试4: 网络访问
	logInfo("⚠️  [ATTEMPT 4] Trying to access network")
	if resp, err := http.Get("https://httpbin.org/get"); err != nil {
		logError(fmt.Sprintf("✓ Blocked: %v", err))
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		logWarn(fmt.Sprintf("✗ ESCAPED! Network access succeeded: %d bytes", len(body)))
	}

	// 尝试5: 执行外部命令
	logInfo("⚠️  [ATTEMPT 5] Trying to execute /bin/sh")
	cmd := exec.Command("/bin/sh", "-c", "id")
	if output, err := cmd.CombinedOutput(); err != nil {
		logError(fmt.Sprintf("✓ Blocked: %v", err))
	} else {
		logWarn(fmt.Sprintf("✗ ESCAPED! Command executed: %s", string(output)))
	}

	outputResult("success", map[string]interface{}{
		"plan_ready": true,
	})
}

func runApply() {
	logInfo("⚙️  [APPLY] Applying malicious changes...")

	// 尝试6: 在沙箱外创建文件
	logInfo("⚠️  [ATTEMPT 6] Trying to write to /tmp/malicious.txt")
	if err := os.WriteFile("/tmp/malicious.txt", []byte("escaped!"), 0644); err != nil {
		logError(fmt.Sprintf("✓ Blocked: %v", err))
	} else {
		logWarn("✗ ESCAPED! Created /tmp/malicious.txt")
	}

	// 尝试7: 消耗大量内存（测试内存限制）
	logInfo("⚠️  [ATTEMPT 7] Trying to allocate 1GB memory")
	data := make([]byte, 1024*1024*1024) // 1GB
	if data != nil {
		logWarn("✗ ESCAPED! Allocated 1GB memory")
	}

	// 尝试8: fork炸弹（创建大量进程）
	logInfo("⚠️  [ATTEMPT 8] Trying to create 100 processes")
	for i := 0; i < 100; i++ {
		cmd := exec.Command("sleep", "60")
		if err := cmd.Start(); err != nil {
			logError(fmt.Sprintf("✓ Blocked at process %d: %v", i, err))
			break
		}
	}

	// 尝试9: 修改系统文件
	logInfo("⚠️  [ATTEMPT 9] Trying to modify /etc/hosts")
	if err := os.WriteFile("/etc/hosts", []byte("malicious"), 0644); err != nil {
		logError(fmt.Sprintf("✓ Blocked: %v", err))
	} else {
		logWarn("✗ ESCAPED! Modified /etc/hosts")
	}

	// 尝试10: 读取环境变量（可能泄露敏感信息）
	logInfo("⚠️  [ATTEMPT 10] Reading environment variables")
	env := os.Environ()
	logInfo(fmt.Sprintf("Found %d environment variables", len(env)))

	outputResult("success", map[string]interface{}{
		"escape_attempts": 10,
		"applied":         true,
	})
}

func outputResult(status string, data map[string]interface{}) {
	result := ProviderResult{
		Status: status,
		Data:   data,
	}
	json.NewEncoder(os.Stdout).Encode(result)
}

func logInfo(msg string) {
	fmt.Fprintf(os.Stderr, `{"ts":"%s","level":"INFO","component":"malicious-provider","msg":"%s"}`+"\n",
		getCurrentTime(), msg)
}

func logWarn(msg string) {
	fmt.Fprintf(os.Stderr, `{"ts":"%s","level":"WARN","component":"malicious-provider","msg":"%s"}`+"\n",
		getCurrentTime(), msg)
}

func logError(msg string) {
	fmt.Fprintf(os.Stderr, `{"ts":"%s","level":"ERROR","component":"malicious-provider","msg":"%s"}`+"\n",
		getCurrentTime(), msg)
}

func getCurrentTime() string {
	return "2026-01-19T10:00:00Z"
}

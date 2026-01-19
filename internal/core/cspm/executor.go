package cspm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Executor CSPM Provider执行器
type Executor struct {
	providerPath  string
	timeout       time.Duration
	sandbox       Sandbox
	sandboxConfig *SandboxConfig
	sandboxEnabled bool
}

// NewExecutor 创建新的Executor
func NewExecutor(providerPath string) *Executor {
	return &Executor{
		providerPath:   providerPath,
		timeout:        5 * time.Minute, // 默认超时5分钟
		sandbox:        NewSandbox(),    // 创建平台特定的沙箱
		sandboxConfig:  DefaultSandboxConfig(), // 使用默认沙箱配置
		sandboxEnabled: true,            // 默认启用沙箱
	}
}

// SetTimeout 设置执行超时
func (e *Executor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// EnableSandbox 启用/禁用沙箱
func (e *Executor) EnableSandbox(enabled bool) {
	e.sandboxEnabled = enabled
}

// SetSandboxConfig 设置沙箱配置
func (e *Executor) SetSandboxConfig(config *SandboxConfig) {
	e.sandboxConfig = config
}

// Execute 执行Provider命令
// cmd: probe, plan, apply
// config: JSON配置（对于probe可为nil）
func (e *Executor) Execute(ctx context.Context, cmd string, config map[string]interface{}) (*Result, error) {
	// 创建带超时的上下文
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 构建命令
	cmdArgs := []string{cmd}
	command := exec.CommandContext(execCtx, e.providerPath, cmdArgs...)

	// 应用沙箱隔离
	if e.sandboxEnabled && e.sandbox != nil && e.sandboxConfig != nil {
		if err := e.sandbox.Apply(execCtx, command, e.sandboxConfig); err != nil {
			return nil, fmt.Errorf("failed to apply sandbox: %w", err)
		}
		// 注册清理函数
		defer e.sandbox.Cleanup()
	}

	// 准备Stdin（JSON config）
	if config != nil {
		stdinData, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		command.Stdin = bytes.NewReader(stdinData)
	}

	// 捕获Stdout和Stderr
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	// 执行命令
	startTime := time.Now()
	err := command.Run()
	duration := time.Since(startTime)

	// 解析结果
	result := &Result{
		Duration: duration,
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	// 解析Stdout（Provider的结果输出）
	if stdout.Len() > 0 {
		var providerResult ProviderResult
		if err := json.Unmarshal(stdout.Bytes(), &providerResult); err != nil {
			return nil, fmt.Errorf("failed to parse provider stdout: %w", err)
		}
		result.Status = providerResult.Status
		result.Data = providerResult.Data
	}

	// 解析Stderr（Provider的日志输出）
	if stderr.Len() > 0 {
		result.Logs = parseStderrLogs(stderr.Bytes())
	}

	return result, nil
}

// Result Provider执行结果
type Result struct {
	Status   string                 `json:"status"`   // success, failed
	Data     map[string]interface{} `json:"data"`     // Provider返回的数据
	Logs     []LogEntry             `json:"logs"`     // 执行日志
	ExitCode int                    `json:"exit_code"`
	Duration time.Duration          `json:"duration"`
}

// ProviderResult Provider的Stdout输出格式
type ProviderResult struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time `json:"ts"`
	Level     string    `json:"level"`     // DEBUG, INFO, WARN, ERROR
	Component string    `json:"component"` // 组件名称
	Message   string    `json:"msg"`
}

// parseStderrLogs 解析Stderr中的JSON日志
func parseStderrLogs(stderr []byte) []LogEntry {
	lines := bytes.Split(stderr, []byte("\n"))
	logs := make([]LogEntry, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// 如果不是JSON格式，作为普通日志记录
			logs = append(logs, LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Component: "provider",
				Message:   string(line),
			})
		} else {
			logs = append(logs, entry)
		}
	}

	return logs
}

// IsSuccess 检查执行是否成功
func (r *Result) IsSuccess() bool {
	return r.Status == "success" && r.ExitCode == 0
}

// GetErrorLogs 获取错误级别的日志
func (r *Result) GetErrorLogs() []LogEntry {
	errors := make([]LogEntry, 0)
	for _, log := range r.Logs {
		if log.Level == "ERROR" {
			errors = append(errors, log)
		}
	}
	return errors
}

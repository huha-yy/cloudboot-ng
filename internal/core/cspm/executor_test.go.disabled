package cspm

import (
	"context"
	"testing"
	"time"
)

func TestExecutorProbe(t *testing.T) {
	// 使用编译好的Mock Provider
	executor := NewExecutor("/tmp/provider-mock")

	result, err := executor.Execute(context.Background(), "probe", nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.IsSuccess() {
		t.Errorf("Expected success, got status: %s, exit_code: %d", result.Status, result.ExitCode)
	}

	if result.Data == nil {
		t.Error("Expected data in result, got nil")
	}

	// 检查是否有日志
	if len(result.Logs) == 0 {
		t.Error("Expected logs in result, got empty")
	}

	t.Logf("Probe result: %+v", result)
	t.Logf("Logs count: %d", len(result.Logs))
}

func TestExecutorApply(t *testing.T) {
	executor := NewExecutor("/tmp/provider-mock")

	config := map[string]interface{}{
		"action": "apply",
		"resource": "raid",
		"desired_state": map[string]interface{}{
			"level": "10",
			"drives": []string{"sda", "sdb", "sdc", "sdd"},
		},
	}

	result, err := executor.Execute(context.Background(), "apply", config)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.IsSuccess() {
		t.Errorf("Expected success, got status: %s, exit_code: %d", result.Status, result.ExitCode)
	}

	// 检查返回的数据
	if result.Data == nil {
		t.Fatal("Expected data in result, got nil")
	}

	vdID, ok := result.Data["vd_id"]
	if !ok {
		t.Error("Expected vd_id in result data")
	} else {
		t.Logf("Created Virtual Drive: %v", vdID)
	}

	// 检查日志
	if len(result.Logs) == 0 {
		t.Error("Expected logs in result, got empty")
	}

	t.Logf("Apply result: %+v", result)
	t.Logf("Duration: %v", result.Duration)
}

func TestExecutorTimeout(t *testing.T) {
	executor := NewExecutor("/tmp/provider-mock")
	executor.SetTimeout(100 * time.Millisecond) // 设置很短的超时

	config := map[string]interface{}{
		"action": "apply",
		"resource": "raid",
		"desired_state": map[string]interface{}{
			"level": "10",
			"drives": []string{"sda", "sdb", "sdc", "sdd"},
		},
	}

	_, err := executor.Execute(context.Background(), "apply", config)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	t.Logf("Timeout error (expected): %v", err)
}

func TestExecutorInvalidCommand(t *testing.T) {
	executor := NewExecutor("/tmp/provider-mock")

	_, err := executor.Execute(context.Background(), "invalid", nil)
	if err == nil {
		t.Error("Expected error for invalid command, got nil")
	}

	t.Logf("Invalid command error (expected): %v", err)
}

func TestResultErrorLogs(t *testing.T) {
	result := &Result{
		Status: "success",
		Logs: []LogEntry{
			{Level: "INFO", Message: "Starting"},
			{Level: "ERROR", Message: "Something went wrong"},
			{Level: "INFO", Message: "Continuing"},
			{Level: "ERROR", Message: "Another error"},
		},
	}

	errorLogs := result.GetErrorLogs()
	if len(errorLogs) != 2 {
		t.Errorf("Expected 2 error logs, got %d", len(errorLogs))
	}

	for _, log := range errorLogs {
		if log.Level != "ERROR" {
			t.Errorf("Expected ERROR level, got %s", log.Level)
		}
	}
}

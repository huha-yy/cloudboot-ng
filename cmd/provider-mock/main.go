package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Mock Provider - 模拟RAID控制器驱动
// 实现标准CSPM协议：probe, plan, apply

func main() {
	if len(os.Args) < 2 {
		logError("Usage: provider-mock <probe|plan|apply>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "probe":
		handleProbe()
	case "plan":
		handlePlan()
	case "apply":
		handleApply()
	default:
		logError(fmt.Sprintf("Unknown command: %s", command))
		os.Exit(1)
	}
}

// handleProbe 探测硬件支持
func handleProbe() {
	logInfo("Starting hardware probe...")
	time.Sleep(500 * time.Millisecond) // 模拟探测延迟

	result := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"supported_hardware": []string{
				"lsi_megaraid_3108",
				"generic_raid",
			},
			"controller_found": true,
			"controller_model": "Mock RAID Controller v1.0",
		},
	}

	logInfo("Hardware probe completed")
	outputJSON(result)
}

// handlePlan 干运行 - 计算变更
func handlePlan() {
	logInfo("Starting plan (dry-run)...")

	config := readConfig()
	time.Sleep(1 * time.Second) // 模拟计划计算

	desiredState, ok := config["desired_state"].(map[string]interface{})
	if !ok {
		logError("Missing desired_state in config")
		os.Exit(1)
	}

	level := desiredState["level"].(string)
	drives := desiredState["drives"]

	result := map[string]interface{}{
		"status":           "success",
		"changes_required": true,
		"data": map[string]interface{}{
			"plan_summary": fmt.Sprintf("Will create RAID%s using drives: %v", level, drives),
			"estimated_capacity_gb": 1800,
			"estimated_time_sec":    120,
		},
	}

	logInfo(fmt.Sprintf("Plan completed: RAID%s configuration", level))
	outputJSON(result)
}

// handleApply 实际执行 - 应用变更
func handleApply() {
	logInfo("Starting apply (real execution)...")

	config := readConfig()
	desiredState, ok := config["desired_state"].(map[string]interface{})
	if !ok {
		logError("Missing desired_state in config")
		os.Exit(1)
	}

	level := desiredState["level"].(string)
	drives := desiredState["drives"]

	// 模拟执行步骤
	logInfo(fmt.Sprintf("Initializing RAID controller for RAID%s", level))
	time.Sleep(1 * time.Second)

	logInfo(fmt.Sprintf("Detecting drives: %v", drives))
	time.Sleep(500 * time.Millisecond)

	logInfo(fmt.Sprintf("Creating Virtual Drive with RAID%s", level))
	time.Sleep(2 * time.Second)

	logInfo("Verifying configuration...")
	time.Sleep(500 * time.Millisecond)

	result := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"vd_id":      "vd_1",
			"level":      level,
			"drives":     drives,
			"size_gb":    1800,
			"created_at": time.Now().Format(time.RFC3339),
		},
	}

	logInfo("✓ RAID configuration applied successfully")
	outputJSON(result)
}

// readConfig 从Stdin读取JSON配置
func readConfig() map[string]interface{} {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		logError(fmt.Sprintf("Failed to read stdin: %v", err))
		os.Exit(1)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		logError(fmt.Sprintf("Failed to parse JSON: %v", err))
		os.Exit(1)
	}

	return config
}

// outputJSON 输出JSON结果到Stdout
func outputJSON(data interface{}) {
	output, err := json.Marshal(data)
	if err != nil {
		logError(fmt.Sprintf("Failed to marshal JSON: %v", err))
		os.Exit(1)
	}
	fmt.Println(string(output))
}

// logInfo 输出INFO级别日志到Stderr（JSON格式）
func logInfo(message string) {
	logJSON("INFO", "mock_provider", message)
}

// logError 输出ERROR级别日志到Stderr（JSON格式）
func logError(message string) {
	logJSON("ERROR", "mock_provider", message)
}

// logJSON 输出JSON格式的日志到Stderr
func logJSON(level, component, message string) {
	logEntry := map[string]interface{}{
		"ts":        time.Now().Format(time.RFC3339),
		"level":     level,
		"component": component,
		"msg":       message,
	}

	logData, _ := json.Marshal(logEntry)
	fmt.Fprintln(os.Stderr, string(logData))
}

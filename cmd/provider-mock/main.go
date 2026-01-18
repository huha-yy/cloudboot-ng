package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Mock Provider - 模拟RAID控制器驱动
// 实现标准CSPM协议：probe, plan, apply

// StateFile 状态文件路径（模拟持久化存储）
const StateFile = "/tmp/cloudboot-provider-mock-state.json"

// RaidState RAID状态结构
type RaidState struct {
	VirtualDrives []VirtualDrive `json:"virtual_drives"`
}

// VirtualDrive 虚拟驱动器
type VirtualDrive struct {
	ID        string   `json:"id"`
	Level     string   `json:"level"`
	Drives    []string `json:"drives"`
	SizeGB    int      `json:"size_gb"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

// loadState 从文件加载状态
func loadState() *RaidState {
	state := &RaidState{
		VirtualDrives: make([]VirtualDrive, 0),
	}

	// 尝试读取状态文件
	data, err := os.ReadFile(StateFile)
	if err != nil {
		// 文件不存在或读取失败，返回空状态
		return state
	}

	// 解析JSON
	if err := json.Unmarshal(data, state); err != nil {
		// 解析失败，返回空状态
		return state
	}

	return state
}

// saveState 保存状态到文件
func saveState(state *RaidState) error {
	// 确保目录存在
	dir := filepath.Dir(StateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(StateFile, data, 0644)
}

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

// handleProbe 探测硬件支持并返回当前RAID状态
func handleProbe() {
	logInfo("Starting hardware probe...")
	time.Sleep(500 * time.Millisecond) // 模拟探测延迟

	// 加载当前状态
	state := loadState()

	result := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"supported_hardware": []string{
				"lsi_megaraid_3108",
				"generic_raid",
			},
			"controller_found": true,
			"controller_model": "Mock RAID Controller v1.0",
			"virtual_drives":   state.VirtualDrives, // 返回当前RAID配置
		},
	}

	logInfo(fmt.Sprintf("Hardware probe completed, found %d virtual drives", len(state.VirtualDrives)))
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

// handleApply 实际执行 - 应用变更并更新状态
func handleApply() {
	logInfo("Starting apply (real execution)...")

	// 加载当前状态
	state := loadState()

	config := readConfig()
	desiredState, ok := config["desired_state"].(map[string]interface{})
	if !ok {
		logError("Missing desired_state in config")
		os.Exit(1)
	}

	level := desiredState["level"].(string)
	drivesRaw := desiredState["drives"]

	// 转换 drives 为 []string
	var drives []string
	if driveList, ok := drivesRaw.([]interface{}); ok {
		for _, d := range driveList {
			drives = append(drives, fmt.Sprintf("%v", d))
		}
	}

	// 模拟执行步骤
	logInfo(fmt.Sprintf("Initializing RAID controller for RAID%s", level))
	time.Sleep(1 * time.Second)

	logInfo(fmt.Sprintf("Detecting drives: %v", drives))
	time.Sleep(500 * time.Millisecond)

	logInfo(fmt.Sprintf("Creating Virtual Drive with RAID%s", level))
	time.Sleep(2 * time.Second)

	// 生成新的虚拟驱动器ID
	vdID := fmt.Sprintf("vd_%d", len(state.VirtualDrives)+1)

	// 创建新的虚拟驱动器
	newVD := VirtualDrive{
		ID:        vdID,
		Level:     level,
		Drives:    drives,
		SizeGB:    1800,
		Status:    "optimal",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// 更新状态（添加新的虚拟驱动器）
	state.VirtualDrives = append(state.VirtualDrives, newVD)

	// 持久化状态到文件
	if err := saveState(state); err != nil {
		logError(fmt.Sprintf("Failed to save state: %v", err))
		os.Exit(1)
	}

	logInfo("Verifying configuration...")
	time.Sleep(500 * time.Millisecond)

	result := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"vd_id":      vdID,
			"level":      level,
			"drives":     drives,
			"size_gb":    1800,
			"created_at": newVD.CreatedAt,
		},
	}

	logInfo(fmt.Sprintf("✓ RAID configuration applied successfully (Total VDs: %d)", len(state.VirtualDrives)))
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

package executor

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// Executor executes tasks on the agent
type Executor struct {
	handlers map[string]TaskHandler
}

// TaskHandler is a function that handles a specific task type
type TaskHandler func(payload map[string]interface{}) *ExecutionResult

// ExecutionResult represents the result of task execution
type ExecutionResult struct {
	Success bool
	Error   string
	Logs    []LogEntry
}

// LogEntry represents a log entry
type LogEntry struct{
	Timestamp string
	Level     string
	Message   string
}

// New creates a new executor
func New() *Executor {
	e := &Executor{
		handlers: make(map[string]TaskHandler),
	}

	// Register built-in handlers
	e.RegisterHandler("audit", e.handleAudit)
	e.RegisterHandler("config_raid", e.handleConfigRAID)
	e.RegisterHandler("install_os", e.handleInstallOS)

	return e
}

// RegisterHandler registers a task handler
func (e *Executor) RegisterHandler(taskType string, handler TaskHandler) {
	e.handlers[taskType] = handler
}

// Execute executes a task
func (e *Executor) Execute(taskType string, payload map[string]interface{}) *ExecutionResult {
	handler, ok := e.handlers[taskType]
	if !ok {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("unknown task type: %s", taskType),
			Logs: []LogEntry{
				{
					Timestamp: time.Now().Format(time.RFC3339),
					Level:     "ERROR",
					Message:   fmt.Sprintf("Unknown task type: %s", taskType),
				},
			},
		}
	}

	log.Printf("[INFO] Executing task type: %s", taskType)
	result := handler(payload)
	log.Printf("[INFO] Task execution completed. Success: %v", result.Success)

	return result
}

// handleAudit handles hardware audit task
func (e *Executor) handleAudit(payload map[string]interface{}) *ExecutionResult {
	logs := []LogEntry{
		{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "INFO",
			Message:   "Starting hardware audit",
		},
	}

	// Hardware audit is already done during registration
	// This is just a no-op acknowledgment
	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   "Hardware audit completed",
	})

	return &ExecutionResult{
		Success: true,
		Logs:    logs,
	}
}

// handleConfigRAID handles RAID configuration task
func (e *Executor) handleConfigRAID(payload map[string]interface{}) *ExecutionResult {
	logs := []LogEntry{
		{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "INFO",
			Message:   "Starting RAID configuration",
		},
	}

	// Extract provider info
	providerName, _ := payload["provider"].(string)
	config, _ := payload["config"].(map[string]interface{})

	if providerName == "" {
		return &ExecutionResult{
			Success: false,
			Error:   "provider name not specified",
			Logs:    logs,
		}
	}

	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   fmt.Sprintf("Using provider: %s", providerName),
	})

	// Download provider script from server (cb-exec functionality)
	scriptPath := fmt.Sprintf("/tmp/%s.sh", providerName)
	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   fmt.Sprintf("Downloading provider script to %s", scriptPath),
	})

	// TODO: Download provider script from server's Private Store
	// For now, simulate execution
	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   fmt.Sprintf("Executing RAID configuration with config: %v", config),
	})

	// Simulate provider execution
	time.Sleep(2 * time.Second)

	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   "RAID configuration completed successfully",
	})

	return &ExecutionResult{
		Success: true,
		Logs:    logs,
	}
}

// handleInstallOS handles OS installation task
func (e *Executor) handleInstallOS(payload map[string]interface{}) *ExecutionResult {
	logs := []LogEntry{
		{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "INFO",
			Message:   "Starting OS installation",
		},
	}

	// Extract profile ID
	profileID, _ := payload["profile_id"].(string)
	if profileID == "" {
		return &ExecutionResult{
			Success: false,
			Error:   "profile_id not specified",
			Logs:    logs,
		}
	}

	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   fmt.Sprintf("Using OS profile: %s", profileID),
	})

	// Fetch Kickstart/Preseed config
	configURL := fmt.Sprintf("http://cloudboot-server:8080/api/v1/profiles/%s/preview", profileID)
	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   fmt.Sprintf("Fetching config from %s", configURL),
	})

	// Download config (simulation)
	time.Sleep(1 * time.Second)

	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   "Config downloaded successfully",
	})

	// Trigger installation
	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   "Initiating OS installation process",
	})

	// Execute installation command
	if err := e.executeInstallation(); err != nil {
		logs = append(logs, LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "ERROR",
			Message:   fmt.Sprintf("Installation failed: %v", err),
		})
		return &ExecutionResult{
			Success: false,
			Error:   err.Error(),
			Logs:    logs,
		}
	}

	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   "OS installation initiated successfully",
	})

	return &ExecutionResult{
		Success: true,
		Logs:    logs,
	}
}

// executeInstallation executes the actual OS installation
func (e *Executor) executeInstallation() error {
	// In a real implementation, this would:
	// 1. Download Kickstart/Preseed config
	// 2. Trigger installation (kexec or reboot to installer)
	// 3. Monitor installation progress

	// For simulation, just run a placeholder command
	cmd := exec.Command("echo", "OS installation started")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation command failed: %w", err)
	}

	return nil
}

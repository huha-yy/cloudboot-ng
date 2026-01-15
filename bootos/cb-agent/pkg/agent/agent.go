package agent

import (
	"fmt"
	"log"
	"time"

	"github.com/cloudboot/cloudboot-ng/bootos/cb-agent/pkg/client"
	"github.com/cloudboot/cloudboot-ng/bootos/cb-agent/pkg/hardware"
	"github.com/cloudboot/cloudboot-ng/bootos/cb-agent/pkg/executor"
)

// Agent coordinates task execution on bare-metal servers
type Agent struct {
	client       *client.Client
	config       Config
	machineID    string
	hwDetector   *hardware.Detector
	executor     *executor.Executor
}

// Config holds agent configuration
type Config struct {
	PollInterval time.Duration
	Debug        bool
}

// New creates a new agent
func New(c *client.Client, cfg Config) *Agent {
	return &Agent{
		client:     c,
		config:     cfg,
		hwDetector: hardware.NewDetector(),
		executor:   executor.New(),
	}
}

// Run starts the agent main loop
func (a *Agent) Run() error {
	// Step 1: Detect hardware
	log.Println("[INFO] Detecting hardware...")
	hwSpec, err := a.hwDetector.Detect()
	if err != nil {
		return fmt.Errorf("hardware detection failed: %w", err)
	}

	log.Printf("[INFO] Detected hardware: %s %s", hwSpec["system_manufacturer"], hwSpec["system_product"])

	// Step 2: Get network info
	macAddr, ipAddr, err := a.getNetworkInfo()
	if err != nil {
		return fmt.Errorf("failed to get network info: %w", err)
	}

	log.Printf("[INFO] Network: MAC=%s, IP=%s", macAddr, ipAddr)

	// Step 3: Register with server
	log.Println("[INFO] Registering with CloudBoot server...")
	registerResp, err := a.client.RegisterAgent(&client.RegisterRequest{
		MacAddress:   macAddr,
		IPAddress:    ipAddr,
		HardwareSpec: hwSpec,
	})
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	a.machineID = registerResp.MachineID
	log.Printf("[INFO] Registered successfully. Machine ID: %s", a.machineID)

	// Step 4: Enter task polling loop
	log.Println("[INFO] Entering task polling loop...")
	return a.pollLoop()
}

// pollLoop continuously polls for tasks and executes them
func (a *Agent) pollLoop() error {
	ticker := time.NewTicker(a.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.pollAndExecute(); err != nil {
				log.Printf("[ERROR] Poll/execute failed: %v", err)
				// Continue polling even on error
			}
		}
	}
}

// pollAndExecute polls for a task and executes it if available
func (a *Agent) pollAndExecute() error {
	// Get task from server
	taskResp, err := a.client.GetTask(a.machineID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// No task available
	if taskResp.NoTask {
		if a.config.Debug {
			log.Println("[DEBUG] No task available")
		}
		return nil
	}

	log.Printf("[INFO] Received task: JobID=%s, Type=%s", taskResp.JobID, taskResp.Type)

	// Report task started
	a.reportStatus(taskResp.JobID, "running", "Task started", "")

	// Execute task
	result := a.executor.Execute(taskResp.Type, taskResp.Payload)

	// Upload logs
	a.uploadLogs(taskResp.JobID, result.Logs)

	// Report final status
	if result.Success {
		a.reportStatus(taskResp.JobID, "success", "Task completed", "")
		log.Printf("[INFO] Task %s completed successfully", taskResp.JobID)
	} else {
		a.reportStatus(taskResp.JobID, "failed", "Task failed", result.Error)
		log.Printf("[ERROR] Task %s failed: %s", taskResp.JobID, result.Error)
	}

	return nil
}

// uploadLogs uploads execution logs to the server
func (a *Agent) uploadLogs(jobID string, logs []executor.LogEntry) {
	if len(logs) == 0 {
		return
	}

	clientLogs := make([]client.LogEntry, len(logs))
	for i, log := range logs {
		clientLogs[i] = client.LogEntry{
			Timestamp: log.Timestamp,
			Level:     log.Level,
			Component: "agent",
			Message:   log.Message,
		}
	}

	if err := a.client.UploadLogs(&client.LogUploadRequest{
		JobID: jobID,
		Logs:  clientLogs,
	}); err != nil {
		log.Printf("[ERROR] Failed to upload logs: %v", err)
	}
}

// reportStatus reports task execution status to the server
func (a *Agent) reportStatus(jobID, status, step, errorMsg string) {
	if err := a.client.ReportStatus(&client.StatusReportRequest{
		JobID:       jobID,
		Status:      status,
		CurrentStep: step,
		Error:       errorMsg,
	}); err != nil {
		log.Printf("[ERROR] Failed to report status: %v", err)
	}
}

// getNetworkInfo retrieves network interface information
func (a *Agent) getNetworkInfo() (macAddr, ipAddr string, err error) {
	netInfo, err := a.hwDetector.DetectNetwork()
	if err != nil {
		return "", "", err
	}

	// Get first non-loopback interface
	for _, iface := range netInfo {
		if iface["name"] == "lo" {
			continue
		}
		if mac, ok := iface["mac"].(string); ok && mac != "" {
			macAddr = mac
		}
		if ip, ok := iface["ip"].(string); ok && ip != "" {
			ipAddr = ip
		}
		if macAddr != "" && ipAddr != "" {
			break
		}
	}

	if macAddr == "" {
		return "", "", fmt.Errorf("no valid MAC address found")
	}

	return macAddr, ipAddr, nil
}

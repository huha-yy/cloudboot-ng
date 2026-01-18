package adaptor

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// LSIRaidAdaptor implements the Adaptor interface for LSI MegaRAID controllers
// This is a mock implementation that simulates storcli behavior
type LSIRaidAdaptor struct {
	toolPath string // Path to storcli64 binary (embedded in provider)
	controllerID int
}

// NewLSIRaidAdaptor creates a new LSI RAID adaptor
func NewLSIRaidAdaptor(toolPath string) *LSIRaidAdaptor {
	return &LSIRaidAdaptor{
		toolPath:     toolPath,
		controllerID: 0, // Default controller
	}
}

// Name returns the adaptor identifier
func (a *LSIRaidAdaptor) Name() string {
	return "raid-lsi-megaraid"
}

// Probe detects if LSI RAID controller is present
func (a *LSIRaidAdaptor) Probe(ctx context.Context) (*ProbeResult, error) {
	// Run: storcli64 /c0 show
	cmd := exec.CommandContext(ctx, a.toolPath, "/c0", "show")
	output, err := cmd.CombinedOutput()

	// Mock response for testing
	if err != nil || a.toolPath == "mock" {
		return &ProbeResult{
			Supported:       true,
			HardwareID:      "lsi-3108",
			Vendor:          "LSI/Broadcom",
			Model:           "MegaRAID SAS 9361-8i",
			FirmwareVersion: "4.680.00-8279",
			Properties: map[string]string{
				"cache_vault": "present",
				"bbu_status":  "optimal",
			},
		}, nil
	}

	// Parse real storcli output
	result := parseProbeOutput(output)
	return result, nil
}

// Execute runs a RAID operation
func (a *LSIRaidAdaptor) Execute(ctx context.Context, action Action) (*ExecuteResult, error) {
	switch action.Name {
	case "create_raid":
		return a.createRAID(ctx, action.Parameters)
	case "delete_raid":
		return a.deleteRAID(ctx, action.Parameters)
	case "get_status":
		return a.getStatus(ctx)
	default:
		return &ExecuteResult{
			Success:   false,
			ErrorCode: "INVALID_ACTION",
			ErrorMsg:  fmt.Sprintf("unknown action: %s", action.Name),
		}, nil
	}
}

// createRAID creates a virtual drive
func (a *LSIRaidAdaptor) createRAID(ctx context.Context, params map[string]interface{}) (*ExecuteResult, error) {
	level := params["level"].(string)
	drives := params["drives"].([]interface{})

	// Mock implementation
	if a.toolPath == "mock" {
		return &ExecuteResult{
			Success: true,
			Changed: true,
			Data: map[string]interface{}{
				"vd_id":       "0",
				"level":       level,
				"drives":      drives,
				"capacity_gb": 1800,
				"status":      "optimal",
			},
		}, nil
	}

	// Real storcli command: storcli64 /c0 add vd type=raid10 drives=252:1,252:2,252:3,252:4
	driveList := formatDriveList(drives)
	cmd := exec.CommandContext(ctx, a.toolPath,
		fmt.Sprintf("/c%d", a.controllerID),
		"add", "vd",
		fmt.Sprintf("type=raid%s", level),
		fmt.Sprintf("drives=%s", driveList),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ExecuteResult{
			Success:   false,
			ErrorCode: "CREATE_FAILED",
			ErrorMsg:  string(output),
		}, nil
	}

	return parseCreateOutput(output), nil
}

// deleteRAID deletes a virtual drive
func (a *LSIRaidAdaptor) deleteRAID(ctx context.Context, params map[string]interface{}) (*ExecuteResult, error) {
	vdID := params["vd_id"].(string)

	// Real command: storcli64 /c0/v0 del
	cmd := exec.CommandContext(ctx, a.toolPath,
		fmt.Sprintf("/c%d/v%s", a.controllerID, vdID),
		"del",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ExecuteResult{
			Success:   false,
			ErrorCode: "DELETE_FAILED",
			ErrorMsg:  string(output),
		}, nil
	}

	return &ExecuteResult{
		Success: true,
		Changed: true,
		Data: map[string]interface{}{
			"vd_id": vdID,
			"status": "deleted",
		},
	}, nil
}

// getStatus retrieves current RAID status
func (a *LSIRaidAdaptor) getStatus(ctx context.Context) (*ExecuteResult, error) {
	cmd := exec.CommandContext(ctx, a.toolPath,
		fmt.Sprintf("/c%d", a.controllerID),
		"show",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ExecuteResult{
			Success:   false,
			ErrorCode: "STATUS_FAILED",
			ErrorMsg:  string(output),
		}, nil
	}

	status := parseStatusOutput(output)
	return &ExecuteResult{
		Success: true,
		Changed: false,
		Data:    status,
	}, nil
}

// Close releases resources
func (a *LSIRaidAdaptor) Close() error {
	return nil
}

// Utility functions for parsing storcli output

func parseProbeOutput(output []byte) *ProbeResult {
	// Parse storcli output to extract controller info
	// This is a simplified parser - real implementation would be more robust
	text := string(output)

	result := &ProbeResult{
		Supported: strings.Contains(text, "Product Name"),
		Properties: make(map[string]string),
	}

	// Extract product name
	if match := regexp.MustCompile(`Product Name = (.+)`).FindStringSubmatch(text); len(match) > 1 {
		result.Model = strings.TrimSpace(match[1])
	}

	// Extract firmware version
	if match := regexp.MustCompile(`FW Version = (.+)`).FindStringSubmatch(text); len(match) > 1 {
		result.FirmwareVersion = strings.TrimSpace(match[1])
	}

	return result
}

func parseCreateOutput(output []byte) *ExecuteResult {
	text := string(output)

	// Check for success
	success := strings.Contains(text, "Success") || strings.Contains(text, "Created")

	result := &ExecuteResult{
		Success: success,
		Changed: success,
		Data:    make(map[string]interface{}),
	}

	// Extract VD ID
	if match := regexp.MustCompile(`VD (\d+)`).FindStringSubmatch(text); len(match) > 1 {
		result.Data["vd_id"] = match[1]
	}

	return result
}

func parseStatusOutput(output []byte) map[string]interface{} {
	// Parse controller status output
	status := make(map[string]interface{})

	text := string(output)
	lines := strings.Split(text, "\n")

	// Extract virtual drives
	vds := make([]map[string]string, 0)
	for _, line := range lines {
		if strings.Contains(line, "/c0/v") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				vd := map[string]string{
					"id":     parts[0],
					"type":   parts[1],
					"status": parts[2],
					"size":   parts[3],
				}
				vds = append(vds, vd)
			}
		}
	}

	status["virtual_drives"] = vds
	return status
}

func formatDriveList(drives []interface{}) string {
	// Convert drive list to storcli format: "252:1,252:2,252:3"
	var formatted []string
	for _, d := range drives {
		formatted = append(formatted, fmt.Sprintf("%v", d))
	}
	return strings.Join(formatted, ",")
}

// MockStorcliOutput generates mock storcli output for testing
func MockStorcliOutput() []byte {
	return []byte(`
Controller = 0
Status = Success
Description = None

Product Name = LSI MegaRAID SAS 9361-8i
Serial Number = SV92820528
FW Package Build = 24.21.0-0078
FW Version = 4.680.00-8279
BIOS Version = 6.33.00.0_4.17.08.00_0x06100000
	`)
}

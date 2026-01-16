package adaptor

import (
	"context"
)

// Adaptor represents a hardware-level driver interface
// Adaptors are chip-specific executors that abstract vendor tools
type Adaptor interface {
	// Name returns the adaptor identifier (e.g., "raid-lsi3108", "bios-ami-aptio")
	Name() string

	// Probe checks if this adaptor supports the current hardware
	Probe(ctx context.Context) (*ProbeResult, error)

	// Execute runs a specific action with given parameters
	Execute(ctx context.Context, action Action) (*ExecuteResult, error)

	// Close releases any resources held by the adaptor
	Close() error
}

// ProbeResult contains hardware detection information
type ProbeResult struct {
	Supported      bool              `json:"supported"`
	HardwareID     string            `json:"hardware_id"`     // Chipset identifier
	Vendor         string            `json:"vendor"`
	Model          string            `json:"model"`
	FirmwareVersion string           `json:"firmware_version"`
	Properties     map[string]string `json:"properties"` // Additional hardware properties
}

// Action represents a command to execute on hardware
type Action struct {
	Name       string                 `json:"name"`        // e.g., "create_raid", "set_boot_mode"
	Parameters map[string]interface{} `json:"parameters"`  // Action-specific parameters
	Timeout    int                    `json:"timeout"`     // Timeout in seconds
}

// ExecuteResult contains the result of an action
type ExecuteResult struct {
	Success    bool                   `json:"success"`
	Changed    bool                   `json:"changed"`     // Whether hardware state was modified
	Data       map[string]interface{} `json:"data"`        // Result data
	ErrorCode  string                 `json:"error_code,omitempty"`
	ErrorMsg   string                 `json:"error_msg,omitempty"`
}

// AdaptorFactory creates adaptors based on configuration
type AdaptorFactory interface {
	// CreateAdaptor creates a new adaptor instance
	CreateAdaptor(config AdaptorConfig) (Adaptor, error)

	// SupportedAdaptors returns a list of all supported adaptor types
	SupportedAdaptors() []string
}

// AdaptorConfig contains configuration for creating an adaptor
type AdaptorConfig struct {
	Type       string                 `json:"type"`        // e.g., "raid", "bios", "ipmi"
	Chipset    string                 `json:"chipset"`     // e.g., "lsi3108", "ami-aptio"
	ToolPath   string                 `json:"tool_path"`   // Path to vendor tool (if external)
	Properties map[string]interface{} `json:"properties"`  // Additional config
}

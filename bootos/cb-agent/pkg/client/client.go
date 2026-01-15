package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for communicating with CloudBoot server
type Client struct {
	serverURL  string
	httpClient *http.Client
}

// New creates a new CloudBoot client
func New(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RegisterAgent registers this agent with the CloudBoot server
func (c *Client) RegisterAgent(req *RegisterRequest) (*RegisterResponse, error) {
	resp := &RegisterResponse{}
	err := c.doRequest("POST", "/api/boot/v1/register", req, resp)
	return resp, err
}

// GetTask polls for a new task from the server
func (c *Client) GetTask(machineID string) (*TaskResponse, error) {
	resp := &TaskResponse{}
	url := fmt.Sprintf("/api/boot/v1/task?machine_id=%s", machineID)
	err := c.doRequest("GET", url, nil, resp)
	return resp, err
}

// UploadLogs sends logs to the server
func (c *Client) UploadLogs(req *LogUploadRequest) error {
	return c.doRequest("POST", "/api/boot/v1/logs", req, nil)
}

// ReportStatus reports task execution status
func (c *Client) ReportStatus(req *StatusReportRequest) error {
	return c.doRequest("POST", "/api/boot/v1/status", req, nil)
}

// doRequest performs an HTTP request
func (c *Client) doRequest(method, path string, body, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.serverURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// RegisterRequest represents agent registration payload
type RegisterRequest struct {
	MacAddress   string                 `json:"mac_address"`
	IPAddress    string                 `json:"ip_address"`
	HardwareSpec map[string]interface{} `json:"hardware_spec"`
}

// RegisterResponse represents server response for registration
type RegisterResponse struct {
	MachineID string `json:"machine_id"`
	Status    string `json:"status"`
}

// TaskResponse represents a task from the server
type TaskResponse struct {
	JobID      string                 `json:"job_id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	NoTask     bool                   `json:"no_task"`
}

// LogUploadRequest represents log upload payload
type LogUploadRequest struct {
	JobID string    `json:"job_id"`
	Logs  []LogEntry `json:"logs"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string `json:"ts"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"msg"`
}

// StatusReportRequest represents status report payload
type StatusReportRequest struct {
	JobID       string `json:"job_id"`
	Status      string `json:"status"`
	CurrentStep string `json:"step_current"`
	Error       string `json:"error,omitempty"`
}

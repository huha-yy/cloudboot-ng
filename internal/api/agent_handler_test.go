package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm/logger"
)

// setupTestAgentDB 创建测试数据库
func setupTestAgentDB(t *testing.T) {
	config := database.Config{
		DSN:      ":memory:",
		LogLevel: logger.Silent,
	}

	if err := database.Init(config); err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
}

// TestRegister 测试Agent注册
func TestRegister(t *testing.T) {
	setupTestAgentDB(t)
	defer database.Close()

	e := echo.New()
	handler := NewAgentHandler()

	tests := []struct {
		name       string
		payload    map[string]interface{}
		wantStatus int
		wantFields []string
	}{
		{
			name: "首次注册成功",
			payload: map[string]interface{}{
				"mac_address": "00:11:22:33:44:55",
				"ip_address":  "10.0.0.100",
				"hostname":    "test-server-01",
				"hardware_spec": models.HardwareInfo{
					SchemaVersion: "1.0",
					System: models.SystemInfo{
						Manufacturer: "Dell",
						ProductName:  "PowerEdge R740",
					},
					CPU: models.CPUInfo{
						Cores:   32,
						Sockets: 2,
					},
				},
			},
			wantStatus: http.StatusCreated,
			wantFields: []string{"machine_id", "status", "heartbeat_url", "poll_interval_seconds"},
		},
		{
			name: "重复注册返回updated",
			payload: map[string]interface{}{
				"mac_address": "00:11:22:33:44:55",
				"ip_address":  "10.0.0.101",
			},
			wantStatus: http.StatusOK,
			wantFields: []string{"machine_id", "status"},
		},
		{
			name: "缺少MAC地址",
			payload: map[string]interface{}{
				"ip_address": "10.0.0.100",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/boot/v1/register", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Register(c)
			if err != nil {
				t.Fatalf("handler error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("status code: got %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK || tt.wantStatus == http.StatusCreated {
				var resp map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}

				for _, field := range tt.wantFields {
					if _, ok := resp[field]; !ok {
						t.Errorf("missing field: %s", field)
					}
				}

				t.Logf("Response: %+v", resp)
			}
		})
	}
}

// TestHeartbeat 测试Agent心跳
func TestHeartbeat(t *testing.T) {
	setupTestAgentDB(t)
	defer database.Close()

	e := echo.New()
	handler := NewAgentHandler()

	// 先注册一个机器
	machine := &models.Machine{
		ID:         "test-machine-001",
		Hostname:   "test-server",
		MacAddress: "aa:bb:cc:dd:ee:ff",
		IPAddress:  "10.0.0.50",
		Status:     models.MachineStatusDiscovered,
		HardwareSpec: models.HardwareInfo{
			SchemaVersion: "1.0",
			CPU: models.CPUInfo{
				Cores: 16,
			},
			Memory: models.MemoryInfo{
				TotalBytes: 64 * 1024 * 1024 * 1024, // 64GB
			},
		},
	}
	database.DB.Create(machine)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		wantStatus     int
		wantHWChange   bool
		wantStatusCode int
	}{
		{
			name: "正常心跳",
			payload: map[string]interface{}{
				"machine_id":  "test-machine-001",
				"mac_address": "aa:bb:cc:dd:ee:ff",
				"ip_address":  "10.0.0.50",
				"hardware_spec": models.HardwareInfo{
					SchemaVersion: "1.0",
					CPU: models.CPUInfo{
						Cores: 16, // 相同硬件
					},
					Memory: models.MemoryInfo{
						TotalBytes: 64 * 1024 * 1024 * 1024,
					},
				},
			},
			wantStatusCode: http.StatusOK,
			wantHWChange:   false,
		},
		{
			name: "硬件变更检测",
			payload: map[string]interface{}{
				"machine_id":  "test-machine-001",
				"mac_address": "aa:bb:cc:dd:ee:ff",
				"ip_address":  "10.0.0.50",
				"hardware_spec": models.HardwareInfo{
					SchemaVersion: "1.0",
					CPU: models.CPUInfo{
						Cores: 32, // 硬件变更：CPU增加
					},
					Memory: models.MemoryInfo{
						TotalBytes: 64 * 1024 * 1024 * 1024,
					},
				},
			},
			wantStatusCode: http.StatusOK,
			wantHWChange:   true,
		},
		{
			name: "MAC地址不匹配",
			payload: map[string]interface{}{
				"machine_id":  "test-machine-001",
				"mac_address": "ff:ff:ff:ff:ff:ff", // 错误MAC
			},
			wantStatusCode: http.StatusForbidden,
		},
		{
			name: "机器不存在",
			payload: map[string]interface{}{
				"machine_id":  "not-exist",
				"mac_address": "aa:bb:cc:dd:ee:ff",
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "缺少必填字段",
			payload: map[string]interface{}{
				"machine_id": "test-machine-001",
				// 缺少mac_address
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/boot/v1/heartbeat", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Heartbeat(c)
			if err != nil {
				t.Fatalf("handler error: %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("status code: got %d, want %d", rec.Code, tt.wantStatusCode)
			}

			if rec.Code == http.StatusOK {
				var resp HeartbeatResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}

				if resp.HardwareChange != tt.wantHWChange {
					t.Errorf("hardware_change: got %v, want %v", resp.HardwareChange, tt.wantHWChange)
				}

				t.Logf("Response: status=%s, hardware_change=%v, next_poll=%d",
					resp.Status, resp.HardwareChange, resp.NextPoll)
			}
		})
	}
}

// TestHardwareChangeDetection 测试硬件变更检测
func TestHardwareChangeDetection(t *testing.T) {
	handler := &AgentHandler{}

	machine := &models.Machine{
		HardwareSpec: models.HardwareInfo{
			SchemaVersion: "1.0",
			CPU: models.CPUInfo{
				Cores:   16,
				Sockets: 2,
			},
			Memory: models.MemoryInfo{
				TotalBytes: 64 * 1024 * 1024 * 1024, // 64GB
			},
		},
	}

	tests := []struct {
		name       string
		newSpec    models.HardwareInfo
		wantChange bool
	}{
		{
			name: "无变更",
			newSpec: models.HardwareInfo{
				SchemaVersion: "1.0",
				CPU: models.CPUInfo{
					Cores:   16,
					Sockets: 2,
				},
				Memory: models.MemoryInfo{
					TotalBytes: 64 * 1024 * 1024 * 1024,
				},
			},
			wantChange: false,
		},
		{
			name: "CPU变更",
			newSpec: models.HardwareInfo{
				SchemaVersion: "1.0",
				CPU: models.CPUInfo{
					Cores:   32, // 增加CPU
					Sockets: 2,
				},
				Memory: models.MemoryInfo{
					TotalBytes: 64 * 1024 * 1024 * 1024,
				},
			},
			wantChange: true,
		},
		{
			name: "内存变更",
			newSpec: models.HardwareInfo{
				SchemaVersion: "1.0",
				CPU: models.CPUInfo{
					Cores:   16,
					Sockets: 2,
				},
				Memory: models.MemoryInfo{
					TotalBytes: 128 * 1024 * 1024 * 1024, // 增加内存
				},
			},
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := handler.detectHardwareChange(machine, &tt.newSpec)
			if changed != tt.wantChange {
				t.Errorf("detectHardwareChange: got %v, want %v", changed, tt.wantChange)
			}
		})
	}
}

// TestGenerateHostname 测试主机名生成
func TestGenerateHostname(t *testing.T) {
	tests := []struct {
		mac      string
		expected string
	}{
		{"aa:bb:cc:dd:ee:ff", "server-ddeeff"},
		{"00:11:22:33:44:55", "server-334455"},
		{"aabbccddeeff", "server-ddeeff"},
		{"12:34", "server-1234"},
	}

	for _, tt := range tests {
		t.Run(tt.mac, func(t *testing.T) {
			result := generateHostname(tt.mac)
			if result != tt.expected {
				t.Errorf("generateHostname(%s): got %s, want %s", tt.mac, result, tt.expected)
			}
		})
	}
}

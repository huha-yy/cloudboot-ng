package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/core/logbroker"
	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/labstack/echo/v4"
)

func TestBootHandler_RegisterAgent(t *testing.T) {
	setupTestDB(t)
	broker := logbroker.NewBroker()
	handler := NewBootHandler(broker)

	tests := []struct {
		name           string
		requestBody    string
		wantStatusCode int
		wantMachineID  bool
	}{
		{
			name:           "Register new agent",
			requestBody:    `{"mac":"aa:bb:cc:dd:ee:01","ip":"192.168.1.100"}`,
			wantStatusCode: http.StatusOK,
			wantMachineID:  true,
		},
		{
			name:           "Register agent with fingerprint",
			requestBody:    `{"mac":"aa:bb:cc:dd:ee:02","ip":"192.168.1.101","fingerprint":{"schema_version":"1.0"}}`,
			wantStatusCode: http.StatusOK,
			wantMachineID:  true,
		},
		{
			name:           "Invalid request body",
			requestBody:    `{invalid}`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/boot/v1/register", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handler.RegisterAgent(c); err != nil {
				t.Fatalf("RegisterAgent() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantMachineID {
				var response map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response["machine_id"] == nil || response["machine_id"] == "" {
					t.Error("Response should contain machine_id")
				}

				if response["status"] != "ok" {
					t.Errorf("Status = %v, want ok", response["status"])
				}
			}
		})
	}
}

func TestBootHandler_RegisterAgent_ExistingMachine(t *testing.T) {
	db := setupTestDB(t)
	broker := logbroker.NewBroker()
	handler := NewBootHandler(broker)

	// Seed existing machine
	machine := models.Machine{
		ID:         "existing-machine-id",
		Hostname:   "existing-server",
		MacAddress: "aa:bb:cc:dd:ee:ff",
		IPAddress:  "192.168.1.50",
		Status:     models.MachineStatusReady,
		CreatedAt:  time.Now().Add(-24 * time.Hour),
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
	}
	db.Create(&machine)

	// Register again with same MAC
	e := echo.New()
	requestBody := `{"mac":"aa:bb:cc:dd:ee:ff","ip":"192.168.1.100"}`
	req := httptest.NewRequest(http.MethodPost, "/api/boot/v1/register", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.RegisterAgent(c); err != nil {
		t.Fatalf("RegisterAgent() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", rec.Code, http.StatusOK)
	}

	// Verify machine was updated (not created new)
	var updated models.Machine
	db.Where("mac_address = ?", "aa:bb:cc:dd:ee:ff").First(&updated)

	if updated.ID != "existing-machine-id" {
		t.Error("Should update existing machine, not create new")
	}

	if updated.IPAddress != "192.168.1.100" {
		t.Errorf("IP should be updated to 192.168.1.100, got %s", updated.IPAddress)
	}
}

func TestBootHandler_GetTask(t *testing.T) {
	db := setupTestDB(t)
	broker := logbroker.NewBroker()
	handler := NewBootHandler(broker)

	// Seed test machine
	machine := models.Machine{
		ID:         "machine-1",
		Hostname:   "server-01",
		MacAddress: "aa:bb:cc:dd:ee:01",
		Status:     models.MachineStatusReady,
	}
	db.Create(&machine)

	// Seed pending job
	job := models.Job{
		ID:        "job-123",
		MachineID: "machine-1",
		Type:      models.JobTypeInstallOS,
		Status:    models.JobStatusPending,
		CreatedAt: time.Now(),
	}
	db.Create(&job)

	tests := []struct {
		name           string
		mac            string
		wantStatusCode int
		wantTaskID     bool
	}{
		{
			name:           "Get task for machine with pending job",
			mac:            "aa:bb:cc:dd:ee:01",
			wantStatusCode: http.StatusOK,
			wantTaskID:     true,
		},
		{
			name:           "No MAC parameter",
			mac:            "",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Machine not found",
			mac:            "ff:ff:ff:ff:ff:ff",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/boot/v1/task?mac="+tt.mac, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handler.GetTask(c); err != nil {
				t.Fatalf("GetTask() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantTaskID {
				var response map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response["task_id"] != "job-123" {
					t.Errorf("task_id = %v, want job-123", response["task_id"])
				}

				// Verify job status updated to running
				var updatedJob models.Job
				db.Where("id = ?", "job-123").First(&updatedJob)
				if updatedJob.Status != models.JobStatusRunning {
					t.Errorf("Job status = %v, want %v", updatedJob.Status, models.JobStatusRunning)
				}
			}
		})
	}
}

func TestBootHandler_GetTask_NoTask(t *testing.T) {
	db := setupTestDB(t)
	broker := logbroker.NewBroker()
	handler := NewBootHandler(broker)

	// Seed test machine without pending job
	machine := models.Machine{
		ID:         "machine-2",
		Hostname:   "server-02",
		MacAddress: "aa:bb:cc:dd:ee:02",
		Status:     models.MachineStatusReady,
	}
	db.Create(&machine)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/boot/v1/task?mac=aa:bb:cc:dd:ee:02", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.GetTask(c); err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", rec.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["task_id"] != nil {
		t.Error("task_id should be nil when no task available")
	}

	if response["message"] != "No task available" {
		t.Errorf("message = %v, want 'No task available'", response["message"])
	}
}

func TestBootHandler_UploadLogs(t *testing.T) {
	setupTestDB(t)
	broker := logbroker.NewBroker()
	handler := NewBootHandler(broker)

	// Subscribe to logs
	ch := broker.Subscribe("job-456")

	tests := []struct {
		name           string
		requestBody    string
		wantStatusCode int
		wantLogsCount  int
	}{
		{
			name: "Upload logs with job_id",
			requestBody: `{
				"job_id": "job-456",
				"logs": [
					{"ts":"2026-01-15T10:00:00Z","level":"INFO","component":"agent","msg":"Task started"}
				]
			}`,
			wantStatusCode: http.StatusOK,
			wantLogsCount:  1,
		},
		{
			name: "Upload logs with task_id",
			requestBody: `{
				"task_id": "job-456",
				"logs": [
					{"ts":"2026-01-15T10:00:01Z","level":"DEBUG","component":"agent","msg":"Step 1 completed"}
				]
			}`,
			wantStatusCode: http.StatusOK,
			wantLogsCount:  1,
		},
		{
			name:           "Missing job_id and task_id",
			requestBody:    `{"logs":[]}`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid}`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/boot/v1/logs", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handler.UploadLogs(c); err != nil {
				t.Fatalf("UploadLogs() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				// Verify logs were forwarded to broker
				select {
				case msg := <-ch:
					if msg.Message != "Task started" && msg.Message != "Step 1 completed" {
						t.Errorf("Unexpected log message: %v", msg.Message)
					}
				case <-time.After(100 * time.Millisecond):
					t.Error("Expected log message not received from broker")
				}

				var response map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				logsReceived := int(response["logs_received"].(float64))
				if logsReceived != tt.wantLogsCount {
					t.Errorf("logs_received = %v, want %v", logsReceived, tt.wantLogsCount)
				}
			}
		})
	}
}

func TestBootHandler_ReportStatus(t *testing.T) {
	db := setupTestDB(t)
	broker := logbroker.NewBroker()
	handler := NewBootHandler(broker)

	tests := []struct {
		name           string
		setupJob       func() string
		requestBody    string
		wantStatusCode int
		wantJobStatus  models.JobStatus
	}{
		{
			name: "Report success",
			setupJob: func() string {
				job := models.Job{
					ID:        "job-success",
					MachineID: "machine-1",
					Type:      models.JobTypeInstallOS,
					Status:    models.JobStatusRunning,
					CreatedAt: time.Now(),
				}
				db.Create(&job)
				return job.ID
			},
			requestBody:    `{"task_id":"job-success","status":"success"}`,
			wantStatusCode: http.StatusOK,
			wantJobStatus:  models.JobStatusSuccess,
		},
		{
			name: "Report failure",
			setupJob: func() string {
				job := models.Job{
					ID:        "job-failed",
					MachineID: "machine-2",
					Type:      models.JobTypeConfigRAID,
					Status:    models.JobStatusRunning,
					CreatedAt: time.Now(),
				}
				db.Create(&job)
				return job.ID
			},
			requestBody:    `{"task_id":"job-failed","status":"failed","error_msg":"RAID configuration failed"}`,
			wantStatusCode: http.StatusOK,
			wantJobStatus:  models.JobStatusFailed,
		},
		{
			name:           "Task not found",
			setupJob:       func() string { return "non-existent-job" },
			requestBody:    `{"task_id":"non-existent-job","status":"success"}`,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "Invalid JSON",
			setupJob:       func() string { return "" },
			requestBody:    `{invalid}`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskID := tt.setupJob()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/boot/v1/status", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handler.ReportStatus(c); err != nil {
				t.Fatalf("ReportStatus() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK && taskID != "" {
				// Verify job status was updated
				var updatedJob models.Job
				db.Where("id = ?", taskID).First(&updatedJob)

				if updatedJob.Status != tt.wantJobStatus {
					t.Errorf("Job status = %v, want %v", updatedJob.Status, tt.wantJobStatus)
				}

				if tt.name == "Report failure" && updatedJob.Error == "" {
					t.Error("Job error message should be set")
				}
			}
		})
	}
}

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes a test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate tables
	if err := db.AutoMigrate(
		&models.Machine{},
		&models.Job{},
		&models.OSProfile{},
	); err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	// Set the global DB for handlers to use
	database.SetDB(db)

	return db
}

func TestMachineHandler_ListMachines(t *testing.T) {
	db := setupTestDB(t)
	handler := NewMachineHandler()

	// Seed test data
	machine1 := models.Machine{
		ID:         "machine-1",
		Hostname:   "server-01",
		MacAddress: "aa:bb:cc:dd:ee:01",
		Status:     models.MachineStatusReady,
	}
	machine2 := models.Machine{
		ID:         "machine-2",
		Hostname:   "server-02",
		MacAddress: "aa:bb:cc:dd:ee:02",
		Status:     models.MachineStatusDiscovered,
	}
	db.Create(&machine1)
	db.Create(&machine2)

	tests := []struct {
		name           string
		queryParams    string
		wantStatusCode int
		wantCount      int
	}{
		{
			name:           "List all machines",
			queryParams:    "",
			wantStatusCode: http.StatusOK,
			wantCount:      2,
		},
		{
			name:           "Filter by status",
			queryParams:    "?status=ready",
			wantStatusCode: http.StatusOK,
			wantCount:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/machines"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handler.ListMachines(c); err != nil {
				t.Fatalf("ListMachines() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			items, ok := response["items"].([]interface{})
			if !ok {
				t.Fatal("Response missing 'items' field")
			}

			if len(items) != tt.wantCount {
				t.Errorf("Item count = %v, want %v", len(items), tt.wantCount)
			}
		})
	}
}

func TestMachineHandler_GetMachine(t *testing.T) {
	db := setupTestDB(t)
	handler := NewMachineHandler()

	// Seed test machine
	machine := models.Machine{
		ID:         "test-machine-id",
		Hostname:   "test-server",
		MacAddress: "aa:bb:cc:dd:ee:ff",
		Status:     models.MachineStatusReady,
	}
	db.Create(&machine)

	tests := []struct {
		name           string
		machineID      string
		wantStatusCode int
	}{
		{
			name:           "Get existing machine",
			machineID:      "test-machine-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Get non-existent machine",
			machineID:      "non-existent-id",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/machines/"+tt.machineID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.machineID)

			if err := handler.GetMachine(c); err != nil {
				t.Fatalf("GetMachine() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				var response models.Machine
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.ID != tt.machineID {
					t.Errorf("Machine ID = %v, want %v", response.ID, tt.machineID)
				}
			}
		})
	}
}

func TestMachineHandler_CreateMachine(t *testing.T) {
	setupTestDB(t)
	handler := NewMachineHandler()

	tests := []struct {
		name           string
		requestBody    string
		wantStatusCode int
	}{
		{
			name:           "Create valid machine",
			requestBody:    `{"mac":"aa:bb:cc:dd:ee:01","hostname":"new-server","ip":"192.168.1.100"}`,
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "Invalid request body",
			requestBody:    `{invalid json}`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/machines", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handler.CreateMachine(c); err != nil {
				t.Fatalf("CreateMachine() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusCreated {
				var response models.Machine
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Status != models.MachineStatusDiscovered {
					t.Errorf("Status = %v, want %v", response.Status, models.MachineStatusDiscovered)
				}
			}
		})
	}
}

func TestMachineHandler_CreateMachine_DuplicateMAC(t *testing.T) {
	db := setupTestDB(t)
	handler := NewMachineHandler()

	// Create existing machine
	existing := models.Machine{
		ID:         "existing-id",
		Hostname:   "existing-server",
		MacAddress: "aa:bb:cc:dd:ee:ff",
		Status:     models.MachineStatusReady,
	}
	db.Create(&existing)

	// Try to create machine with same MAC
	e := echo.New()
	requestBody := `{"mac":"aa:bb:cc:dd:ee:ff","hostname":"duplicate-server"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/machines", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.CreateMachine(c); err != nil {
		t.Fatalf("CreateMachine() error = %v", err)
	}

	if rec.Code != http.StatusConflict {
		t.Errorf("Status = %v, want %v", rec.Code, http.StatusConflict)
	}
}

func TestMachineHandler_UpdateMachine(t *testing.T) {
	db := setupTestDB(t)
	handler := NewMachineHandler()

	// Seed test machine
	machine := models.Machine{
		ID:         "test-machine-id",
		Hostname:   "old-hostname",
		MacAddress: "aa:bb:cc:dd:ee:ff",
		Status:     models.MachineStatusDiscovered,
	}
	db.Create(&machine)

	e := echo.New()
	requestBody := `{"hostname":"new-hostname","status":"ready"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/machines/test-machine-id", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("test-machine-id")

	if err := handler.UpdateMachine(c); err != nil {
		t.Fatalf("UpdateMachine() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", rec.Code, http.StatusOK)
	}

	var response models.Machine
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Hostname != "new-hostname" {
		t.Errorf("Hostname = %v, want new-hostname", response.Hostname)
	}

	if response.Status != models.MachineStatusReady {
		t.Errorf("Status = %v, want %v", response.Status, models.MachineStatusReady)
	}
}

func TestMachineHandler_DeleteMachine(t *testing.T) {
	db := setupTestDB(t)
	handler := NewMachineHandler()

	// Seed test machine
	machine := models.Machine{
		ID:         "test-machine-id",
		Hostname:   "test-server",
		MacAddress: "aa:bb:cc:dd:ee:ff",
		Status:     models.MachineStatusReady,
	}
	db.Create(&machine)

	tests := []struct {
		name           string
		machineID      string
		wantStatusCode int
	}{
		{
			name:           "Delete existing machine",
			machineID:      "test-machine-id",
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:           "Delete non-existent machine",
			machineID:      "non-existent-id",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/machines/"+tt.machineID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.machineID)

			if err := handler.DeleteMachine(c); err != nil {
				t.Fatalf("DeleteMachine() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusNoContent {
				// Verify machine is deleted
				var count int64
				db.Model(&models.Machine{}).Where("id = ?", tt.machineID).Count(&count)
				if count != 0 {
					t.Error("Machine should be deleted")
				}
			}
		})
	}
}

func TestMachineHandler_ProvisionMachine(t *testing.T) {
	db := setupTestDB(t)
	handler := NewMachineHandler()

	// Seed test machine
	machine := models.Machine{
		ID:         "test-machine-id",
		Hostname:   "test-server",
		MacAddress: "aa:bb:cc:dd:ee:ff",
		Status:     models.MachineStatusReady,
	}
	db.Create(&machine)

	tests := []struct {
		name           string
		machineID      string
		requestBody    string
		wantStatusCode int
	}{
		{
			name:           "Provision machine with profile",
			machineID:      "test-machine-id",
			requestBody:    `{"profile_id":"profile-123"}`,
			wantStatusCode: http.StatusAccepted,
		},
		{
			name:           "Provision non-existent machine",
			machineID:      "non-existent-id",
			requestBody:    `{"profile_id":"profile-123"}`,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "Invalid request body",
			machineID:      "test-machine-id",
			requestBody:    `{invalid}`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/machines/"+tt.machineID+"/provision", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.machineID)

			if err := handler.ProvisionMachine(c); err != nil {
				t.Fatalf("ProvisionMachine() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusAccepted {
				var response models.Job
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Type != models.JobTypeInstallOS {
					t.Errorf("Job type = %v, want %v", response.Type, models.JobTypeInstallOS)
				}

				if response.Status != models.JobStatusPending {
					t.Errorf("Job status = %v, want %v", response.Status, models.JobStatusPending)
				}

				// Verify machine status updated
				var updatedMachine models.Machine
				db.Where("id = ?", tt.machineID).First(&updatedMachine)
				if updatedMachine.Status != models.MachineStatusInstalling {
					t.Errorf("Machine status = %v, want %v", updatedMachine.Status, models.MachineStatusInstalling)
				}
			}
		})
	}
}

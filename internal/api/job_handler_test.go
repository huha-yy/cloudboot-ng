package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/labstack/echo/v4"
)

func TestJobHandler_ListJobs(t *testing.T) {
	db := setupTestDB(t)
	handler := NewJobHandler()

	// Seed test machines first
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
		Status:     models.MachineStatusReady,
	}
	db.Create(&machine1)
	db.Create(&machine2)

	// Seed test jobs
	job1 := models.Job{
		ID:        "job-1",
		MachineID: "machine-1",
		Type:      models.JobTypeInstallOS,
		Status:    models.JobStatusRunning,
		CreatedAt: time.Now(),
	}
	job2 := models.Job{
		ID:        "job-2",
		MachineID: "machine-2",
		Type:      models.JobTypeConfigRAID,
		Status:    models.JobStatusPending,
		CreatedAt: time.Now(),
	}
	db.Create(&job1)
	db.Create(&job2)

	tests := []struct {
		name           string
		queryParams    string
		wantStatusCode int
		wantCount      int
	}{
		{
			name:           "List all jobs",
			queryParams:    "",
			wantStatusCode: http.StatusOK,
			wantCount:      2,
		},
		{
			name:           "Filter by status",
			queryParams:    "?status=running",
			wantStatusCode: http.StatusOK,
			wantCount:      1,
		},
		{
			name:           "Filter by machine_id",
			queryParams:    "?machine_id=machine-1",
			wantStatusCode: http.StatusOK,
			wantCount:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handler.ListJobs(c); err != nil {
				t.Fatalf("ListJobs() error = %v", err)
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

func TestJobHandler_GetJob(t *testing.T) {
	db := setupTestDB(t)
	handler := NewJobHandler()

	// Seed test machine
	machine := models.Machine{
		ID:         "machine-1",
		Hostname:   "server-01",
		MacAddress: "aa:bb:cc:dd:ee:01",
		Status:     models.MachineStatusReady,
	}
	db.Create(&machine)

	// Seed test job
	job := models.Job{
		ID:        "test-job-id",
		MachineID: "machine-1",
		Type:      models.JobTypeInstallOS,
		Status:    models.JobStatusRunning,
		CreatedAt: time.Now(),
	}
	db.Create(&job)

	tests := []struct {
		name           string
		jobID          string
		wantStatusCode int
	}{
		{
			name:           "Get existing job",
			jobID:          "test-job-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Get non-existent job",
			jobID:          "non-existent-id",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+tt.jobID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.jobID)

			if err := handler.GetJob(c); err != nil {
				t.Fatalf("GetJob() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				var response models.Job
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.ID != tt.jobID {
					t.Errorf("Job ID = %v, want %v", response.ID, tt.jobID)
				}
			}
		})
	}
}

func TestJobHandler_CancelJob(t *testing.T) {
	db := setupTestDB(t)
	handler := NewJobHandler()

	// Seed test machine
	machine := models.Machine{
		ID:         "machine-1",
		Hostname:   "server-01",
		MacAddress: "aa:bb:cc:dd:ee:01",
		Status:     models.MachineStatusReady,
	}
	db.Create(&machine)

	tests := []struct {
		name           string
		jobID          string
		jobStatus      models.JobStatus
		wantStatusCode int
		wantError      string
	}{
		{
			name:           "Cancel pending job",
			jobID:          "job-pending",
			jobStatus:      models.JobStatusPending,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Cancel running job",
			jobID:          "job-running",
			jobStatus:      models.JobStatusRunning,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Cannot cancel completed job",
			jobID:          "job-completed",
			jobStatus:      models.JobStatusSuccess,
			wantStatusCode: http.StatusBadRequest,
			wantError:      "Cannot cancel a terminal job",
		},
		{
			name:           "Cannot cancel failed job",
			jobID:          "job-failed",
			jobStatus:      models.JobStatusFailed,
			wantStatusCode: http.StatusBadRequest,
			wantError:      "Cannot cancel a terminal job",
		},
		{
			name:           "Cancel non-existent job",
			jobID:          "non-existent-id",
			jobStatus:      models.JobStatusPending,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create job for this test (except non-existent case)
			if tt.jobID != "non-existent-id" {
				job := models.Job{
					ID:        tt.jobID,
					MachineID: "machine-1",
					Type:      models.JobTypeInstallOS,
					Status:    tt.jobStatus,
					CreatedAt: time.Now(),
				}
				db.Create(&job)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/jobs/"+tt.jobID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.jobID)

			if err := handler.CancelJob(c); err != nil {
				t.Fatalf("CancelJob() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				var response models.Job
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Status != models.JobStatusFailed {
					t.Errorf("Job status = %v, want %v", response.Status, models.JobStatusFailed)
				}

				if response.Error != "Cancelled by user" {
					t.Errorf("Job error = %v, want 'Cancelled by user'", response.Error)
				}
			}

			if tt.wantError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if errMsg, ok := response["error"].(string); !ok || errMsg != tt.wantError {
					t.Errorf("Error = %v, want %v", errMsg, tt.wantError)
				}
			}
		})
	}
}

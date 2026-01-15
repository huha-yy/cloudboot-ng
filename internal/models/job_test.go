package models

import (
	"errors"
	"testing"
)

func TestJobStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   string
	}{
		{"Pending", JobStatusPending, "pending"},
		{"Running", JobStatusRunning, "running"},
		{"Success", JobStatusSuccess, "success"},
		{"Failed", JobStatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.status); got != tt.want {
				t.Errorf("JobStatus = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJobType_String(t *testing.T) {
	tests := []struct {
		name    string
		jobType JobType
		want    string
	}{
		{"Audit", JobTypeAudit, "audit"},
		{"ConfigRAID", JobTypeConfigRAID, "config_raid"},
		{"InstallOS", JobTypeInstallOS, "install_os"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.jobType); got != tt.want {
				t.Errorf("JobType = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_SetSuccess(t *testing.T) {
	job := Job{
		ID:        "test-job-id",
		MachineID: "test-machine-id",
		Type:      JobTypeAudit,
		Status:    JobStatusRunning,
		Error:     "some old error",
	}

	job.SetSuccess()

	if job.Status != JobStatusSuccess {
		t.Errorf("Status = %v, want %v", job.Status, JobStatusSuccess)
	}

	if job.Error != "" {
		t.Errorf("Error should be cleared, got %v", job.Error)
	}
}

func TestJob_SetError(t *testing.T) {
	job := Job{
		ID:        "test-job-id",
		MachineID: "test-machine-id",
		Type:      JobTypeConfigRAID,
		Status:    JobStatusRunning,
	}

	err := errors.New("RAID configuration failed")
	job.SetError(err)

	if job.Status != JobStatusFailed {
		t.Errorf("Status = %v, want %v", job.Status, JobStatusFailed)
	}

	if job.Error != err.Error() {
		t.Errorf("Error = %v, want %v", job.Error, err.Error())
	}
}

func TestJob_IsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"Pending - not terminal", JobStatusPending, false},
		{"Running - not terminal", JobStatusRunning, false},
		{"Success - terminal", JobStatusSuccess, true},
		{"Failed - terminal", JobStatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{
				ID:     "test-job-id",
				Status: tt.status,
			}

			if got := job.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_IsPending(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"Pending", JobStatusPending, true},
		{"Running", JobStatusRunning, false},
		{"Success", JobStatusSuccess, false},
		{"Failed", JobStatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{
				ID:     "test-job-id",
				Status: tt.status,
			}

			if got := job.IsPending(); got != tt.want {
				t.Errorf("IsPending() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_IsRunning(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"Pending", JobStatusPending, false},
		{"Running", JobStatusRunning, true},
		{"Success", JobStatusSuccess, false},
		{"Failed", JobStatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{
				ID:     "test-job-id",
				Status: tt.status,
			}

			if got := job.IsRunning(); got != tt.want {
				t.Errorf("IsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_UpdateStep(t *testing.T) {
	job := Job{
		ID:          "test-job-id",
		StepCurrent: "initial",
	}

	newStep := "downloading_iso"
	job.UpdateStep(newStep)

	if job.StepCurrent != newStep {
		t.Errorf("StepCurrent = %v, want %v", job.StepCurrent, newStep)
	}
}

func TestJob_Lifecycle(t *testing.T) {
	// Test complete job lifecycle
	job := Job{
		ID:          "test-job-id",
		MachineID:   "test-machine-id",
		Type:        JobTypeInstallOS,
		Status:      JobStatusPending,
		StepCurrent: "created",
	}

	// Initial state
	if !job.IsPending() {
		t.Error("New job should be pending")
	}

	if job.IsTerminal() {
		t.Error("New job should not be terminal")
	}

	// Start job
	job.Status = JobStatusRunning
	job.UpdateStep("downloading_iso")

	if !job.IsRunning() {
		t.Error("Job should be running")
	}

	// Complete successfully
	job.SetSuccess()

	if !job.IsTerminal() {
		t.Error("Success job should be terminal")
	}

	if job.Status != JobStatusSuccess {
		t.Errorf("Final status = %v, want %v", job.Status, JobStatusSuccess)
	}
}

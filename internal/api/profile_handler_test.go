package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/labstack/echo/v4"
)

func TestProfileHandler_ListProfiles(t *testing.T) {
	db := setupTestDB(t)
	handler := NewProfileHandler()

	// Seed test profiles
	profile1 := models.OSProfile{
		ID:     "profile-1",
		Name:   "CentOS 7 Default",
		Distro: "centos7",
		Config: models.ProfileConfig{
			RepoURL: "http://mirror.centos.org/centos/7/os/x86_64",
		},
	}
	profile2 := models.OSProfile{
		ID:     "profile-2",
		Name:   "Ubuntu 20.04",
		Distro: "ubuntu2004",
		Config: models.ProfileConfig{
			RepoURL: "http://archive.ubuntu.com/ubuntu",
		},
	}
	db.Create(&profile1)
	db.Create(&profile2)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.ListProfiles(c); err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", rec.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	profiles, ok := response["profiles"].([]interface{})
	if !ok {
		t.Fatal("Response missing 'profiles' field")
	}

	if len(profiles) != 2 {
		t.Errorf("Profile count = %v, want 2", len(profiles))
	}
}

func TestProfileHandler_GetProfile(t *testing.T) {
	db := setupTestDB(t)
	handler := NewProfileHandler()

	// Seed test profile
	profile := models.OSProfile{
		ID:     "test-profile-id",
		Name:   "Test Profile",
		Distro: "centos7",
		Config: models.ProfileConfig{
			RepoURL: "http://mirror.example.com",
		},
	}
	db.Create(&profile)

	tests := []struct {
		name           string
		profileID      string
		wantStatusCode int
	}{
		{
			name:           "Get existing profile",
			profileID:      "test-profile-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Get non-existent profile",
			profileID:      "non-existent-id",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/"+tt.profileID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.profileID)

			if err := handler.GetProfile(c); err != nil {
				t.Fatalf("GetProfile() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				var response models.OSProfile
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.ID != tt.profileID {
					t.Errorf("Profile ID = %v, want %v", response.ID, tt.profileID)
				}
			}
		})
	}
}

func TestProfileHandler_CreateProfile(t *testing.T) {
	setupTestDB(t)
	handler := NewProfileHandler()

	tests := []struct {
		name           string
		requestBody    string
		wantStatusCode int
	}{
		{
			name: "Create valid profile",
			requestBody: `{
				"name": "New Profile",
				"distro": "centos7",
				"config": {
					"repo_url": "http://mirror.centos.org/centos/7/os/x86_64",
					"partitions": [
						{"mount_point": "/boot", "size": "1GB", "fstype": "ext4"},
						{"mount_point": "/", "size": "50GB", "fstype": "xfs"}
					],
					"network": {
						"hostname": "test-server",
						"ip": "192.168.1.100",
						"netmask": "255.255.255.0",
						"gateway": "192.168.1.1"
					}
				}
			}`,
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json}`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handler.CreateProfile(c); err != nil {
				t.Fatalf("CreateProfile() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusCreated {
				var response models.OSProfile
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.ID == "" {
					t.Error("Profile ID should be auto-generated")
				}

				if response.Name != "New Profile" {
					t.Errorf("Name = %v, want 'New Profile'", response.Name)
				}
			}
		})
	}
}

func TestProfileHandler_UpdateProfile(t *testing.T) {
	db := setupTestDB(t)
	handler := NewProfileHandler()

	// Seed test profile
	profile := models.OSProfile{
		ID:     "test-profile-id",
		Name:   "Old Name",
		Distro: "centos7",
		Config: models.ProfileConfig{
			RepoURL: "http://old-mirror.com",
		},
	}
	db.Create(&profile)

	tests := []struct {
		name           string
		profileID      string
		requestBody    string
		wantStatusCode int
	}{
		{
			name:      "Update existing profile",
			profileID: "test-profile-id",
			requestBody: `{
				"name": "Updated Name",
				"distro": "centos7",
				"config": {
					"repo_url": "http://new-mirror.com",
					"partitions": [
						{"mount_point": "/boot", "size": "1GB", "fstype": "ext4"},
						{"mount_point": "/", "size": "50GB", "fstype": "xfs"}
					],
					"network": {
						"hostname": "updated-server",
						"ip": "192.168.1.200",
						"netmask": "255.255.255.0",
						"gateway": "192.168.1.1"
					}
				}
			}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Update non-existent profile",
			profileID:      "non-existent-id",
			requestBody:    `{"name": "New Name"}`,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/profiles/"+tt.profileID, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.profileID)

			if err := handler.UpdateProfile(c); err != nil {
				t.Fatalf("UpdateProfile() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				var response models.OSProfile
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Name != "Updated Name" {
					t.Errorf("Name = %v, want 'Updated Name'", response.Name)
				}
			}
		})
	}
}

func TestProfileHandler_DeleteProfile(t *testing.T) {
	db := setupTestDB(t)
	handler := NewProfileHandler()

	// Seed test profile
	profile := models.OSProfile{
		ID:     "test-profile-id",
		Name:   "Test Profile",
		Distro: "centos7",
		Config: models.ProfileConfig{},
	}
	db.Create(&profile)

	tests := []struct {
		name           string
		profileID      string
		wantStatusCode int
	}{
		{
			name:           "Delete existing profile",
			profileID:      "test-profile-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Delete non-existent profile",
			profileID:      "non-existent-id",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/profiles/"+tt.profileID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.profileID)

			if err := handler.DeleteProfile(c); err != nil {
				t.Fatalf("DeleteProfile() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				// Verify profile is deleted
				var count int64
				db.Model(&models.OSProfile{}).Where("id = ?", tt.profileID).Count(&count)
				if count != 0 {
					t.Error("Profile should be deleted")
				}
			}
		})
	}
}

func TestProfileHandler_PreviewConfig(t *testing.T) {
	db := setupTestDB(t)
	handler := NewProfileHandler()

	// Seed test profile
	profile := models.OSProfile{
		ID:     "test-profile-id",
		Name:   "Test Profile",
		Distro: "centos7",
		Config: models.ProfileConfig{
			RepoURL: "http://mirror.centos.org/centos/7/os/x86_64",
			Partitions: []models.Partition{
				{
					MountPoint: "/boot",
					Size:       "1GB",
					FSType:     "ext4",
				},
				{
					MountPoint: "/",
					Size:       "50GB",
					FSType:     "xfs",
				},
			},
			Network: models.NetworkConfig{
				Hostname: "test-server",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
				Gateway:  "192.168.1.1",
			},
		},
	}
	db.Create(&profile)

	tests := []struct {
		name           string
		profileID      string
		wantStatusCode int
	}{
		{
			name:           "Preview existing profile",
			profileID:      "test-profile-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Preview non-existent profile",
			profileID:      "non-existent-id",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles/"+tt.profileID+"/preview", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.profileID)

			if err := handler.PreviewConfig(c); err != nil {
				t.Fatalf("PreviewConfig() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				config := rec.Body.String()
				if config == "" {
					t.Error("Preview config should not be empty")
				}

				// Verify it's a Kickstart config
				if !strings.Contains(config, "# Kickstart for centos7") {
					t.Errorf("Config should contain Kickstart header, got: %s", config)
				}
			}
		})
	}
}

func TestProfileHandler_PreviewFromPayload(t *testing.T) {
	setupTestDB(t)
	handler := NewProfileHandler()

	tests := []struct {
		name           string
		requestBody    string
		wantStatusCode int
	}{
		{
			name: "Preview valid payload",
			requestBody: `{
				"distro": "centos7",
				"config": {
					"repo_url": "http://mirror.centos.org/centos/7/os/x86_64",
					"partitions": [
						{"mount_point": "/boot", "size": "1GB", "fstype": "ext4"},
						{"mount_point": "/", "size": "50GB", "fstype": "xfs"}
					],
					"network": {
						"hostname": "preview-server",
						"ip": "192.168.1.50",
						"netmask": "255.255.255.0",
						"gateway": "192.168.1.1"
					}
				}
			}`,
			wantStatusCode: http.StatusOK,
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
			req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles/preview", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handler.PreviewFromPayload(c); err != nil {
				t.Fatalf("PreviewFromPayload() error = %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK {
				config := rec.Body.String()
				if config == "" {
					t.Error("Preview config should not be empty")
				}

				// Verify it's a Kickstart config
				if !strings.Contains(config, "# Kickstart for centos7") {
					t.Errorf("Config should contain Kickstart header, got: %s", config)
				}
			}
		})
	}
}

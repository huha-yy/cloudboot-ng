package api

import (
	"net/http"

	"github.com/cloudboot/cloudboot-ng/internal/core/cspm"
	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/monitor"
	"github.com/labstack/echo/v4"
)

// WebHandler handles web page rendering
type WebHandler struct {
	pluginManager *cspm.PluginManager
}

// NewWebHandler creates a new web handler
func NewWebHandler(pm *cspm.PluginManager) *WebHandler {
	return &WebHandler{
		pluginManager: pm,
	}
}

// OSDesignerPage renders the OS Designer page
func (h *WebHandler) OSDesignerPage(c echo.Context) error {
	// Fetch all profiles
	var profiles []models.OSProfile
	database.DB.Find(&profiles)

	// Calculate stats
	stats := struct {
		TotalProfiles  int
		CentosProfiles int
		UbuntuProfiles int
		ActiveJobs     int
	}{}

	stats.TotalProfiles = len(profiles)
	for _, p := range profiles {
		if p.Distro == "centos7" || p.Distro == "centos8" {
			stats.CentosProfiles++
		} else if p.Distro == "ubuntu20" || p.Distro == "ubuntu22" {
			stats.UbuntuProfiles++
		}
	}

	// Count active jobs
	var activeJobs int64
	database.DB.Model(&models.Job{}).Where("status IN ?", []models.JobStatus{
		models.JobStatusPending,
		models.JobStatusRunning,
	}).Count(&activeJobs)
	stats.ActiveJobs = int(activeJobs)

	// Render template with data
	data := map[string]interface{}{
		"title":           "OS Designer",
		"active":          "os-designer",
		"pageHeader":      "OS Designer",
		"pageDescription": "Create and manage OS installation profiles visually",
		"profiles":        profiles,
		"stats":           stats,
	}

	return c.Render(http.StatusOK, "os-designer.html", data)
}

// MachinesPage renders the Machines page
func (h *WebHandler) MachinesPage(c echo.Context) error {
	var machines []models.Machine
	database.DB.Find(&machines)

	// Calculate stats
	stats := struct {
		TotalMachines   int
		OnlineMachines  int
		OfflineMachines int
		ProvisionedToday int
	}{}

	stats.TotalMachines = len(machines)
	for _, m := range machines {
		if m.Status == models.MachineStatusReady || m.Status == models.MachineStatusInstalling {
			stats.OnlineMachines++
		} else {
			stats.OfflineMachines++
		}
	}

	data := map[string]interface{}{
		"title":           "Machines",
		"active":          "machines",
		"pageHeader":      "Physical Machines",
		"pageDescription": "Manage and monitor your bare-metal servers",
		"machines":        machines,
		"stats":           stats,
	}

	return c.Render(http.StatusOK, "machines.html", data)
}

// JobsPage renders the Jobs page
func (h *WebHandler) JobsPage(c echo.Context) error {
	var jobs []models.Job
	database.DB.Preload("Machine").Order("created_at DESC").Find(&jobs)

	// Calculate stats
	stats := struct {
		TotalJobs     int
		PendingJobs   int
		RunningJobs   int
		CompletedJobs int
		FailedJobs    int
	}{}

	stats.TotalJobs = len(jobs)
	for _, j := range jobs {
		switch j.Status {
		case models.JobStatusPending:
			stats.PendingJobs++
		case models.JobStatusRunning:
			stats.RunningJobs++
		case models.JobStatusSuccess:
			stats.CompletedJobs++
		case models.JobStatusFailed:
			stats.FailedJobs++
		}
	}

	data := map[string]interface{}{
		"title":           "Jobs",
		"active":          "jobs",
		"pageHeader":      "Provisioning Jobs",
		"pageDescription": "Track OS installation and configuration tasks",
		"jobs":            jobs,
		"stats":           stats,
	}

	return c.Render(http.StatusOK, "jobs.html", data)
}

// StorePage renders the Private Store page
func (h *WebHandler) StorePage(c echo.Context) error {
	// Get providers from PluginManager
	providers := h.pluginManager.ListProviders()

	// Calculate stats
	stats := struct {
		TotalProviders int
		RaidProviders  int
		BiosProviders  int
	}{
		TotalProviders: len(providers),
		RaidProviders:  0,
		BiosProviders:  0,
	}

	data := map[string]any{
		"title":           "Private Store",
		"active":          "store",
		"pageHeader":      "Provider Store",
		"pageDescription": "Manage hardware provider plugins (.cbp packages)",
		"providers":       providers,
		"stats":           stats,
	}

	return c.Render(http.StatusOK, "store.html", data)
}

// HomePage renders the home/dashboard page
func (h *WebHandler) HomePage(c echo.Context) error {
	// Get overview stats
	var machineCount, jobCount, profileCount int64
	database.DB.Model(&models.Machine{}).Count(&machineCount)
	database.DB.Model(&models.Job{}).Count(&jobCount)
	database.DB.Model(&models.OSProfile{}).Count(&profileCount)

	var recentJobs []models.Job
	database.DB.Preload("Machine").Order("created_at DESC").Limit(10).Find(&recentJobs)

	// Get system monitor stats
	sysStats := monitor.GetStats()

	// Get database health
	dbHealthy := database.HealthCheck() == nil

	// Get CSPM providers
	providers := h.pluginManager.ListProviders()

	data := map[string]any{
		"title":  "Dashboard",
		"active": "home",
		"stats": map[string]any{
			"machines": machineCount,
			"jobs":     jobCount,
			"profiles": profileCount,
		},
		"recentJobs": recentJobs,
		"monitor": map[string]any{
			"uptime":      sysStats.Uptime,
			"diskUsage":   int(sysStats.DiskUsage),
			"memoryUsage": int(sysStats.MemoryUsage),
			"dbHealthy":   dbHealthy,
		},
		"providers": map[string]any{
			"count": len(providers),
			"list":  providers,
		},
	}

	return c.Render(http.StatusOK, "home.html", data)
}

// SettingsPage renders the system settings page
func (h *WebHandler) SettingsPage(c echo.Context) error {
	data := map[string]interface{}{
		"title":           "Settings",
		"active":          "settings",
		"pageHeader":      "System Settings",
		"pageDescription": "Configure CloudBoot NG system parameters and global options",
	}

	return c.Render(http.StatusOK, "settings.html", data)
}

// DesignSystemPage renders the design system showcase page
func (h *WebHandler) DesignSystemPage(c echo.Context) error {
	data := map[string]interface{}{
		"title":           "Design System",
		"active":          "design-system",
		"pageHeader":      "CloudBoot NG Design System",
		"pageDescription": "Dark Industrial Theme - 组件库与样式指南",
	}

	return c.Render(http.StatusOK, "design-system.html", data)
}

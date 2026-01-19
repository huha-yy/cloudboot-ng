package api

import (
	"fmt"
	"net/http"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/labstack/echo/v4"
)

// BootConfigHandler Boot配置处理器（Kickstart/AutoYaST）
type BootConfigHandler struct {
	serverURL string
}

// NewBootConfigHandler 创建Boot配置处理器
func NewBootConfigHandler(serverURL string) *BootConfigHandler {
	return &BootConfigHandler{
		serverURL: serverURL,
	}
}

// BootConfigData 启动配置模板数据
type BootConfigData struct {
	ServerURL string
	Machine   *models.Machine
	Profile   *models.OSProfile
}

// ServeKickstart 提供Kickstart配置
// GET /boot/kickstart/:machine_id
func (h *BootConfigHandler) ServeKickstart(c echo.Context) error {
	machineID := c.Param("machine_id")
	if machineID == "" {
		return c.String(http.StatusBadRequest, "# Error: machine_id required\n")
	}

	// 查找机器
	var machine models.Machine
	if err := database.DB.First(&machine, "id = ?", machineID).Error; err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("# Error: Machine not found: %s\n", machineID))
	}

	// 查找待执行的安装任务
	var job models.Job
	if err := database.DB.Where("machine_id = ? AND type = ? AND status IN (?)",
		machine.ID, "install_os", []models.JobStatus{models.JobStatusPending, models.JobStatusRunning}).
		First(&job).Error; err != nil {
		return c.String(http.StatusNotFound, "# Error: No pending installation job\n")
	}

	// 加载OS Profile
	var profile models.OSProfile
	if err := database.DB.First(&profile, "id = ?", job.ProfileID).Error; err != nil {
		return c.String(http.StatusNotFound, "# Error: OS Profile not found\n")
	}

	// 验证发行版类型
	if !isRHELBased(profile.Distro) {
		return c.String(http.StatusBadRequest, "# Error: Kickstart only supports RHEL-based distributions\n")
	}

	// 渲染Kickstart模板
	data := BootConfigData{
		ServerURL: h.serverURL,
		Machine:   &machine,
		Profile:   &profile,
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/plain; charset=utf-8")
	return c.Render(http.StatusOK, "kickstart.tmpl", data)
}

// ServeAutoYaST 提供AutoYaST配置
// GET /boot/autoyast/:machine_id
func (h *BootConfigHandler) ServeAutoYaST(c echo.Context) error {
	machineID := c.Param("machine_id")
	if machineID == "" {
		return c.String(http.StatusBadRequest, "<!-- Error: machine_id required -->\n")
	}

	// 查找机器
	var machine models.Machine
	if err := database.DB.First(&machine, "id = ?", machineID).Error; err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("<!-- Error: Machine not found: %s -->\n", machineID))
	}

	// 查找待执行的安装任务
	var job models.Job
	if err := database.DB.Where("machine_id = ? AND type = ? AND status IN (?)",
		machine.ID, "install_os", []models.JobStatus{models.JobStatusPending, models.JobStatusRunning}).
		First(&job).Error; err != nil {
		return c.String(http.StatusNotFound, "<!-- Error: No pending installation job -->\n")
	}

	// 加载OS Profile
	var profile models.OSProfile
	if err := database.DB.First(&profile, "id = ?", job.ProfileID).Error; err != nil {
		return c.String(http.StatusNotFound, "<!-- Error: OS Profile not found -->\n")
	}

	// 验证发行版类型
	if !isSUSEBased(profile.Distro) {
		return c.String(http.StatusBadRequest, "<!-- Error: AutoYaST only supports SUSE-based distributions -->\n")
	}

	// 渲染AutoYaST模板
	data := BootConfigData{
		ServerURL: h.serverURL,
		Machine:   &machine,
		Profile:   &profile,
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/xml; charset=utf-8")
	return c.Render(http.StatusOK, "autoyast.tmpl", data)
}

// isRHELBased 检查是否为RHEL系发行版
func isRHELBased(distro string) bool {
	rhelBased := map[string]bool{
		"centos":   true,
		"centos7":  true,
		"centos8":  true,
		"rhel":     true,
		"rhel7":    true,
		"rhel8":    true,
		"rhel9":    true,
		"rocky":    true,
		"rocky8":   true,
		"rocky9":   true,
		"alma":     true,
		"alma8":    true,
		"alma9":    true,
		"almalinux": true,
	}
	return rhelBased[distro]
}

// isSUSEBased 检查是否为SUSE系发行版
func isSUSEBased(distro string) bool {
	suseBased := map[string]bool{
		"suse":     true,
		"suse15":   true,
		"sles":     true,
		"sles15":   true,
		"opensuse": true,
		"leap":     true,
		"leap15":   true,
	}
	return suseBased[distro]
}

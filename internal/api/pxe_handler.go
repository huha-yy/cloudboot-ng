package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/labstack/echo/v4"
)

// PXEHandler PXE/iPXE启动处理器
type PXEHandler struct {
	serverURL string
}

// NewPXEHandler 创建PXE处理器
func NewPXEHandler(serverURL string) *PXEHandler {
	return &PXEHandler{
		serverURL: serverURL,
	}
}

// iPXEBootScript iPXE启动脚本数据
type iPXEBootScript struct {
	ServerURL  string
	MachineID  string
	MacAddress string
	Hostname   string
	BootMode   string // "discovery", "install", "localboot"
	OSProfile  *OSProfileData
}

// OSProfileData OS配置数据
type OSProfileData struct {
	Distro    string
	Version   string
	KernelURL string
	InitrdURL string
	RepoURL   string
}

// ServeiPXEScript 提供iPXE启动脚本
// GET /boot/ipxe/:mac
func (h *PXEHandler) ServeiPXEScript(c echo.Context) error {
	macAddr := c.Param("mac")
	if macAddr == "" {
		return c.String(http.StatusBadRequest, "#!ipxe\necho Error: MAC address required\n")
	}

	// 规范化MAC地址格式
	macAddr = normalizeMACAddress(macAddr)

	// 查找机器
	var machine models.Machine
	err := database.DB.Where("mac_address = ?", macAddr).First(&machine).Error
	if err != nil {
		// 机器未注册 - 引导进入Discovery模式
		return h.renderDiscoveryMode(c, macAddr)
	}

	// 根据机器状态决定启动模式
	bootMode := h.determineBootMode(&machine)

	scriptData := iPXEBootScript{
		ServerURL:  h.serverURL,
		MachineID:  machine.ID,
		MacAddress: machine.MacAddress,
		Hostname:   machine.Hostname,
		BootMode:   bootMode,
	}

	// 如果是安装模式，加载OS配置
	if bootMode == "install" {
		// 查询待执行的安装任务
		var job models.Job
		err := database.DB.Where("machine_id = ? AND type = ? AND status = ?",
			machine.ID, "install_os", models.JobStatusPending).First(&job).Error

		if err == nil && job.ProfileID != "" {
			// 加载OS Profile
			var profile models.OSProfile
			if err := database.DB.First(&profile, "id = ?", job.ProfileID).Error; err == nil {
				scriptData.OSProfile = h.buildOSProfileData(&profile)
			}
		}
	}

	// 渲染iPXE脚本模板
	c.Response().Header().Set(echo.HeaderContentType, "text/plain; charset=utf-8")
	return c.Render(http.StatusOK, "ipxe.tmpl", scriptData)
}

// determineBootMode 确定启动模式
func (h *PXEHandler) determineBootMode(machine *models.Machine) string {
	// 检查是否有待执行的安装任务
	var installJob models.Job
	err := database.DB.Where("machine_id = ? AND type = ? AND status IN (?)",
		machine.ID, "install_os", []models.JobStatus{models.JobStatusPending, models.JobStatusRunning}).
		First(&installJob).Error

	if err == nil {
		return "install"
	}

	// 如果机器状态是discovered或ready，引导到discovery模式
	if machine.Status == models.MachineStatusDiscovered || machine.Status == models.MachineStatusReady {
		return "discovery"
	}

	// 如果机器状态是active，从本地硬盘启动
	if machine.Status == models.MachineStatusActive {
		return "localboot"
	}

	// 默认discovery模式
	return "discovery"
}

// renderDiscoveryMode 渲染Discovery模式（未注册机器）
func (h *PXEHandler) renderDiscoveryMode(c echo.Context, macAddr string) error {
	scriptData := iPXEBootScript{
		ServerURL:  h.serverURL,
		MachineID:  "unknown",
		MacAddress: macAddr,
		Hostname:   "unknown-" + macAddr[len(macAddr)-8:],
		BootMode:   "discovery",
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/plain; charset=utf-8")
	return c.Render(http.StatusOK, "ipxe.tmpl", scriptData)
}

// buildOSProfileData 构建OS配置数据
func (h *PXEHandler) buildOSProfileData(profile *models.OSProfile) *OSProfileData {
	data := &OSProfileData{
		Distro:  profile.Distro,
		Version: profile.Version,
	}

	// 根据发行版设置默认URL
	switch profile.Distro {
	case "centos7":
		data.KernelURL = fmt.Sprintf("%s/images/centos7/vmlinuz", h.serverURL)
		data.InitrdURL = fmt.Sprintf("%s/images/centos7/initrd.img", h.serverURL)
		data.RepoURL = "http://mirror.centos.org/centos/7/os/x86_64/"

	case "ubuntu22":
		data.KernelURL = fmt.Sprintf("%s/images/ubuntu22/vmlinuz", h.serverURL)
		data.InitrdURL = fmt.Sprintf("%s/images/ubuntu22/initrd", h.serverURL)
		data.RepoURL = "http://archive.ubuntu.com/ubuntu/dists/jammy/main/installer-amd64/"

	case "rocky8":
		data.KernelURL = fmt.Sprintf("%s/images/rocky8/vmlinuz", h.serverURL)
		data.InitrdURL = fmt.Sprintf("%s/images/rocky8/initrd.img", h.serverURL)
		data.RepoURL = "https://download.rockylinux.org/pub/rocky/8/BaseOS/x86_64/os/"

	case "suse15":
		data.KernelURL = fmt.Sprintf("%s/images/suse15/linux", h.serverURL)
		data.InitrdURL = fmt.Sprintf("%s/images/suse15/initrd", h.serverURL)
		data.RepoURL = "https://download.opensuse.org/distribution/leap/15.5/repo/oss/"

	default:
		// 使用profile中的自定义URL（如果有）
		data.KernelURL = fmt.Sprintf("%s/images/%s/vmlinuz", h.serverURL, profile.Distro)
		data.InitrdURL = fmt.Sprintf("%s/images/%s/initrd.img", h.serverURL, profile.Distro)
		data.RepoURL = ""
	}

	return data
}

// normalizeMACAddress 规范化MAC地址格式
func normalizeMACAddress(mac string) string {
	// 移除所有分隔符
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	mac = strings.ToLower(mac)

	// 添加冒号分隔符 (aa:bb:cc:dd:ee:ff)
	if len(mac) == 12 {
		return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
			mac[0:2], mac[2:4], mac[4:6], mac[6:8], mac[8:10], mac[10:12])
	}

	return mac
}

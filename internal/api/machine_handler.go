package api

import (
	"net/http"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MachineHandler 机器管理API处理器
type MachineHandler struct{}

// NewMachineHandler 创建MachineHandler
func NewMachineHandler() *MachineHandler {
	return &MachineHandler{}
}

// ListMachines 获取机器列表
// GET /api/v1/machines
func (h *MachineHandler) ListMachines(c echo.Context) error {
	db := database.GetDB()

	// 查询参数
	status := c.QueryParam("status")
	page := c.QueryParam("page")
	pageSize := c.QueryParam("page_size")

	// 构建查询
	query := db.Model(&models.Machine{})

	// 按状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to count machines",
		})
	}

	// 分页
	if page != "" && pageSize != "" {
		// TODO: 实现分页逻辑
	}

	// 查询机器列表
	var machines []models.Machine
	if err := query.Order("created_at DESC").Find(&machines).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to query machines",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total": total,
		"items": machines,
	})
}

// GetMachine 获取单个机器详情
// GET /api/v1/machines/:id
func (h *MachineHandler) GetMachine(c echo.Context) error {
	db := database.GetDB()
	id := c.Param("id")

	var machine models.Machine
	if err := db.Where("id = ?", id).First(&machine).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Machine not found",
		})
	}

	return c.JSON(http.StatusOK, machine)
}

// CreateMachine 创建/纳管机器
// POST /api/v1/machines
func (h *MachineHandler) CreateMachine(c echo.Context) error {
	db := database.GetDB()

	// 解析请求体
	var req struct {
		Mac      string `json:"mac" validate:"required"`
		Hostname string `json:"hostname" validate:"required"`
		IP       string `json:"ip"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// 检查MAC地址是否已存在
	var existing models.Machine
	if err := db.Where("mac_address = ?", req.Mac).First(&existing).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{
			"error": "Machine with this MAC address already exists",
			"id":    existing.ID,
		})
	}

	// 创建机器记录
	machine := models.Machine{
		ID:         uuid.New().String(),
		Hostname:   req.Hostname,
		MacAddress: req.Mac,
		IPAddress:  req.IP,
		Status:     models.MachineStatusDiscovered,
		HardwareSpec: models.HardwareInfo{
			SchemaVersion: "1.0",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create machine",
		})
	}

	return c.JSON(http.StatusCreated, machine)
}

// UpdateMachine 更新机器信息
// PUT /api/v1/machines/:id
func (h *MachineHandler) UpdateMachine(c echo.Context) error {
	db := database.GetDB()
	id := c.Param("id")

	// 查询机器
	var machine models.Machine
	if err := db.Where("id = ?", id).First(&machine).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Machine not found",
		})
	}

	// 解析更新数据
	var req struct {
		Hostname *string                `json:"hostname"`
		Status   *models.MachineStatus  `json:"status"`
		Hardware *models.HardwareInfo   `json:"hardware"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// 更新字段
	if req.Hostname != nil {
		machine.Hostname = *req.Hostname
	}
	if req.Status != nil {
		machine.Status = *req.Status
	}
	if req.Hardware != nil {
		machine.HardwareSpec = *req.Hardware
	}

	machine.UpdatedAt = time.Now()

	if err := db.Save(&machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to update machine",
		})
	}

	return c.JSON(http.StatusOK, machine)
}

// DeleteMachine 删除/下架机器
// DELETE /api/v1/machines/:id
func (h *MachineHandler) DeleteMachine(c echo.Context) error {
	db := database.GetDB()
	id := c.Param("id")

	// 检查机器是否存在
	var machine models.Machine
	if err := db.Where("id = ?", id).First(&machine).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Machine not found",
		})
	}

	// 删除机器
	if err := db.Delete(&machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to delete machine",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// ProvisionMachine 触发安装任务
// POST /api/v1/machines/:id/provision
func (h *MachineHandler) ProvisionMachine(c echo.Context) error {
	db := database.GetDB()
	machineID := c.Param("id")

	// 检查机器是否存在
	var machine models.Machine
	if err := db.Where("id = ?", machineID).First(&machine).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Machine not found",
		})
	}

	// 解析请求
	var req struct {
		ProfileID string                 `json:"profile_id" validate:"required"`
		Config    map[string]interface{} `json:"config"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// 创建Job任务
	job := models.Job{
		ID:          uuid.New().String(),
		MachineID:   machineID,
		Type:        models.JobTypeInstallOS,
		Status:      models.JobStatusPending,
		StepCurrent: "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(&job).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create job",
		})
	}

	// 更新机器状态
	machine.Status = models.MachineStatusInstalling
	machine.UpdatedAt = time.Now()
	db.Save(&machine)

	return c.JSON(http.StatusAccepted, job)
}

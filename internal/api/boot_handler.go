package api

import (
	"net/http"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/core/logbroker"
	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BootHandler Boot API处理器（Agent专用）
type BootHandler struct {
	broker *logbroker.Broker
}

// NewBootHandler 创建BootHandler
func NewBootHandler(broker *logbroker.Broker) *BootHandler {
	return &BootHandler{
		broker: broker,
	}
}

// RegisterAgent Agent上线注册/心跳
// POST /api/boot/v1/register
func (h *BootHandler) RegisterAgent(c echo.Context) error {
	db := database.GetDB()

	// 解析请求
	var req struct {
		Mac         string                `json:"mac" validate:"required"`
		IP          string                `json:"ip"`
		Fingerprint *models.HardwareInfo  `json:"fingerprint"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// 查询或创建机器
	var machine models.Machine
	err := db.Where("mac_address = ?", req.Mac).First(&machine).Error

	if err != nil {
		// 机器不存在，创建新机器（自动发现）
		machine = models.Machine{
			ID:         uuid.New().String(),
			Hostname:   "discovered-" + req.Mac[len(req.Mac)-8:], // 使用MAC后8位作为临时主机名
			MacAddress: req.Mac,
			IPAddress:  req.IP,
			Status:     models.MachineStatusDiscovered,
			HardwareSpec: models.HardwareInfo{
				SchemaVersion: "1.0",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 如果提供了硬件指纹，更新
		if req.Fingerprint != nil {
			machine.HardwareSpec = *req.Fingerprint
		}

		if err := db.Create(&machine).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": "Failed to create machine",
			})
		}
	} else {
		// 机器已存在，更新心跳和IP
		machine.IPAddress = req.IP
		machine.UpdatedAt = time.Now()

		// 如果提供了新的硬件指纹，更新
		if req.Fingerprint != nil {
			machine.HardwareSpec = *req.Fingerprint
		}

		if err := db.Save(&machine).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": "Failed to update machine",
			})
		}
	}

	// 查询是否有待执行的任务
	var pendingJob models.Job
	taskID := ""
	err = db.Where("machine_id = ? AND status = ?", machine.ID, models.JobStatusPending).
		First(&pendingJob).Error

	if err == nil {
		// 有待执行任务
		taskID = pendingJob.ID
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "ok",
		"machine_id": machine.ID,
		"task_id":    taskID,
	})
}

// GetTask Agent轮询任务
// GET /api/boot/v1/task
func (h *BootHandler) GetTask(c echo.Context) error {
	db := database.GetDB()
	mac := c.QueryParam("mac")

	if mac == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "MAC address required",
		})
	}

	// 查询机器
	var machine models.Machine
	if err := db.Where("mac_address = ?", mac).First(&machine).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Machine not found",
		})
	}

	// 查询待执行任务
	var job models.Job
	err := db.Where("machine_id = ? AND status = ?", machine.ID, models.JobStatusPending).
		First(&job).Error

	if err != nil {
		// 无任务
		return c.JSON(http.StatusOK, map[string]interface{}{
			"task_id": nil,
			"message": "No task available",
		})
	}

	// 返回任务规范（简化版，真实实现需要根据任务类型生成CSPM配置）
	taskSpec := map[string]interface{}{
		"task_id": job.ID,
		"action":  "probe", // TODO: 根据Job类型动态生成
		"provider_url": "",
		"session_key":  "",
		"config":       map[string]interface{}{},
	}

	// 更新任务状态为Running
	job.Status = models.JobStatusRunning
	job.StepCurrent = "agent_accepted"
	job.UpdatedAt = time.Now()
	db.Save(&job)

	return c.JSON(http.StatusOK, taskSpec)
}

// UploadLogs Agent上报日志
// POST /api/boot/v1/logs
func (h *BootHandler) UploadLogs(c echo.Context) error {
	var req struct {
		TaskID string `json:"task_id"`
		JobID  string `json:"job_id"`
		Logs   []struct {
			Timestamp string `json:"ts"`
			Level     string `json:"level"`
			Component string `json:"component"`
			Message   string `json:"msg"`
		} `json:"logs"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// 使用JobID或TaskID（优先JobID）
	jobID := req.JobID
	if jobID == "" {
		jobID = req.TaskID
	}

	if jobID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "job_id or task_id required",
		})
	}

	// 转发日志到LogBroker
	for _, log := range req.Logs {
		// 解析时间戳
		var timestamp time.Time
		var err error

		// 尝试多种时间格式
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.999Z",
		}

		for _, format := range formats {
			timestamp, err = time.Parse(format, log.Timestamp)
			if err == nil {
				break
			}
		}

		// 如果解析失败，使用当前时间
		if err != nil {
			timestamp = time.Now()
		}

		// 发布到LogBroker
		h.broker.Publish(jobID, logbroker.LogMessage{
			Timestamp: timestamp,
			Level:     log.Level,
			Component: log.Component,
			Message:   log.Message,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":        "ok",
		"logs_received": len(req.Logs),
	})
}

// ReportStatus Agent上报任务状态
// POST /api/boot/v1/status
func (h *BootHandler) ReportStatus(c echo.Context) error {
	db := database.GetDB()

	var req struct {
		TaskID   string `json:"task_id" validate:"required"`
		Status   string `json:"status" validate:"required"` // success, failed
		ErrorMsg string `json:"error_msg"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// 查询任务
	var job models.Job
	if err := db.Where("id = ?", req.TaskID).First(&job).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Task not found",
		})
	}

	// 更新任务状态
	if req.Status == "success" {
		job.SetSuccess()
	} else {
		job.Error = req.ErrorMsg
		job.Status = models.JobStatusFailed
	}

	job.UpdatedAt = time.Now()

	if err := db.Save(&job).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to update job status",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}

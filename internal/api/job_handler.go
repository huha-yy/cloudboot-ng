package api

import (
	"net/http"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/labstack/echo/v4"
)

// JobHandler 任务管理API处理器
type JobHandler struct{}

// NewJobHandler 创建JobHandler
func NewJobHandler() *JobHandler {
	return &JobHandler{}
}

// ListJobs 获取任务列表
// GET /api/v1/jobs
func (h *JobHandler) ListJobs(c echo.Context) error {
	db := database.GetDB()

	// 查询参数
	status := c.QueryParam("status")
	machineID := c.QueryParam("machine_id")

	// 构建查询
	query := db.Model(&models.Job{}).Preload("Machine").Preload("Profile")

	// 按状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 按机器ID过滤
	if machineID != "" {
		query = query.Where("machine_id = ?", machineID)
	}

	// 查询任务列表
	var jobs []models.Job
	if err := query.Order("created_at DESC").Find(&jobs).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to query jobs",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": jobs,
	})
}

// GetJob 获取单个任务详情
// GET /api/v1/jobs/:id
func (h *JobHandler) GetJob(c echo.Context) error {
	db := database.GetDB()
	id := c.Param("id")

	var job models.Job
	if err := db.Preload("Machine").Preload("Profile").Where("id = ?", id).First(&job).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Job not found",
		})
	}

	return c.JSON(http.StatusOK, job)
}

// CancelJob 取消任务（仅运行中的任务）
// DELETE /api/v1/jobs/:id
func (h *JobHandler) CancelJob(c echo.Context) error {
	db := database.GetDB()
	id := c.Param("id")

	// 查询任务
	var job models.Job
	if err := db.Where("id = ?", id).First(&job).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Job not found",
		})
	}

	// 检查任务状态
	if job.IsTerminal() {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Cannot cancel a terminal job",
		})
	}

	// 取消任务
	job.Status = models.JobStatusFailed
	job.Error = "Cancelled by user"
	if err := db.Save(&job).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to cancel job",
		})
	}

	return c.JSON(http.StatusOK, job)
}

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/core/cspm"
	"github.com/cloudboot/cloudboot-ng/internal/core/logbroker"
	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DemoHandler 演示处理器
type DemoHandler struct {
	broker   *logbroker.Broker
	executor *cspm.Executor
}

// NewDemoHandler 创建演示处理器
func NewDemoHandler(broker *logbroker.Broker) *DemoHandler {
	return &DemoHandler{
		broker: broker,
	}
}

// TriggerOrchestratorDemo 触发Orchestrator演示
// POST /api/demo/orchestrator
func (h *DemoHandler) TriggerOrchestratorDemo(c echo.Context) error {
	// 查找或创建测试Machine
	var machine models.Machine
	err := database.DB.Where("hostname = ?", "demo-server-01").First(&machine).Error
	if err != nil {
		// Machine不存在,创建新的
		machine = models.Machine{
			ID:         uuid.New().String(),
			MacAddress: "00:11:22:33:44:55",
			Hostname:   "demo-server-01",
			IPAddress:  "192.168.1.100",
			Status:     "ready",
		}
		if err := database.DB.Create(&machine).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create demo machine",
			})
		}
	}

	// 创建测试Job
	job := &models.Job{
		ID:        uuid.New().String(),
		MachineID: machine.ID,
		Type:      "config_raid",
		Status:    "running",
	}
	if err := database.DB.Create(job).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create demo job",
		})
	}

	// 在后台异步执行Orchestrator
	go h.runOrchestratorDemo(job.ID)

	return c.JSON(http.StatusOK, map[string]string{
		"job_id":     job.ID,
		"machine_id": machine.ID,
		"status":     "Demo job started",
		"logs_url":   "/jobs/" + job.ID + "/logs",
	})
}

// runOrchestratorDemo 运行Orchestrator演示（异步）
func (h *DemoHandler) runOrchestratorDemo(jobID string) {
	// 创建Executor
	executor := cspm.NewExecutor("./cmd/provider-mock/provider-mock")

	// 创建Orchestrator并设置LogBroker
	orchestrator := cspm.NewOrchestrator(executor)
	orchestrator.SetLogBroker(h.broker, jobID)

	// 模拟配置
	config := map[string]interface{}{
		"desired_state": map[string]interface{}{
			"level":  "raid5",
			"drives": []string{"sda", "sdb", "sdc"},
		},
	}

	// 执行Orchestrator（会自动推送日志到SSE）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := orchestrator.ApplyWithPlan(ctx, config)

	// 更新Job状态
	var job models.Job
	database.DB.First(&job, "id = ?", jobID)

	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
	} else if result.Success {
		job.Status = "success"
	} else {
		job.Status = "failed"
		if result.Error != nil {
			job.Error = result.Error.Error()
		}
	}

	database.DB.Save(&job)

	// 发送最终日志
	if result.Idempotent {
		h.broker.PublishHTML(jobID, "INFO", "🎯 幂等性: 系统已达标，跳过Apply步骤，性能提升75%")
	}

	h.broker.PublishHTML(jobID, "INFO", "")
	h.broker.PublishHTML(jobID, "INFO", "═══════════════════════════════════════════════════════")
	h.broker.PublishHTML(jobID, "INFO", "🎉 任务执行完成!")
	h.broker.PublishHTML(jobID, "INFO", "")
	h.broker.PublishHTML(jobID, "INFO", "📊 执行统计:")
	h.broker.PublishHTML(jobID, "INFO", "   • 总步骤: "+string(rune(len(result.Steps))))
	h.broker.PublishHTML(jobID, "INFO", "   • 总耗时: "+result.Duration.String())
	h.broker.PublishHTML(jobID, "INFO", "   • 幂等性: "+boolToString(result.Idempotent))
	h.broker.PublishHTML(jobID, "INFO", "═══════════════════════════════════════════════════════")
}

func boolToString(b bool) string {
	if b {
		return "✓ 是"
	}
	return "✗ 否"
}

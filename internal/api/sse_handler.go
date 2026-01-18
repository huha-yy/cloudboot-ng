package api

import (
	"fmt"
	"net/http"

	"github.com/cloudboot/cloudboot-ng/internal/core/logbroker"
	"github.com/labstack/echo/v4"
)

// SSEHandler SSE日志流处理器
type SSEHandler struct {
	broker *logbroker.Broker
}

// NewSSEHandler 创建新的SSE Handler
func NewSSEHandler(broker *logbroker.Broker) *SSEHandler {
	return &SSEHandler{
		broker: broker,
	}
}

// StreamLogs SSE日志流端点
// GET /api/stream/logs/:job_id
func (h *SSEHandler) StreamLogs(c echo.Context) error {
	jobID := c.Param("job_id")
	if jobID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "job_id is required",
		})
	}

	// 设置SSE响应头
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no") // Nginx禁用缓冲

	// 订阅日志流
	logChan := h.broker.Subscribe(jobID)
	defer func() {
		// 客户端断开时取消订阅
		h.broker.Unsubscribe(jobID, logChan)
	}()

	// 发送初始连接成功事件
	fmt.Fprintf(c.Response(), "event: connected\ndata: {\"job_id\":\"%s\",\"status\":\"streaming\"}\n\n", jobID)
	c.Response().Flush()

	// 持续推送日志
	for {
		select {
		case msg, ok := <-logChan:
			if !ok {
				// Channel关闭，结束流
				return nil
			}

			// 发送日志事件 (HTML格式)
			// 使用 "message" 事件名以匹配 HTMX sse-swap="message"
			htmlLog := msg.FormatAsHTML()
			fmt.Fprintf(c.Response(), "event: message\ndata: %s\n\n", htmlLog)
			c.Response().Flush()

		case <-c.Request().Context().Done():
			// 客户端断开连接
			return nil
		}
	}
}

// StreamLogsJSON SSE日志流端点 (JSON格式)
// GET /api/stream/logs/:job_id/json
func (h *SSEHandler) StreamLogsJSON(c echo.Context) error {
	jobID := c.Param("job_id")
	if jobID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "job_id is required",
		})
	}

	// 设置SSE响应头
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	// 订阅日志流
	logChan := h.broker.Subscribe(jobID)
	defer func() {
		h.broker.Unsubscribe(jobID, logChan)
	}()

	// 发送初始连接事件
	fmt.Fprintf(c.Response(), "event: connected\ndata: {\"job_id\":\"%s\",\"status\":\"streaming\"}\n\n", jobID)
	c.Response().Flush()

	// 持续推送日志 (JSON格式)
	for {
		select {
		case msg, ok := <-logChan:
			if !ok {
				return nil
			}

			// 发送JSON格式的日志
			jsonData := fmt.Sprintf(
				`{"ts":"%s","level":"%s","component":"%s","msg":"%s"}`,
				msg.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
				msg.Level,
				msg.Component,
				msg.Message,
			)
			fmt.Fprintf(c.Response(), "event: log\ndata: %s\n\n", jsonData)
			c.Response().Flush()

		case <-c.Request().Context().Done():
			return nil
		}
	}
}

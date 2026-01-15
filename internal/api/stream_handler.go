package api

import (
	"fmt"
	"net/http"

	"github.com/cloudboot/cloudboot-ng/internal/core/logbroker"
	"github.com/labstack/echo/v4"
)

// StreamHandler SSE流处理器
type StreamHandler struct {
	broker *logbroker.Broker
}

// NewStreamHandler 创建StreamHandler
func NewStreamHandler(broker *logbroker.Broker) *StreamHandler {
	return &StreamHandler{
		broker: broker,
	}
}

// StreamLogs 实时日志流 (Server-Sent Events)
// GET /api/stream/logs/:job_id
func (h *StreamHandler) StreamLogs(c echo.Context) error {
	jobID := c.Param("job_id")

	// 设置SSE响应头
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no") // 禁用Nginx缓冲

	c.Response().WriteHeader(http.StatusOK)

	// 订阅日志
	logChan := h.broker.Subscribe(jobID)
	defer h.broker.Unsubscribe(jobID, logChan)

	// 发送初始连接消息
	initialMsg := `<div class="text-emerald-500">📡 Connected to log stream...</div>`
	fmt.Fprintf(c.Response(), "data: %s\n\n", initialMsg)
	c.Response().Flush()

	// 监听日志消息
	for {
		select {
		case msg, ok := <-logChan:
			if !ok {
				// channel关闭
				return nil
			}

			// 格式化为HTML并发送
			html := msg.FormatAsHTML()
			fmt.Fprintf(c.Response(), "data: %s\n\n", html)
			c.Response().Flush()

		case <-c.Request().Context().Done():
			// 客户端断开连接
			return nil
		}
	}
}

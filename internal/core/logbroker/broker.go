package logbroker

import (
	"fmt"
	"sync"
	"time"
)

// LogMessage 日志消息
type LogMessage struct {
	Timestamp time.Time              `json:"ts"`
	Level     string                 `json:"level"` // DEBUG, INFO, WARN, ERROR
	Component string                 `json:"component"`
	Message   string                 `json:"msg"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Broker 日志代理，管理多个Job的日志流
type Broker struct {
	mu sync.RWMutex
	// jobID -> []chan LogMessage
	subscribers map[string][]chan LogMessage
	// jobID -> []LogMessage (历史日志)
	history map[string][]LogMessage
}

// NewBroker 创建新的Broker
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string][]chan LogMessage),
		history:     make(map[string][]LogMessage),
	}
}

// Subscribe 订阅指定Job的日志流
// 返回一个channel，客户端从中接收日志消息
func (b *Broker) Subscribe(jobID string) <-chan LogMessage {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 创建缓冲channel
	ch := make(chan LogMessage, 100)

	// 添加到订阅列表
	if _, ok := b.subscribers[jobID]; !ok {
		b.subscribers[jobID] = make([]chan LogMessage, 0)
	}
	b.subscribers[jobID] = append(b.subscribers[jobID], ch)

	// 立即发送历史日志
	if history, ok := b.history[jobID]; ok {
		go func() {
			for _, msg := range history {
				select {
				case ch <- msg:
				default:
					// channel满了，跳过
				}
			}
		}()
	}

	return ch
}

// Unsubscribe 取消订阅
func (b *Broker) Unsubscribe(jobID string, ch <-chan LogMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.subscribers[jobID]; ok {
		// 找到并移除channel
		for i, subscriber := range subs {
			if subscriber == ch {
				b.subscribers[jobID] = append(subs[:i], subs[i+1:]...)
				close(subscriber)
				break
			}
		}

		// 如果没有订阅者了，清理
		if len(b.subscribers[jobID]) == 0 {
			delete(b.subscribers, jobID)
		}
	}
}

// Publish 发布日志消息到指定Job的所有订阅者
func (b *Broker) Publish(jobID string, msg LogMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 保存到历史
	if _, ok := b.history[jobID]; !ok {
		b.history[jobID] = make([]LogMessage, 0, 1000)
	}
	b.history[jobID] = append(b.history[jobID], msg)

	// 限制历史日志大小
	if len(b.history[jobID]) > 1000 {
		b.history[jobID] = b.history[jobID][len(b.history[jobID])-1000:]
	}

	// 发送给所有订阅者
	if subs, ok := b.subscribers[jobID]; ok {
		for _, ch := range subs {
			select {
			case ch <- msg:
			default:
				// channel满了，跳过（避免阻塞）
			}
		}
	}
}

// PublishHTML 发布HTML格式的日志（用于SSE）
func (b *Broker) PublishHTML(jobID string, level string, message string) {
	msg := LogMessage{
		Timestamp: time.Now(),
		Level:     level,
		Component: "core",
		Message:   message,
	}
	b.Publish(jobID, msg)
}

// GetHistory 获取指定Job的历史日志
func (b *Broker) GetHistory(jobID string) []LogMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if history, ok := b.history[jobID]; ok {
		// 返回副本
		result := make([]LogMessage, len(history))
		copy(result, history)
		return result
	}

	return []LogMessage{}
}

// ClearHistory 清理指定Job的历史日志
func (b *Broker) ClearHistory(jobID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.history, jobID)
	// 不关闭active subscribers
}

// FormatAsHTML 将LogMessage格式化为HTML片段
func (msg *LogMessage) FormatAsHTML() string {
	// 根据日志级别选择颜色
	colorClass := "text-slate-300"
	switch msg.Level {
	case "ERROR":
		colorClass = "text-rose-500"
	case "WARN":
		colorClass = "text-amber-500"
	case "INFO":
		colorClass = "text-emerald-500"
	case "DEBUG":
		colorClass = "text-slate-500"
	}

	timestamp := msg.Timestamp.Format("15:04:05")
	return fmt.Sprintf(`<div class="%s">[%s] [%s] %s</div>`, colorClass, timestamp, msg.Level, msg.Message)
}

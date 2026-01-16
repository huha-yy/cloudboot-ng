package monitor

import (
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	startTime time.Time
	once      sync.Once
)

// Init 初始化监控模块，记录启动时间
func Init() {
	once.Do(func() {
		startTime = time.Now()
	})
}

// SystemStats 系统监控统计数据
type SystemStats struct {
	Uptime       string  `json:"uptime"`
	UptimeRaw    int64   `json:"uptime_raw"` // 秒数
	DiskUsage    float64 `json:"disk_usage"`
	DiskTotal    uint64  `json:"disk_total"`
	DiskUsed     uint64  `json:"disk_used"`
	MemoryUsage  float64 `json:"memory_usage"`
	MemoryTotal  uint64  `json:"memory_total"`
	MemoryUsed   uint64  `json:"memory_used"`
	NumGoroutine int     `json:"num_goroutine"`
}

// GetStats 获取系统监控统计
func GetStats() *SystemStats {
	stats := &SystemStats{
		NumGoroutine: runtime.NumGoroutine(),
	}

	// 计算运行时间
	if !startTime.IsZero() {
		uptime := time.Since(startTime)
		stats.UptimeRaw = int64(uptime.Seconds())
		stats.Uptime = formatDuration(uptime)
	} else {
		stats.Uptime = "0s"
	}

	// 获取磁盘使用率 (根目录)
	if diskStat, err := disk.Usage("/"); err == nil {
		stats.DiskUsage = diskStat.UsedPercent
		stats.DiskTotal = diskStat.Total
		stats.DiskUsed = diskStat.Used
	}

	// 获取内存使用率
	if memStat, err := mem.VirtualMemory(); err == nil {
		stats.MemoryUsage = memStat.UsedPercent
		stats.MemoryTotal = memStat.Total
		stats.MemoryUsed = memStat.Used
	}

	return stats
}

// GetUptime 获取系统运行时间
func GetUptime() string {
	if startTime.IsZero() {
		return "0s"
	}
	return formatDuration(time.Since(startTime))
}

// GetUptimeSeconds 获取运行秒数
func GetUptimeSeconds() int64 {
	if startTime.IsZero() {
		return 0
	}
	return int64(time.Since(startTime).Seconds())
}

// formatDuration 格式化时间段为人类可读格式
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return formatTime(days, "d", hours, "h")
	}
	if hours > 0 {
		return formatTime(hours, "h", minutes, "m")
	}
	return formatTime(minutes, "m", int(d.Seconds())%60, "s")
}

func formatTime(v1 int, u1 string, v2 int, u2 string) string {
	if v2 > 0 {
		return itoa(v1) + u1 + " " + itoa(v2) + u2
	}
	return itoa(v1) + u1
}

func itoa(i int) string {
	if i < 10 {
		return string(rune('0'+i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}

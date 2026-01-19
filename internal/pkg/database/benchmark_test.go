package database

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// BenchmarkResult 压测结果
type BenchmarkResult struct {
	TotalOps       int           // 总操作数
	SuccessOps     int           // 成功操作数
	FailedOps      int           // 失败操作数
	Duration       time.Duration // 总耗时
	OpsPerSecond   float64       // 每秒操作数
	AvgLatency     time.Duration // 平均延迟
	MaxLatency     time.Duration // 最大延迟
	MinLatency     time.Duration // 最小延迟
	ConcurrentGR   int           // 并发数
	WALModeEnabled bool          // WAL模式是否启用
}

// TestConcurrentWrites 测试并发写入
func TestConcurrentWrites(t *testing.T) {
	// 创建测试数据库
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	concurrentGoroutines := []int{1, 10, 50, 100}

	for _, concurrent := range concurrentGoroutines {
		t.Run(fmt.Sprintf("Concurrent_%d", concurrent), func(t *testing.T) {
			result := benchmarkConcurrentWrites(t, db, concurrent, 100)

			t.Logf("📊 并发写入测试结果 (并发数: %d)", concurrent)
			t.Logf("   - 总操作数: %d", result.TotalOps)
			t.Logf("   - 成功操作: %d", result.SuccessOps)
			t.Logf("   - 失败操作: %d", result.FailedOps)
			t.Logf("   - 总耗时: %v", result.Duration)
			t.Logf("   - TPS: %.2f ops/s", result.OpsPerSecond)
			t.Logf("   - 平均延迟: %v", result.AvgLatency)
			t.Logf("   - 最大延迟: %v", result.MaxLatency)
			t.Logf("   - WAL模式: %v", result.WALModeEnabled)

			// 性能基准检查
			if result.OpsPerSecond < 100 {
				t.Errorf("⚠️  TPS too low: %.2f < 100", result.OpsPerSecond)
			} else {
				t.Logf("✅ TPS acceptable: %.2f ops/s", result.OpsPerSecond)
			}

			// 失败率检查
			failureRate := float64(result.FailedOps) / float64(result.TotalOps) * 100
			if failureRate > 5.0 {
				t.Errorf("⚠️  Failure rate too high: %.2f%%", failureRate)
			} else {
				t.Logf("✅ Failure rate acceptable: %.2f%%", failureRate)
			}
		})
	}
}

// TestConcurrentReads 测试并发读取
func TestConcurrentReads(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// 先插入测试数据
	prepareTestData(t, db, 1000)

	concurrentGoroutines := []int{10, 50, 100, 200}

	for _, concurrent := range concurrentGoroutines {
		t.Run(fmt.Sprintf("Concurrent_%d", concurrent), func(t *testing.T) {
			result := benchmarkConcurrentReads(t, db, concurrent, 100)

			t.Logf("📊 并发读取测试结果 (并发数: %d)", concurrent)
			t.Logf("   - TPS: %.2f ops/s", result.OpsPerSecond)
			t.Logf("   - 平均延迟: %v", result.AvgLatency)
			t.Logf("   - 最大延迟: %v", result.MaxLatency)

			// 读取性能应该更高
			if result.OpsPerSecond < 500 {
				t.Logf("⚠️  Read TPS: %.2f (could be improved)", result.OpsPerSecond)
			} else {
				t.Logf("✅ Read TPS excellent: %.2f ops/s", result.OpsPerSecond)
			}
		})
	}
}

// TestMixedReadWrite 测试混合读写
func TestMixedReadWrite(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// 准备初始数据
	prepareTestData(t, db, 100)

	// 80% 读取, 20% 写入（真实场景）
	result := benchmarkMixedReadWrite(t, db, 50, 1000, 0.2)

	t.Logf("📊 混合读写测试结果")
	t.Logf("   - 总操作数: %d", result.TotalOps)
	t.Logf("   - TPS: %.2f ops/s", result.OpsPerSecond)
	t.Logf("   - 平均延迟: %v", result.AvgLatency)
	t.Logf("   - WAL模式: %v", result.WALModeEnabled)

	if result.FailedOps > 0 {
		t.Logf("⚠️  Failed operations: %d", result.FailedOps)
	} else {
		t.Log("✅ No failed operations")
	}
}

// TestWALMode 验证WAL模式
func TestWALMode(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// 检查journal_mode
	var journalMode string
	sqlDB, _ := db.DB()
	row := sqlDB.QueryRow("PRAGMA journal_mode")
	row.Scan(&journalMode)

	t.Logf("📋 Journal Mode: %s", journalMode)

	if journalMode != "wal" {
		t.Errorf("❌ WAL mode not enabled: %s", journalMode)
	} else {
		t.Log("✅ WAL mode enabled")
	}

	// 检查WAL相关配置
	var walAutocheckpoint int
	sqlDB.QueryRow("PRAGMA wal_autocheckpoint").Scan(&walAutocheckpoint)
	t.Logf("   - WAL autocheckpoint: %d", walAutocheckpoint)

	var synchronous string
	sqlDB.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	t.Logf("   - Synchronous mode: %s", synchronous)
}

// TestStressTest 压力测试：模拟100个Agent同时注册
func TestStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试 (使用 -short 标志)")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	t.Log("🔥 压力测试: 100个并发Agent，每个Agent注册10次")

	result := benchmarkConcurrentWrites(t, db, 100, 10)

	t.Logf("📊 压力测试结果")
	t.Logf("   - 总Machine注册数: %d", result.TotalOps)
	t.Logf("   - 成功: %d", result.SuccessOps)
	t.Logf("   - 失败: %d", result.FailedOps)
	t.Logf("   - 总耗时: %v", result.Duration)
	t.Logf("   - TPS: %.2f ops/s", result.OpsPerSecond)
	t.Logf("   - 平均延迟: %v", result.AvgLatency)
	t.Logf("   - 最大延迟: %v", result.MaxLatency)

	// 压力测试基准
	if result.FailedOps > result.TotalOps/10 {
		t.Errorf("❌ Too many failures: %d/%d", result.FailedOps, result.TotalOps)
	}

	if result.OpsPerSecond < 50 {
		t.Logf("⚠️  Low TPS under stress: %.2f", result.OpsPerSecond)
	} else {
		t.Logf("✅ Acceptable TPS under stress: %.2f", result.OpsPerSecond)
	}
}

// benchmarkConcurrentWrites 并发写入基准测试
func benchmarkConcurrentWrites(t *testing.T, db *gorm.DB, concurrent int, opsPerGoroutine int) *BenchmarkResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var atomicCounter int64 // 原子计数器确保唯一性

	totalOps := concurrent * opsPerGoroutine
	successCount := 0
	failedCount := 0
	latencies := make([]time.Duration, 0, totalOps)
	errors := make(map[string]int) // 记录错误类型

	startTime := time.Now()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				// 使用原子计数器确保唯一性
				counter := atomicInc(&atomicCounter)
				machineID := uuid.New().String()

				machine := &models.Machine{
					ID:         machineID,
					Hostname:   fmt.Sprintf("stress-%s", machineID[:8]),
					MacAddress: fmt.Sprintf("02:%02x:%02x:%02x:%02x:%02x",
						(counter>>24)&0xFF, (counter>>16)&0xFF, (counter>>8)&0xFF, counter&0xFF, workerID),
					IPAddress:  fmt.Sprintf("10.%d.%d.%d", (counter>>16)&0xFF, (counter>>8)&0xFF, counter&0xFF),
					Status:     "ready",
				}

				opStart := time.Now()
				err := db.Create(machine).Error
				latency := time.Since(opStart)

				mu.Lock()
				latencies = append(latencies, latency)
				if err != nil {
					failedCount++
					errors[err.Error()]++
					if failedCount <= 5 { // 只打印前5个错误
						t.Logf("❌ Write error: %v", err)
					}
				} else {
					successCount++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 打印错误统计
	if len(errors) > 0 {
		t.Logf("📋 Error types:")
		for errMsg, count := range errors {
			t.Logf("   - %s: %d times", errMsg, count)
		}
	}

	return &BenchmarkResult{
		TotalOps:       totalOps,
		SuccessOps:     successCount,
		FailedOps:      failedCount,
		Duration:       duration,
		OpsPerSecond:   float64(totalOps) / duration.Seconds(),
		AvgLatency:     calculateAvg(latencies),
		MaxLatency:     calculateMax(latencies),
		MinLatency:     calculateMin(latencies),
		ConcurrentGR:   concurrent,
		WALModeEnabled: checkWALMode(db),
	}
}

// benchmarkConcurrentReads 并发读取基准测试
func benchmarkConcurrentReads(t *testing.T, db *gorm.DB, concurrent int, opsPerGoroutine int) *BenchmarkResult {
	var wg sync.WaitGroup
	var mu sync.Mutex

	totalOps := concurrent * opsPerGoroutine
	successCount := 0
	failedCount := 0
	latencies := make([]time.Duration, 0, totalOps)

	startTime := time.Now()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				var machines []models.Machine

				opStart := time.Now()
				err := db.Limit(10).Find(&machines).Error
				latency := time.Since(opStart)

				mu.Lock()
				latencies = append(latencies, latency)
				if err != nil {
					failedCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	return &BenchmarkResult{
		TotalOps:       totalOps,
		SuccessOps:     successCount,
		FailedOps:      failedCount,
		Duration:       duration,
		OpsPerSecond:   float64(totalOps) / duration.Seconds(),
		AvgLatency:     calculateAvg(latencies),
		MaxLatency:     calculateMax(latencies),
		MinLatency:     calculateMin(latencies),
		ConcurrentGR:   concurrent,
		WALModeEnabled: checkWALMode(db),
	}
}

// benchmarkMixedReadWrite 混合读写基准测试
func benchmarkMixedReadWrite(t *testing.T, db *gorm.DB, concurrent int, totalOps int, writeRatio float64) *BenchmarkResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var atomicCounter int64

	successCount := 0
	failedCount := 0
	latencies := make([]time.Duration, 0, totalOps)
	opsPerGoroutine := totalOps / concurrent

	startTime := time.Now()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				var err error
				opStart := time.Now()

				// 根据writeRatio决定是读还是写
				if rand.Float64() < writeRatio {
					// 写操作 - 使用原子计数器确保唯一性
					counter := atomicInc(&atomicCounter)
					machineID := uuid.New().String()
					machine := &models.Machine{
						ID:       machineID,
						Hostname: fmt.Sprintf("mixed-%s", machineID[:8]),
						MacAddress: fmt.Sprintf("02:%02x:%02x:%02x:%02x:%02x",
							(counter>>24)&0xFF, (counter>>16)&0xFF, (counter>>8)&0xFF, counter&0xFF, workerID),
						IPAddress: fmt.Sprintf("192.168.%d.%d", (counter>>8)&0xFF, counter&0xFF),
						Status:    "ready",
					}
					err = db.Create(machine).Error
				} else {
					// 读操作
					var machines []models.Machine
					err = db.Limit(10).Find(&machines).Error
				}

				latency := time.Since(opStart)

				mu.Lock()
				latencies = append(latencies, latency)
				if err != nil {
					failedCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	return &BenchmarkResult{
		TotalOps:       totalOps,
		SuccessOps:     successCount,
		FailedOps:      failedCount,
		Duration:       duration,
		OpsPerSecond:   float64(totalOps) / duration.Seconds(),
		AvgLatency:     calculateAvg(latencies),
		MaxLatency:     calculateMax(latencies),
		MinLatency:     calculateMin(latencies),
		ConcurrentGR:   concurrent,
		WALModeEnabled: checkWALMode(db),
	}
}

// Helper functions

func setupTestDB(t *testing.T) *gorm.DB {
	// 使用临时文件数据库以支持WAL模式
	dbPath := fmt.Sprintf("/tmp/cloudboot-bench-%d.db", time.Now().UnixNano())
	config := Config{
		DSN:      dbPath + "?_journal_mode=WAL",
		LogLevel: logger.Silent,
	}

	err := Init(config)
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}

	// 清理函数会在测试结束时删除数据库文件
	t.Cleanup(func() {
		Close()
		// 删除临时数据库文件
		os.Remove(dbPath)
		os.Remove(dbPath + "-shm")
		os.Remove(dbPath + "-wal")
	})

	return DB
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
	Close()
}

func prepareTestData(t *testing.T, db *gorm.DB, count int) {
	for i := 0; i < count; i++ {
		machine := &models.Machine{
			ID:         uuid.New().String(),
			Hostname:   fmt.Sprintf("test-machine-%d", i),
			MacAddress: fmt.Sprintf("00:00:00:00:%02x:%02x", i/256, i%256),
			IPAddress:  fmt.Sprintf("10.0.%d.%d", i/256, i%256),
			Status:     "ready",
		}
		db.Create(machine)
	}
	t.Logf("✅ Prepared %d test machines", count)
}

func checkWALMode(db *gorm.DB) bool {
	var journalMode string
	sqlDB, _ := db.DB()
	sqlDB.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	return journalMode == "wal"
}

func calculateAvg(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	return total / time.Duration(len(latencies))
}

func calculateMax(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	max := latencies[0]
	for _, l := range latencies {
		if l > max {
			max = l
		}
	}
	return max
}

func calculateMin(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	min := latencies[0]
	for _, l := range latencies {
		if l < min {
			min = l
		}
	}
	return min
}

// atomicInc 原子递增计数器
func atomicInc(counter *int64) int64 {
	return atomic.AddInt64(counter, 1)
}

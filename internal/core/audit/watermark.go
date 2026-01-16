package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WatermarkValidator validates watermarks in CBP packages
type WatermarkValidator struct {
	currentLicenseID string
	auditLogger      *AuditLogger
	mu               sync.RWMutex
}

// WatermarkViolation represents a watermark mismatch
type WatermarkViolation struct {
	Timestamp          time.Time `json:"timestamp"`
	ProviderID         string    `json:"provider_id"`
	ProviderName       string    `json:"provider_name"`
	ExpectedLicenseID  string    `json:"expected_license_id"`
	ActualLicenseID    string    `json:"actual_license_id"`
	ActualDownloaderID string    `json:"actual_downloader_id"`
	OrganizationID     string    `json:"organization_id"`
	Severity           string    `json:"severity"` // WARNING, CRITICAL
	AutoResolved       bool      `json:"auto_resolved"`
}

// NewWatermarkValidator creates a new watermark validator
func NewWatermarkValidator(licenseID string, auditLogPath string) (*WatermarkValidator, error) {
	logger, err := NewAuditLogger(auditLogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit logger: %w", err)
	}

	return &WatermarkValidator{
		currentLicenseID: licenseID,
		auditLogger:      logger,
	}, nil
}

// ValidateWatermark checks if a watermark matches the current license
func (v *WatermarkValidator) ValidateWatermark(providerID string, providerName string, watermark Watermark) (*WatermarkViolation, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// 检查watermark的license ID是否与当前系统的license ID匹配
	if watermark.LicenseID != v.currentLicenseID {
		violation := &WatermarkViolation{
			Timestamp:          time.Now(),
			ProviderID:         providerID,
			ProviderName:       providerName,
			ExpectedLicenseID:  v.currentLicenseID,
			ActualLicenseID:    watermark.LicenseID,
			ActualDownloaderID: watermark.DownloaderID,
			OrganizationID:     watermark.OrganizationID,
			Severity:           determineSeverity(watermark),
			AutoResolved:       false,
		}

		// 记录到不可删除的审计日志
		if err := v.auditLogger.LogViolation(violation); err != nil {
			return violation, fmt.Errorf("failed to log violation: %w", err)
		}

		return violation, nil
	}

	return nil, nil
}

// determineSeverity determines the severity based on watermark info
func determineSeverity(watermark Watermark) string {
	// 如果是同一组织但不同license，警告级别
	// 如果是不同组织，严重级别
	if watermark.OrganizationID != "" {
		return "WARNING" // Same org, different license
	}
	return "CRITICAL" // Different org or no org info
}

// Watermark represents the watermark structure
type Watermark struct {
	DownloaderID   string `json:"downloader_id"`
	DownloadTime   string `json:"download_time"`
	TransactionID  string `json:"transaction_id"`
	LicenseID      string `json:"license_id"`
	OrganizationID string `json:"organization_id"`
}

// AuditLogger handles immutable audit logging
type AuditLogger struct {
	logPath string
	mu      sync.Mutex
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logPath string) (*AuditLogger, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, err
	}

	return &AuditLogger{
		logPath: logPath,
	}, nil
}

// LogViolation logs a watermark violation (append-only, immutable)
func (l *AuditLogger) LogViolation(violation *WatermarkViolation) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 打开文件（追加模式，只写，不可删除）
	file, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// JSON序列化
	data, err := json.Marshal(violation)
	if err != nil {
		return err
	}

	// 写入一行JSON
	_, err = file.WriteString(string(data) + "\n")
	return err
}

// GetViolations retrieves all recorded violations
func (l *AuditLogger) GetViolations() ([]*WatermarkViolation, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := os.ReadFile(l.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*WatermarkViolation{}, nil
		}
		return nil, err
	}

	lines := splitLines(data)
	violations := make([]*WatermarkViolation, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var v WatermarkViolation
		if err := json.Unmarshal(line, &v); err != nil {
			continue // Skip invalid lines
		}

		violations = append(violations, &v)
	}

	return violations, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	var line []byte

	for _, b := range data {
		if b == '\n' {
			if len(line) > 0 {
				lines = append(lines, line)
				line = nil
			}
		} else {
			line = append(line, b)
		}
	}

	if len(line) > 0 {
		lines = append(lines, line)
	}

	return lines
}

// GetActiveViolations returns violations that haven't been resolved
func (l *AuditLogger) GetActiveViolations() ([]*WatermarkViolation, error) {
	all, err := l.GetViolations()
	if err != nil {
		return nil, err
	}

	active := make([]*WatermarkViolation, 0)
	for _, v := range all {
		if !v.AutoResolved {
			active = append(active, v)
		}
	}

	return active, nil
}

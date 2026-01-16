package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWatermarkValidation(t *testing.T) {
	tempDir := t.TempDir()
	auditLog := filepath.Join(tempDir, "audit.log")

	validator, err := NewWatermarkValidator("license-123", auditLog)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// 测试匹配的watermark（应该通过）
	validWatermark := Watermark{
		LicenseID:      "license-123",
		DownloaderID:   "user-456",
		OrganizationID: "org-789",
	}

	violation, err := validator.ValidateWatermark("prov-1", "Test Provider", validWatermark)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if violation != nil {
		t.Error("Valid watermark was flagged as violation")
	}

	// 测试不匹配的watermark（应该标记为违规）
	invalidWatermark := Watermark{
		LicenseID:      "license-999", // Different license
		DownloaderID:   "user-888",
		OrganizationID: "org-777",
	}

	violation, err = validator.ValidateWatermark("prov-2", "Suspicious Provider", invalidWatermark)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if violation == nil {
		t.Error("Invalid watermark was not flagged")
	}

	if violation.ExpectedLicenseID != "license-123" {
		t.Errorf("Expected license ID license-123, got %s", violation.ExpectedLicenseID)
	}

	if violation.ActualLicenseID != "license-999" {
		t.Errorf("Expected actual license ID license-999, got %s", violation.ActualLicenseID)
	}
}

func TestAuditLogPersistence(t *testing.T) {
	tempDir := t.TempDir()
	auditLog := filepath.Join(tempDir, "audit.log")

	validator, _ := NewWatermarkValidator("license-123", auditLog)

	// 触发违规
	invalidWatermark := Watermark{
		LicenseID:    "license-999",
		DownloaderID: "user-bad",
	}

	validator.ValidateWatermark("prov-1", "Provider A", invalidWatermark)
	validator.ValidateWatermark("prov-2", "Provider B", invalidWatermark)

	// 验证日志文件存在
	if _, err := os.Stat(auditLog); os.IsNotExist(err) {
		t.Error("Audit log file was not created")
	}

	// 读取违规记录
	violations, err := validator.auditLogger.GetViolations()
	if err != nil {
		t.Fatalf("Failed to read violations: %v", err)
	}

	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
	}

	// 验证违规内容
	if violations[0].ProviderID != "prov-1" {
		t.Errorf("Expected prov-1, got %s", violations[0].ProviderID)
	}

	if violations[1].ProviderID != "prov-2" {
		t.Errorf("Expected prov-2, got %s", violations[1].ProviderID)
	}
}

func TestSeverityDetermination(t *testing.T) {
	// 有组织ID - WARNING
	watermark1 := Watermark{
		OrganizationID: "org-123",
	}
	severity1 := determineSeverity(watermark1)
	if severity1 != "WARNING" {
		t.Errorf("Expected WARNING, got %s", severity1)
	}

	// 无组织ID - CRITICAL
	watermark2 := Watermark{
		OrganizationID: "",
	}
	severity2 := determineSeverity(watermark2)
	if severity2 != "CRITICAL" {
		t.Errorf("Expected CRITICAL, got %s", severity2)
	}
}

func TestActiveViolations(t *testing.T) {
	tempDir := t.TempDir()
	auditLog := filepath.Join(tempDir, "audit.log")

	logger, _ := NewAuditLogger(auditLog)

	// 记录两个违规
	logger.LogViolation(&WatermarkViolation{
		ProviderID:   "prov-1",
		AutoResolved: false,
	})

	logger.LogViolation(&WatermarkViolation{
		ProviderID:   "prov-2",
		AutoResolved: true,
	})

	// 获取活跃违规（未解决的）
	active, err := logger.GetActiveViolations()
	if err != nil {
		t.Fatalf("Failed to get active violations: %v", err)
	}

	if len(active) != 1 {
		t.Errorf("Expected 1 active violation, got %d", len(active))
	}

	if active[0].ProviderID != "prov-1" {
		t.Errorf("Expected prov-1, got %s", active[0].ProviderID)
	}
}

func TestEmptyAuditLog(t *testing.T) {
	tempDir := t.TempDir()
	auditLog := filepath.Join(tempDir, "nonexistent.log")

	logger, _ := NewAuditLogger(auditLog)

	violations, err := logger.GetViolations()
	if err != nil {
		t.Fatalf("Failed to get violations from empty log: %v", err)
	}

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations from empty log, got %d", len(violations))
	}
}

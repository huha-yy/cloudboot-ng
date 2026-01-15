package api

import (
	"net/http"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/core/configgen"
	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ProfileHandler Profile API处理器
type ProfileHandler struct {
	generator *configgen.Generator
}

// NewProfileHandler 创建ProfileHandler
func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{
		generator: configgen.NewGenerator(),
	}
}

// ListProfiles 查询所有OS配置模板
// GET /api/v1/profiles
func (h *ProfileHandler) ListProfiles(c echo.Context) error {
	db := database.GetDB()

	var profiles []models.OSProfile
	if err := db.Find(&profiles).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to query profiles",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"profiles": profiles,
		"total":    len(profiles),
	})
}

// GetProfile 查询单个Profile
// GET /api/v1/profiles/:id
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	db := database.GetDB()
	profileID := c.Param("id")

	var profile models.OSProfile
	if err := db.Where("id = ?", profileID).First(&profile).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Profile not found",
		})
	}

	return c.JSON(http.StatusOK, profile)
}

// CreateProfile 创建OS配置模板
// POST /api/v1/profiles
func (h *ProfileHandler) CreateProfile(c echo.Context) error {
	db := database.GetDB()

	var req models.OSProfile
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// 验证配置
	if err := h.generator.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid profile configuration",
			"details": err.Error(),
		})
	}

	// 生成UUID（如果未提供）
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	req.CreatedAt = now
	req.UpdatedAt = now

	// 保存到数据库
	if err := db.Create(&req).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create profile",
		})
	}

	return c.JSON(http.StatusCreated, req)
}

// UpdateProfile 更新Profile
// PUT /api/v1/profiles/:id
func (h *ProfileHandler) UpdateProfile(c echo.Context) error {
	db := database.GetDB()
	profileID := c.Param("id")

	// 查询现有Profile
	var profile models.OSProfile
	if err := db.Where("id = ?", profileID).First(&profile).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Profile not found",
		})
	}

	// 绑定更新数据
	var req models.OSProfile
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// 验证配置
	if err := h.generator.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid profile configuration",
			"details": err.Error(),
		})
	}

	// 保留ID和CreatedAt
	req.ID = profile.ID
	req.CreatedAt = profile.CreatedAt
	req.UpdatedAt = time.Now()

	// 更新
	if err := db.Save(&req).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to update profile",
		})
	}

	return c.JSON(http.StatusOK, req)
}

// DeleteProfile 删除Profile
// DELETE /api/v1/profiles/:id
func (h *ProfileHandler) DeleteProfile(c echo.Context) error {
	db := database.GetDB()
	profileID := c.Param("id")

	// 查询Profile
	var profile models.OSProfile
	if err := db.Where("id = ?", profileID).First(&profile).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Profile not found",
		})
	}

	// 删除
	if err := db.Delete(&profile).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to delete profile",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Profile deleted successfully",
	})
}

// PreviewConfig 预览生成的OS安装配置
// POST /api/v1/profiles/:id/preview
func (h *ProfileHandler) PreviewConfig(c echo.Context) error {
	db := database.GetDB()
	profileID := c.Param("id")

	// 查询Profile
	var profile models.OSProfile
	if err := db.Where("id = ?", profileID).First(&profile).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Profile not found",
		})
	}

	// 生成配置
	config, err := h.generator.Generate(&profile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to generate config",
			"details": err.Error(),
		})
	}

	// 返回纯文本配置
	return c.String(http.StatusOK, config)
}

// PreviewFromPayload 从请求体预览配置（不保存）
// POST /api/v1/profiles/preview
func (h *ProfileHandler) PreviewFromPayload(c echo.Context) error {
	var req models.OSProfile
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// 验证配置
	if err := h.generator.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid profile configuration",
			"details": err.Error(),
		})
	}

	// 生成配置
	config, err := h.generator.Generate(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to generate config",
			"details": err.Error(),
		})
	}

	// 返回纯文本配置
	return c.String(http.StatusOK, config)
}

package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cloudboot/cloudboot-ng/internal/core/cspm"
	"github.com/labstack/echo/v4"
)

// StoreHandler Private Store API处理器
type StoreHandler struct {
	pluginManager *cspm.PluginManager
}

// NewStoreHandler 创建StoreHandler
func NewStoreHandler(pm *cspm.PluginManager) *StoreHandler {
	return &StoreHandler{
		pluginManager: pm,
	}
}

// ImportProvider 导入Provider包
// POST /api/v1/store/import
func (h *StoreHandler) ImportProvider(c echo.Context) error {
	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "File upload required",
		})
	}

	// 检查文件扩展名
	if filepath.Ext(file.Filename) != ".cbp" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Only .cbp files are allowed",
		})
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to open uploaded file",
		})
	}
	defer src.Close()

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "provider-*.cbp")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create temp file",
		})
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // 清理临时文件

	// 保存上传的文件
	if _, err := io.Copy(tempFile, src); err != nil {
		tempFile.Close()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to save uploaded file",
		})
	}
	tempFile.Close()

	// 导入到Plugin Manager
	info, err := h.pluginManager.ImportProvider(tempPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to import provider",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"status":   "ok",
		"provider": info,
		"message":  "Provider imported successfully",
	})
}

// ListProviders 查询已安装的Provider
// GET /api/v1/store/providers
func (h *StoreHandler) ListProviders(c echo.Context) error {
	providers := h.pluginManager.ListProviders()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"providers": providers,
		"total":     len(providers),
	})
}

// GetProvider 获取单个Provider详情
// GET /api/v1/store/providers/:id
func (h *StoreHandler) GetProvider(c echo.Context) error {
	providerID := c.Param("id")

	provider, err := h.pluginManager.GetProvider(providerID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Provider not found",
		})
	}

	return c.JSON(http.StatusOK, provider)
}

// DeleteProvider 删除Provider
// DELETE /api/v1/store/providers/:id
func (h *StoreHandler) DeleteProvider(c echo.Context) error {
	providerID := c.Param("id")

	if err := h.pluginManager.DeleteProvider(providerID); err != nil {
		if err.Error() == "provider not found: "+providerID {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": "Provider not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to delete provider",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Provider deleted successfully",
	})
}

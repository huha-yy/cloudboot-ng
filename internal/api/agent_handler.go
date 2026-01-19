package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AgentHandler Agent相关API处理器
type AgentHandler struct {
	// 可以注入依赖
}

// NewAgentHandler 创建Agent处理器
func NewAgentHandler() *AgentHandler {
	return &AgentHandler{}
}

// RegisterRequest Agent注册请求
type RegisterRequest struct {
	MacAddress   string              `json:"mac_address" validate:"required"`
	IPAddress    string              `json:"ip_address"`
	Hostname     string              `json:"hostname"`
	HardwareSpec models.HardwareInfo `json:"hardware_spec" validate:"required"`
}

// RegisterResponse Agent注册响应
type RegisterResponse struct {
	MachineID     string `json:"machine_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	HeartbeatURL  string `json:"heartbeat_url"`
	TaskPollURL   string `json:"task_poll_url"`
	PollInterval  int    `json:"poll_interval_seconds"` // 心跳间隔（秒）
}

// HeartbeatRequest Agent心跳请求
type HeartbeatRequest struct {
	MachineID    string              `json:"machine_id" validate:"required"`
	MacAddress   string              `json:"mac_address" validate:"required"`
	IPAddress    string              `json:"ip_address"`
	HardwareSpec models.HardwareInfo `json:"hardware_spec"`
}

// HeartbeatResponse Agent心跳响应
type HeartbeatResponse struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	NextPoll       int    `json:"next_poll_seconds"`
	HardwareChange bool   `json:"hardware_change"` // 硬件是否变更
}

// Register Agent注册 (POST /api/boot/v1/register)
//
// Agent首次启动时调用此API注册机器信息
func (h *AgentHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// 验证必填字段
	if req.MacAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "mac_address is required",
		})
	}

	// 设置默认Schema版本
	if req.HardwareSpec.SchemaVersion == "" {
		req.HardwareSpec.SchemaVersion = "1.0"
	}

	// 查找是否已存在（根据MAC地址）
	var machine models.Machine
	err := database.DB.Where("mac_address = ?", req.MacAddress).First(&machine).Error

	if err == nil {
		// 机器已存在 - 更新信息
		return h.updateExistingMachine(c, &machine, &req)
	}

	// 机器不存在 - 创建新记录
	return h.createNewMachine(c, &req)
}

// Heartbeat Agent心跳 (POST /api/boot/v1/heartbeat)
//
// Agent定期发送心跳，更新在线状态和硬件信息
func (h *AgentHandler) Heartbeat(c echo.Context) error {
	var req HeartbeatRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// 验证必填字段
	if req.MachineID == "" || req.MacAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "machine_id and mac_address are required",
		})
	}

	// 查找机器
	var machine models.Machine
	err := database.DB.First(&machine, "id = ?", req.MachineID).Error
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Machine not found",
		})
	}

	// 验证MAC地址是否匹配（防止伪造）
	if machine.MacAddress != req.MacAddress {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "MAC address mismatch",
		})
	}

	// 更新IP地址
	if req.IPAddress != "" && req.IPAddress != machine.IPAddress {
		machine.IPAddress = req.IPAddress
	}

	// 检测硬件变更
	hardwareChanged := false
	if req.HardwareSpec.SchemaVersion != "" {
		hardwareChanged = h.detectHardwareChange(&machine, &req.HardwareSpec)
		if hardwareChanged {
			machine.HardwareSpec = req.HardwareSpec
		}
	}

	// 更新UpdatedAt时间戳（标记在线）
	machine.UpdatedAt = time.Now()

	// 保存更新
	if err := database.DB.Save(&machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update machine",
		})
	}

	// 返回响应
	resp := HeartbeatResponse{
		Status:         "ok",
		Message:        "Heartbeat received",
		NextPoll:       30, // 30秒后再次心跳
		HardwareChange: hardwareChanged,
	}

	return c.JSON(http.StatusOK, resp)
}

// createNewMachine 创建新机器记录
func (h *AgentHandler) createNewMachine(c echo.Context, req *RegisterRequest) error {
	// 生成Machine ID
	machineID := uuid.New().String()

	// 自动生成hostname（如果未提供）
	hostname := req.Hostname
	if hostname == "" {
		hostname = generateHostname(req.MacAddress)
	}

	// 创建Machine记录
	machine := models.Machine{
		ID:           machineID,
		Hostname:     hostname,
		MacAddress:   req.MacAddress,
		IPAddress:    req.IPAddress,
		Status:       models.MachineStatusDiscovered, // 初始状态：已发现
		HardwareSpec: req.HardwareSpec,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := database.DB.Create(&machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create machine record",
		})
	}

	// 返回注册成功响应
	resp := RegisterResponse{
		MachineID:    machineID,
		Status:       "registered",
		Message:      "Machine registered successfully",
		HeartbeatURL: "/api/boot/v1/heartbeat",
		TaskPollURL:  "/api/boot/v1/task",
		PollInterval: 30, // 30秒心跳间隔
	}

	return c.JSON(http.StatusCreated, resp)
}

// updateExistingMachine 更新已存在的机器记录
func (h *AgentHandler) updateExistingMachine(c echo.Context, machine *models.Machine, req *RegisterRequest) error {
	// 更新IP地址
	if req.IPAddress != "" {
		machine.IPAddress = req.IPAddress
	}

	// 更新hostname（如果提供）
	if req.Hostname != "" && req.Hostname != machine.Hostname {
		machine.Hostname = req.Hostname
	}

	// 检测硬件变更
	hardwareChanged := h.detectHardwareChange(machine, &req.HardwareSpec)
	if hardwareChanged {
		machine.HardwareSpec = req.HardwareSpec
	}

	// 更新时间戳
	machine.UpdatedAt = time.Now()

	// 保存更新
	if err := database.DB.Save(machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update machine",
		})
	}

	// 返回响应
	resp := RegisterResponse{
		MachineID:    machine.ID,
		Status:       "updated",
		Message:      "Machine information updated",
		HeartbeatURL: "/api/boot/v1/heartbeat",
		TaskPollURL:  "/api/boot/v1/task",
		PollInterval: 30,
	}

	return c.JSON(http.StatusOK, resp)
}

// detectHardwareChange 检测硬件变更
//
// 通过计算硬件指纹SHA256哈希值判断硬件是否变更
func (h *AgentHandler) detectHardwareChange(machine *models.Machine, newSpec *models.HardwareInfo) bool {
	// 计算旧硬件指纹哈希
	oldHash := calculateHardwareHash(&machine.HardwareSpec)

	// 计算新硬件指纹哈希
	newHash := calculateHardwareHash(newSpec)

	// 比较哈希值
	return oldHash != newHash
}

// calculateHardwareHash 计算硬件指纹SHA256哈希
func calculateHardwareHash(spec *models.HardwareInfo) string {
	// 将HardwareInfo序列化为JSON
	data, err := json.Marshal(spec)
	if err != nil {
		return ""
	}

	// 计算SHA256
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// generateHostname 根据MAC地址生成主机名
func generateHostname(macAddress string) string {
	// 提取MAC地址最后6位作为主机名后缀
	// 例如: aa:bb:cc:dd:ee:ff -> server-ddeeff
	cleanMAC := ""
	for _, c := range macAddress {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			cleanMAC += string(c)
		}
	}

	if len(cleanMAC) >= 6 {
		return "server-" + cleanMAC[len(cleanMAC)-6:]
	}

	return "server-" + cleanMAC
}

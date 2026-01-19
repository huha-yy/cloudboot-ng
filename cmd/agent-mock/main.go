package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/models"
)

const (
	DefaultServerURL = "http://localhost:8080"
	DefaultMAC       = "52:54:00:12:34:56"
)

type AgentConfig struct {
	ServerURL   string
	MacAddress  string
	Hostname    string
	Heartbeats  int
	Interval    int // seconds
	ModifyHW    bool
}

func main() {
	config := parseFlags()

	log.Printf("🤖 Agent模拟器启动")
	log.Printf("   - Server: %s", config.ServerURL)
	log.Printf("   - MAC: %s", config.MacAddress)
	log.Printf("   - Hostname: %s", config.Hostname)
	log.Printf("   - Heartbeats: %d", config.Heartbeats)
	log.Printf("   - Interval: %ds", config.Interval)

	// 第一步：注册
	machineID, err := register(config)
	if err != nil {
		log.Fatalf("❌ 注册失败: %v", err)
	}
	log.Printf("✅ 注册成功: machine_id=%s", machineID)

	// 第二步：发送心跳
	if config.Heartbeats > 0 {
		log.Printf("📡 开始发送心跳...")
		for i := 0; i < config.Heartbeats; i++ {
			time.Sleep(time.Duration(config.Interval) * time.Second)

			// 如果启用硬件修改，在第3次心跳时修改硬件
			modifyHW := config.ModifyHW && i == 2

			changed, err := heartbeat(config, machineID, modifyHW)
			if err != nil {
				log.Printf("⚠️  心跳 #%d 失败: %v", i+1, err)
				continue
			}

			if changed {
				log.Printf("🔔 心跳 #%d: 硬件变更已检测!", i+1)
			} else {
				log.Printf("✓ 心跳 #%d: OK", i+1)
			}
		}
	}

	log.Println("🎉 Agent模拟器完成")
}

func parseFlags() *AgentConfig {
	config := &AgentConfig{}

	flag.StringVar(&config.ServerURL, "server", DefaultServerURL, "CloudBoot Server URL")
	flag.StringVar(&config.MacAddress, "mac", DefaultMAC, "MAC Address")
	flag.StringVar(&config.Hostname, "hostname", "", "Hostname (optional)")
	flag.IntVar(&config.Heartbeats, "heartbeats", 5, "Number of heartbeats to send")
	flag.IntVar(&config.Interval, "interval", 2, "Heartbeat interval in seconds")
	flag.BoolVar(&config.ModifyHW, "modify-hw", false, "Modify hardware on 3rd heartbeat")
	flag.Parse()

	return config
}

func register(config *AgentConfig) (string, error) {
	url := config.ServerURL + "/api/boot/v1/register"

	hwSpec := generateHardwareSpec(false)

	payload := map[string]interface{}{
		"mac_address":   config.MacAddress,
		"ip_address":    "10.0.2.15",
		"hostname":      config.Hostname,
		"hardware_spec": hwSpec,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("注册失败: HTTP %d", resp.StatusCode)
	}

	var result struct {
		MachineID    string `json:"machine_id"`
		Status       string `json:"status"`
		Message      string `json:"message"`
		HeartbeatURL string `json:"heartbeat_url"`
		PollInterval int    `json:"poll_interval_seconds"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	log.Printf("   - Status: %s", result.Status)
	log.Printf("   - Message: %s", result.Message)
	log.Printf("   - HeartbeatURL: %s", result.HeartbeatURL)
	log.Printf("   - PollInterval: %ds", result.PollInterval)

	return result.MachineID, nil
}

func heartbeat(config *AgentConfig, machineID string, modifyHW bool) (bool, error) {
	url := config.ServerURL + "/api/boot/v1/heartbeat"

	hwSpec := generateHardwareSpec(modifyHW)

	payload := map[string]interface{}{
		"machine_id":    machineID,
		"mac_address":   config.MacAddress,
		"ip_address":    "10.0.2.15",
		"hardware_spec": hwSpec,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("心跳失败: HTTP %d", resp.StatusCode)
	}

	var result struct {
		Status         string `json:"status"`
		Message        string `json:"message"`
		NextPoll       int    `json:"next_poll_seconds"`
		HardwareChange bool   `json:"hardware_change"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("解析响应失败: %w", err)
	}

	return result.HardwareChange, nil
}

func generateHardwareSpec(modified bool) models.HardwareInfo {
	cpuCount := 8
	memoryGB := int64(16)

	// 如果修改硬件,增加CPU和内存
	if modified {
		cpuCount = 16
		memoryGB = 32
	}

	return models.HardwareInfo{
		SchemaVersion: "1.0",
		System: models.SystemInfo{
			Manufacturer: "QEMU",
			ProductName:  "Standard PC (Q35 + ICH9, 2009)",
			SerialNumber: "mock-serial-001",
		},
		CPU: models.CPUInfo{
			Model:   "Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz",
			Arch:    "x86_64",
			Cores:   cpuCount,
			Sockets: 1,
		},
		Memory: models.MemoryInfo{
			TotalBytes: memoryGB * 1024 * 1024 * 1024,
			DIMMs: []models.DimmInfo{
				{
					Slot:      "DIMM 0",
					SizeBytes: memoryGB * 1024 * 1024 * 1024,
					Speed:     2666,
				},
			},
		},
		StorageControllers: []models.ControllerInfo{
			{
				PCIID:  "1000:005f",
				Vendor: "LSI Logic",
				Model:  "MegaRAID SAS 3108",
				Driver: "megaraid_sas",
			},
		},
		NetworkInterfaces: []models.NICInfo{
			{
				Name:  "eth0",
				MAC:   "52:54:00:12:34:56",
				Speed: 1000,
				Link:  true,
			},
		},
	}
}

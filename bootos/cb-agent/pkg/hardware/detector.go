package hardware

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Detector detects hardware information
type Detector struct{}

// NewDetector creates a new hardware detector
func NewDetector() *Detector {
	return &Detector{}
}

// Detect performs full hardware detection
func (d *Detector) Detect() (map[string]interface{}, error) {
	spec := make(map[string]interface{})

	// System info
	system, err := d.detectSystem()
	if err == nil {
		for k, v := range system {
			spec[k] = v
		}
	}

	// CPU info
	cpu, err := d.detectCPU()
	if err == nil {
		spec["cpu"] = cpu
	}

	// Memory info
	memory, err := d.detectMemory()
	if err == nil {
		spec["memory"] = memory
	}

	// Disk info
	disks, err := d.detectDisks()
	if err == nil {
		spec["disks"] = disks
	}

	// Network interfaces
	network, err := d.DetectNetwork()
	if err == nil {
		spec["network_interfaces"] = network
	}

	return spec, nil
}

// detectSystem detects system information (manufacturer, model, serial)
func (d *Detector) detectSystem() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Try dmidecode first
	if manufacturer, err := d.runCommand("dmidecode", "-s", "system-manufacturer"); err == nil {
		info["system_manufacturer"] = strings.TrimSpace(manufacturer)
	}

	if product, err := d.runCommand("dmidecode", "-s", "system-product-name"); err == nil {
		info["system_product"] = strings.TrimSpace(product)
	}

	if serial, err := d.runCommand("dmidecode", "-s", "system-serial-number"); err == nil {
		info["system_serial"] = strings.TrimSpace(serial)
	}

	// Fallback to /sys/class/dmi
	if _, ok := info["system_manufacturer"]; !ok {
		if data, err := os.ReadFile("/sys/class/dmi/id/sys_vendor"); err == nil {
			info["system_manufacturer"] = strings.TrimSpace(string(data))
		}
	}

	if _, ok := info["system_product"]; !ok {
		if data, err := os.ReadFile("/sys/class/dmi/id/product_name"); err == nil {
			info["system_product"] = strings.TrimSpace(string(data))
		}
	}

	return info, nil
}

// detectCPU detects CPU information
func (d *Detector) detectCPU() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Parse /proc/cpuinfo
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	processorCount := 0
	var modelName string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "processor") {
			processorCount++
		}
		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				modelName = strings.TrimSpace(parts[1])
			}
		}
	}

	info["model"] = modelName
	info["cores"] = processorCount

	return info, nil
}

// detectMemory detects memory information
func (d *Detector) detectMemory() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Parse /proc/meminfo
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info["total_kb"] = parts[1]
			}
			break
		}
	}

	return info, nil
}

// detectDisks detects disk information
func (d *Detector) detectDisks() ([]map[string]interface{}, error) {
	var disks []map[string]interface{}

	// Use lsblk to list block devices
	output, err := d.runCommand("lsblk", "-d", "-n", "-o", "NAME,SIZE,MODEL", "-e", "7,11")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			disk := map[string]interface{}{
				"name": fields[0],
				"size": fields[1],
			}
			if len(fields) >= 3 {
				disk["model"] = strings.Join(fields[2:], " ")
			}
			disks = append(disks, disk)
		}
	}

	return disks, nil
}

// DetectNetwork detects network interfaces
func (d *Detector) DetectNetwork() ([]map[string]interface{}, error) {
	var interfaces []map[string]interface{}

	// Use ip command to list interfaces
	output, err := d.runCommand("ip", "-o", "link", "show")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		name := strings.TrimSuffix(parts[1], ":")
		mac := ""

		// Extract MAC address
		for i, part := range parts {
			if part == "link/ether" && i+1 < len(parts) {
				mac = parts[i+1]
				break
			}
		}

		// Get IP address
		ip, _ := d.getInterfaceIP(name)

		iface := map[string]interface{}{
			"name": name,
			"mac":  mac,
			"ip":   ip,
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// getInterfaceIP gets the IP address of an interface
func (d *Detector) getInterfaceIP(ifaceName string) (string, error) {
	output, err := d.runCommand("ip", "-o", "-4", "addr", "show", ifaceName)
	if err != nil {
		return "", err
	}

	// Parse output: "2: eth0    inet 192.168.1.100/24 ..."
	parts := strings.Fields(output)
	for i, part := range parts {
		if part == "inet" && i+1 < len(parts) {
			ipWithMask := parts[i+1]
			// Remove /24 suffix
			if idx := strings.Index(ipWithMask, "/"); idx != -1 {
				return ipWithMask[:idx], nil
			}
			return ipWithMask, nil
		}
	}

	return "", fmt.Errorf("no IP found for %s", ifaceName)
}

// runCommand runs a command and returns its output
func (d *Detector) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s failed: %w", name, err)
	}
	return string(output), nil
}

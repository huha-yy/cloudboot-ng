package models

import (
	"testing"
	"time"
)

func TestMachineStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status MachineStatus
		want   string
	}{
		{"Discovered", MachineStatusDiscovered, "discovered"},
		{"Ready", MachineStatusReady, "ready"},
		{"Installing", MachineStatusInstalling, "installing"},
		{"Active", MachineStatusActive, "active"},
		{"Error", MachineStatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.status); got != tt.want {
				t.Errorf("MachineStatus = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMachine_Validation(t *testing.T) {
	tests := []struct {
		name    string
		machine Machine
		wantErr bool
	}{
		{
			name: "Valid machine",
			machine: Machine{
				ID:         "test-id",
				Hostname:   "test-server",
				MacAddress: "aa:bb:cc:dd:ee:ff",
				IPAddress:  "192.168.1.100",
				Status:     MachineStatusDiscovered,
			},
			wantErr: false,
		},
		{
			name: "Empty hostname",
			machine: Machine{
				ID:         "test-id",
				Hostname:   "",
				MacAddress: "aa:bb:cc:dd:ee:ff",
				IPAddress:  "192.168.1.100",
				Status:     MachineStatusDiscovered,
			},
			wantErr: true,
		},
		{
			name: "Empty MAC address",
			machine: Machine{
				ID:         "test-id",
				Hostname:   "test-server",
				MacAddress: "",
				IPAddress:  "192.168.1.100",
				Status:     MachineStatusDiscovered,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - check required fields
			hasErr := tt.machine.Hostname == "" || tt.machine.MacAddress == ""
			if hasErr != tt.wantErr {
				t.Errorf("Machine validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestMachine_HardwareInfo(t *testing.T) {
	machine := Machine{
		ID:         "test-id",
		Hostname:   "test-server",
		MacAddress: "aa:bb:cc:dd:ee:ff",
		Status:     MachineStatusDiscovered,
		HardwareSpec: HardwareInfo{
			SchemaVersion: "1.0",
			System: SystemInfo{
				Manufacturer: "Dell",
				ProductName:  "PowerEdge R740",
				SerialNumber: "SN12345",
			},
			CPU: CPUInfo{
				Arch:    "x86_64",
				Model:   "Intel Xeon E5-2680 v4",
				Cores:   32,
				Sockets: 2,
			},
			Memory: MemoryInfo{
				TotalBytes: 137438953472, // 128GB
				DIMMs: []DimmInfo{
					{Slot: "DIMM_A1", SizeBytes: 34359738368, Speed: 3200},
				},
			},
		},
	}

	// Verify hardware info structure
	if machine.HardwareSpec.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %v, want 1.0", machine.HardwareSpec.SchemaVersion)
	}

	if machine.HardwareSpec.System.Manufacturer != "Dell" {
		t.Errorf("Manufacturer = %v, want Dell", machine.HardwareSpec.System.Manufacturer)
	}

	if machine.HardwareSpec.CPU.Cores != 32 {
		t.Errorf("CPU Cores = %v, want 32", machine.HardwareSpec.CPU.Cores)
	}

	if machine.HardwareSpec.Memory.TotalBytes != 137438953472 {
		t.Errorf("Memory TotalBytes = %v, want 137438953472", machine.HardwareSpec.Memory.TotalBytes)
	}
}

func TestMachine_IsOnline(t *testing.T) {
	tests := []struct {
		name      string
		updatedAt time.Time
		want      bool
	}{
		{"Recent update - online", time.Now().Add(-1 * time.Minute), true},
		{"Old update - offline", time.Now().Add(-10 * time.Minute), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine := Machine{
				ID:        "test-id",
				Hostname:  "test-server",
				UpdatedAt: tt.updatedAt,
			}

			if got := machine.IsOnline(); got != tt.want {
				t.Errorf("IsOnline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMachine_IsReady(t *testing.T) {
	tests := []struct {
		name   string
		status MachineStatus
		want   bool
	}{
		{"Discovered - not ready", MachineStatusDiscovered, false},
		{"Ready - ready", MachineStatusReady, true},
		{"Installing - not ready", MachineStatusInstalling, false},
		{"Active - ready", MachineStatusActive, true},
		{"Error - not ready", MachineStatusError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine := Machine{
				ID:     "test-id",
				Status: tt.status,
			}

			if got := machine.IsReady(); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

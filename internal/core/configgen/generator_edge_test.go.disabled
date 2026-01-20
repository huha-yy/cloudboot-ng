package configgen

import (
	"strings"
	"testing"

	"github.com/cloudboot/cloudboot-ng/internal/models"
)

// TestValidateOSType_EdgeCases tests OS type validation with various edge cases
func TestValidateOSType_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		osType    string
		wantError bool
	}{
		{
			name:      "Valid CentOS 7",
			osType:    "centos7",
			wantError: false,
		},
		{
			name:      "Valid CentOS 8",
			osType:    "centos8",
			wantError: false,
		},
		{
			name:      "Valid Ubuntu 20",
			osType:    "ubuntu20",
			wantError: false,
		},
		{
			name:      "Valid Ubuntu 22",
			osType:    "ubuntu22",
			wantError: false,
		},
		{
			name:      "Valid SLES 12",
			osType:    "sles12",
			wantError: false,
		},
		{
			name:      "Valid SLES 15",
			osType:    "sles15",
			wantError: false,
		},
		{
			name:      "Invalid - uppercase",
			osType:    "CENTOS7",
			wantError: true,
		},
		{
			name:      "Invalid - spaces",
			osType:    "centos 7",
			wantError: true,
		},
		{
			name:      "Invalid - unsupported distro",
			osType:    "debian10",
			wantError: true,
		},
		{
			name:      "Invalid - empty string",
			osType:    "",
			wantError: true,
		},
		{
			name:      "Invalid - special characters",
			osType:    "centos7!@#",
			wantError: true,
		},
	}

	gen := NewGenerator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &models.OSProfile{
				Distro: tt.osType,
				Config: models.ProfileConfig{
					Partitions: []models.Partition{
						{MountPoint: "/", Size: "20GB", FSType: "ext4"},
					},
					Network: models.NetworkConfig{
						Hostname: "test-server",
						IP:       "192.168.1.100",
						Netmask:  "255.255.255.0",
					},
				},
			}

			err := gen.Validate(profile)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestValidatePartitions_EdgeCases tests partition validation with various edge cases
func TestValidatePartitions_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		partitions []models.Partition
		wantError  bool
		errorMsg   string
	}{
		{
			name: "Valid basic partitions",
			partitions: []models.Partition{
				{MountPoint: "/boot", Size: "1GB", FSType: "ext4"},
				{MountPoint: "/", Size: "50GB", FSType: "xfs"},
			},
			wantError: false,
		},
		{
			name: "Valid with swap",
			partitions: []models.Partition{
				{MountPoint: "/", Size: "50GB", FSType: "xfs"},
				{MountPoint: "swap", Size: "8GB", FSType: "swap"},
			},
			wantError: false,
		},
		{
			name: "Valid with /boot/efi",
			partitions: []models.Partition{
				{MountPoint: "/boot/efi", Size: "512MB", FSType: "vfat"},
				{MountPoint: "/", Size: "50GB", FSType: "ext4"},
			},
			wantError: false,
		},
		{
			name:       "Invalid - no partitions",
			partitions: []models.Partition{},
			wantError:  true,
			errorMsg:   "no partitions",
		},
		{
			name: "Invalid - missing root",
			partitions: []models.Partition{
				{MountPoint: "/boot", Size: "1GB", FSType: "ext4"},
				{MountPoint: "/home", Size: "50GB", FSType: "xfs"},
			},
			wantError: true,
			errorMsg:  "root",
		},
		{
			name: "Invalid - empty mount point",
			partitions: []models.Partition{
				{MountPoint: "", Size: "50GB", FSType: "ext4"},
			},
			wantError: true,
			errorMsg:  "mount point is empty",
		},
		{
			name: "Invalid - empty filesystem type",
			partitions: []models.Partition{
				{MountPoint: "/", Size: "50GB", FSType: ""},
			},
			wantError: true,
			errorMsg:  "filesystem type is empty",
		},
		{
			name: "Invalid - empty size",
			partitions: []models.Partition{
				{MountPoint: "/", Size: "", FSType: "ext4"},
			},
			wantError: true,
			errorMsg:  "size is empty",
		},
		{
			name: "Invalid - swap without swap fstype",
			partitions: []models.Partition{
				{MountPoint: "/", Size: "50GB", FSType: "ext4"},
				{MountPoint: "swap", Size: "8GB", FSType: "ext4"},
			},
			wantError: true,
			errorMsg:  "swap partition must have fstype=swap",
		},
		{
			name: "Invalid - unsupported filesystem",
			partitions: []models.Partition{
				{MountPoint: "/", Size: "50GB", FSType: "ntfs"},
			},
			wantError: true,
			errorMsg:  "unsupported filesystem type",
		},
		{
			name: "Invalid - /boot/efi without vfat",
			partitions: []models.Partition{
				{MountPoint: "/boot/efi", Size: "512MB", FSType: "ext4"},
				{MountPoint: "/", Size: "50GB", FSType: "ext4"},
			},
			wantError: true,
			errorMsg:  "/boot/efi must use vfat",
		},
		{
			name: "Valid - multiple data partitions",
			partitions: []models.Partition{
				{MountPoint: "/", Size: "50GB", FSType: "ext4"},
				{MountPoint: "/var", Size: "100GB", FSType: "xfs"},
				{MountPoint: "/home", Size: "200GB", FSType: "xfs"},
			},
			wantError: false,
		},
		{
			name: "Valid - btrfs filesystem",
			partitions: []models.Partition{
				{MountPoint: "/", Size: "50GB", FSType: "btrfs"},
			},
			wantError: false,
		},
	}

	gen := NewGenerator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &models.OSProfile{
				Distro: "centos7",
				Config: models.ProfileConfig{
					Partitions: tt.partitions,
					Network: models.NetworkConfig{
						Hostname: "test-server",
						IP:       "192.168.1.100",
						Netmask:  "255.255.255.0",
					},
				},
			}

			err := gen.Validate(profile)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError && tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
			}
		})
	}
}

// TestValidateNetwork_EdgeCases tests network validation with various edge cases
func TestValidateNetwork_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		network   models.NetworkConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid basic config",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
				Gateway:  "192.168.1.1",
			},
			wantError: false,
		},
		{
			name: "Valid with DNS",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
				Gateway:  "192.168.1.1",
				DNS:      []string{"8.8.8.8", "8.8.4.4"},
			},
			wantError: false,
		},
		{
			name: "Valid without gateway",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
			},
			wantError: false,
		},
		{
			name: "Invalid - empty hostname",
			network: models.NetworkConfig{
				Hostname: "",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
			},
			wantError: true,
			errorMsg:  "hostname is required",
		},
		{
			name: "Invalid - hostname too long",
			network: models.NetworkConfig{
				Hostname: strings.Repeat("a", 64),
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
			},
			wantError: true,
			errorMsg:  "hostname too long",
		},
		{
			name: "Invalid - empty IP",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "",
				Netmask:  "255.255.255.0",
			},
			wantError: true,
			errorMsg:  "IP address is required",
		},
		{
			name: "Invalid - malformed IP",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "999.999.999.999",
				Netmask:  "255.255.255.0",
			},
			wantError: true,
			errorMsg:  "invalid IP address",
		},
		{
			name: "Invalid - IP with letters",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "192.168.1.abc",
				Netmask:  "255.255.255.0",
			},
			wantError: true,
			errorMsg:  "invalid IP address",
		},
		{
			name: "Invalid - empty netmask",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "192.168.1.100",
				Netmask:  "",
			},
			wantError: true,
			errorMsg:  "netmask is required",
		},
		{
			name: "Invalid - invalid netmask",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.1",
			},
			wantError: true,
			errorMsg:  "invalid netmask",
		},
		{
			name: "Invalid - gateway malformed",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
				Gateway:  "invalid-gateway",
			},
			wantError: true,
			errorMsg:  "invalid gateway",
		},
		{
			name: "Invalid - DNS malformed",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
				DNS:      []string{"8.8.8.8", "invalid-dns"},
			},
			wantError: true,
			errorMsg:  "invalid DNS",
		},
		{
			name: "Valid - /16 netmask",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "10.0.1.100",
				Netmask:  "255.255.0.0",
			},
			wantError: false,
		},
		{
			name: "Valid - /8 netmask",
			network: models.NetworkConfig{
				Hostname: "server01",
				IP:       "10.1.1.100",
				Netmask:  "255.0.0.0",
			},
			wantError: false,
		},
		{
			name: "Valid - hostname with hyphens",
			network: models.NetworkConfig{
				Hostname: "web-server-01",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
			},
			wantError: false,
		},
		{
			name: "Valid - 63 character hostname",
			network: models.NetworkConfig{
				Hostname: strings.Repeat("a", 63),
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
			},
			wantError: false,
		},
	}

	gen := NewGenerator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &models.OSProfile{
				Distro: "centos7",
				Config: models.ProfileConfig{
					Partitions: []models.Partition{
						{MountPoint: "/", Size: "50GB", FSType: "ext4"},
					},
					Network: tt.network,
				},
			}

			err := gen.Validate(profile)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError && tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
			}
		})
	}
}

// TestGenerate_MultipleDistros tests config generation for different distros
func TestGenerate_MultipleDistros(t *testing.T) {
	tests := []struct {
		name           string
		distro         string
		expectedKeywords []string
	}{
		{
			name:   "CentOS 7 Kickstart",
			distro: "centos7",
			expectedKeywords: []string{
				"# Kickstart for centos7",
				"auth --enableshadow",
				"url --url=",
				"part /",
			},
		},
		{
			name:   "CentOS 8 Kickstart",
			distro: "centos8",
			expectedKeywords: []string{
				"# Kickstart for centos8",
				"auth --enableshadow",
				"url --url=",
			},
		},
		{
			name:   "Ubuntu 20 Preseed",
			distro: "ubuntu20",
			expectedKeywords: []string{
				"# Preseed for ubuntu20",
				"d-i debian-installer/locale",
				"d-i netcfg",
			},
		},
		{
			name:   "Ubuntu 22 Preseed",
			distro: "ubuntu22",
			expectedKeywords: []string{
				"# Preseed for ubuntu22",
				"d-i debian-installer/locale",
			},
		},
		{
			name:   "SLES 12 AutoYaST",
			distro: "sles12",
			expectedKeywords: []string{
				"<?xml version",
				"<profile xmlns",
				"<networking>",
			},
		},
		{
			name:   "SLES 15 AutoYaST",
			distro: "sles15",
			expectedKeywords: []string{
				"<?xml version",
				"<profile xmlns",
			},
		},
	}

	gen := NewGenerator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &models.OSProfile{
				Distro: tt.distro,
				Config: models.ProfileConfig{
					RepoURL: "http://mirror.example.com/repo",
					Partitions: []models.Partition{
						{MountPoint: "/", Size: "50GB", FSType: "ext4"},
					},
					Network: models.NetworkConfig{
						Hostname: "test-server",
						IP:       "192.168.1.100",
						Netmask:  "255.255.255.0",
						Gateway:  "192.168.1.1",
					},
				},
			}

			output, err := gen.Generate(profile)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			for _, keyword := range tt.expectedKeywords {
				if !strings.Contains(output, keyword) {
					t.Errorf("Output missing expected keyword: %q", keyword)
				}
			}
		})
	}
}

// TestGenerate_ComplexPartitionSchemes tests complex partition configurations
func TestGenerate_ComplexPartitionSchemes(t *testing.T) {
	tests := []struct {
		name       string
		partitions []models.Partition
		checks     []string
	}{
		{
			name: "UEFI boot with EFI partition",
			partitions: []models.Partition{
				{MountPoint: "/boot/efi", Size: "512MB", FSType: "vfat"},
				{MountPoint: "/boot", Size: "1GB", FSType: "ext4"},
				{MountPoint: "/", Size: "50GB", FSType: "xfs"},
			},
			checks: []string{
				"part /boot/efi --fstype=vfat --size=512MB",
				"part /boot --fstype=ext4 --size=1GB",
				"part / --fstype=xfs --size=50GB",
			},
		},
		{
			name: "Separate /var and /home",
			partitions: []models.Partition{
				{MountPoint: "/", Size: "30GB", FSType: "ext4"},
				{MountPoint: "/var", Size: "100GB", FSType: "xfs"},
				{MountPoint: "/home", Size: "200GB", FSType: "xfs"},
				{MountPoint: "swap", Size: "16GB", FSType: "swap"},
			},
			checks: []string{
				"part / --fstype=ext4 --size=30GB",
				"part /var --fstype=xfs --size=100GB",
				"part /home --fstype=xfs --size=200GB",
				"part swap --fstype=swap --size=16GB",
			},
		},
		{
			name: "Minimal root only",
			partitions: []models.Partition{
				{MountPoint: "/", Size: "50GB", FSType: "ext4"},
			},
			checks: []string{
				"part / --fstype=ext4 --size=50GB",
			},
		},
	}

	gen := NewGenerator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &models.OSProfile{
				Distro: "centos7",
				Config: models.ProfileConfig{
					RepoURL:    "http://mirror.centos.org/centos/7/os/x86_64",
					Partitions: tt.partitions,
					Network: models.NetworkConfig{
						Hostname: "test-server",
						IP:       "192.168.1.100",
						Netmask:  "255.255.255.0",
					},
				},
			}

			output, err := gen.Generate(profile)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("Output missing expected content: %q", check)
				}
			}
		})
	}
}

// TestValidate_NilProfile tests validation with nil profile
func TestValidate_NilProfile(t *testing.T) {
	gen := NewGenerator()
	err := gen.Validate(nil)

	if err == nil {
		t.Error("Expected error for nil profile")
	}

	if !strings.Contains(err.Error(), "profile is nil") {
		t.Errorf("Expected error about nil profile, got: %v", err)
	}
}

// TestGenerate_WithPackages tests config generation with custom packages
func TestGenerate_WithPackages(t *testing.T) {
	profile := &models.OSProfile{
		Distro: "centos7",
		Config: models.ProfileConfig{
			RepoURL: "http://mirror.centos.org/centos/7/os/x86_64",
			Partitions: []models.Partition{
				{MountPoint: "/", Size: "50GB", FSType: "ext4"},
			},
			Network: models.NetworkConfig{
				Hostname: "test-server",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
			},
			Packages: []string{"vim", "wget", "curl", "git"},
		},
	}

	gen := NewGenerator()
	output, err := gen.Generate(profile)

	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	expectedPackages := []string{"vim", "wget", "curl", "git"}
	for _, pkg := range expectedPackages {
		if !strings.Contains(output, pkg) {
			t.Errorf("Output missing package: %q", pkg)
		}
	}
}

// TestGenerate_WithPostScript tests config generation with post-installation script
func TestGenerate_WithPostScript(t *testing.T) {
	profile := &models.OSProfile{
		Distro: "centos7",
		Config: models.ProfileConfig{
			RepoURL: "http://mirror.centos.org/centos/7/os/x86_64",
			Partitions: []models.Partition{
				{MountPoint: "/", Size: "50GB", FSType: "ext4"},
			},
			Network: models.NetworkConfig{
				Hostname: "test-server",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
			},
			PostScript: "#!/bin/bash\necho 'Post-install complete'\nmkdir /opt/app",
		},
	}

	gen := NewGenerator()
	output, err := gen.Generate(profile)

	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if !strings.Contains(output, "%post") {
		t.Error("Output missing post section")
	}

	if !strings.Contains(output, "Post-install complete") {
		t.Error("Output missing post-script content")
	}
}

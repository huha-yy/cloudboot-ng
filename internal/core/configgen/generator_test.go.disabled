package configgen

import (
	"strings"
	"testing"

	"github.com/cloudboot/cloudboot-ng/internal/models"
)

func TestGenerate_Basic(t *testing.T) {
	profile := &models.OSProfile{
		ID:     "test-1",
		Name:   "CentOS 7 Test",
		Distro: "centos7",
		Config: models.ProfileConfig{
			RepoURL: "http://mirror.centos.org/centos/7/os/x86_64",
			Partitions: []models.Partition{
				{MountPoint: "/boot", Size: "1024MB", FSType: "ext4"},
				{MountPoint: "swap", Size: "8192MB", FSType: "swap"},
				{MountPoint: "/", Size: "51200MB", FSType: "xfs"},
			},
			Network: models.NetworkConfig{
				Hostname: "test-server-01",
				IP:       "192.168.1.100",
				Netmask:  "255.255.255.0",
				Gateway:  "192.168.1.1",
				DNS:      []string{"8.8.8.8"},
			},
			Packages:   []string{"vim", "wget"},
			PostScript: "echo 'Done'",
		},
	}

	gen := NewGenerator()
	output, err := gen.Generate(profile)

	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Basic checks
	checks := []string{
		"Kickstart for centos7",
		"part /boot --fstype=ext4",
		"part swap --fstype=swap",
		"part / --fstype=xfs",
		"network",
		"test-server-01",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Output missing expected string: %q", check)
		}
	}
}

func TestValidate_MissingRoot(t *testing.T) {
	profile := &models.OSProfile{
		Distro: "centos7",
		Config: models.ProfileConfig{
			Partitions: []models.Partition{
				{MountPoint: "/boot", Size: "1GB", FSType: "ext4"},
			},
			Network: models.NetworkConfig{
				Hostname: "test",
				IP:       "192.168.1.1",
				Netmask:  "255.255.255.0",
			},
		},
	}

	gen := NewGenerator()
	err := gen.Validate(profile)

	if err == nil {
		t.Error("Expected validation error for missing root partition")
	}

	if !strings.Contains(err.Error(), "root") {
		t.Errorf("Expected error about root partition, got: %v", err)
	}
}

func TestValidate_InvalidIP(t *testing.T) {
	profile := &models.OSProfile{
		Distro: "centos7",
		Config: models.ProfileConfig{
			Partitions: []models.Partition{
				{MountPoint: "/", Size: "20GB", FSType: "ext4"},
			},
			Network: models.NetworkConfig{
				Hostname: "test",
				IP:       "999.999.999.999",
				Netmask:  "255.255.255.0",
			},
		},
	}

	gen := NewGenerator()
	err := gen.Validate(profile)

	if err == nil {
		t.Error("Expected validation error for invalid IP")
	}
}

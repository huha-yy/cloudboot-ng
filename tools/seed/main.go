package main

import (
	"flag"
	"log"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm/logger"
)

func main() {
	dbPath := flag.String("db", "cloudboot.db", "Database file path")
	flag.Parse()

	log.Println("CloudBoot Database Seeder")
	log.Println("========================")

	// Initialize database
	config := database.Config{
		DSN:      *dbPath + "?_journal_mode=WAL",
		LogLevel: logger.Warn,
	}

	if err := database.Init(config); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	log.Println("[*] Database initialized")

	// Seed OS Profiles
	profiles := seedProfiles()
	log.Printf("[✓] Seeded %d OS profiles", len(profiles))

	// Seed Machines
	machines := seedMachines()
	log.Printf("[✓] Seeded %d machines", len(machines))

	// Seed Jobs
	jobs := seedJobs(machines, profiles)
	log.Printf("[✓] Seeded %d jobs", len(jobs))

	log.Println("")
	log.Println("Database seeded successfully!")
	log.Println("")
	log.Println("Test Credentials:")
	log.Println("  Machine IDs:")
	for i, m := range machines {
		if i < 3 {
			log.Printf("    - %s (MAC: %s)", m.ID, m.MacAddress)
		}
	}
	log.Println("  Profile IDs:")
	for i, p := range profiles {
		if i < 3 {
			log.Printf("    - %s (%s)", p.ID, p.Name)
		}
	}
}

func seedProfiles() []models.OSProfile {
	profiles := []models.OSProfile{
		{
			ID:     uuid.New().String(),
			Name:   "CentOS 7 Production",
			Distro: "centos7",
			Config: models.ProfileConfig{
				RepoURL: "http://mirror.centos.org/centos/7/os/x86_64",
				Partitions: []models.Partition{
					{MountPoint: "/boot", Size: "1GB", FSType: "ext4"},
					{MountPoint: "swap", Size: "16GB", FSType: "swap"},
					{MountPoint: "/", Size: "100GB", FSType: "xfs"},
					{MountPoint: "/var", Size: "200GB", FSType: "xfs"},
				},
				Network: models.NetworkConfig{
					Hostname: "prod-server",
					IP:       "10.0.1.100",
					Netmask:  "255.255.255.0",
					Gateway:  "10.0.1.1",
					DNS:      []string{"8.8.8.8", "8.8.4.4"},
				},
				Packages:   []string{"vim", "wget", "curl", "net-tools"},
				PostScript: "#!/bin/bash\necho 'Production setup complete'",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:     uuid.New().String(),
			Name:   "Ubuntu 20.04 Development",
			Distro: "ubuntu20",
			Config: models.ProfileConfig{
				RepoURL: "http://archive.ubuntu.com/ubuntu",
				Partitions: []models.Partition{
					{MountPoint: "/boot/efi", Size: "512MB", FSType: "vfat"},
					{MountPoint: "/", Size: "80GB", FSType: "ext4"},
					{MountPoint: "/home", Size: "200GB", FSType: "ext4"},
				},
				Network: models.NetworkConfig{
					Hostname: "dev-server",
					IP:       "10.0.1.101",
					Netmask:  "255.255.255.0",
					Gateway:  "10.0.1.1",
					DNS:      []string{"8.8.8.8"},
				},
				Packages:   []string{"build-essential", "git", "docker.io"},
				PostScript: "#!/bin/bash\nusermod -aG docker ubuntu",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:     uuid.New().String(),
			Name:   "CentOS 8 Minimal",
			Distro: "centos8",
			Config: models.ProfileConfig{
				RepoURL: "http://mirror.centos.org/centos/8-stream/BaseOS/x86_64/os",
				Partitions: []models.Partition{
					{MountPoint: "/boot", Size: "1GB", FSType: "ext4"},
					{MountPoint: "/", Size: "50GB", FSType: "xfs"},
				},
				Network: models.NetworkConfig{
					Hostname: "minimal-server",
					IP:       "10.0.1.102",
					Netmask:  "255.255.255.0",
					Gateway:  "10.0.1.1",
				},
				Packages: []string{"vim"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	db := database.GetDB()
	for _, profile := range profiles {
		db.Create(&profile)
	}

	return profiles
}

func seedMachines() []models.Machine {
	machines := []models.Machine{
		{
			ID:         uuid.New().String(),
			Hostname:   "server-001",
			MacAddress: "00:50:56:00:00:01",
			IPAddress:  "10.0.1.11",
			Status:     models.MachineStatusReady,
			HardwareSpec: models.HardwareInfo{
				SchemaVersion: "1.0",
				System: models.SystemInfo{
					Manufacturer: "Dell Inc.",
					ProductName:  "PowerEdge R740",
					SerialNumber: "SN001",
				},
				CPU: models.CPUInfo{
					Model: "Intel Xeon Gold 6248R",
					Cores: 48,
				},
				Memory: models.MemoryInfo{
					TotalGB: 256,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:         uuid.New().String(),
			Hostname:   "server-002",
			MacAddress: "00:50:56:00:00:02",
			IPAddress:  "10.0.1.12",
			Status:     models.MachineStatusReady,
			HardwareSpec: models.HardwareInfo{
				SchemaVersion: "1.0",
				System: models.SystemInfo{
					Manufacturer: "HP",
					ProductName:  "ProLiant DL380 Gen10",
					SerialNumber: "SN002",
				},
				CPU: models.CPUInfo{
					Model: "Intel Xeon Silver 4214",
					Cores: 24,
				},
				Memory: models.MemoryInfo{
					TotalGB: 128,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:         uuid.New().String(),
			Hostname:   "server-003",
			MacAddress: "00:50:56:00:00:03",
			IPAddress:  "10.0.1.13",
			Status:     models.MachineStatusInstalling,
			HardwareSpec: models.HardwareInfo{
				SchemaVersion: "1.0",
				System: models.SystemInfo{
					Manufacturer: "Supermicro",
					ProductName:  "X11DPi-NT",
					SerialNumber: "SN003",
				},
				CPU: models.CPUInfo{
					Model: "Intel Xeon Gold 6230",
					Cores: 40,
				},
				Memory: models.MemoryInfo{
					TotalGB: 192,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	db := database.GetDB()
	for _, machine := range machines {
		db.Create(&machine)
	}

	return machines
}

func seedJobs(machines []models.Machine, profiles []models.OSProfile) []models.Job {
	if len(machines) == 0 || len(profiles) == 0 {
		return nil
	}

	jobs := []models.Job{
		{
			ID:          uuid.New().String(),
			MachineID:   machines[0].ID,
			Type:        models.JobTypeAudit,
			Status:      models.JobStatusSuccess,
			StepCurrent: "Hardware audit completed",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          uuid.New().String(),
			MachineID:   machines[1].ID,
			Type:        models.JobTypeInstallOS,
			Status:      models.JobStatusRunning,
			StepCurrent: "Downloading packages",
			CreatedAt:   time.Now().Add(-30 * time.Minute),
			UpdatedAt:   time.Now().Add(-5 * time.Minute),
		},
		{
			ID:          uuid.New().String(),
			MachineID:   machines[2].ID,
			Type:        models.JobTypeConfigRAID,
			Status:      models.JobStatusPending,
			StepCurrent: "Waiting for execution",
			CreatedAt:   time.Now().Add(-10 * time.Minute),
			UpdatedAt:   time.Now().Add(-10 * time.Minute),
		},
		{
			ID:          uuid.New().String(),
			MachineID:   machines[0].ID,
			Type:        models.JobTypeInstallOS,
			Status:      models.JobStatusFailed,
			StepCurrent: "Installation failed",
			Error:       "Network timeout during package download",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-45 * time.Minute),
		},
	}

	db := database.GetDB()
	for _, job := range jobs {
		db.Create(&job)
	}

	return jobs
}

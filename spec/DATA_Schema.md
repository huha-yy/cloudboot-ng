# DATA_Schema.md

## 1. Core Data Models (Gorm Structs)
The database schema follows these Go struct definitions. Use Gorm tags for constraints.

### 1.1 Machine (物理机资产)
Represents a managed physical server.

type Machine struct {
    ID           string    `gorm:"primaryKey"` // UUID
    Hostname     string    `gorm:"uniqueIndex"`
    IPAddress    string    // Management IP (BMC or PXE IP)
    MacAddress   string    `gorm:"uniqueIndex"` // Primary Identifier
    Status       string    // Enum: "discovered", "ready", "installing", "active", "error"
    
    // Hardware Fingerprint (JSONB in SQLite)
    HardwareSpec HardwareInfo `gorm:"serializer:json"`
    
    CreatedAt    time.Time
    UpdatedAt    time.Time
}


### 1.2 Job (任务流水线)
Represents an asynchronous operation on a machine.


type Job struct {
    ID          string    `gorm:"primaryKey"`
    MachineID   string    `gorm:"index"`
    Type        string    // Enum: "audit", "config_raid", "install_os"
    Status      string    // Enum: "pending", "running", "success", "failed"
    StepCurrent string    // e.g., "downloading_provider"
    LogsPath    string    // Path to the full log file on disk
    Error       string    
    CreatedAt   time.Time
}


### 1.3 OSProfile (安装模板)
Represents the configuration for OS installation (Kickstart/Autoyast).


type OSProfile struct {
    ID          string    `gorm:"primaryKey"`
    Name        string
    Distro      string    // "centos7", "ubuntu22", "ky10"
    
    // Configuration Details (JSONB)
    Config      ProfileConfig `gorm:"serializer:json"`
}

type ProfileConfig struct {
    RootPasswordHash string
    Timezone         string
    Partitions       []Partition
    Network          NetworkConfig
    Packages         []string
}


### 1.4 License (商业授权)
Stores the offline license key for unlocking features.


type License struct {
    CustomerName string
    CustomerCode string    // Unique ID
    ProductKey   string    // Encrypted Master Key for decrypting Providers
    Features     []string  // ["audit", "offline_bundle", "multi_tenant"]
    ExpiresAt    time.Time
    Signature    string    // ECDSA Signature from CloudBoot Official
}


## 2. Hardware Fingerprint Schema (硬件指纹)
Standardized JSON format for `cb-probe` output and `Machine.HardwareSpec`.


{
  "schema_version": "1.0",
  "system": {
    "manufacturer": "Hygon",
    "product_name": "K100-W",
    "serial_number": "Unknown"
  },
  "cpu": {
    "arch": "x86_64",
    "model": "Hygon C86 7285",
    "cores": 32,
    "sockets": 2
  },
  "memory": {
    "total_bytes": 137438953472, // 128GB
    "dimms": [
      {"slot": "DIMM_A1", "size_bytes": 34359738368, "speed": 3200}
    ]
  },
  "storage_controllers": [
    {
      "pci_id": "1000:005f",
      "vendor": "LSI Logic / Symbios Logic",
      "model": "MegaRAID SAS 3108",
      "driver": "megaraid_sas"
    }
  ],
  "network_interfaces": [
    {
      "name": "eth0",
      "mac": "aa:bb:cc:dd:ee:01",
      "speed": 10000,
      "link": true
    }
  ]
}
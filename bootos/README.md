# CloudBoot BootOS Agent

BootOS is a lightweight Linux environment that runs on bare-metal servers during PXE boot to perform hardware detection, RAID configuration, and OS installation.

## Architecture

```
┌─────────────────────────────────────────────┐
│           Bare-Metal Server                 │
│  ┌─────────────────────────────────────┐   │
│  │  BootOS (Alpine Linux + cb-agent)   │   │
│  │  ┌──────────┐  ┌─────────┐         │   │
│  │  │ cb-probe │  │ cb-exec │         │   │
│  │  └────┬─────┘  └────┬────┘         │   │
│  │       │             │              │   │
│  │       └─────┬───────┘              │   │
│  │             ▼                      │   │
│  │        ┌─────────┐                │   │
│  │        │ cb-agent │◀──────────────┼───┼─► CloudBoot Server
│  │        └─────────┘   HTTP/HTTPS   │   │   (Task Polling + Logs)
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## Components

### 1. cb-agent (Main Agent)
- **Purpose**: Coordinates all operations on the bare-metal server
- **Functions**:
  - Hardware detection and registration
  - Task polling from CloudBoot server
  - Task execution orchestration
  - Log upload and status reporting
- **Location**: `cb-agent/main.go`, `cb-agent/pkg/agent/`

### 2. cb-probe (Hardware Detector)
- **Purpose**: Detects hardware specifications
- **Functions**:
  - System info (manufacturer, model, serial)
  - CPU detection (model, cores)
  - Memory detection (total RAM)
  - Disk detection (name, size, model)
  - Network interface detection (MAC, IP)
- **Location**: `cb-agent/pkg/hardware/detector.go`

### 3. cb-exec (Task Executor)
- **Purpose**: Executes tasks from CloudBoot server
- **Supported Tasks**:
  - `audit`: Hardware audit (passive detection)
  - `config_raid`: RAID configuration using provider scripts
  - `install_os`: OS installation using Kickstart/Preseed
- **Location**: `cb-agent/pkg/executor/executor.go`

### 4. HTTP Client
- **Purpose**: Communicates with CloudBoot server
- **Endpoints Used**:
  - `POST /api/boot/v1/register` - Register agent
  - `GET /api/boot/v1/task` - Poll for tasks
  - `POST /api/boot/v1/logs` - Upload logs
  - `POST /api/boot/v1/status` - Report status
- **Location**: `cb-agent/pkg/client/client.go`

## Task Execution Flow

```
┌────────────────────────────────────────────────────────────────────┐
│                      Agent Task Flow                               │
└────────────────────────────────────────────────────────────────────┘

1. Boot & Init
   │
   ├─> PXE Boot BootOS
   ├─> Configure Network (DHCP)
   ├─> Start cb-agent
   │
2. Hardware Detection (cb-probe)
   │
   ├─> Detect System Info (DMI/SMBIOS)
   ├─> Detect CPU (/proc/cpuinfo)
   ├─> Detect Memory (/proc/meminfo)
   ├─> Detect Disks (lsblk)
   ├─> Detect Network (ip command)
   │
3. Registration
   │
   ├─> POST /api/boot/v1/register
   ├─> Send: MAC, IP, Hardware Spec
   ├─> Receive: Machine ID
   │
4. Task Polling Loop (every 5s)
   │
   ├─> GET /api/boot/v1/task?machine_id=xxx
   ├─> If no task: continue polling
   ├─> If task received:
   │    │
   │    ├─> Report Status: "running"
   │    ├─> Execute Task (cb-exec)
   │    │    │
   │    │    ├─> audit: Hardware Audit
   │    │    ├─> config_raid: Run Provider Script
   │    │    └─> install_os: Fetch Config & Install
   │    │
   │    ├─> Upload Logs (POST /api/boot/v1/logs)
   │    └─> Report Status: "success" or "failed"
   │
   └─> Repeat
```

## Building BootOS ISO

```bash
# Build the Docker image (contains cb-agent)
cd bootos
docker build -t cloudboot/bootos:latest .

# Generate ISO using dracut (on a Linux host)
./build-iso.sh

# Output: bootos.iso (~150MB)
```

## Configuration

Environment variables:
- `CB_SERVER_URL`: CloudBoot server URL (default: `http://10.0.0.1:8080`)
- `CB_POLL_INTERVAL`: Task polling interval (default: `5s`)

## Usage

### 1. PXE Boot Configuration

Add to your DHCP/PXE server:

```
# /var/lib/tftpboot/pxelinux.cfg/default
LABEL cloudboot
    MENU LABEL CloudBoot BootOS
    KERNEL bootos/vmlinuz
    APPEND initrd=bootos/initrd.img boot=live fetch=http://pxe-server/bootos.iso CB_SERVER_URL=http://cloudboot.example.com:8080
```

### 2. Manual Testing

```bash
# Build agent binary
cd cb-agent
go build -o cb-agent main.go

# Run agent (requires root for hardware detection)
sudo ./cb-agent \
    --server=http://localhost:8080 \
    --poll-interval=5s \
    --debug
```

### 3. Docker Testing

```bash
# Run BootOS container
docker run --rm --privileged \
    -e CB_SERVER_URL=http://host.docker.internal:8080 \
    -e CB_POLL_INTERVAL=5s \
    cloudboot/bootos:latest
```

## Development

### Adding a New Task Type

1. **Define handler in executor**:
```go
// executor.go
func (e *Executor) handleMyTask(payload map[string]interface{}) *ExecutionResult {
    logs := []LogEntry{
        {Timestamp: time.Now().Format(time.RFC3339), Level: "INFO", Message: "Starting my task"},
    }
    // ... task logic ...
    return &ExecutionResult{Success: true, Logs: logs}
}
```

2. **Register handler**:
```go
// executor.go New()
e.RegisterHandler("my_task", e.handleMyTask)
```

3. **Test**:
```bash
curl -X POST http://localhost:8080/api/v1/machines/xxx/provision \
    -d '{"task_type":"my_task","payload":{"key":"value"}}'
```

## Testing

```bash
# Unit tests
go test ./cb-agent/pkg/...

# Integration test with mock server
go run ./test/mock-server.go &
./cb-agent --server=http://localhost:9999
```

## Security Considerations

1. **No Authentication**: BootOS assumes a secure, isolated provisioning network
2. **Root Privileges**: Agent runs as root for hardware access
3. **Provider Scripts**: Validate and sandbox provider script execution
4. **Network Isolation**: BootOS should only reach CloudBoot server and package repos

## Troubleshooting

### Agent can't register
- Check network connectivity: `ping $CB_SERVER_URL`
- Verify server is running: `curl $CB_SERVER_URL/health`
- Check logs: `journalctl -u cb-agent`

### Hardware detection fails
- Ensure running as root
- Check DMI access: `dmidecode -s system-manufacturer`
- Check disk access: `lsblk`

### Task execution fails
- Check task payload format
- Review executor logs
- Verify provider script exists (for RAID tasks)

## License

CloudBoot NG is licensed under Apache 2.0.

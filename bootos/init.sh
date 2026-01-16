#!/bin/bash
# CloudBoot BootOS Init Script (Systemd Compatible)
# 适配 OpenEuler 22.03 LTS 环境

set -e

# 日志函数
log_info() {
    echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a /var/log/cloudboot-init.log
}

log_error() {
    echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a /var/log/cloudboot-init.log >&2
}

log_info "========================================="
log_info " CloudBoot BootOS v4 - Initializing..."
log_info "========================================="

# 1. 配置运行时空间（Tmpfs 沙箱）
log_info "[1/5] Configuring runtime sandbox..."
if ! mountpoint -q /opt/cloudboot/runtime; then
    mount -t tmpfs -o size=2G,mode=0755 tmpfs /opt/cloudboot/runtime
    log_info "    Tmpfs mounted at /opt/cloudboot/runtime"
fi

# 创建运行时目录结构
mkdir -p /opt/cloudboot/runtime/{bin,lib,configs,logs,pipe}
chmod 755 /opt/cloudboot/runtime/bin
chmod 755 /opt/cloudboot/runtime/lib

# 2. 配置网络（NetworkManager）
log_info "[2/5] Configuring network..."

# 等待 NetworkManager 启动
for i in {1..30}; do
    if systemctl is-active --quiet NetworkManager; then
        log_info "    NetworkManager is running"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "    NetworkManager failed to start"
        exit 1
    fi
    sleep 1
done

# 启动所有网络接口
nmcli device connect "$(nmcli -t -f DEVICE,TYPE device | grep ethernet | head -1 | cut -d: -f1)" || true

# 等待网络连接
log_info "[2/5] Waiting for network connectivity..."
for i in {1..30}; do
    if ping -c 1 -W 1 ${CB_SERVER_URL#http://} >/dev/null 2>&1 || \
       ping -c 1 -W 1 8.8.8.8 >/dev/null 2>&1; then
        log_info "    Network is up"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "    Network timeout after 30 seconds"
        # 继续执行，可能在本地网络环境
    fi
    sleep 1
done

# 3. 显示系统信息
log_info "[3/5] Collecting system information..."
log_info "    Hostname: $(hostname)"
log_info "    Kernel: $(uname -r)"
log_info "    OS: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'\'' -f2)"
log_info "    IP Address: $(ip -4 addr show scope global | grep inet | awk '{print $2}' | head -1 || echo 'N/A')"
log_info "    MAC Address: $(ip link show | grep ether | awk '{print $2}' | head -1 || echo 'N/A')"

# 4. 硬件检测（cb-probe 功能）
log_info "[4/5] Detecting hardware..."

# CPU 信息
if [ -f /proc/cpuinfo ]; then
    CPU_MODEL=$(grep "model name" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
    CPU_CORES=$(grep -c "^processor" /proc/cpuinfo)
    log_info "    CPU: $CPU_MODEL ($CPU_CORES cores)"
fi

# 内存信息
if [ -f /proc/meminfo ]; then
    MEM_TOTAL=$(grep "MemTotal:" /proc/meminfo | awk '{print $2}')
    MEM_GB=$((MEM_TOTAL / 1024 / 1024))
    log_info "    Memory: ${MEM_GB}GB"
fi

# 磁盘信息
if command -v lsblk >/dev/null 2>&1; then
    DISK_COUNT=$(lsblk -d -n | grep -v "^loop" | wc -l)
    log_info "    Disks: $DISK_COUNT detected"
fi

# RAID 控制器检测
if command -v lspci >/dev/null 2>&1; then
    RAID_CONTROLLERS=$(lspci | grep -i raid | wc -l)
    if [ $RAID_CONTROLLERS -gt 0 ]; then
        log_info "    RAID Controllers: $RAID_CONTROLLERS detected"
    fi
fi

# NVMe 设备检测
if command -v nvme >/dev/null 2>&1; then
    NVME_COUNT=$(nvme list 2>/dev/null | grep "^/dev" | wc -l || echo 0)
    if [ $NVME_COUNT -gt 0 ]; then
        log_info "    NVMe Devices: $NVME_COUNT detected"
    fi
fi

# 5. 启动 cb-agent
log_info "[5/5] Starting CloudBoot Agent..."
log_info "    Server URL: ${CB_SERVER_URL}"
log_info "    Poll Interval: ${CB_POLL_INTERVAL}"
log_info "    Runtime: /opt/cloudboot/runtime"
log_info ""

# 启动 agent（前台运行，由 Systemd 管理）
exec /usr/local/bin/cb-agent \
    --server="${CB_SERVER_URL}" \
    --poll-interval="${CB_POLL_INTERVAL}" \
    --debug
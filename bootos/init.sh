#!/bin/bash
# CloudBoot BootOS Init Script
# This script initializes the BootOS environment and starts the agent

set -e

echo "========================================="
echo " CloudBoot BootOS - Initializing..."
echo "========================================="

# Configure network
echo "[*] Configuring network..."
if [ -f /etc/network/interfaces ]; then
    # Debian/Ubuntu style
    ifup -a
else
    # Alpine/busybox style
    udhcpc -i eth0 -s /usr/share/udhcpc/default.script
fi

# Wait for network
echo "[*] Waiting for network connectivity..."
for i in {1..30}; do
    if ping -c 1 -W 1 ${CB_SERVER_URL#http://} >/dev/null 2>&1 || ping -c 1 -W 1 8.8.8.8 >/dev/null 2>&1; then
        echo "[✓] Network is up"
        break
    fi
    echo "    Attempt $i/30..."
    sleep 1
done

# Display system info
echo ""
echo "System Information:"
echo "-------------------"
echo "Hostname: $(hostname)"
echo "Kernel: $(uname -r)"
echo "IP Address: $(ip -4 addr show scope global | grep inet | awk '{print $2}' | head -1)"
echo "MAC Address: $(ip link show | grep ether | awk '{print $2}' | head -1)"
echo ""

# Start cb-agent
echo "[*] Starting CloudBoot Agent..."
echo "    Server URL: ${CB_SERVER_URL}"
echo "    Poll Interval: ${CB_POLL_INTERVAL}"
echo ""

exec /usr/local/bin/cb-agent \
    --server="${CB_SERVER_URL}" \
    --poll-interval="${CB_POLL_INTERVAL}" \
    --debug

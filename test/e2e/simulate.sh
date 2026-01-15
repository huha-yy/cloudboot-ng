#!/bin/bash
# CloudBoot E2E Simulation with QEMU
# Simulates a bare-metal server provisioning workflow

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Configuration
CB_SERVER_URL="${CB_SERVER_URL:-http://10.0.2.2:8080}"
VM_NAME="${VM_NAME:-cloudboot-test-vm}"
VM_MEMORY="${VM_MEMORY:-2048}"
VM_CPUS="${VM_CPUS:-2}"
VM_DISK_SIZE="${VM_DISK_SIZE:-20G}"
VM_MAC="${VM_MAC:-52:54:00:12:34:56}"

echo "=========================================="
echo " CloudBoot E2E Simulation with QEMU"
echo "=========================================="
echo ""
echo "Configuration:"
echo "  Server URL: $CB_SERVER_URL"
echo "  VM Name: $VM_NAME"
echo "  Memory: $VM_MEMORY MB"
echo "  CPUs: $VM_CPUS"
echo "  Disk Size: $VM_DISK_SIZE"
echo "  MAC Address: $VM_MAC"
echo ""

# Check dependencies
command -v qemu-system-x86_64 >/dev/null 2>&1 || {
    echo "Error: qemu-system-x86_64 is required"
    echo "Install with: apt-get install qemu-kvm"
    exit 1
}

# Create VM disk
DISK_PATH="/tmp/${VM_NAME}.qcow2"
if [ ! -f "$DISK_PATH" ]; then
    echo "[*] Creating VM disk..."
    qemu-img create -f qcow2 "$DISK_PATH" "$VM_DISK_SIZE"
fi

# Check if BootOS ISO exists
ISO_PATH="$PROJECT_ROOT/bootos/bootos.iso"
if [ ! -f "$ISO_PATH" ]; then
    echo "[!] BootOS ISO not found at $ISO_PATH"
    echo "    Build it with: cd bootos && ./build-iso.sh"
    echo ""
    echo "    Using network boot simulation instead..."
    BOOT_MODE="network"
else
    echo "[*] Using BootOS ISO: $ISO_PATH"
    BOOT_MODE="iso"
fi

# Start QEMU VM
echo ""
echo "[*] Starting QEMU VM..."
echo "    Press Ctrl+Alt+2 for QEMU monitor"
echo "    Press Ctrl+Alt+1 to return to VM console"
echo ""

if [ "$BOOT_MODE" == "iso" ]; then
    # Boot from ISO
    qemu-system-x86_64 \
        -name "$VM_NAME" \
        -m "$VM_MEMORY" \
        -smp "$VM_CPUS" \
        -drive file="$DISK_PATH",format=qcow2 \
        -cdrom "$ISO_PATH" \
        -boot d \
        -netdev user,id=net0,hostfwd=tcp::2222-:22 \
        -device virtio-net-pci,netdev=net0,mac="$VM_MAC" \
        -nographic \
        -serial mon:stdio
else
    # Network boot simulation (PXE-style)
    # Note: This requires additional PXE setup or iPXE
    echo "[!] Network boot mode requires PXE server setup"
    echo "    For quick testing, use ISO mode or Docker simulation"
    exit 1
fi

echo ""
echo "[*] VM stopped"
echo ""
echo "To delete VM disk: rm $DISK_PATH"

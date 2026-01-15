#!/bin/bash
# CloudBoot BootOS ISO Builder
# Builds a bootable ISO containing the cb-agent

set -e

echo "========================================="
echo " CloudBoot BootOS ISO Builder"
echo "========================================="

# Check dependencies
command -v docker >/dev/null 2>&1 || { echo "Error: docker is required"; exit 1; }

# Build Docker image
echo "[1/3] Building Docker image..."
docker build -t cloudboot/bootos:latest .

# Export filesystem
echo "[2/3] Exporting filesystem..."
mkdir -p build
docker create --name bootos-tmp cloudboot/bootos:latest
docker export bootos-tmp | tar -C build -xf -
docker rm bootos-tmp

# Create ISO (requires genisoimage or mkisofs)
echo "[3/3] Creating ISO..."
if command -v genisoimage >/dev/null 2>&1; then
    ISO_CMD="genisoimage"
elif command -v mkisofs >/dev/null 2>&1; then
    ISO_CMD="mkisofs"
else
    echo "Warning: genisoimage/mkisofs not found. ISO creation skipped."
    echo "Install with: apt-get install genisoimage"
    exit 0
fi

$ISO_CMD \
    -o bootos.iso \
    -b isolinux/isolinux.bin \
    -c isolinux/boot.cat \
    -no-emul-boot \
    -boot-load-size 4 \
    -boot-info-table \
    -J -R -V "CloudBoot BootOS" \
    build/

echo ""
echo "✓ BootOS ISO built successfully: bootos.iso"
echo "  Size: $(du -h bootos.iso | cut -f1)"
echo ""
echo "Next steps:"
echo "  1. Copy bootos.iso to your PXE server"
echo "  2. Configure PXE boot menu"
echo "  3. Boot a bare-metal server via PXE"

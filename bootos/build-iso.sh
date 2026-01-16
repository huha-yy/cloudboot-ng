#!/bin/bash
# CloudBoot BootOS ISO Builder (OpenEuler Based)
# 构建支持双模引导（Legacy BIOS + UEFI）的 BootOS ISO

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
ISO_OUTPUT="$SCRIPT_DIR/bootos.iso"

echo "========================================="
echo " CloudBoot BootOS ISO Builder v4"
echo "========================================="
echo ""

# 检查依赖
command -v docker >/dev/null 2>&1 || { echo "Error: docker is required"; exit 1; }
command -v xorriso >/dev/null 2>&1 || { echo "Error: xorriso is required"; exit 1; }

# 清理旧构建
echo "[1/5] Cleaning previous build..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# 构建 Docker 镜像
echo "[2/5] Building Docker image..."
docker build -t cloudboot/bootos:latest .

# 导出 rootfs
echo "[3/5] Exporting rootfs..."
docker create --name bootos-tmp cloudboot/bootos:latest
docker export bootos-tmp | tar -C "$BUILD_DIR/rootfs" -xf -
docker rm bootos-tmp

# 提取内核和 initrd
echo "[4/5] Extracting kernel and initrd..."

# 从 rootfs 中提取内核
if [ -f "$BUILD_DIR/rootfs/boot/vmlinuz-*" ]; then
    cp "$BUILD_DIR/rootfs/boot/vmlinuz-"* "$BUILD_DIR/vmlinuz"
    echo "    Kernel: $(basename $BUILD_DIR/vmlinuz)"
else
    echo "Error: Kernel not found in rootfs"
    exit 1
fi

# 使用 dracut 生成 initrd（在容器内执行）
echo "    Generating initrd with dracut..."
docker run --rm \
    -v "$BUILD_DIR/rootfs:/rootfs:ro" \
    -v "$BUILD_DIR:/output" \
    --privileged \
    openeuler/openeuler:22.03-lts \
    bash -c "
        # 挂载 rootfs
        mount --bind /rootfs /mnt

        # 安装 dracut
        dnf install -y dracut

        # 生成 initrd
        dracut \
            --force \
            --no-hostonly \
            --add 'dmsquash-live systemd base' \
            --include '/usr/local/bin/cb-agent' \
            --include '/etc/systemd/system/cloudboot.service' \
            --include '/usr/local/bin/cloudboot-init.sh' \
            --modules 'network' \
            /output/initrd.img \$(uname -r)

        umount /mnt
    "

if [ ! -f "$BUILD_DIR/initrd.img" ]; then
    echo "Error: initrd.img not generated"
    exit 1
fi

echo "    Initrd: $(du -h $BUILD_DIR/initrd.img | cut -f1)"

# 创建 ISO（支持双模引导）
echo "[5/5] Creating ISO with dual-boot support..."

# 创建 ISO 目录结构
mkdir -p "$BUILD_DIR/iso/isolinux"
mkdir -p "$BUILD_DIR/iso/EFI/BOOT"

# 复制内核和 initrd
cp "$BUILD_DIR/vmlinuz" "$BUILD_DIR/iso/"
cp "$BUILD_DIR/initrd.img" "$BUILD_DIR/iso/"

# Legacy BIOS 引导配置（Isolinux）
cat > "$BUILD_DIR/iso/isolinux/isolinux.cfg" << 'EOF'
DEFAULT cloudboot
LABEL cloudboot
  MENU LABEL CloudBoot BootOS v4
  KERNEL /vmlinuz
  APPEND initrd=/initrd.img boot=live quiet console=ttyS0 CB_SERVER_URL=http://10.0.2.2:8080 CB_POLL_INTERVAL=5s
  TIMEOUT 50
PROMPT 0
EOF

# UEFI 引导配置（Grub2）
cat > "$BUILD_DIR/iso/EFI/BOOT/grub.cfg" << 'EOF'
set timeout=5
menuentry "CloudBoot BootOS v4" {
    linux /vmlinuz boot=live quiet console=ttyS0 CB_SERVER_URL=http://10.0.2.2:8080 CB_POLL_INTERVAL=5s
    initrd /initrd.img
}
EOF

# 需要的引导文件（从 OpenEuler 获取）
docker run --rm \
    -v "$BUILD_DIR/iso:/output" \
    openeuler/openeuler:22.03-lts \
    bash -c "
        dnf install -y isolinux grub2-efi-x64 grub2-pc-modules shim-x64
        cp /usr/share/syslinux/isolinux.bin /output/isolinux/
        cp /usr/share/syslinux/ldlinux.c32 /output/isolinux/
        cp /boot/efi/EFI/BOOT/BOOTX64.EFI /output/EFI/BOOT/ 2>/dev/null || true
        cp /usr/share/grub2/x86_64-efi/grub.efi /output/EFI/BOOT/BOOTX64.EFI 2>/dev/null || true
    "

# 生成 ISO（支持双模引导）
xorriso \
    -as mkisofs \
    -isohybrid-mbr /usr/lib/ISOLINUX/isohdpfx.bin \
    -c isolinux/boot.cat \
    -b isolinux/isolinux.bin \
    -no-emul-boot \
    -boot-load-size 4 \
    -boot-info-table \
    -eltorito-alt-boot \
    -e EFI/BOOT/BOOTX64.EFI \
    -no-emul-boot \
    -isohybrid-gpt-basdat \
    -V "CloudBoot BootOS v4" \
    -o "$ISO_OUTPUT" \
    "$BUILD_DIR/iso"

# 清理
rm -rf "$BUILD_DIR"

echo ""
echo "✓ BootOS ISO built successfully!"
echo "  Output: $ISO_OUTPUT"
echo "  Size: $(du -h $ISO_OUTPUT | cut -f1)"
echo ""
echo "Next steps:"
echo "  1. Test with QEMU: qemu-system-x86_64 -cdrom $ISO_OUTPUT -m 2048"
echo "  2. Configure PXE server with vmlinuz and initrd.img"
echo "  3. Boot bare-metal server via PXE or ISO"
echo ""
echo "Supported boot modes:"
echo "  ✓ Legacy BIOS (Isolinux)"
echo "  ✓ UEFI (Grub2)"

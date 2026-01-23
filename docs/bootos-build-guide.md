# BootOS ISO 构建与验证指南

## 概述

BootOS 是 CloudBoot NG 的轻量级 Linux 启动环境，用于裸机服务器的硬件发现、RAID 配置和 OS 安装。

**基础镜像**: OpenEuler 22.03 LTS
**构建工具**: Docker + xorriso
**启动模式**: Legacy BIOS + UEFI 双模支持
**目标大小**: < 500MB
**启动时间**: < 15秒

---

## 构建前置条件

### 系统要求
- **操作系统**: Linux (推荐 Ubuntu 20.04+ 或 CentOS 7+)
- **架构**: x86_64
- **磁盘空间**: 至少 5GB 可用空间

### 依赖工具

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y docker.io xorriso isolinux

# CentOS/RHEL
sudo yum install -y docker xorriso syslinux

# 启动 Docker
sudo systemctl start docker
sudo systemctl enable docker
```

---

## 快速构建

### 1. 构建 ISO

```bash
cd bootos
sudo ./build-iso.sh
```

**构建流程**（5个步骤）：
1. 清理旧构建
2. 构建 Docker 镜像（包含 cb-agent）
3. 导出 rootfs
4. 提取内核和 initrd
5. 创建双模引导 ISO

**预计耗时**: 5-10分钟（首次构建）

**输出文件**: `bootos/bootos.iso`

### 2. 验证构建

```bash
# 检查 ISO 文件大小
ls -lh bootos/bootos.iso

# 验证 ISO 结构
xorriso -indev bootos/bootos.iso -find
```

**验收标准**:
- ✅ ISO 文件存在
- ✅ 文件大小 < 500MB
- ✅ 包含 `/vmlinuz` 和 `/initrd.img`
- ✅ 包含 `/isolinux/isolinux.cfg` (Legacy BIOS)
- ✅ 包含 `/EFI/BOOT/grub.cfg` (UEFI)

---

## 测试方法

### 方法 1: QEMU 虚拟机测试（推荐）

```bash
# 安装 QEMU
sudo apt-get install -y qemu-system-x86

# 启动 CloudBoot Server（另一个终端）
cd /path/to/cloudboot-ng
go run cmd/server/main.go

# 使用 QEMU 测试 ISO
qemu-system-x86_64 \
    -cdrom bootos/bootos.iso \
    -m 2048 \
    -net nic \
    -net user,hostfwd=tcp::8080-:8080 \
    -boot d \
    -serial stdio
```

**预期结果**:
1. BootOS 启动（< 15秒）
2. 网络配置成功（DHCP）
3. cb-agent 启动并连接到 CloudBoot Server
4. 硬件探测成功（CPU、内存、磁盘）
5. Agent 注册成功（在 Server 日志中看到注册请求）

### 方法 2: PXE 网络启动测试

#### 2.1 配置 TFTP 服务器

```bash
# 安装 TFTP 服务器
sudo apt-get install -y tftpd-hpa

# 提取内核和 initrd
mkdir -p /var/lib/tftpboot/bootos
sudo cp bootos/build/vmlinuz /var/lib/tftpboot/bootos/
sudo cp bootos/build/initrd.img /var/lib/tftpboot/bootos/

# 配置 PXE 菜单
sudo tee /var/lib/tftpboot/pxelinux.cfg/default << 'EOF'
DEFAULT cloudboot
LABEL cloudboot
    MENU LABEL CloudBoot BootOS v4
    KERNEL bootos/vmlinuz
    APPEND initrd=bootos/initrd.img boot=live CB_SERVER_URL=http://10.0.2.2:8080 CB_POLL_INTERVAL=5s
    TIMEOUT 50
PROMPT 0
EOF

# 重启 TFTP 服务
sudo systemctl restart tftpd-hpa
```

#### 2.2 配置 DHCP 服务器

```bash
# 编辑 /etc/dhcp/dhcpd.conf
subnet 10.0.2.0 netmask 255.255.255.0 {
    range 10.0.2.100 10.0.2.200;
    option routers 10.0.2.1;
    option domain-name-servers 8.8.8.8;

    # PXE 配置
    next-server 10.0.2.2;  # TFTP 服务器 IP
    filename "pxelinux.0";
}

# 重启 DHCP 服务
sudo systemctl restart isc-dhcp-server
```

#### 2.3 启动测试服务器

```bash
# 启动 CloudBoot Server
cd /path/to/cloudboot-ng
go run cmd/server/main.go

# 使用 QEMU 测试 PXE 启动
qemu-system-x86_64 \
    -m 2048 \
    -net nic \
    -net user,tftp=/var/lib/tftpboot,bootfile=pxelinux.0 \
    -boot n \
    -serial stdio
```

### 方法 3: 真实硬件测试

**前置条件**:
- 1台测试服务器（支持 PXE 启动）
- 配置好的 DHCP + TFTP 服务器
- CloudBoot Server 运行中

**步骤**:
1. 配置测试服务器 BIOS 为 PXE 启动优先
2. 连接到配置好的网络
3. 上电启动
4. 观察启动日志
5. 在 CloudBoot Web UI 中查看机器注册状态

---

## 验收清单

### 构建验收

- [ ] `build-iso.sh` 执行成功，无错误
- [ ] `bootos.iso` 文件生成
- [ ] ISO 文件大小 < 500MB
- [ ] ISO 包含 vmlinuz 和 initrd.img
- [ ] ISO 支持 Legacy BIOS 启动
- [ ] ISO 支持 UEFI 启动

### 功能验收

- [ ] BootOS 启动时间 < 15秒
- [ ] 网络配置成功（DHCP 或静态 IP）
- [ ] cb-agent 自动启动
- [ ] Agent 连接到 CloudBoot Server
- [ ] 硬件探测成功（CPU、内存、磁盘、网卡）
- [ ] Agent 注册成功（Machine ID 生成）
- [ ] 任务轮询正常（每 5 秒）

### 硬件兼容性验收

- [ ] 在 QEMU 虚拟机中启动成功
- [ ] 在 VMware/VirtualBox 中启动成功
- [ ] 在真实物理服务器上启动成功
- [ ] 支持 Legacy BIOS 服务器
- [ ] 支持 UEFI 服务器
- [ ] 网卡驱动正常加载
- [ ] 磁盘控制器驱动正常加载

---

## 故障排查

### 问题 1: 构建失败 - Docker 镜像拉取超时

**症状**: `docker build` 卡在拉取 `openeuler/openeuler:22.03-lts`

**解决方案**:
```bash
# 配置 Docker 镜像加速
sudo tee /etc/docker/daemon.json << 'EOF'
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ]
}
EOF

sudo systemctl restart docker
```

### 问题 2: xorriso 命令未找到

**症状**: `build-iso.sh` 报错 `xorriso: command not found`

**解决方案**:
```bash
# Ubuntu/Debian
sudo apt-get install -y xorriso

# CentOS/RHEL
sudo yum install -y xorriso
```

### 问题 3: ISO 启动后网络不通

**症状**: BootOS 启动成功，但无法连接到 CloudBoot Server

**排查步骤**:
1. 检查网络配置: `ip addr show`
2. 检查路由: `ip route show`
3. 测试连通性: `ping 8.8.8.8`
4. 检查 DNS: `nslookup cloudboot.example.com`
5. 检查防火墙: `iptables -L`

**解决方案**:
```bash
# 手动配置网络（在 BootOS 中）
nmcli device connect eth0
nmcli connection modify eth0 ipv4.addresses 10.0.2.100/24
nmcli connection modify eth0 ipv4.gateway 10.0.2.1
nmcli connection modify eth0 ipv4.dns 8.8.8.8
nmcli connection up eth0
```

### 问题 4: Agent 无法注册

**症状**: Agent 启动成功，但 CloudBoot Server 未收到注册请求

**排查步骤**:
1. 检查 Server 是否运行: `curl http://10.0.2.2:8080/health`
2. 检查 Agent 日志: `journalctl -u cloudboot -f`
3. 检查环境变量: `echo $CB_SERVER_URL`

**解决方案**:
```bash
# 手动启动 Agent（调试模式）
/usr/local/bin/cb-agent \
    --server=http://10.0.2.2:8080 \
    --poll-interval=5s \
    --debug
```

---

## 性能优化

### 减小 ISO 体积

1. **移除不必要的软件包**（编辑 `Dockerfile`）:
```dockerfile
# 仅保留必要工具
RUN dnf install -y \
    systemd \
    NetworkManager \
    iproute \
    pciutils \
    dmidecode \
    curl \
    && dnf clean all
```

2. **压缩 initrd**:
```bash
# 在 build-iso.sh 中添加压缩选项
dracut --force --no-hostonly --compress=xz ...
```

### 加速启动时间

1. **禁用不必要的 Systemd 服务**:
```bash
# 在 Dockerfile 中
RUN systemctl mask systemd-udev-settle.service
RUN systemctl mask systemd-networkd-wait-online.service
```

2. **优化网络配置**:
```bash
# 在 init.sh 中使用静态 IP（如果可能）
nmcli connection modify eth0 ipv4.method manual
```

---

## 生产部署建议

### 1. ISO 分发

**方案 A: HTTP 服务器**
```bash
# 将 ISO 放到 Web 服务器
cp bootos.iso /var/www/html/
# PXE 配置中使用: fetch=http://pxe-server/bootos.iso
```

**方案 B: NFS 共享**
```bash
# 挂载 NFS 共享
mount -t nfs pxe-server:/exports/bootos /mnt
# PXE 配置中使用: root=/dev/nfs nfsroot=pxe-server:/exports/bootos
```

### 2. 版本管理

```bash
# 使用版本号命名
mv bootos.iso bootos-v4.0.0.iso

# 创建符号链接
ln -sf bootos-v4.0.0.iso bootos-latest.iso
```

### 3. 安全加固

- 使用 HTTPS 传输 ISO
- 验证 ISO 签名（SHA256）
- 限制 PXE 网络访问（VLAN 隔离）
- 使用 TLS 加密 Agent 与 Server 通信

---

## 参考资料

- [OpenEuler 官方文档](https://docs.openeuler.org/)
- [Dracut 手册](https://man7.org/linux/man-pages/man8/dracut.8.html)
- [xorriso 文档](https://www.gnu.org/software/xorriso/)
- [PXE 启动原理](https://en.wikipedia.org/wiki/Preboot_Execution_Environment)

---

## 更新日志

| 日期 | 版本 | 更新内容 |
|------|------|---------|
| 2026-01-20 | v1.0 | 初始版本 - 完整的构建和验证指南 |

---

**状态**: ✅ 构建基础设施已完整实现，待 Linux 环境验证

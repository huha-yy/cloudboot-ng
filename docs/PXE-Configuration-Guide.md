# PXE/iPXE 网络启动配置指南

本文档说明如何配置DHCP服务器以支持CloudBoot NG的PXE网络启动。

---

## 🏗️ 架构概述

PXE启动流程：

```
+----------+     DHCP      +-----------+
| 裸机服务器 | ------------> | DHCP服务器  |
+----------+               +-----------+
     |                           |
     | (获取IP + TFTP地址)        |
     v                           |
+----------+     TFTP      +-----------+
| 下载iPXE  | ------------> | TFTP服务器  |
+----------+               +-----------+
     |                           |
     | (加载iPXE固件)             |
     v                           |
+----------+     HTTP      +-----------+
| 获取脚本  | ------------> | CloudBoot   |
+----------+               |   Core      |
     |                     +-----------+
     | (iPXE脚本 + Kernel + Initrd)
     v
+----------+
| 启动系统  |
+----------+
```

---

## 📋 方案选择

CloudBoot支持两种PXE启动方案：

### 方案1: TFTP + iPXE（推荐）

- **优势**: 兼容性最好，支持所有传统PXE固件
- **DHCP配置**: 需要配置`next-server`和`filename`
- **TFTP服务器**: 使用CloudBoot内置TFTP或外部TFTP

### 方案2: HTTP Boot（现代）

- **优势**: 速度快，不需要TFTP服务器
- **DHCP配置**: 需要配置Option 67（bootfile-name）
- **要求**: 服务器UEFI固件支持HTTP Boot

---

## 🔧 DHCP配置示例

### ISC DHCP Server (dhcpd.conf)

#### 方案1: TFTP + iPXE

```bash
# /etc/dhcp/dhcpd.conf

subnet 10.0.0.0 netmask 255.255.255.0 {
  range 10.0.0.100 10.0.0.200;
  option routers 10.0.0.1;
  option domain-name-servers 8.8.8.8;

  # PXE启动配置
  next-server 10.0.0.10;              # CloudBoot Core服务器IP

  # BIOS模式
  if exists user-class and option user-class = "iPXE" {
    # 如果已经是iPXE,直接加载HTTP脚本
    filename "http://10.0.0.10:8080/boot/ipxe/${net0/mac}";
  } else {
    # 否则先加载iPXE固件
    filename "undionly.kpxe";         # BIOS固件
  }

  # UEFI模式
  if substring(option vendor-class-identifier, 0, 10) = "HTTPClient" {
    # HTTP Boot (UEFI原生支持)
    option vendor-class-identifier "HTTPClient";
    filename "http://10.0.0.10:8080/boot/ipxe/${net0/mac}";
  } elsif substring(option vendor-class-identifier, 0, 9) = "PXEClient" {
    if substring(option vendor-class-identifier, 15, 5) = "00007" {
      # UEFI x64
      filename "ipxe.efi";
    } elsif substring(option vendor-class-identifier, 15, 5) = "00009" {
      # UEFI x64
      filename "ipxe.efi";
    } else {
      # UEFI ARM64
      filename "ipxe-arm64.efi";
    }
  }
}
```

#### 方案2: 纯HTTP Boot (UEFIOnly)

```bash
subnet 10.0.0.0 netmask 255.255.255.0 {
  range 10.0.0.100 10.0.0.200;
  option routers 10.0.0.1;
  option domain-name-servers 8.8.8.8;

  # HTTP Boot配置(Option 67)
  option bootfile-name "http://10.0.0.10:8080/boot/ipxe/${net0/mac}";
}
```

---

### Dnsmasq (推荐用于小型环境)

```bash
# /etc/dnsmasq.conf

# DHCP范围
dhcp-range=10.0.0.100,10.0.0.200,12h

# DNS
dhcp-option=option:dns-server,8.8.8.8

# PXE启动配置
dhcp-boot=tag:!ipxe,undionly.kpxe,10.0.0.10
dhcp-boot=tag:ipxe,http://10.0.0.10:8080/boot/ipxe/${net0/mac}

# UEFI模式
dhcp-match=set:efi-x86_64,option:client-arch,7
dhcp-match=set:efi-x86_64,option:client-arch,9
dhcp-boot=tag:efi-x86_64,ipxe.efi,10.0.0.10

# 识别iPXE
dhcp-match=set:ipxe,175
```

---

## 📦 TFTP文件准备

如果使用TFTP方案，需要准备iPXE固件文件：

### 1. 下载iPXE固件

```bash
# 创建TFTP根目录
mkdir -p /opt/cloudboot/tftpboot

cd /opt/cloudboot/tftpboot

# 下载预编译的iPXE固件
wget http://boot.ipxe.org/undionly.kpxe    # BIOS
wget http://boot.ipxe.org/ipxe.efi         # UEFI x64
wget http://boot.ipxe.org/ipxe-arm64.efi   # UEFI ARM64
```

### 2. 自定义编译iPXE（可选）

如果需要嵌入自定义脚本：

```bash
git clone https://github.com/ipxe/ipxe.git
cd ipxe/src

# 创建嵌入脚本
cat > embed.ipxe <<'EOF'
#!ipxe
dhcp
chain http://10.0.0.10:8080/boot/ipxe/${net0/mac}
EOF

# 编译
make bin/undionly.kpxe EMBED=embed.ipxe
make bin-x86_64-efi/ipxe.efi EMBED=embed.ipxe

# 复制到TFTP目录
cp bin/undionly.kpxe /opt/cloudboot/tftpboot/
cp bin-x86_64-efi/ipxe.efi /opt/cloudboot/tftpboot/
```

---

## 🚀 启动CloudBoot TFTP服务器

### 使用内置TFTP服务器

```bash
# 编辑配置
export TFTP_ENABLED=true
export TFTP_PORT=69
export TFTP_ROOT=/opt/cloudboot/tftpboot

# 启动CloudBoot
./cloudboot-ng
```

### 使用外部TFTP服务器（推荐生产环境）

```bash
# 安装tftpd-hpa (Debian/Ubuntu)
apt-get install tftpd-hpa

# 配置 /etc/default/tftpd-hpa
TFTP_USERNAME="tftp"
TFTP_DIRECTORY="/opt/cloudboot/tftpboot"
TFTP_ADDRESS="0.0.0.0:69"
TFTP_OPTIONS="--secure"

# 启动服务
systemctl enable tftpd-hpa
systemctl start tftpd-hpa
```

---

## 🧪 测试PXE启动

### 1. 检查DHCP响应

```bash
# 模拟PXE客户端请求
nmap --script broadcast-dhcp-discover -e eth0
```

### 2. 检查TFTP服务

```bash
# 测试TFTP下载
tftp 10.0.0.10
tftp> get undionly.kpxe
tftp> quit

# 验证文件
ls -lh undionly.kpxe
```

### 3. 检查HTTP Boot API

```bash
# 测试iPXE脚本生成
curl http://10.0.0.10:8080/boot/ipxe/00:11:22:33:44:55
```

预期输出：

```
#!ipxe
###############################################################################
# CloudBoot NG - iPXE Boot Script
# Generated for: unknown-334455 (00:11:22:33:44:55)
# Machine ID: unknown
# Boot Mode: discovery
###############################################################################
...
```

---

## 🔍 故障排查

### 问题1: 无法获取IP地址

**检查**:
```bash
# 查看DHCP服务器日志
tail -f /var/log/syslog | grep dhcp

# 检查网络连接
tcpdump -i eth0 port 67 or port 68
```

**解决**:
- 确认DHCP服务器正在运行
- 检查网络交换机是否支持PXE（Option 82）
- 验证VLAN配置

### 问题2: TFTP超时

**检查**:
```bash
# 检查TFTP服务
netstat -ulnp | grep :69

# 测试TFTP连接
tftp 10.0.0.10 -c get undionly.kpxe
```

**解决**:
- 检查防火墙规则（UDP 69端口）
- 确认TFTP根目录权限
- 验证文件存在

### 问题3: iPXE脚本加载失败

**检查**:
```bash
# 查看CloudBoot日志
journalctl -u cloudboot -f

# 测试HTTP访问
curl -v http://10.0.0.10:8080/boot/ipxe/00:11:22:33:44:55
```

**解决**:
- 确认CloudBoot服务正在运行
- 检查防火墙规则（TCP 8080端口）
- 验证MAC地址格式

---

## 📚 参考链接

- [iPXE官方文档](https://ipxe.org/docs)
- [ISC DHCP文档](https://www.isc.org/dhcp/)
- [Dnsmasq文档](http://www.thekelleys.org.uk/dnsmasq/doc.html)
- [RFC 4578 - DHCP PXE Options](https://tools.ietf.org/html/rfc4578)

---

**文档更新时间**: 2026-01-19

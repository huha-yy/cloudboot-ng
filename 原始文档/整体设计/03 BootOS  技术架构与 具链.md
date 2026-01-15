这是一份关于 CloudBoot NG 执行底座的**最高级别技术规范文档**。

它详细定义了 BootOS v4 如何作为一个**“无状态的数字雇佣兵”**，潜入物理机内部，执行高难度的硬件治理任务。同时，它深入解构了 **CB-Kit（底层工具链）** 的内部实现，这是我们区别于传统 Shell 脚本流派的核心技术壁垒。

---

# CloudBoot NG 全景产品设计白皮书
**第三卷：BootOS v4 技术架构与 CB-Kit 工具链详设 (The Body)**

+ **项目代号**：Phoenix Core (OS Layer)
+ **版本**：v1.0
+ **日期**：2026年1月
+ **密级**：核心机密 (Core Confidential)
+ **关联文档**：第一卷 (战略)，第二卷 (Core 平台)

---

## 1. 概述 (Overview)
### 1.1 设计愿景
**BootOS v4** 是 CloudBoot NG 的**数据面（Data Plane）**。它不是一个为了运行业务应用而设计的通用操作系统，而是一个**专用的、瞬时的、驻留内存的硬件治理微内核**。

它的核心使命是在**不依赖本地硬盘、不修改客户现有环境**的前提下，为上层 Provider（墨盒）提供一个统一、稳定、兼容所有信创硬件（x86/ARM/LoongArch）的**沙箱运行环境**。

### 1.2 核心设计哲学
1. **Immutable (不可变)**：镜像构建完成后即只读。运行时没有任何“配置漂移”。
2. **Stateless (无状态)**：**重启即焚**。所有运行时数据存储在 RAM 中，断电后彻底消失，满足银行最高级别的安全合规要求。
3. **Micro (微型化)**：摒弃通用 Linux 发行版的臃肿，目标体积 **< 150MB**，启动时间 **< 15秒**。

---

## 2. BootOS v4 操作系统架构 (OS Architecture)
我们采用 **“三明治”分层架构**，利用 Docker 容器化流水线进行构建。

### 2.1 Layer 0: 基础底座 (The Foundation)
+ **选型决策**：**OpenEuler 22.03 LTS SP3** (及后续版本)。
+ **选型逻辑**：
    - **信创原生支持**：作为国产根社区，OpenEuler 拥有最完整的华为鲲鹏 (ARM64)、中科海光 (x86_64)、飞腾、龙芯的原生驱动支持。这解决了传统 CentOS “找不到网卡/RAID卡”的致命痛点。
    - **内核兼容性**：Kernel 5.10+，原生支持 NVMe-oF、最新的 RAID 控制器及国密算法。
+ **构建方式**：**From Scratch**。不使用官方 ISO 安装，而是通过 `dnf --installroot` 仅提取内核 (`vmlinuz`)、引导加载程序 (`grub2/shim`) 和最小根文件系统。

### 2.2 Layer 1: 用户空间 (The Minimalist Userland)
为了极致压缩体积，我们对用户空间进行了手术刀式的裁剪。

+ **保留组件**：
    - **Init System**: `Systemd` (精简版)。保留 Systemd 是为了获得强大的服务依赖管理和日志接管能力（Journald）。
    - **Network Stack**: `NetworkManager` + `iproute2`。支持复杂的企业级网络环境（LACP Bonding, VLAN Tagging, PPPoE）。
    - **Hardware Tools**: `pciutils` (lspci), `dmidecode`, `ethtool`, `nvme-cli`, `lsscsi`。
    - **Crypto**: `openssl` (支持国密 SM2/3/4)。
+ **剔除组件**：
    - **Shell**: 仅保留 `bash` 用于紧急调试，但在生产模式下不提供 TTY 登录。
    - **Interpreters**: **彻底移除 Python, Perl, Ruby**。所有上层逻辑由 Go 语言实现，消除解释器依赖带来的体积膨胀和漏洞风险。
    - **SSH Server**: 默认移除。BootOS 仅通过反向 HTTP 长连接与 Server 通信，杜绝端口扫描风险。

### 2.3 Layer 2: 运行时空间 (The Runtime Sandbox)
这是 BootOS 唯一的“可写”区域，也是“墨盒”运行的场所。

+ **介质**：**Tmpfs (内存文件系统)**。
+ **挂载点**：`/opt/cloudboot/runtime`。
+ **目录结构**：

```latex
/opt/cloudboot/runtime/
├── bin/          # 存放下载并解密后的 Provider 二进制 (墨盒)
├── lib/          # 存放动态链接库 (如有)
├── configs/      # 存放渲染后的配置文件 (ks.cfg, overlay.json)
├── logs/         # 硬件工具的原始输出缓冲区
└── pipe/         # 进程间通信管道
```

---

## 3. CB-Kit 底层工具集详设 (The Toolkit)
为了贯彻 **Unix 哲学** 和 **单体交付** 策略，我们将所有逻辑封装在一个静态编译的 Go 二进制文件 `cb-tools` 中。运行时通过软链接分身为四个微工具。

### 3.1 cb-agent (总控 / The Brain)
**角色**：BootOS 的 PID 1 守护进程。

+ **核心职责**：
    1. **握手与保活**：启动时解析 Kernel Cmdline (`cb_server=...`)，建立 WebSocket/HTTP2 长连接。
    2. **配置融合 (Config Injection)**：
        * 接收 Server 下发的 **Task Spec**（标准配置）。
        * 接收 **User Overlay**（用户微调）。
        * **逻辑**：将两者 Deep Merge，生成最终的 `effective_config.json` 传递给执行器。
    3. **看门狗**：监控子进程资源使用，防止 OOM。

### 3.2 cb-probe (探针 / The Eyes)
**角色**：硬件指纹采集器。

+ **核心职责**：
    1. **深度扫描**：
        * 调用 `lspci -nnvv` 获取 PCI 树。
        * 调用 `dmidecode` 获取 SMBIOS。
        * **透传扫描**：尝试通过 `smartctl` 穿透 RAID 卡读取物理磁盘 SN。
    2. **数据清洗 (Normalization)**：
        * 将杂乱的厂商字符串（如 "Huawei Technologies Co., Ltd."）标准化为枚举值（"Huawei"）。
        * 输出标准 JSON 指纹数据，供 Server 端资产库比对。

### 3.3 cb-fetch (物流 / The Hands)
**角色**：安全下载与 DRM 执行器。

+ **核心职责**：
    1. **流式下载**：支持断点续传，不将大文件读入内存变量，而是直接通过 `io.Copy` 写入 Tmpfs 文件句柄。
    2. **动态解密 (DRM)**：
        * Provider 在 Store 中是加密存储的。
        * `cb-fetch` 握手获取 **Session Key**。
        * 边下载、边解密、边写入。**密钥只在内存中存在，文件只在 Tmpfs 中存在。**
    3. **完整性校验**：计算 SHA256 签名，防止中间人篡改。

### 3.4 cb-exec (执行 / The Fist)
**角色**：沙箱运行器。

+ **核心职责**：
    1. **环境隔离**：设置纯净的 `ENV`，切换工作目录。
    2. **IO 捕获与流化**：
        * 启动 Provider 子进程。
        * 实时读取 `Stdout` 和 `Stderr` 管道。
        * **日志结构化**：将原始文本封装为 JSON (`{ts, level, source, msg}`)。
        * 通过 HTTP SSE 通道实时推送到 Server，实现 Web 端“黑客帝国”式的滚动日志效果。
    3. **资源限制**：利用 `cgroups` 限制 Provider 的 CPU 和 内存配额。

---

## 4. 关键技术突破 (Key Technical Breakthroughs)
### 4.1 ZRAM 内存倍增技术
针对老旧服务器内存小（<8GB）及大文件传输场景的防线。

+ **机制**：BootOS 启动时加载 `zram` 模块，将 50% 物理内存划分为压缩块设备并挂载为 Swap。
+ **算法**：使用 `zstd` (Zstandard)，提供极高的压缩比和极低的 CPU 开销。
+ **效果**：物理 8GB 内存可承载逻辑 12GB+ 数据，有效防止解压大型固件包时 OOM。

### 4.2 双模引导加载器 (Hybrid Bootloader)
为了制作一个“通吃”的 ISO 镜像。

+ **Legacy BIOS**: 使用 `Isolinux/Pxelinux`。
+ **UEFI**: 使用 `Grub2-EFI`。
    - **安全启动**：集成 Microsoft 签名的 `Shim` 加载器，确保能通过银行服务器开启了 Secure Boot 的验证。

### 4.3 离线与在线的统一逻辑
+ `cb-fetch` 支持多协议：
    - `http://` -> 从 CloudBoot Core 下载。
    - `file://` -> 从本地挂载点（U盘 `CLOUDBOOT_DATA` 分区）读取。
+ 这使得 **CloudBoot PE (单机版)** 和 **CloudBoot Core (网络版)** 共用同一套代码逻辑，仅配置不同。

---

## 5. 接口定义 (Interface Specifications)
### 5.1 硬件指纹 Schema (cb-probe Output)
```json
{
  "schema_version": "v1.0",
  "basic": {
    "manufacturer": "Hygon",
    "product_name": "K100",
    "serial_number": "SN-12345678",
    "uuid": "..."
  },
  "cpu": { "arch": "x86_64", "model": "Hygon C86 7285", "cores": 32 },
  "memory": { "total_gb": 256, "dimms": [...] },
  "storage": {
    "controllers": [
      { "pci_id": "1000:005f", "vendor": "LSI", "model": "MegaRAID 3108" }
    ]
  },
  "network": [...]
}
```

### 5.2 任务执行 Schema (cb-exec Input)
```json
{
  "task_id": "job-8848",
  "action": "apply",
  "provider_path": "/opt/cloudboot/runtime/bin/provider-raid",
  "config": {
    // 标准配置 + 用户 Overlay 合并后的结果
    "raid_level": "10",
    "drives": "all",
    "quirks": { "init_timeout": 600 }
  }
}
```

---

## 6. 构建流水线 (Build Pipeline)
BootOS v4 的构建实现了 **Infrastructure as Code**。

1. **Compile Tools**: 静态编译 `cb-tools` (CGO_ENABLED=0)。
2. **Build Rootfs**: 使用 Docker 运行 `dnf --installroot` 构建基础环境。
3. **Generate Initrd**: 在容器内运行 `dracut`，注入 `cb-tools` 和 `systemd unit`。
4. **Pack ISO**: 使用 `xorriso` 生成支持双模引导的 ISO。
5. **Test**: 自动在 QEMU 中启动 ISO，验证 Agent 是否能向 Mock Server 发送心跳。

---

## 7. 结语
BootOS v4 是 CloudBoot NG 体系中最坚硬的**“钻头”**。

它通过 **OpenEuler** 解决了信创兼容的广度，通过 **微内核架构** 解决了部署的灵活性，通过 **CB-Kit 工具链** 解决了商业模式的闭环。它是我们“以代码定义基础设施”愿景在物理世界的真实投射。

---

_(文档结束)_


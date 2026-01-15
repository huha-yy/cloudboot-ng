---
status: Approved
author: 架构师 (System Architect - Claude)
reviewers: [技术负责人]
created: 2026-01-15
updated: 2026-01-15
version: 1.0
depends_on: [../requirements/PRD.md]
---

# CloudBoot NG 架构设计文档

## 1. 架构概述

CloudBoot NG 采用**"单体交付、无状态运行、软硬解耦"**的设计哲学。系统分为两个核心部分：

- **控制面（Core）**: CloudBoot Server，部署在管理节点，负责资产管理、任务编排、日志聚合
- **数据面（Data Plane）**: BootOS + Agent，运行在待部署物理机，负责硬件探测和指令执行

通过 **CSPM（CloudBoot Server Provider Mechanism）协议**实现软硬解耦——Core 只知道"做什么"，Provider（硬件驱动）只负责"怎么做"。

## 2. 技术栈（GOTH Stack）

| 层级 | 组件 | 选型 | 原因 |
|------|------|------|------|
| **语言** | - | Go 1.22+ | 静态编译、并发优势、单二进制 |
| **Web框架** | 核心 | Echo v4 | 轻量、快速、中间件完善 |
| **数据库** | 存储 | SQLite3 + Gorm ORM | 零外部依赖、嵌入式、WAL并发 |
| **模板** | SSR | html/template | 标准库、安全性好 |
| **样式** | UI | Tailwind CSS | 实用优先、CLI编译嵌入 |
| **交互（宏）** | AJAX | HTMX | 服务端驱动、替代传统AJAX |
| **交互（微）** | 响应式 | Alpine.js | 轻量级客户端响应，无构建步骤 |
| **构建工具** | 开发 | Air + Makefile | 热重载、任务编排 |
| **构建工具** | 样式 | Tailwind CLI | 生成CSS后嵌入二进制 |

## 3. 系统架构（C4模型）

### 3.1 容器图（Container Diagram）

```
┌──────────────────────────────────────────────────────────────────────────┐
│                    CloudBoot Core (Server Tier)                          │
│  [单体 Go 二进制 + 嵌入式 SQLite + 前端资源]                               │
│                                                                          │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────────────┐    │
│  │  Web UI     │◀─┤ HTMX/SSE API │  │   Private Store             │    │
│  │  (嵌入)     │  │  Handler     │  │ (DRM/水印/License验证)     │    │
│  └─────────────┘  └──────┬───────┘  └──────────┬──────────────────┘    │
│                          │                     │                       │
│                   ┌──────▼──────┐              │                       │
│                   │ Orchestrator│              │ .cbp 文件             │
│                   │ (状态机)    │              │ (AES-256加密)        │
│                   └──────┬──────┘              │                       │
│                          │                     │                       │
│                   ┌──────▼──────┐              │                       │
│                   │ LogBroker   │              │                       │
│                   │ (日志聚合)  │              │                       │
│                   └──────┬──────┘              │                       │
└─────────────────────────┼──────────────────────┼───────────────────────┘
                          │ HTTP/WebSocket     │ HTTP Stream
                          │ (任务指令/日志)     │ (Provider+Session Key)
                          │                     │
┌─────────────────────────▼─────────────────────▼───────────────────────┐
│                  BootOS (Client Tier)                                  │
│  [Tmpfs/RamDisk - 无持久化存储]                                        │
│                                                                        │
│  ┌─────────────┐   ┌──────────────┐   ┌───────────────────────────┐  │
│  │  cb-agent   │──▶│  cb-fetch    │──▶│      cb-exec              │  │
│  │  (脑部)     │   │  (解密器)    │   │   (沙箱执行器)           │  │
│  └─────────────┘   └──────────────┘   └───────────┬───────────────┘  │
│                                                    │ Stdin/Stdout     │
│                                            ┌───────▼────────┐        │
│                                            │   Provider     │        │
│                                            │  (驱动程序)   │        │
│                                            │  (内存解密)   │        │
│                                            └────────────────┘        │
└────────────────────────────────────────────────────────────────────┘
```

### 3.2 组件描述

**CloudBoot Core（单体二进制）：**
- 嵌入所有静态资源（HTML/CSS/JS）via `//go:embed`
- 内置SQLite3数据库（无外部MySQL/PG依赖）
- 提供Web UI + HTTP API
- 运行Orchestrator状态机（任务编排）
- 运行LogBroker（SSE日志流）
- 管理Private Store（Provider库）

**BootOS（内存操作系统）：**
- 基于OpenEuler 22.03，通过PXE/ISO启动
- 全量运行于Tmpfs（内存），无硬盘持久化
- 启动时加载cb-agent（初始化和轮询）

**Agent（cb-agent）：**
- HTTP客户端，向Core注册和轮询任务
- 下载Provider二进制和Session Key
- 调用cb-fetch进行内存解密
- 监控任务状态，上报日志

**Fetch（cb-fetch）：**
- 内存解密工具
- 使用Session Key解密Provider至`/dev/shm/provider`
- Provider执行后立即删除（内存焚毁）

**Exec（cb-exec）：**
- 沙箱执行引擎
- 调起Provider进程
- 捕获Stdout/Stderr，转发日志至Agent
- 隔离Provider的文件系统访问

## 4. 核心子系统设计

### 4.1 CSPM 协议（CloudBoot Server Provider Mechanism）

**设计理念**：解耦硬件特性。Core不关心RAID卡型号，只需发送"配置RAID"指令；Provider知道如何与具体硬件通信。

**Protocol Stack**：
```
Host (Agent)
    ↓ HTTP POST /api/boot/v1/task
Core (Orchestrator) ← Server端任务编排
    ↓ JSON over Stdin
Provider (驱动)  ← 在BootOS内沙箱执行
    ↓ JSON over Stdout
Core (LogBroker) ← 捕获Provider输出
    ↓ SSE
Browser (Terminal View) ← 实时显示
```

**Provider CLI 规范**：
```bash
provider probe                    # 探测硬件
provider plan [JSON config]      # 干运行（计算变更）
provider apply [JSON config]     # 真实执行（应用变更）
```

**DRM/加密流程**：
1. **Store（官方云端）**: 使用全局Master Key加密Provider
2. **Core（客户部署）**:
   - 导入.cbp包，验证watermark签名
   - 使用License解密Master Key（信封加密）
   - 生成临时Session Key，重新加密Provider
3. **BootOS（内存）**:
   - cb-fetch使用Session Key解密至内存
   - Provider执行后立即销毁，无磁盘留痕

### 4.2 任务编排状态机

```
┌─────────┐
│Pending  │  任务创建，等待Agent上线
└────┬────┘
     │
     ▼
┌──────────┐
│Handshake │  Agent确认，交换Session Key
└────┬─────┘
     │
     ▼
┌────────┐
│Probing │  运行 provider probe，检测硬件
└────┬───┘
     │
     ├─→ 硬件不支持 ──→ ┌──────┐
     │                 │Error │
     │                 └──────┘
     │
     ▼
┌──────────────┐
│Provisioning  │  运行 provider plan + apply
└────┬─────────┘
     │
     ├─→ 应用失败 ──→ ┌──────┐
     │               │Error │
     │               └──────┘
     │
     ▼
┌─────────┐
│Reporting│  Agent上报最终状态
└────┬────┘
     │
     ▼
┌────────┐
│Success │  任务完成
└────────┘
```

**LogBroker**：每个状态转移时，Agent上报Stderr日志（JSON格式）：
```json
{"ts": "2026-01-15T14:23:45Z", "level": "INFO", "component": "provider_raid", "msg": "RAID config success"}
```

Core聚合日志，通过SSE实时推送至Browser的Matrix Terminal组件。

### 4.3 资产库与硬件指纹

**影子资产**：
- Agent首次通过PXE启动，向Core发送`POST /api/boot/v1/register`
- Core创建Machine记录，status="discovered"
- UI中即时展示该机器（灰色、可见但不可操作）

**硬件指纹**：
- cb-probe采集CPU型号/核数、内存条数、RAID卡信息等
- 生成标准化JSON（schema v1.0）
- 每次任务前，比对当前指纹与历史指纹，检测硬件变更
- 防止硬件篡改（偷梁换柱）

**数据模型**：
```go
type Machine struct {
    ID            string         // UUID
    Hostname      string
    MacAddress    string         // 主标识符
    Status        string         // discovered|ready|installing|active|error
    HardwareSpec  HardwareInfo   // JSONB: CPU/内存/存储控制器
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type HardwareInfo struct {
    CPU struct {
        Model  string
        Cores  int
        Sockets int
    }
    Memory struct {
        TotalBytes int64
        DIMMs      []DimmInfo
    }
    StorageControllers []ControllerInfo
    NetworkInterfaces  []NICInfo
}
```

## 5. 通信协议

### 5.1 Agent Boot API（Agent ↔ Core）

**Endpoint**: `/api/boot/v1/...`

```yaml
POST /api/boot/v1/register
  Description: Agent上线注册/心跳
  Body:
    mac: "aa:bb:cc:dd:ee:ff"
    ip: "192.168.1.100"
    fingerprint: {...}  # Hardware fingerprint
  Response:
    status: "ok"
    task_id: "task_xyz"  # 如果有待执行任务

GET /api/boot/v1/task?mac=aa:bb:cc:dd:ee:ff
  Description: Agent轮询待执行任务
  Response:
    task_id: "task_xyz"
    action: "probe" | "apply"
    provider_url: "http://core/private-store/raid-lsi.cbp"
    session_key: "<encrypted session key>"
    config: {...}  # CSPM JSON config
```

### 5.2 外部API（Terraform/UI）

**Endpoint**: `/api/v1/...`（REST + JSON）

```yaml
GET /api/v1/machines
  Description: 查询所有物理机
  Response: [{id, hostname, mac, status, hardware}, ...]

POST /api/v1/machines
  Description: 纳管新机器（或通过PXE自动发现）

POST /api/v1/machines/{id}/provision
  Description: 触发安装任务
  Body:
    profile_id: "os_centos7_raid10"
    config: {...}  # Optional overlay
  Response:
    job_id: "job_abc"
    status: "pending"

GET /api/v1/profiles
  Description: 查询OS安装模板
```

### 5.3 SSE 流（实时日志）

**Endpoint**: `/api/stream/logs/{job_id}` （Server-Sent Events）

```
data: <div class="text-emerald-500">[RAID] Initializing...</div>
data: <div class="text-emerald-500">[RAID] Config success</div>
...
```

浏览器接收SSE，动态append到Matrix Terminal组件，无需页面刷新。

## 6. 部署架构

### 6.1 CloudBoot Core 部署

**最小化部署**：
```
单机运行:
  ./cloudboot-core --listen=0.0.0.0:8080 --db=data.db
```

**依赖项**：
- Linux内核 > 4.4
- libsqlite3（若启用CGO）

**网络要求**：
- DHCP (port 67/UDP)
- TFTP (port 69/UDP)  ← BootOS PXE启动
- HTTP (port 80 或 8080) ← Web UI + API

### 6.2 BootOS 发布形式

**ISO/USB镜像**：
- Kernel + Initrd（包含cb-agent、cb-probe、cb-exec）
- 大小目标: < 500MB
- 启动时间: < 15秒

**PXE启动**：
- TFTP服务从Core提供
- Agent自动向Core注册

## 7. 安全设计

### 7.1 Provider 沙箱隔离

- Provider 进程在独立命名空间运行
- 文件系统限制：只允许写入 `/opt/cloudboot/runtime`
- 内存解密后立即执行，不落磁盘

### 7.2 DRM 机制

- Provider 密钥分离（Master Key + Session Key）
- Watermark 校验（追踪下载来源）
- License 签名验证（ECDSA）

### 7.3 Web 安全

- HTMX 接口：仅返回HTML fragments（不含script tags）
- 上传接口：Content-Type检查、文件类型白名单
- CSRF防护：Token验证
- 日志不包含敏感信息（密钥、密码）

## 8. 可扩展性考量

### 8.1 并发支持

- SQLite WAL 模式：支持并发读，单一写锁
- 目标：支持 500+ 并发节点安装

### 8.2 Provider 生态

- Private Store 存储多个Provider二进制
- User Overlay 机制允许客户自定义参数
- Provider versioning（manifest.json）

### 8.3 多租户（未来）

- License 多客户隔离
- 数据库tenant_id分区

## 9. 决策记录

| ID | 决策 | 原因 | 备选方案 |
|----|----|------|--------|
| AD-001 | Go单体二进制 | 满足"零依赖"原则，便于离线部署 | Java/Python (需运行时) |
| AD-002 | Echo而非Gin | 更轻量、官方维护好 | Gin (性能差异可忽略) |
| AD-003 | SQLite嵌入式 | 无外部DB，符合单体原则 | PostgreSQL (需网络) |
| AD-004 | HTMX+Alpine | 减少JS依赖，SSR安全 | React+TypeScript (体积大) |
| AD-005 | CSPM JSON协议 | 通用、易于扩展 | Protocol Buffer (复杂) |
| AD-006 | Provider内存解密 | 防止密钥落盘 | 永久存储 (安全风险) |

---

**文档交接**：架构师 → 技术负责人
- 产出文件：ARCHITECTURE.md, API-SPEC.yaml
- 文档状态：Approved
- 核心内容：系统架构、CSPM协议、任务编排、部署指南

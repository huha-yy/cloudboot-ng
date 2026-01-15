---
status: Approved
author: 技术负责人 (Tech Lead - Claude)
reviewers: [前端开发, 后端开发]
created: 2026-01-15
updated: 2026-01-15
version: 1.0
depends_on: [../requirements/PRD.md, ../design/ARCHITECTURE.md, ../api/API-SPEC.yaml]
---

# CloudBoot NG 任务分解文档

## 1. 开发策略

基于"作战地图"，将开发周期拆分为 **6个特种行动（Phases）**。

**核心策略**：
- **交互开发**（视觉/体验）与 **挂机开发**（逻辑/底层）交替进行
- **先快后准**：先搭骨架快速出原型，再补齐单元测试和边缘情况
- **并行作业**：前后端在API规范确定后可独立并行开发

## 2. 分阶段任务明细

### Phase 1: 创世纪 (Genesis) - 骨架与视觉

**目标**: 建立项目基础设施，完成UI设计系统和基础组件

| 任务ID | 任务名称 | 负责人 | 估时 | 依赖 | 交付物 |
|--------|---------|--------|------|------|--------|
| G-01 | 项目基建 | 全栈 | 0.5d | - | `go.mod`, `Makefile`, Tailwind CLI配置, 目录结构 |
| G-02 | UI组件库 | 全栈 | 0.5d | G-01 | Card, Button, Badge, Terminal 组件 (HTML+Tailwind) |
| G-03 | 布局实现 | 全栈 | 0.5d | G-02 | Sidebar, Topbar, 响应式框架 |
| G-04 | Design System页 | 全栈 | 0.2d | G-03 | `/design-system` 路由及展示页面 |

**详细说明**：

#### G-01: 项目基建
```bash
# 目标：建立标准Go项目结构
cloudboot-ng/
├── cmd/
│   ├── server/main.go          # CloudBoot Core入口
│   ├── agent/main.go            # BootOS Agent入口
│   └── provider-mock/main.go    # Mock Provider入口
├── internal/
│   ├── core/
│   ├── models/
│   ├── api/
│   └── pkg/
├── web/
│   ├── static/
│   │   ├── css/                 # Tailwind输出
│   │   └── js/                  # htmx.min.js, alpine.min.js
│   └── templates/
│       ├── components/
│       ├── layouts/
│       └── views/
├── scripts/
├── go.mod
├── Makefile
└── README.md
```

**Makefile 目标**：
```makefile
.PHONY: dev build test lint

dev:
	@echo "启动开发环境..."
	tailwindcss -i web/static/css/input.css -o web/static/css/output.css --watch &
	air

build:
	@echo "构建生产二进制..."
	tailwindcss -i web/static/css/input.css -o web/static/css/output.css --minify
	CGO_ENABLED=1 go build -ldflags="-s -w" -o cloudboot-core cmd/server/main.go

test:
	go test -v ./...

lint:
	golangci-lint run
```

**go.mod 依赖（初步）**：
```
github.com/labstack/echo/v4
gorm.io/gorm
gorm.io/driver/sqlite
github.com/mattn/go-sqlite3
```

#### G-02: UI组件库

基于 `spec/UI_Design_System.md`，实现以下组件：

1. **Glass Card** (`web/templates/components/card.html`)
2. **Primary Button** (`web/templates/components/button.html`)
3. **Status Badge** (`web/templates/components/badge.html`)
4. **Matrix Terminal** (`web/templates/components/terminal.html`)

#### G-03: 布局实现

1. **Base Layout** (`web/templates/layouts/base.html`)
   - `<head>`: 加载Tailwind CSS, HTMX, Alpine.js
   - Sidebar + Topbar + Main Content 区域
2. **Sidebar Component** (`web/templates/components/sidebar.html`)
   - 可折叠/展开
   - 导航项：Dashboard, Machines, Jobs, Profiles, Store
3. **Topbar Component** (`web/templates/components/topbar.html`)
   - Glassmorphism效果
   - 用户信息、通知图标

#### G-04: Design System 页

创建一个展示所有组件的页面，方便前端开发时参考：
- 路由: `/design-system`
- 展示所有组件的不同状态和变体

---

### Phase 2: 核心脏器 (Core Organs) - 后端逻辑

**目标**: 实现数据模型、CSPM引擎、Mock Provider

| 任务ID | 任务名称 | 负责人 | 估时 | 依赖 | 交付物 |
|--------|---------|--------|------|------|--------|
| C-01 | 数据层 | 后端 | 0.5d | G-01 | Gorm Models, SQLite WAL初始化 |
| C-02 | CSPM引擎 | 后端 | 1.0d | C-01 | PluginManager, Executor |
| C-03 | Mock Provider | 后端 | 0.5d | C-02 | 模拟RAID配置的Go二进制 |
| C-04 | 单元测试 | 后端 | 0.5d | C-03 | Core逻辑与Mock Provider集成测试 |

**详细说明**：

#### C-01: 数据层

**文件**: `internal/models/machine.go`, `job.go`, `profile.go`, `license.go`

```go
// internal/models/machine.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Machine struct {
    ID            string         `gorm:"primaryKey"`
    Hostname      string         `gorm:"uniqueIndex"`
    MacAddress    string         `gorm:"uniqueIndex"`
    IPAddress     string
    Status        string         // discovered|ready|installing|active|error
    HardwareSpec  HardwareInfo   `gorm:"serializer:json"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type HardwareInfo struct {
    SchemaVersion string `json:"schema_version"`
    System        struct {
        Manufacturer string `json:"manufacturer"`
        ProductName  string `json:"product_name"`
        SerialNumber string `json:"serial_number"`
    } `json:"system"`
    CPU struct {
        Arch    string `json:"arch"`
        Model   string `json:"model"`
        Cores   int    `json:"cores"`
        Sockets int    `json:"sockets"`
    } `json:"cpu"`
    Memory struct {
        TotalBytes int64      `json:"total_bytes"`
        DIMMs      []DimmInfo `json:"dimms"`
    } `json:"memory"`
    StorageControllers []ControllerInfo  `json:"storage_controllers"`
    NetworkInterfaces  []NICInfo         `json:"network_interfaces"`
}

// ... 其他结构体定义
```

**SQLite WAL配置** (`internal/pkg/database/db.go`):
```go
db, err := gorm.Open(sqlite.Open("cloudboot.db?_journal_mode=WAL"), &gorm.Config{})
```

#### C-02: CSPM引擎

**文件**:
- `internal/core/cspm/plugin_manager.go`: Provider库管理、加密/解密
- `internal/core/cspm/executor.go`: 执行Provider二进制，捕获Stdin/Stdout

**核心接口**:
```go
package cspm

type ProviderExecutor interface {
    // 执行Provider命令
    Execute(ctx context.Context, cmd string, config map[string]interface{}) (*Result, error)
}

type Result struct {
    Status   string                 `json:"status"`
    Data     map[string]interface{} `json:"data"`
    Logs     []LogEntry             `json:"logs"`
    ExitCode int
}

type LogEntry struct {
    Timestamp time.Time `json:"ts"`
    Level     string    `json:"level"`
    Component string    `json:"component"`
    Message   string    `json:"msg"`
}
```

**Executor实现（伪代码）**:
```go
func (e *Executor) Execute(ctx context.Context, cmd string, config map[string]interface{}) (*Result, error) {
    // 1. 构建Provider执行命令
    cmdArgs := []string{e.providerPath, cmd}

    // 2. 准备Stdin（JSON config）
    stdinData, _ := json.Marshal(config)

    // 3. 启动进程，捕获Stdout/Stderr
    cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
    cmd.Stdin = bytes.NewReader(stdinData)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    // 4. 执行并等待
    err := cmd.Run()

    // 5. 解析结果
    var result Result
    json.Unmarshal(stdout.Bytes(), &result)

    // 6. 解析日志（Stderr）
    result.Logs = parseStderrLogs(stderr.Bytes())

    return &result, err
}
```

#### C-03: Mock Provider

**文件**: `cmd/provider-mock/main.go`

实现标准CSPM CLI接口：
```bash
provider-mock probe
provider-mock plan < config.json
provider-mock apply < config.json
```

**probe 输出示例**:
```json
{
  "status": "success",
  "supported_hardware": ["lsi_megaraid_3108", "generic_raid"]
}
```

**apply 输入/输出**:
```json
// Input (Stdin)
{
  "action": "apply",
  "resource": "raid",
  "desired_state": {"level": "10", "drives": ["sda", "sdb", "sdc", "sdd"]}
}

// Output (Stdout)
{
  "status": "success",
  "data": {"vd_id": "vd_1", "level": "10", "size_gb": 1800}
}

// Logs (Stderr)
{"ts": "2026-01-15T14:30:00Z", "level": "INFO", "component": "raid", "msg": "Initializing RAID controller"}
{"ts": "2026-01-15T14:30:02Z", "level": "INFO", "component": "raid", "msg": "Creating VD with RAID10"}
```

**Mock实现**: 使用`time.Sleep`模拟硬件操作延迟，返回预定义JSON

#### C-04: 单元测试

**文件**: `internal/core/cspm/executor_test.go`

```go
func TestExecutorWithMockProvider(t *testing.T) {
    tests := []struct {
        name       string
        cmd        string
        config     map[string]interface{}
        wantStatus string
        wantError  bool
    }{
        {
            name: "probe success",
            cmd:  "probe",
            config: nil,
            wantStatus: "success",
            wantError: false,
        },
        {
            name: "apply raid10",
            cmd: "apply",
            config: map[string]interface{}{
                "action": "apply",
                "resource": "raid",
                "desired_state": map[string]interface{}{
                    "level": "10",
                    "drives": []string{"sda", "sdb", "sdc", "sdd"},
                },
            },
            wantStatus: "success",
            wantError: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            executor := NewExecutor("../../../cmd/provider-mock/provider-mock")
            result, err := executor.Execute(context.Background(), tt.cmd, tt.config)

            if (err != nil) != tt.wantError {
                t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
                return
            }

            if result.Status != tt.wantStatus {
                t.Errorf("Execute() status = %v, want %v", result.Status, tt.wantStatus)
            }
        })
    }
}
```

---

### Phase 3: 杀手级体验 (Killer Experience) - 前端交互

**目标**: 实现SSE日志流、OS Designer

| 任务ID | 任务名称 | 负责人 | 估时 | 依赖 | 交付物 |
|--------|---------|--------|------|------|--------|
| K-01 | SSE管道 | 全栈 | 0.5d | C-02 | LogBroker, HTMX SSE集成 |
| K-02 | OS Designer | 全栈 | 1.0d | G-02 | Alpine.js分区拖拽与状态管理 |
| K-03 | 实时预览 | 全栈 | 0.5d | K-02 | 后端Template渲染接口 + 前端hx-post |
| K-04 | 联调Demo | 全栈 | 0.5d | K-01, C-03 | 端到端测试流程 |

**详细说明**：

#### K-01: SSE管道

**后端**: `internal/core/logbroker/broker.go`

```go
package logbroker

type Broker struct {
    clients map[string]chan LogMessage
    mu      sync.RWMutex
}

func (b *Broker) Subscribe(jobID string) <-chan LogMessage {
    ch := make(chan LogMessage, 100)
    b.mu.Lock()
    b.clients[jobID] = ch
    b.mu.Unlock()
    return ch
}

func (b *Broker) Publish(jobID string, msg LogMessage) {
    b.mu.RLock()
    if ch, ok := b.clients[jobID]; ok {
        ch <- msg
    }
    b.mu.RUnlock()
}
```

**前端HTMX**: `web/templates/views/job_detail.html`

```html
<div hx-ext="sse" sse-connect="/api/stream/logs/{{.JobID}}" sse-swap="message">
    <div id="log-container" class="terminal">
        <!-- 日志动态append到这里 -->
    </div>
</div>
```

#### K-02: OS Designer

**前端**: `web/templates/views/os_designer.html`

使用Alpine.js实现分区编辑器：
```html
<div x-data="partitionEditor()">
    <!-- 分区列表 -->
    <template x-for="(part, idx) in partitions" :key="idx">
        <div class="partition-card">
            <input x-model="part.mount_point" placeholder="挂载点">
            <input x-model="part.size" placeholder="大小">
            <select x-model="part.fstype">
                <option value="ext4">ext4</option>
                <option value="xfs">xfs</option>
            </select>
            <button @click="removePartition(idx)">删除</button>
        </div>
    </template>

    <button @click="addPartition()">添加分区</button>

    <!-- 实时预览 -->
    <button hx-post="/api/v1/profiles/preview" hx-vals="partitionsJSON()" hx-target="#preview">生成预览</button>
    <pre id="preview"></pre>
</div>

<script>
function partitionEditor() {
    return {
        partitions: [
            { mount_point: '/', size: '50GB', fstype: 'ext4' }
        ],
        addPartition() {
            this.partitions.push({ mount_point: '', size: '', fstype: 'ext4' });
        },
        removePartition(idx) {
            this.partitions.splice(idx, 1);
        },
        partitionsJSON() {
            return JSON.stringify({ partitions: this.partitions });
        }
    }
}
</script>
```

---

### Phase 4: 配置生成引擎 (Compiler)

**目标**: Kickstart/Autoyast模板生成

| 任务ID | 任务名称 | 负责人 | 估时 | 依赖 | 交付物 |
|--------|---------|--------|------|------|--------|
| CP-01 | 模板库 | 后端 | 0.5d | - | CentOS 7/8, Ubuntu, SUSE模板 |
| CP-02 | 渲染引擎 | 后端 | 0.5d | CP-01 | ConfigGen接口, Helper Functions |
| CP-03 | 校验器 | 后端 | 0.5d | CP-02 | 分区/网络逻辑校验 |
| CP-04 | Table-Driven测试 | 后端 | 0.5d | CP-03 | 20+用例覆盖边缘场景 |

**详细说明**：

#### CP-01: 模板库

**文件**: `internal/core/configgen/templates/centos7.ks.tmpl`

```bash
# Kickstart for CentOS 7
install
auth --useshadow --passalgo=sha512
bootloader --location=mbr --boot-drive={{.BootDrive}}
timezone {{.Timezone}}

# 分区
clearpart --all --initlabel
{{range .Partitions}}
part {{.MountPoint}} --fstype={{.FSType}} --size={{.Size}}
{{end}}

# 网络
{{if .Network.DHCP}}
network --bootproto=dhcp
{{else}}
network --bootproto=static --ip={{.Network.IP}} --netmask={{.Network.Netmask}}
{{end}}

# 软件包
%packages
{{range .Packages}}
{{.}}
{{end}}
%end
```

---

### Phase 5: 数据面 (Data Plane - BootOS)

**目标**: 实现BootOS Agent、探测工具、构建流程

| 任务ID | 任务名称 | 负责人 | 估时 | 依赖 | 交付物 |
|--------|---------|--------|------|------|--------|
| D-01 | cb-agent | 后端 | 1.0d | C-02 | HTTP客户端、任务轮询、下载器 |
| D-02 | cb-probe/exec | 后端 | 1.0d | D-01 | 硬件探测、沙箱执行 |
| D-03 | 构建工厂 | DevOps | 1.0d | D-02 | Dockerfile, dracut, ISO打包 |
| D-04 | hw-init TUI | 后端 | 0.5d | D-02 | Bubbletea工具封装 |

---

### Phase 6: 全链路仿真 (Simulation)

**目标**: QEMU仿真测试、集成验收

| 任务ID | 任务名称 | 负责人 | 估时 | 依赖 | 交付物 |
|--------|---------|--------|------|------|--------|
| S-01 | Seed Tool | 后端 | 0.2d | C-01 | 数据库预置Mock Provider数据 |
| S-02 | QEMU脚本 | DevOps | 0.5d | D-03 | simulate.sh, 网络配置 |
| S-03 | 集成验收 | Tech Lead | 0.5d | ALL | E2E场景测试 |

---

## 3. 人员分配与并行路径

### 并行开发路径

```
Phase 1 (G-01 → G-04)  [全栈] ─┐
                                ├─→ Phase 3 (K-01 → K-04)  [全栈]
Phase 2 (C-01 → C-04)  [后端] ─┘

Phase 4 (CP-01 → CP-04) [后端] ─→ 可与Phase 3并行

Phase 5 (D-01 → D-04)   [后端+DevOps] ─→ 依赖Phase 2完成

Phase 6 (S-01 → S-03)   [全员] ─→ 最终集成
```

### 人员角色定义

- **全栈开发**: 负责Phase 1, 3（UI+后端接口）
- **后端开发**: 负责Phase 2, 4, 5（核心逻辑）
- **DevOps**: 负责Phase 5-D03, Phase 6（构建与部署）
- **Tech Lead**: 负责代码审查、集成验收

---

## 4. 关键里程碑与验收标准

| 里程碑 | 完成时间 | 验收标准 |
|--------|---------|---------|
| M1: 项目骨架 | Phase 1完成 | `make dev` 运行成功，Design System页面可访问 |
| M2: 核心逻辑 | Phase 2完成 | 单元测试通过率 > 80%，Mock Provider可执行 |
| M3: UI交互 | Phase 3完成 | SSE日志实时显示，OS Designer可生成预览 |
| M4: E2E验收 | Phase 6完成 | QEMU环境下完整流程通过 |

---

## 5. 风险与缓解措施

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|---------|
| SQLite WAL并发问题 | 中 | 高 | 提前压测，必要时切换PostgreSQL |
| Provider加密性能 | 低 | 中 | 使用Go标准库crypto，避免外部依赖 |
| QEMU环境搭建复杂 | 高 | 低 | 提供Docker镜像简化环境 |
| 前后端联调阻塞 | 中 | 中 | API规范先行，Mock数据独立开发 |

---

## 📋 文档交接

### 交接方: 技术负责人
- 产出文档: `docs/dev/TASK-BREAKDOWN.md`
- 文档状态: Approved
- 核心内容摘要: 6个Phase任务分解、人员分配、并行路径、验收标准

### 接收方: 前端开发 + 后端开发
- 待产出文档: 代码实现 + `FRONTEND-IMPL.md` + `BACKEND-IMPL.md`
- 依赖内容: API-SPEC.yaml、任务清单、验收标准
- 开始时间: Phase 1 (前后端可并行启动基建工作)

### 注意事项
- 前后端必须严格遵循 `API-SPEC.yaml` 接口定义
- 所有UI组件必须使用 `spec/UI_Design_System.md` 规定的颜色和样式
- 每个Phase完成后提交代码审查，通过后进入下一Phase
- 单元测试覆盖率目标: > 80%

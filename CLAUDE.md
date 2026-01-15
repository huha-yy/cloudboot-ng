# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**CloudBoot NG** is a next-generation bare metal infrastructure provisioning platform - described as "Terraform for Bare Metal" and a "Digital Visa Officer for Infrastructure". It aims to automate hardware lifecycle management with standardized, code-defined workflows.

### Key Principles (Inviolable)

The project strictly adheres to these supreme principles:

1. **Single Binary Delivery**: Entire CloudBoot Core MUST compile into one static Go binary
   - NO external runtime dependencies (Node.js, Python, Java, Nginx, Systemd)
   - Use `//go:embed` for all static assets (HTML/CSS/SQL)

2. **Zero Dependencies**: Binary runs on any modern Linux (x86/ARM) with just `chmod +x`
   - Embedded SQLite (WAL mode) - no external database required
   - No cgit, no container runtime needed

3. **Stateless Execution**:
   - BootOS runs entirely in RAM (Tmpfs), reboots to clean slate
   - Agent doesn't persist data locally - all state pushed to Core

4. **Decoupled Architecture (CSPM Protocol)**:
   - Core handles Intent (What to do)
   - Hardware Providers handle Instruction (How to do it)
   - Communication: JSON over Stdin/Stdout with structured JSON logs on Stderr

## Architecture & Tech Stack (GOTH Stack)

| Component | Choice | Notes |
|-----------|--------|-------|
| Language | Go 1.22+ | Static binary, concurrency |
| Web Framework | Echo v4 | Fast, minimal, robust middleware |
| Database | SQLite3 (Gorm ORM) | Embedded, WAL mode for concurrency |
| Templating | html/template | Standard Go templates, SSR |
| Styling | Tailwind CSS | Utility-first, embedded during build |
| Macro-Interaction | HTMX | Server-driven UI (AJAX replacement) |
| Micro-Interaction | Alpine.js | Client-side reactivity without build steps |

### System Structure

```
cloudboot-ng/
├── cmd/
│   ├── server/           # CloudBoot Core main entry point
│   ├── agent/            # BootOS Agent
│   ├── provider-mock/    # Mock Hardware Provider (for testing)
│   └── tools/            # Helper utilities
├── internal/
│   ├── core/             # Core business logic (Service Layer)
│   │   ├── machine/      # Machine lifecycle management
│   │   ├── job/          # Task scheduler & orchestration
│   │   └── cspm/         # Plugin loading & execution engine
│   ├── models/           # Gorm data models
│   ├── api/              # HTTP Handlers (HTMX & JSON)
│   └── pkg/              # Shared utilities (Logger, Crypto)
├── web/
│   ├── static/           # Raw assets (JS libs, CSS, images)
│   └── templates/        # HTML templates (components, layouts, views)
├── scripts/              # Build scripts (Makefile, Dockerfile)
├── go.mod
└── README.md
```

## Communication Protocols

**HTTP / HTMX**:
- SSR endpoints return full pages via `c.Render(...)`
- HTMX endpoints return HTML fragments for dynamic updates
- API endpoints at `/api/v1/...` return JSON for external integrations (Terraform)

**Server-Sent Events (SSE)**:
- Real-time streaming of installation logs
- Endpoint: `/api/stream/logs?job_id=xyz`
- Format: `data: {"ts": "...", "level": "INFO", "msg": "..."}\n\n`

**CSPM (Provider Protocol)**:
- JSON over Stdin/Stdout between Core and Provider binaries
- All Providers MUST implement: `probe`, `plan`, `apply` subcommands
- Structured JSON logs on Stderr
- Supports `overlay` field for configuration injection (quirks/overrides)

## Build & Development Commands

**Expected Makefile targets** (when project reaches code phase):

```bash
make dev          # Run Tailwind watch + Air (live reload Go)
make build        # Build production binary (runs tailwind build --minify first)
make test         # Run unit tests with mock providers
make lint         # Lint Go code
```

**Key Requirements for Builds**:
- MUST run `tailwind build --minify` before building binary
- MUST use `-ldflags="-s -w"` to strip symbols for smaller binary
- SQLite requires CGO enabled
- Target: Binary < 60MB, BootOS startup < 15 seconds

## Data Models (Core Gorm Structs)

### Machine
- UUID-based unique identifier
- Status: "discovered" → "ready" → "installing" → "active" or "error"
- HardwareSpec: JSONB field storing standardized hardware fingerprint
- MAC address: Primary identifier from network discovery

### Job
- References Machine by ID
- Type: "audit", "config_raid", "install_os"
- Status: "pending" → "running" → "success" or "failed"
- StepCurrent: tracks progress within long-running jobs (e.g., "downloading_provider")
- LogsPath: file path to full execution logs

### OSProfile
- Distro: "centos7", "ubuntu22", "ky10", etc.
- Config: JSONB with Partitions array, NetworkConfig, packages list
- Used to generate Kickstart/Autoyast files

### License
- ProductKey: Encrypted Master Key for decrypting Providers (envelope encryption)
- Features array: what capabilities are enabled
- ECDSA Signature from official CloudBoot for validation

### Hardware Fingerprint (Standard JSON Schema v1.0)
Must include: system info, CPU specs, memory DIMMs, storage controllers (with driver info), network interfaces. Used for hardware change detection.

## CSPM / Provider Security (DRM Protocol)

**Artifact Format (.cbp - CloudBoot Package)**:
- Standard ZIP containing: `manifest.json`, `watermark.json` (download trace), `provider.enc` (AES-256 encrypted binary)

**Decryption Flow**:
1. Core imports .cbp, verifies watermark signature
2. Core uses Customer License Key to decrypt Master Key (envelope encryption)
3. Core generates Session Key, re-encrypts binary, sends to BootOS
4. BootOS `cb-fetch` decrypts to `/dev/shm/provider` (memory only)
5. Provider executes, deletes immediately after exit (no persistent decrypted copy)

## UI Design System

**Theme**: "Dark Industrial" - cockpit precision with hacker aesthetic

**Strict Color Palette** (Tailwind):
- Canvas: `bg-slate-950` - Global background
- Surface: `bg-slate-900` - Cards, sidebar
- Border: `border-slate-800`
- Primary (Success/Active): `emerald-500`
- Accent (AI/Magic): `violet-500`
- Destructive: `rose-500`
- Text Main: `text-slate-200`
- Text Muted: `text-slate-400`

**Typography**:
- UI Font: Sans-serif (Inter or System)
- Data Font: **MANDATORY** `font-mono` (JetBrains Mono) for: UUIDs, IPs, MACs, logs, version numbers, config code

**Key Components**:
- Glass Cards: `bg-slate-900/50 backdrop-blur-sm border border-slate-800`
- Primary Buttons: Emerald with subtle glow (`shadow-emerald-900/20`)
- Matrix Terminal: Black background, monospace, green text for success logs
- Status Badge: With pulsing dot for "Online" state

**Layout Structure**:
- Sidebar: 64px (collapsed) / 240px (expanded), darker bg
- Topbar: Sticky with glassmorphism
- Main Content: `max-w-7xl mx-auto p-6`

## Development Rules (AI Agent Guidelines)

When generating code:

1. **TDD (Test-Driven Development)**:
   - Write `_test.go` files before or simultaneously with implementation
   - Tests must cover logic, not just syntax
   - Use table-driven tests for edge cases

2. **Mock Everything**:
   - No physical servers in dev - mock all hardware interactions (RAID, IPMI, PXE)
   - Providers should be testable against mock harnesses

3. **No Hallucinations**:
   - Use only standard library or widely-used stable packages
   - Approved libraries: `gin`, `gorm`, `cobra`, `mattn/go-sqlite3`, `golang-jwt`, `bubbletea` (TUI)
   - No invented Go libraries

4. **Strict Go Idioms**:
   - Explicit error handling (wrap errors with context, don't just return)
   - Comments required for all exported functions
   - Follow effective Go conventions

## API Endpoints (Boot & External)

**Agent Boot API** (`/api/boot/v1/...`):
- `POST /register`: Agent heartbeat/registration with hardware fingerprint
- `GET /task`: Agent polls for task instructions (returns TaskSpec JSON)

**External API** (`/api/v1/...`):
- `GET/POST /servers`: Machine list and creation (Terraform-friendly)
- `POST /servers/{id}/provision`: Trigger installation job (accepts profile_id)
- Returns JSON, supports pagination and filtering

**Stream API**:
- `GET /api/stream/logs/{job_id}`: Server-Sent Events for real-time logs

## Testing Strategy

**Unit Tests**:
- Core logic with mock providers (run in CI via `go test ./...`)
- Table-driven tests for config generation edge cases

**Integration Tests**:
- CSPM protocol parsing and execution
- Provider encryption/decryption flows
- SQLite WAL concurrency

**E2E Tests** (via simulation):
- QEMU environment with mock hardware
- Test Agent registration → Task download → Provider execution → Log streaming
- Verify hardened sandbox prevents Provider from writing outside `/opt/cloudboot/runtime`

## 工作流程与协作指南 (Workflow & Collaboration Guide)

### 核心工作原则

1. **文档驱动开发 (Documentation-Driven Development)**
   - 始终依据 `@指引文档.md` 展开工作
   - 在开始任何编码前，确保完全理解相关的PRD、架构设计、API规范、任务分解和测试计划
   - 所有实现决策必须有文档依据

2. **团队协作模式**
   - 使用 `/elite-dev-team` skill 进行虚拟团队协作
   - 触发时机：
     - 完成需求分析后
     - 完成架构/技术设计后
     - 完成任务分工后
     - 完成核心功能开发里程碑后
     - 完成测试阶段后
     - 项目交付时
   - 该skill会在框架内自动调度各专业角色（产品经理、架构师、开发工程师、测试工程师、DevOps等）协作

3. **进度跟踪与汇报**
   - 所有进度更新必须同步至 `TODO.md` 文件
   - 格式要求：包含具体进展描述和精确到分钟的时间戳 (HH:MM)
   - 进度项示例：
     ```
     - [x] 实现Machine数据模型 | 2026-01-15 14:23 | 完成
     - [ ] 编写Machine单元测试 | 2026-01-15 14:25 | 进行中
     ```
   - 适时使用 `/dingtalk-progress` 汇报关键进展至钉钉群
   - 汇报触发时机：
     - 完成重要开发里程碑
     - 发现并解决关键阻塞问题
     - 完成单元测试通过率达成目标
     - 完成集成测试验证
     - 完成功能对标文档检查

4. **自主决策与全自动运行**
   - **全自动运行**：过程中无需询问用户问题，自主进行技术决策
   - 对文档有合理解释的范围内，做出最佳工程决策
   - 对不确定的需求，基于项目原则和上下文推导决策
   - 记录所有自主决策的理由至TODO.md的决策日志部分
   - **待人类确认机制**：如存在需要与用户沟的内容（如重大架构变更、潜在风险警告等），将内容记录至 `待人类确认.md` 文件，继续开发不阻塞

5. **网络资源访问**
   - 如果需要下载依赖或资源，可使用网络代理：
     ```bash
     export https_proxy=http://127.0.0.1:7897
     export http_proxy=http://127.0.0.1:7897
     export all_proxy=socks5://127.0.0.1:7897
     ```
   - 设置后可正常执行 `go get`、`npm install` 等命令

6. **完工验收与迭代 (Completion & Iteration)**
   - 完成所有工作后，形成详细的 `IMPLEMENTATION_REPORT.md` 文档，包括：
     - 实现的功能列表
     - 与原始文档（PRD、架构设计、API规范等）的对标检查
     - 发现的差异和偏差说明
     - 完成度统计 (%)
   - 仔细核对功能实现和原始文档间的**每一项**要求
   - 如发现差异，自动进入迭代周期：
     - 更新TODO.md标记缺陷项
     - 继续开发补齐缺口
     - 直至达到100%符合文档要求

### TODO.md 文件格式

创建 `TODO.md` 在项目根目录，包含以下结构：

```markdown
# CloudBoot NG 开发进度追踪

## 任务总览
- 总任务数: X
- 已完成: Y
- 进行中: Z
- 完成率: Y/X

## 任务列表

### Phase 1: 项目基建
- [x] Go项目初始化 | 2026-01-15 10:30 | 完成
- [ ] Makefile编写 | 2026-01-15 10:45 | 进行中

### Phase 2: 核心脏器
...

## 决策日志
- 2026-01-15 10:30 | 选择使用mattn/go-sqlite3而非其他SQLite驱动 | 原因：官方推荐、广泛使用、性能好

## 遇到的问题与解决方案
- [问题] 某某功能实现困难 | [解决] 采取方案X | 时间: 2026-01-15 11:00

## 待办事项（当前会话）
- [ ] 任务1
- [ ] 任务2
```

## Important Context

This repository is specification-phase - the codebase is being designed per document-driven development model. Reference documents in `/spec/`:
- `PROJECT_Manifest.md`: Core principles and business model
- `ARCH_Stack.md`: Technology stack justification and directory layout
- `DATA_Schema.md`: Gorm model definitions and hardware fingerprint schema
- `CSPM_Protocol.md`: Provider interaction protocol and DRM security flow
- `UI_Design_System.md`: Complete design system with component snippets

The `指引文档.md` (in Chinese) contains full PRD, architecture design, API spec, task breakdown, and test plan generated by the Elite Dev Team.

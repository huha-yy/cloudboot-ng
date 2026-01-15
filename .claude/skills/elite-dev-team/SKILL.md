---
name: elite-dev-team
description: 虚拟世界级软件开发团队，采用文档驱动的协作模式。团队由产品经理、架构师、技术负责人、前端开发、后端开发、测试工程师、DevOps组成。核心特点：(1) 文档驱动 - 每个角色有明确的文档产物，文档是角色间协作的唯一接口；(2) 模板标准化 - 提供PRD、架构设计、API规范、任务分解、测试计划、部署文档等标准模板；(3) 流程规范 - 需求→设计→开发→测试→部署的完整链路。触发场景：当用户提到"团队开发"、"企业级项目"、"完整软件开发"、"文档驱动"、"多角色协作"时使用此skill。
---

# Elite Dev Team - 文档驱动的精英开发团队

## 核心理念

**文档是角色协作的唯一接口**。每个角色通过产出标准化文档来驱动下游工作，确保信息传递的完整性和可追溯性。

## 文档驱动协作流程（含并行与回环）

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                            文档驱动开发流水线 (v2.0)                                  │
└─────────────────────────────────────────────────────────────────────────────────────┘

  产品经理         架构师          技术负责人        前端 ∥ 后端        测试          DevOps
     │               │                │                 │               │              │
     ▼               ▼                ▼                 ▼               ▼              ▼
 ┌───────┐      ┌───────┐        ┌───────┐        ┌─────────┐     ┌───────┐      ┌───────┐
 │ PRD   │─────▶│架构设计│───┬───▶│任务分解│───────▶│ 并行开发 │────▶│ 测试  │─────▶│ 部署  │
 │ 文档  │      │ 文档  │   │    │ 文档  │        │ ┌───┬───┐│     │ 执行  │      │ 发布  │
 └───────┘      └───────┘   │    └───────┘        │ │FE │BE ││     └───┬───┘      └───┬───┘
     │               │      │                     │ └───┴───┘│         │              │
     │               ▼      │                     └────┬─────┘         │              │
     │          ┌───────┐   │                          │               │              │
     │          │API规范│───┘                          ▼               │              │
     │          └───────┘                      ┌──────────────┐        │              │
     │               │                         │ 代码审查回环 │◀───────┤              │
     │               │                         └──────┬───────┘        │              │
     │               │                                │                │              │
     └───────────────┴────────────────────────────────┴────────────────┴──────────────┘
                                                                       │
                                    ┌──────────────────────────────────┘
                                    │
                              ┌─────▼─────┐     ┌──────────┐     ┌──────────┐
                              │Bug修复回环│────▶│部署回滚回环│───▶│产品验收回环│
                              └───────────┘     └──────────┘     └──────────┘
```

### 可并行任务

| 并行组 | 任务A | 任务B | 同步点 |
|--------|-------|-------|--------|
| 设计阶段 | 架构设计 | 测试计划(初稿) | API规范完成 |
| 开发阶段 | **前端开发** | **后端开发** | 联调测试 |
| 测试阶段 | 单元测试(开发) | 集成测试(QA) | 测试报告 |

### 反馈回环

| 回环 | 触发条件 | 文档 | 责任角色 |
|------|----------|------|----------|
| 代码审查 | PR提交 | PR-REVIEW.md | 技术负责人→开发 |
| Bug修复 | 测试发现Bug | BUG-TRACKING.md | 测试→开发→测试 |
| 部署回滚 | 部署失败 | ROLLBACK.md | DevOps→开发 |
| 产品验收 | 发布后 | UAT-RESULT.md | 产品→全团队 |

## 角色与文档产物

| 角色 | 输入文档 | 输出文档 | 模板位置 |
|------|----------|----------|----------|
| 产品经理 | 业务需求 | PRD.md, USER-STORIES.md | `assets/templates/PRD.md` |
| 架构师 | PRD.md | ARCHITECTURE.md, API-SPEC.yaml | `assets/templates/ARCHITECTURE.md` |
| 技术负责人 | PRD.md, ARCHITECTURE.md | TASK-BREAKDOWN.md | `assets/templates/TASK-BREAKDOWN.md` |
| 前端开发 | API-SPEC.yaml, TASK-BREAKDOWN.md | 代码, FRONTEND-IMPL.md | - |
| 后端开发 | API-SPEC.yaml, TASK-BREAKDOWN.md | 代码, BACKEND-IMPL.md | - |
| 测试工程师 | PRD.md, 实现文档 | TEST-PLAN.md, TEST-REPORT.md | `assets/templates/TEST-PLAN.md` |
| DevOps | ARCHITECTURE.md, TEST-REPORT.md | DEPLOYMENT.md | `assets/templates/DEPLOYMENT.md` |

## 项目文档目录结构

```
your-project/
├── docs/
│   ├── requirements/          # 需求文档
│   │   ├── PRD.md            # 产品需求文档
│   │   ├── USER-STORIES.md   # 用户故事
│   │   └── ACCEPTANCE.md     # 验收标准
│   ├── design/               # 设计文档
│   │   ├── ARCHITECTURE.md   # 架构设计
│   │   └── DATABASE.md       # 数据库设计
│   ├── api/                  # API文档
│   │   └── API-SPEC.yaml     # OpenAPI规范
│   ├── dev/                  # 开发文档
│   │   ├── TASK-BREAKDOWN.md # 任务分解
│   │   └── CODING-STANDARDS.md
│   ├── impl/                 # 实现文档
│   │   ├── FRONTEND-IMPL.md
│   │   └── BACKEND-IMPL.md
│   ├── test/                 # 测试文档
│   │   ├── TEST-PLAN.md
│   │   ├── TEST-CASES.md
│   │   └── TEST-REPORT.md
│   ├── ops/                  # 运维文档
│   │   ├── DEPLOYMENT.md
│   │   └── RUNBOOK.md
│   └── adr/                  # 架构决策记录
│       └── ADR-001.md
└── src/                      # 源代码
```

## 角色工作指南

### 产品经理
阅读 [product-manager.md](references/product-manager.md) 了解详情。

**核心产出：**
1. 使用 `assets/templates/PRD.md` 模板创建 `docs/requirements/PRD.md`
2. 完成后设置文档状态为 `In Review`
3. 通知架构师进行评审

### 架构师  
阅读 [architect.md](references/architect.md) 了解详情。

**核心产出：**
1. 基于PRD.md，使用 `assets/templates/ARCHITECTURE.md` 创建架构文档
2. 使用 `assets/templates/API-SPEC.yaml` 创建API规范
3. 完成后通知技术负责人

### 技术负责人
阅读 [tech-lead.md](references/tech-lead.md) 了解详情。

**核心产出：**
1. 基于架构文档，使用 `assets/templates/TASK-BREAKDOWN.md` 创建任务分解
2. 分配任务给前端和后端开发
3. 组织代码审查

### 前端开发
阅读 [frontend-dev.md](references/frontend-dev.md) 了解详情。

**核心产出：**
1. 根据API-SPEC.yaml和TASK-BREAKDOWN.md实现前端代码
2. 完成后更新FRONTEND-IMPL.md记录实现细节
3. 通知测试工程师

### 后端开发
阅读 [backend-dev.md](references/backend-dev.md) 了解详情。

**核心产出：**
1. 根据API-SPEC.yaml和TASK-BREAKDOWN.md实现后端代码
2. 完成后更新BACKEND-IMPL.md记录实现细节
3. 通知测试工程师

### 测试工程师
阅读 [qa-engineer.md](references/qa-engineer.md) 了解详情。

**核心产出：**
1. 基于PRD.md，使用 `assets/templates/TEST-PLAN.md` 创建测试计划
2. 执行测试后产出TEST-REPORT.md
3. 完成后通知DevOps准备部署

### DevOps
阅读 [devops.md](references/devops.md) 了解详情。

**核心产出：**
1. 基于架构文档和测试报告，使用 `assets/templates/DEPLOYMENT.md` 创建部署文档
2. 执行部署并更新RUNBOOK.md

## 文档交接协议

角色切换时必须执行文档交接：

```markdown
## 📋 文档交接

### 交接方: [角色名]
- 产出文档: [文档路径]
- 文档状态: Approved
- 核心内容摘要: [简述]

### 接收方: [角色名]  
- 待产出文档: [文档名]
- 依赖内容: [从交接文档中需要的信息]

### 注意事项
- [特别说明]
```

## 快速开始

### 启动新项目

```
用户: 帮我开发一个[项目描述]

团队响应（含并行与回环）:

阶段1: 需求定义
├─ [产品经理] 创建 docs/requirements/PRD.md
└─ 文档交接 → 架构师

阶段2: 设计（部分并行）
├─ [架构师] 创建 ARCHITECTURE.md + API-SPEC.yaml
├─ [测试工程师] 开始准备 TEST-PLAN.md (并行)
└─ 文档交接 → 技术负责人

阶段3: 任务分解
├─ [技术负责人] 创建 TASK-BREAKDOWN.md
└─ 文档交接 → 前端 + 后端 (并行启动)

阶段4: 开发（并行 + 代码审查回环）
├─ [前端开发] 实现UI ─┬─→ PR提交 → 代码审查 → 合并/返工
├─ [后端开发] 实现API ─┘
├─ 🔄 代码审查回环: 技术负责人审查 → 通过/返工
└─ 联调完成 → 文档交接 → 测试

阶段5: 测试（Bug修复回环）
├─ [测试工程师] 执行测试
├─ 🔄 Bug修复回环: 发现Bug → 开发修复 → 测试验证
├─ 所有Bug关闭 → TEST-REPORT.md
└─ 文档交接 → DevOps

阶段6: 部署（部署回滚回环）
├─ [DevOps] 执行部署
├─ 🔄 部署回滚回环: 失败 → 回滚 → 修复 → 重新部署
└─ 部署成功 → 通知全团队

阶段7: 验收（产品验收回环）
├─ [产品经理] UAT验收
├─ 🔄 验收回环: 问题 → 新需求/优化 → 回到阶段1/4
└─ 验收通过 → 项目完成
```

## 参考资源

- **文档流转规范**: [document-flow.md](references/document-flow.md)
- **工作流程**: [workflows.md](references/workflows.md)
- **项目模板**: [project-templates.md](references/project-templates.md)

# 系统架构师 (System Architect)

## 角色职责

系统架构师负责将产品需求转化为可落地的技术架构，确保系统的可扩展性、可维护性和高可用性。

## 📋 文档产物（必须输出）

| 文档 | 输出路径 | 模板 | 下游消费者 |
|------|----------|------|------------|
| **ARCHITECTURE.md** | `docs/design/ARCHITECTURE.md` | `assets/templates/ARCHITECTURE.md` | 技术负责人、DevOps |
| **API-SPEC.yaml** | `docs/api/API-SPEC.yaml` | `assets/templates/API-SPEC.yaml` | 前端、后端开发 |
| DATABASE.md | `docs/design/DATABASE.md` | - | 后端开发 |
| ADR-xxx.md | `docs/adr/` | - | 技术负责人 |

### 输入依赖
- `docs/requirements/PRD.md` (来自产品经理)

### 文档产出流程
```
1. 阅读PRD.md，理解功能和非功能需求
2. 复制模板创建ARCHITECTURE.md
3. 复制模板创建API-SPEC.yaml
4. 设置状态为 "In Review"
5. 执行文档交接给技术负责人
```

## 核心能力

### 1. 架构设计

#### 分层架构设计
```
┌─────────────────────────────────────────────┐
│              Presentation Layer              │
│         (Web/Mobile/API Gateway)             │
├─────────────────────────────────────────────┤
│              Application Layer               │
│           (Business Logic/Services)          │
├─────────────────────────────────────────────┤
│                Domain Layer                  │
│          (Domain Models/Rules)               │
├─────────────────────────────────────────────┤
│             Infrastructure Layer             │
│      (Database/Cache/Message Queue)          │
└─────────────────────────────────────────────┘
```

#### 微服务架构原则
- 单一职责：每个服务只做一件事
- 自治性：独立部署、独立扩展
- 去中心化：数据分散管理
- 弹性设计：容错与降级机制

### 2. 技术选型

| 层级 | 考量因素 | 常见方案 |
|------|----------|----------|
| 前端 | 复杂度、团队技能 | React/Vue/Angular |
| 后端 | 性能、生态系统 | Go/Java/Node.js/Python |
| 数据库 | 数据模型、规模 | PostgreSQL/MongoDB/Redis |
| 消息队列 | 吞吐量、可靠性 | Kafka/RabbitMQ/Redis |
| 容器化 | 编排需求 | Docker/Kubernetes |

### 3. 架构文档

```markdown
## 架构设计文档模板

### 1. 架构概览
- 系统上下文图
- 容器图
- 组件图

### 2. 关键架构决策(ADR)
| 决策ID | 问题 | 决策 | 理由 | 影响 |
|--------|------|------|------|------|
| ADR001 |      |      |      |      |

### 3. 服务设计
- 服务边界
- API契约
- 数据模型

### 4. 非功能性架构
- 性能设计
- 安全设计
- 可观测性设计

### 5. 部署架构
- 环境规划
- 网络拓扑
- 资源规格
```

### 4. API设计原则

#### RESTful设计规范
- 资源命名：复数名词，如 `/users`, `/orders`
- HTTP方法：GET查询/POST创建/PUT更新/DELETE删除
- 版本控制：`/api/v1/`
- 分页设计：`?page=1&size=20`
- 错误响应：统一错误码体系

#### 接口契约示例
```yaml
openapi: 3.0.0
paths:
  /users:
    get:
      summary: 获取用户列表
      parameters:
        - name: page
          in: query
          schema:
            type: integer
      responses:
        '200':
          description: 成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserList'
```

## 协作接口

### 接收自产品经理
- PRD文档
- 非功能性需求

### 输出给开发团队
- 架构设计文档
- 技术选型决策
- API接口规范
- 数据库设计

### 输出给DevOps
- 部署架构图
- 资源需求清单
- 监控指标定义

## 架构评审检查清单

- [ ] 是否满足业务需求？
- [ ] 是否具备可扩展性？
- [ ] 是否考虑了安全性？
- [ ] 是否有容灾方案？
- [ ] 是否便于运维监控？
- [ ] 技术债务是否可控？

## 文档交接模板

完成架构设计后，使用以下格式交接：

```markdown
## 📋 文档交接 - 架构师 → 技术负责人

### 产出文档
- docs/design/ARCHITECTURE.md (状态: Approved)
- docs/api/API-SPEC.yaml (状态: Approved)
- docs/design/DATABASE.md (状态: Approved)

### 核心内容摘要
- 技术栈: [前端/后端/数据库]
- 架构模式: [单体/微服务]
- API数量: [X个端点]
- 核心实体: [列表]

### 接收方待办
- [ ] 技术负责人评审架构
- [ ] 创建TASK-BREAKDOWN.md
- [ ] 分配开发任务

### 特别注意
- [架构风险或技术难点]
```

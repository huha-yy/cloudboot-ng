---
status: draft
author: DevOps
reviewers: []
created: YYYY-MM-DD
updated: YYYY-MM-DD
version: 0.1
depends_on: [ARCHITECTURE.md, TEST-REPORT.md]
---

# 部署文档

## 1. 文档信息

| 项目 | 内容 |
|------|------|
| 项目名称 | |
| 部署版本 | |
| 关联架构文档 | |
| 关联测试报告 | |

## 2. 部署概述

### 2.1 部署目标
<!-- 基于架构文档的部署目标 -->

### 2.2 部署架构
```
<!-- 部署架构图 -->
┌─────────────────────────────────────────────────────────────┐
│                         Load Balancer                        │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         ┌─────────┐    ┌─────────┐    ┌─────────┐
         │  App 1  │    │  App 2  │    │  App 3  │
         └─────────┘    └─────────┘    └─────────┘
              │               │               │
              └───────────────┼───────────────┘
                              ▼
                    ┌──────────────────┐
                    │    Database      │
                    └──────────────────┘
```

## 3. 环境配置

### 3.1 环境列表

| 环境 | 域名 | 集群 | 配置文件 |
|------|------|------|----------|
| DEV | dev.example.com | dev-cluster | .env.dev |
| QA | qa.example.com | qa-cluster | .env.qa |
| STAGING | staging.example.com | prod-cluster | .env.staging |
| PROD | example.com | prod-cluster | .env.prod |

### 3.2 环境变量

| 变量名 | 描述 | DEV | QA | STAGING | PROD |
|--------|------|-----|-----|---------|------|
| DATABASE_URL | 数据库连接 | | | | |
| REDIS_URL | Redis连接 | | | | |
| JWT_SECRET | JWT密钥 | | | | [Secret] |

### 3.3 资源配置

| 组件 | DEV | QA | STAGING | PROD |
|------|-----|-----|---------|------|
| App副本数 | 1 | 2 | 2 | 3 |
| App CPU | 0.5 | 0.5 | 1 | 2 |
| App内存 | 512Mi | 512Mi | 1Gi | 2Gi |
| DB规格 | t3.micro | t3.small | t3.medium | t3.large |

## 4. CI/CD流水线

### 4.1 流水线概览
```
代码提交 → 构建 → 单元测试 → 构建镜像 → 部署DEV → 集成测试 → 部署QA → 部署PROD
    │                                        │              │            │
    └─ PR触发                                └─ 自动        └─ 自动      └─ 手动审批
```

### 4.2 构建配置

#### Dockerfile
```dockerfile
# 见项目根目录 Dockerfile
```

#### 构建命令
```bash
# 构建镜像
docker build -t app:${VERSION} .

# 推送镜像
docker push registry.example.com/app:${VERSION}
```

### 4.3 部署命令

```bash
# Kubernetes部署
kubectl apply -f k8s/deployment.yaml
kubectl set image deployment/app app=registry.example.com/app:${VERSION}

# 验证部署
kubectl rollout status deployment/app
```

## 5. 数据库变更

### 5.1 迁移脚本

| 版本 | 脚本 | 描述 | 可回滚 |
|------|------|------|--------|
| V001 | migrations/V001_init.sql | 初始化表结构 | 是 |
| V002 | migrations/V002_add_index.sql | 添加索引 | 是 |

### 5.2 迁移执行

```bash
# 执行迁移
./scripts/migrate.sh up

# 回滚迁移
./scripts/migrate.sh down
```

## 6. 发布流程

### 6.1 发布检查清单

#### 发布前
- [ ] TEST-REPORT.md 确认测试通过
- [ ] 数据库迁移脚本已准备
- [ ] 配置变更已准备
- [ ] 回滚方案已确认
- [ ] 发布窗口已确认
- [ ] 相关方已通知

#### 发布中
- [ ] 备份当前版本
- [ ] 执行数据库迁移
- [ ] 部署新版本
- [ ] 验证健康检查
- [ ] 验证核心功能

#### 发布后
- [ ] 监控指标正常
- [ ] 错误率在阈值内
- [ ] 性能指标正常
- [ ] 发布通知发送

### 6.2 灰度策略

| 阶段 | 流量比例 | 持续时间 | 回滚条件 |
|------|----------|----------|----------|
| 1 | 5% | 30分钟 | 错误率 > 1% |
| 2 | 20% | 1小时 | 错误率 > 0.5% |
| 3 | 50% | 2小时 | 错误率 > 0.1% |
| 4 | 100% | - | 错误率 > 0.05% |

### 6.3 回滚流程

```bash
# 1. 停止当前部署
kubectl rollout pause deployment/app

# 2. 回滚到上一版本
kubectl rollout undo deployment/app

# 3. 验证回滚
kubectl rollout status deployment/app

# 4. 如需回滚数据库
./scripts/migrate.sh down
```

## 7. 监控配置

### 7.1 健康检查

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 3000
  initialDelaySeconds: 15
  periodSeconds: 20

readinessProbe:
  httpGet:
    path: /health/ready
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 10
```

### 7.2 告警规则

| 告警名称 | 条件 | 严重级别 | 通知渠道 |
|----------|------|----------|----------|
| 高错误率 | error_rate > 1% | Critical | 电话+钉钉 |
| 高延迟 | p95 > 500ms | Warning | 钉钉 |
| 服务不可用 | up == 0 | Critical | 电话+钉钉 |

### 7.3 监控面板
- Grafana Dashboard: [链接]
- 日志查询: [Kibana链接]

## 8. 运维手册

### 8.1 常见问题处理

| 问题 | 症状 | 处理步骤 |
|------|------|----------|
| OOM | 容器重启 | 1. 查看日志 2. 增加内存限制 |
| 连接池耗尽 | 请求超时 | 1. 检查连接数 2. 重启服务 |

### 8.2 日常运维命令

```bash
# 查看日志
kubectl logs -f deployment/app

# 进入容器
kubectl exec -it deployment/app -- /bin/sh

# 扩容
kubectl scale deployment/app --replicas=5

# 查看资源使用
kubectl top pods
```

---
## 文档审批

| 角色 | 姓名 | 日期 | 状态 |
|------|------|------|------|
| DevOps | | | 已创建 |
| 技术负责人 | | | 待审核 |

## 下游文档
此文档将作为以下文档的输入：
- [ ] RUNBOOK.md (运维手册详情)
- [ ] MONITORING.md (监控配置详情)

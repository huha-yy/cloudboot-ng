# DevOps工程师 (DevOps Engineer)

## 角色职责

DevOps工程师负责构建和维护CI/CD流水线，管理基础设施，确保系统的可靠性、可观测性和安全性。

## 📋 文档产物（必须输出）

| 文档 | 输出路径 | 模板 | 下游消费者 |
|------|----------|------|------------|
| **DEPLOYMENT.md** | `docs/ops/DEPLOYMENT.md` | `assets/templates/DEPLOYMENT.md` | 全团队 |
| RUNBOOK.md | `docs/ops/RUNBOOK.md` | - | 运维团队 |
| MONITORING.md | `docs/ops/MONITORING.md` | - | 全团队 |

### 输入依赖
- `docs/design/ARCHITECTURE.md` (来自架构师)
- `docs/test/TEST-REPORT.md` (来自测试工程师)

### 文档产出流程
```
1. 阅读ARCHITECTURE.md，理解部署架构
2. 阅读TEST-REPORT.md，确认可发布
3. 复制模板创建DEPLOYMENT.md
4. 执行部署
5. 更新RUNBOOK.md
6. 通知全团队部署完成
```

## 核心能力

### 1. CI/CD流水线

#### GitHub Actions示例
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run tests
        run: npm test -- --coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Build Docker image
        run: |
          docker build -t app:${{ github.sha }} .
      
      - name: Push to registry
        run: |
          docker tag app:${{ github.sha }} registry.example.com/app:${{ github.sha }}
          docker push registry.example.com/app:${{ github.sha }}

  deploy:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to production
        run: |
          kubectl set image deployment/app app=registry.example.com/app:${{ github.sha }}
```

### 2. 容器化与编排

#### Dockerfile最佳实践
```dockerfile
# 多阶段构建
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

FROM node:20-alpine AS runner
WORKDIR /app
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

COPY --from=builder /app/node_modules ./node_modules
COPY --chown=nextjs:nodejs . .

USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
```

#### Kubernetes部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: app
        image: registry.example.com/app:latest
        ports:
        - containerPort: 3000
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        readinessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 15
          periodSeconds: 20
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
spec:
  selector:
    app: app
  ports:
  - port: 80
    targetPort: 3000
  type: ClusterIP
```

### 3. 基础设施即代码

#### Terraform示例
```hcl
# AWS EKS集群
module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  version         = "~> 19.0"
  
  cluster_name    = "production-cluster"
  cluster_version = "1.28"
  
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.private_subnets
  
  eks_managed_node_groups = {
    general = {
      desired_size   = 3
      min_size       = 2
      max_size       = 5
      instance_types = ["t3.medium"]
    }
  }
}

# RDS数据库
resource "aws_db_instance" "main" {
  identifier           = "production-db"
  engine               = "postgres"
  engine_version       = "15.4"
  instance_class       = "db.t3.medium"
  allocated_storage    = 100
  storage_encrypted    = true
  
  db_name              = "app"
  username             = var.db_username
  password             = var.db_password
  
  multi_az             = true
  skip_final_snapshot  = false
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
}
```

### 4. 监控与可观测性

#### Prometheus + Grafana
```yaml
# prometheus配置
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'app'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: app

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

rule_files:
  - /etc/prometheus/alerts/*.yml
```

#### 告警规则
```yaml
groups:
  - name: app-alerts
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) 
          / sum(rate(http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "高错误率告警"
          description: "5xx错误率超过5%"
      
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, 
            sum(rate(http_request_duration_seconds_bucket[5m])) 
            by (le)) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "高延迟告警"
          description: "P95延迟超过1秒"
```

### 5. 日志管理

#### ELK Stack配置
```yaml
# Filebeat配置
filebeat.inputs:
  - type: container
    paths:
      - '/var/log/containers/*.log'
    processors:
      - add_kubernetes_metadata:
          host: ${NODE_NAME}
          matchers:
            - logs_path:
                logs_path: "/var/log/containers/"

output.elasticsearch:
  hosts: ['elasticsearch:9200']
  index: "app-logs-%{+yyyy.MM.dd}"
```

### 6. 安全实践

#### 密钥管理
```yaml
# External Secrets Operator
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: app-secrets
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: app-secrets
  data:
    - secretKey: database-url
      remoteRef:
        key: prod/app/database
        property: url
```

## 协作接口

### 接收自架构师
- 部署架构设计
- 资源需求规格
- 监控指标定义

### 接收自开发团队
- 应用配置需求
- 环境变量清单
- 部署依赖说明

### 输出给全团队
- 部署状态通知
- 系统健康报告
- 故障分析报告

### 输出给测试团队
- 测试环境配置
- 环境重置能力
- 部署日志

## 运维手册

### 故障响应流程
1. **检测** → 监控告警触发
2. **响应** → 值班人员确认
3. **诊断** → 查看日志和指标
4. **修复** → 执行修复操作
5. **复盘** → 编写事后分析报告

### 发布检查清单
- [ ] 所有测试通过
- [ ] 配置变更已审核
- [ ] 数据库迁移已准备
- [ ] 回滚方案已就绪
- [ ] 监控告警已配置
- [ ] 通知相关方

## 文档交接模板

完成部署后，使用以下格式通知：

```markdown
## 📋 部署完成通知 - DevOps → 全团队

### 产出文档
- docs/ops/DEPLOYMENT.md (状态: Approved)
- docs/ops/RUNBOOK.md (状态: Approved)

### 部署信息
- 版本: [X.X.X]
- 环境: [QA/STAGING/PROD]
- 时间: [时间]
- 状态: 成功/失败

### 访问地址
- 应用URL: [URL]
- 监控面板: [URL]
- 日志查询: [URL]

### 注意事项
- [需要关注的点]
```

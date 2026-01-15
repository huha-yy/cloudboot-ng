# 项目模板与规范

## 项目启动文档

```markdown
# 项目名称

## 项目概述
- 项目背景
- 业务目标
- 成功指标

## 团队配置
| 角色 | 人员 | 职责 |
|------|------|------|
| 产品经理 |  | 需求管理 |
| 架构师 |  | 技术架构 |
| 技术负责人 |  | 开发管理 |
| 前端开发 |  | UI实现 |
| 后端开发 |  | 服务端开发 |
| 测试工程师 |  | 质量保障 |
| DevOps |  | 运维部署 |

## 里程碑计划
| 阶段 | 目标 | 时间 | 交付物 |
|------|------|------|--------|
| M1 | MVP | | |
| M2 | Beta | | |
| M3 | GA | | |

## 技术栈
- 前端: 
- 后端: 
- 数据库: 
- 基础设施: 

## 风险评估
| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
|      |      |      |          |
```

## 功能开发模板

### 功能设计文档
```markdown
# 功能名称

## 需求背景
- 用户痛点
- 业务价值
- 成功指标

## 功能描述
### 用户故事
As a [用户角色]
I want [功能]
So that [价值]

### 验收标准
- [ ] AC1: 
- [ ] AC2: 
- [ ] AC3: 

## 技术设计
### 数据模型
```

### API设计
```

### 时序图
```

## 测试计划
### 测试场景
| 场景 | 输入 | 预期 |
|------|------|------|
|      |      |      |

## 发布计划
- 灰度策略
- 回滚方案
```

## 代码仓库结构

### 单体应用
```
project/
├── .github/
│   └── workflows/
├── src/
│   ├── api/
│   ├── services/
│   ├── models/
│   └── utils/
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── docs/
├── scripts/
├── docker/
├── .env.example
├── docker-compose.yml
├── Makefile
└── README.md
```

### 微服务项目
```
project/
├── services/
│   ├── user-service/
│   ├── order-service/
│   └── payment-service/
├── libs/
│   ├── common/
│   └── proto/
├── infra/
│   ├── k8s/
│   └── terraform/
├── docs/
│   ├── architecture/
│   └── api/
└── docker-compose.yml
```

### 前端项目
```
frontend/
├── public/
├── src/
│   ├── components/
│   │   ├── ui/
│   │   └── business/
│   ├── pages/
│   ├── hooks/
│   ├── services/
│   ├── stores/
│   ├── utils/
│   ├── styles/
│   └── types/
├── tests/
├── .storybook/
└── package.json
```

## 环境配置模板

### 环境变量示例
```bash
# .env.example

# 应用配置
APP_NAME=my-app
APP_ENV=development
APP_PORT=3000
APP_DEBUG=true

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mydb
DB_USER=
DB_PASSWORD=

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=
JWT_EXPIRES_IN=24h

# 第三方服务
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_REGION=ap-northeast-1

# 监控配置
SENTRY_DSN=
```

### Docker Compose模板
```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=development
    depends_on:
      - db
      - redis
    volumes:
      - .:/app
      - /app/node_modules

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: mydb
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  pgdata:
```

## 文档模板

### API文档模板
```markdown
# API名称

## 接口信息
- **URL**: `/api/v1/resource`
- **Method**: POST
- **Auth**: Bearer Token

## 请求参数
### Headers
| 名称 | 类型 | 必填 | 说明 |
|------|------|------|------|
| Authorization | string | 是 | Bearer token |

### Body
| 名称 | 类型 | 必填 | 说明 |
|------|------|------|------|
| field1 | string | 是 | 字段说明 |

### 示例
```json
{
  "field1": "value"
}
```

## 响应
### 成功响应
```json
{
  "code": 0,
  "data": {},
  "message": "success"
}
```

### 错误响应
| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误 |
| 40101 | 未授权 |
```

## 质量标准

### 代码覆盖率要求
| 类型 | 最低要求 |
|------|----------|
| 行覆盖率 | 80% |
| 分支覆盖率 | 70% |
| 函数覆盖率 | 80% |

### 性能指标
| 指标 | 目标 |
|------|------|
| API响应时间P95 | < 200ms |
| 页面加载时间 | < 3s |
| 错误率 | < 0.1% |
| 可用性 | 99.9% |

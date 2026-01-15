# 测试工程师 (QA Engineer)

## 角色职责

测试工程师是产品质量的守护者，负责设计测试策略、编写测试用例、执行测试并确保产品符合验收标准。

## 📋 文档产物（必须输出）

| 文档 | 输出路径 | 模板 | 下游消费者 |
|------|----------|------|------------|
| **TEST-PLAN.md** | `docs/test/TEST-PLAN.md` | `assets/templates/TEST-PLAN.md` | 开发团队、产品经理 |
| TEST-CASES.md | `docs/test/TEST-CASES.md` | - | 开发团队 |
| **TEST-REPORT.md** | `docs/test/TEST-REPORT.md` | - | 产品经理、DevOps |
| BUG-REPORT.md | `docs/test/BUG-REPORT.md` | - | 开发团队 |

### 输入依赖
- `docs/requirements/PRD.md` (来自产品经理)
- `docs/requirements/ACCEPTANCE.md` (来自产品经理)
- `docs/impl/FRONTEND-IMPL.md` (来自前端开发)
- `docs/impl/BACKEND-IMPL.md` (来自后端开发)

### 文档产出流程
```
1. 阅读PRD.md，理解验收标准
2. 复制模板创建TEST-PLAN.md
3. 执行测试
4. 产出TEST-REPORT.md
5. 测试通过后，执行文档交接给DevOps
```

## 核心能力

### 1. 测试策略设计

#### 测试金字塔
```
            /\
           /  \     E2E Tests (10%)
          /----\    - 关键业务流程
         /      \   - 用户旅程验证
        /--------\  Integration Tests (20%)
       /          \ - API测试
      /------------\- 服务间集成
     /              \ Unit Tests (70%)
    /----------------\ - 函数级测试
                       - 组件测试
```

#### 测试类型覆盖
| 测试类型 | 目的 | 工具 |
|----------|------|------|
| 单元测试 | 验证独立功能单元 | Jest/Pytest |
| 集成测试 | 验证模块间交互 | Supertest/httpx |
| E2E测试 | 验证完整业务流程 | Playwright/Cypress |
| 性能测试 | 验证系统性能指标 | k6/JMeter |
| 安全测试 | 发现安全漏洞 | OWASP ZAP |

### 2. 测试用例设计

#### 用例模板
```markdown
## 测试用例: TC001 - 用户登录

### 前置条件
- 已注册用户存在
- 系统运行正常

### 测试数据
| 字段 | 值 |
|------|-----|
| email | test@example.com |
| password | Test123! |

### 测试步骤
1. 打开登录页面
2. 输入邮箱
3. 输入密码
4. 点击登录按钮

### 预期结果
- 登录成功，跳转到首页
- 显示用户名
- 获得有效token

### 实际结果
[待填写]

### 状态: [通过/失败/阻塞]
```

#### 边界值分析
```markdown
## 密码验证边界测试

| 场景 | 输入 | 预期结果 |
|------|------|----------|
| 最小长度 | 7字符 | 失败：密码太短 |
| 有效边界 | 8字符 | 成功 |
| 最大长度 | 128字符 | 成功 |
| 超出限制 | 129字符 | 失败：密码太长 |
```

### 3. 自动化测试

#### API测试示例
```python
import pytest
import httpx

class TestUserAPI:
    base_url = "http://localhost:8000/api/v1"
    
    @pytest.fixture
    async def client(self):
        async with httpx.AsyncClient() as client:
            yield client
    
    @pytest.fixture
    async def auth_token(self, client):
        response = await client.post(f"{self.base_url}/auth/login", json={
            "email": "test@example.com",
            "password": "Test123!"
        })
        return response.json()["access_token"]
    
    async def test_get_user_profile(self, client, auth_token):
        response = await client.get(
            f"{self.base_url}/users/me",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert "email" in data
        assert "id" in data
```

#### E2E测试示例
```typescript
// Playwright测试
import { test, expect } from '@playwright/test';

test.describe('用户登录流程', () => {
  test('成功登录', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('[data-testid="email"]', 'test@example.com');
    await page.fill('[data-testid="password"]', 'Test123!');
    await page.click('[data-testid="login-btn"]');
    
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('[data-testid="user-name"]')).toBeVisible();
  });
  
  test('登录失败 - 错误密码', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('[data-testid="email"]', 'test@example.com');
    await page.fill('[data-testid="password"]', 'wrong');
    await page.click('[data-testid="login-btn"]');
    
    await expect(page.locator('.error-message')).toContainText('密码错误');
  });
});
```

### 4. 缺陷管理

#### Bug报告模板
```markdown
## Bug: BUG-001

### 标题
用户登录后页面白屏

### 严重程度: Critical
### 优先级: P0

### 环境
- 浏览器: Chrome 120
- 操作系统: macOS 14.0
- 版本: v1.2.3

### 复现步骤
1. 打开登录页面
2. 输入正确的邮箱和密码
3. 点击登录按钮

### 预期结果
跳转到dashboard页面

### 实际结果
页面白屏，控制台报错：
`TypeError: Cannot read property 'user' of undefined`

### 截图/日志
[附件]

### 根因分析
用户信息API返回格式变更，前端未兼容

### 修复建议
检查API响应结构，添加空值保护
```

### 5. 测试报告

```markdown
## 测试报告 - Sprint 2024-W03

### 概要
| 指标 | 数值 |
|------|------|
| 测试用例总数 | 156 |
| 执行用例数 | 150 |
| 通过 | 142 |
| 失败 | 5 |
| 阻塞 | 3 |
| 通过率 | 94.7% |

### 缺陷统计
| 严重程度 | 新增 | 修复 | 遗留 |
|----------|------|------|------|
| Critical | 1 | 1 | 0 |
| Major | 3 | 2 | 1 |
| Minor | 5 | 3 | 2 |

### 风险评估
1. ⚠️ 支付模块存在1个Major缺陷未修复
2. ⚠️ 性能测试未覆盖高并发场景

### 发布建议
建议修复Major缺陷后发布
```

## 协作接口

### 接收自产品经理
- 验收标准
- 业务规则
- 测试场景

### 接收自开发团队
- 可测试构建
- 技术文档
- 变更说明

### 输出给项目团队
- 测试报告
- 缺陷清单
- 质量评估

### 输出给发布决策
- 发布建议
- 风险说明
- 遗留问题清单

## 测试环境管理

### 环境隔离
- DEV：开发自测
- QA：系统测试
- UAT：用户验收
- STAGING：预发布验证

## 文档交接模板

完成测试后，使用以下格式交接：

```markdown
## 📋 文档交接 - 测试工程师 → DevOps

### 产出文档
- docs/test/TEST-PLAN.md (状态: Approved)
- docs/test/TEST-REPORT.md (状态: Approved)

### 核心内容摘要
- 测试用例总数: [X个]
- 通过率: [X%]
- 遗留缺陷: [Critical: 0, Major: X, Minor: X]
- 发布建议: [可发布/需修复后发布]

### 接收方待办
- [ ] DevOps阅读测试报告
- [ ] 创建DEPLOYMENT.md
- [ ] 准备部署

### 特别注意
- [遗留风险]
- [特别关注的功能]
```

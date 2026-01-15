# 后端开发工程师 (Backend Developer)

## 角色职责

后端开发工程师负责实现服务端业务逻辑，设计数据模型，提供稳定可靠的API服务。

## 📋 文档产物（必须输出）

| 文档 | 输出路径 | 下游消费者 |
|------|----------|------------|
| **BACKEND-IMPL.md** | `docs/impl/BACKEND-IMPL.md` | 测试工程师、技术负责人 |
| API变更日志 | `docs/api/CHANGELOG.md` | 前端开发 |

### 输入依赖
- `docs/api/API-SPEC.yaml` (来自架构师)
- `docs/design/DATABASE.md` (来自架构师)
- `docs/dev/TASK-BREAKDOWN.md` (来自技术负责人)

### 文档产出流程
```
1. 阅读API-SPEC.yaml，理解接口契约
2. 按TASK-BREAKDOWN.md分配的任务实现功能
3. 完成后更新BACKEND-IMPL.md记录实现细节
4. 执行文档交接给测试工程师
```

## 核心能力

### 1. 项目结构（Clean Architecture）

```
src/
├── api/                    # API层
│   ├── handlers/          # 请求处理器
│   ├── middleware/        # 中间件
│   └── routes/            # 路由定义
├── application/           # 应用层
│   ├── services/          # 业务服务
│   ├── dto/               # 数据传输对象
│   └── validators/        # 验证器
├── domain/                # 领域层
│   ├── entities/          # 实体
│   ├── repositories/      # 仓储接口
│   └── events/            # 领域事件
├── infrastructure/        # 基础设施层
│   ├── database/          # 数据库实现
│   ├── cache/             # 缓存实现
│   └── external/          # 外部服务
└── config/                # 配置
```

### 2. API实现模式

#### RESTful Controller
```python
# Python FastAPI示例
from fastapi import APIRouter, Depends, HTTPException
from typing import List

router = APIRouter(prefix="/users", tags=["users"])

@router.get("/", response_model=List[UserResponse])
async def list_users(
    page: int = 1,
    size: int = 20,
    service: UserService = Depends(get_user_service)
):
    return await service.list_users(page, size)

@router.post("/", response_model=UserResponse, status_code=201)
async def create_user(
    request: CreateUserRequest,
    service: UserService = Depends(get_user_service)
):
    return await service.create_user(request)
```

#### Service层
```python
class UserService:
    def __init__(self, repo: UserRepository, cache: CacheService):
        self.repo = repo
        self.cache = cache
    
    async def get_user(self, user_id: str) -> User:
        # 先查缓存
        cached = await self.cache.get(f"user:{user_id}")
        if cached:
            return User.parse_raw(cached)
        
        # 查数据库
        user = await self.repo.find_by_id(user_id)
        if not user:
            raise UserNotFoundException(user_id)
        
        # 写入缓存
        await self.cache.set(f"user:{user_id}", user.json(), ttl=3600)
        return user
```

### 3. 数据库设计

#### 实体设计原则
```python
# SQLAlchemy示例
class User(Base):
    __tablename__ = "users"
    
    id = Column(UUID, primary_key=True, default=uuid4)
    email = Column(String(255), unique=True, nullable=False)
    password_hash = Column(String(255), nullable=False)
    status = Column(Enum(UserStatus), default=UserStatus.ACTIVE)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, onupdate=datetime.utcnow)
    
    # 关系
    profile = relationship("UserProfile", back_populates="user", uselist=False)
    orders = relationship("Order", back_populates="user")
```

#### 索引策略
- 主键自动索引
- 外键添加索引
- 高频查询字段添加索引
- 复合索引遵循最左匹配原则

### 4. 错误处理

```python
# 统一错误处理
class AppException(Exception):
    def __init__(self, code: str, message: str, status: int = 400):
        self.code = code
        self.message = message
        self.status = status

class ErrorResponse(BaseModel):
    code: str
    message: str
    details: dict = None

@app.exception_handler(AppException)
async def app_exception_handler(request, exc: AppException):
    return JSONResponse(
        status_code=exc.status,
        content=ErrorResponse(
            code=exc.code,
            message=exc.message
        ).dict()
    )
```

### 5. 安全实践

#### 认证与授权
```python
# JWT认证中间件
async def auth_middleware(
    credentials: HTTPAuthorizationCredentials = Security(bearer_scheme)
):
    token = credentials.credentials
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
        return CurrentUser(**payload)
    except jwt.ExpiredSignatureError:
        raise HTTPException(401, "Token expired")
    except jwt.InvalidTokenError:
        raise HTTPException(401, "Invalid token")
```

#### 输入验证
- 使用Pydantic/marshmallow进行严格的输入验证
- SQL参数化查询防止注入
- 敏感数据加密存储

### 6. 日志与监控

```python
import structlog

logger = structlog.get_logger()

async def create_order(request: CreateOrderRequest):
    logger.info("creating_order", 
        user_id=request.user_id,
        items_count=len(request.items))
    
    try:
        order = await order_service.create(request)
        logger.info("order_created", order_id=order.id)
        return order
    except Exception as e:
        logger.error("order_creation_failed", 
            error=str(e),
            user_id=request.user_id)
        raise
```

## 协作接口

### 接收自架构师
- API设计规范
- 数据库设计
- 技术选型

### 输出给前端
- API文档（OpenAPI/Swagger）
- 错误码文档
- 联调支持

### 输出给测试
- API测试环境
- 测试数据准备
- 接口变更通知

### 输出给DevOps
- 部署配置
- 环境变量清单
- 健康检查接口

## 代码质量

### 单元测试覆盖
```python
@pytest.mark.asyncio
async def test_create_user_success():
    repo = MockUserRepository()
    service = UserService(repo)
    
    request = CreateUserRequest(email="test@example.com", password="secure123")
    user = await service.create_user(request)
    
    assert user.email == "test@example.com"
    assert repo.save_called
```

### 代码审查要点
- 是否符合SOLID原则
- 错误处理是否完善
- 性能是否可接受
- 安全风险是否可控

## 文档交接模板

完成开发后，使用以下格式交接：

```markdown
## 📋 文档交接 - 后端开发 → 测试工程师

### 产出文档
- docs/impl/BACKEND-IMPL.md (状态: Completed)

### 完成的任务
| 任务ID | 任务名称 | 状态 |
|--------|----------|------|
| BE-001 | xxx | 已完成 |
| BE-002 | xxx | 已完成 |

### API测试信息
- 基础URL: [URL]
- Swagger文档: [URL]
- 测试数据: [说明]

### 需要测试的接口
- [ ] POST /api/v1/xxx
- [ ] GET /api/v1/xxx

### 已知问题
- [如有]
```

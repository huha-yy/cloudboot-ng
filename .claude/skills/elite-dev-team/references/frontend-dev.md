# 前端开发工程师 (Frontend Developer)

## 角色职责

前端开发工程师负责实现用户界面，打造流畅的用户体验，确保应用在各端设备上的一致性表现。

## 📋 文档产物（必须输出）

| 文档 | 输出路径 | 下游消费者 |
|------|----------|------------|
| **FRONTEND-IMPL.md** | `docs/impl/FRONTEND-IMPL.md` | 测试工程师、技术负责人 |
| 组件文档 | `src/components/README.md` | 开发团队 |

### 输入依赖
- `docs/api/API-SPEC.yaml` (来自架构师)
- `docs/dev/TASK-BREAKDOWN.md` (来自技术负责人)

### 文档产出流程
```
1. 阅读API-SPEC.yaml，理解接口契约
2. 按TASK-BREAKDOWN.md分配的任务实现功能
3. 完成后更新FRONTEND-IMPL.md记录实现细节
4. 执行文档交接给测试工程师
```

## 核心能力

### 1. 技术栈

#### React生态
```javascript
// 项目结构
src/
├── components/          // 可复用组件
│   ├── ui/             // 基础UI组件
│   └── business/       // 业务组件
├── pages/              // 页面组件
├── hooks/              // 自定义Hooks
├── services/           // API服务
├── stores/             // 状态管理
├── utils/              // 工具函数
└── styles/             // 样式文件
```

#### 状态管理模式
```typescript
// Zustand示例
import { create } from 'zustand';

interface UserStore {
  user: User | null;
  setUser: (user: User) => void;
  logout: () => void;
}

export const useUserStore = create<UserStore>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
  logout: () => set({ user: null }),
}));
```

### 2. 组件设计原则

#### 组件分层
- **UI组件**：纯展示，无业务逻辑
- **容器组件**：负责数据获取和状态管理
- **业务组件**：封装特定业务逻辑

#### 组件规范
```typescript
// 组件模板
interface ButtonProps {
  variant: 'primary' | 'secondary' | 'ghost';
  size: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  onClick?: () => void;
  children: React.ReactNode;
}

export const Button: React.FC<ButtonProps> = ({
  variant = 'primary',
  size = 'md',
  disabled = false,
  loading = false,
  onClick,
  children,
}) => {
  // 实现...
};
```

### 3. 性能优化

#### 渲染优化
- React.memo：避免不必要的重渲染
- useMemo/useCallback：缓存计算结果和回调
- 虚拟列表：大数据量列表优化
- 代码分割：React.lazy + Suspense

#### 资源优化
- 图片懒加载
- 资源压缩与CDN
- Service Worker缓存
- Tree Shaking

### 4. 测试策略

```typescript
// 组件测试示例
import { render, screen, fireEvent } from '@testing-library/react';

describe('Button', () => {
  it('should render correctly', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('should handle click', () => {
    const onClick = jest.fn();
    render(<Button onClick={onClick}>Click</Button>);
    fireEvent.click(screen.getByText('Click'));
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});
```

### 5. 样式方案

#### CSS-in-JS (Styled Components / Emotion)
```typescript
const StyledButton = styled.button<{ $variant: string }>`
  padding: 8px 16px;
  border-radius: 4px;
  background: ${({ $variant }) => 
    $variant === 'primary' ? '#007bff' : '#6c757d'};
`;
```

#### Tailwind CSS
```jsx
<button className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
  Click me
</button>
```

## 协作接口

### 接收自架构师
- API接口文档
- 前端架构规范
- 技术选型决策

### 接收自设计师
- UI设计稿
- 交互规范
- 设计系统

### 输出给测试工程师
- 可测试的构建产物
- 组件使用文档
- 已知问题清单

### 与后端协作
- API联调
- 接口问题反馈
- 数据格式确认

## 代码规范

### ESLint配置要点
- 强制使用TypeScript
- 禁止any类型
- 要求函数返回类型
- 组件命名PascalCase
- 文件命名kebab-case

### Git提交规范
```
feat: 新功能
fix: 修复bug
refactor: 重构
style: 样式调整
docs: 文档更新
test: 测试相关
chore: 构建/工具
```

## 文档交接模板

完成开发后，使用以下格式交接：

```markdown
## 📋 文档交接 - 前端开发 → 测试工程师

### 产出文档
- docs/impl/FRONTEND-IMPL.md (状态: Completed)

### 完成的任务
| 任务ID | 任务名称 | 状态 |
|--------|----------|------|
| FE-001 | xxx | 已完成 |
| FE-002 | xxx | 已完成 |

### 测试入口
- 页面URL: [URL]
- 测试账号: [账号信息]

### 需要测试的功能
- [ ] 功能1
- [ ] 功能2

### 已知问题
- [如有]
```

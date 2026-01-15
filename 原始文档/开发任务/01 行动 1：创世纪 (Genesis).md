这是 **行动 1：创世纪 (Genesis)** 的具体任务执行规格书。

请将以下内容保存为 `TASK_01_Genesis.md`。这是您发给 Claude Code 的第一道战术指令，它将决定整个项目的代码结构和视觉基因。

---

# TASK_01_Genesis.md
**任务名称**：行动 1：创世纪 (骨架构建与视觉定调)  
**执行模式**：⚡️ 交互开发 (Interactive)  
**前置依赖**：必须先读取 `PRODUCT_Blueprint.md`, `ARCH_Stack.md`, `UI_Design_System.md`。

---

## 1. 任务目标 (Objective)
构建 Cloudboot NG 的单体应用骨架，打通 **Go + HTMX + Tailwind + Alpine.js** 技术栈，并实现核心 UI 组件库的可视化展示。

**交付物**：

1. 可编译运行的 Go 项目结构。
2. 自动化构建脚本 (`Makefile`)。
3. `/design-system` 路由，展示深色工业风组件库的实际渲染效果。

---

## 2. 详细执行步骤 (Execution Steps)
### Step 1: 基础设施初始化 (Infrastructure)
+ **初始化 Module**：`go mod init cloudboot-ng`。
+ **目录结构**：严格遵循 `ARCH_Stack.md` 定义的标准目录结构。
+ **Web Server**：
    - 引入 **Echo v4** 框架。
    - 配置 Middleware: Logger, Recover.
    - 配置静态资源服务：映射 `/static` 到 `web/static`。
+ **构建系统**：
    - 创建 `Makefile`。
    - Target `setup`: 下载 `htmx.min.js`, `alpine.min.js` 到 `web/static/js/`。下载 Tailwind Standalone CLI (根据 OS 自动判断)。
    - Target `dev`: 并行运行 Tailwind Watch 模式和 Go Run (推荐使用 `air` 热重载如果可用，否则直接 `go run`)。

### Step 2: 前端工程化 (Frontend Engineering)
+ **Tailwind 配置**：
    - 初始化 `tailwind.config.js`。
    - **关键配置**：
        * `content`: `["./web/templates/**/*.html"]`
        * `theme`: 严格复刻 `UI_Design_System.md` 中的色彩 (Slate-950/900, Emerald-500) 和字体 (Inter, JetBrains Mono)。
        * `plugins`: 引入 `@tailwindcss/forms`, `@tailwindcss/typography` (如果 CLI 支持)。
+ **模板引擎**：
    - 在 `internal/pkg/renderer` 中实现 Echo 的 `Renderer` 接口，解析 `web/templates` 下的所有 HTML 文件。

### Step 3: 核心组件实现 (Atomic Components)
在 `web/templates/components` 目录下，根据 `UI_Design_System.md` 的 HTML 片段实现以下组件（使用 Go `define` 语法）：

1. **Layout (**`layouts/base.html`**)**:
    - 包含 HTML5 Boilerplate。
    - 引入 CSS, HTMX, Alpine。
    - 设置 `body` 背景色为 `bg-slate-950`，文字色 `text-slate-200`。
2. **Navigation (**`components/sidebar.html`**)**:
    - 左侧固定侧边栏。
    - 包含 Logo (Cloudboot NG)。
    - 链接：Dashboard, Assets, Store, Settings。
    - 实现 Active 状态高亮 (Emerald 文字 + 左侧光标)。
3. **UI Primitives**:
    - `components/card.html`: 玻璃拟态卡片。
    - `components/button.html`: 工业风按钮 (Primary/Secondary)。
    - `components/badge.html`: 呼吸灯状态点。
    - `components/terminal.html`: 黑色日志窗口 (带 macOS 风格红黄绿点)。

### Step 4: 设计系统展示页 (The Showcase)
+ **Handler**: 创建 `internal/handlers/design_system.go`。
+ **Route**: 注册 `GET /design-system`。
+ **Template**: 创建 `web/templates/views/design_system.html`。
+ **内容**：
    - 在一个页面内，通过 Grid 布局展示所有组件。
    - 展示按钮的 Hover/Active 状态。
    - 在 Terminal 组件中硬编码几行模拟日志（绿色字体），验证等宽字体渲染是否正确。

---

## 3. 验收标准 (Acceptance Criteria)
当 Claude 完成工作后，我（人类指挥官）将执行 `make dev` 并访问 `http://localhost:8080/design-system`。

**我必须看到：**

1. **无报错**：控制台没有 404 或 500 错误。
2. **深色质感**：背景必须是深邃的蓝黑 (Slate-950)，而不是纯黑。
3. **字体正确**：标题是无衬线体，Terminal 里的日志必须是 **JetBrains Mono**。
4. **交互微效**：鼠标悬停在按钮和卡片上时，必须有微弱的光晕或边框变色效果。

---

### **给 Claude 的启动 Prompt**：
_(请复制以下指令给 Claude)_

> "Claude, 我们开始 **行动 1：创世纪**。  
请读取当前目录下的 `TASK_01_Genesis.md`，并结合之前读取的 `ARCH_Stack.md` 和 `UI_Design_System.md`。
>
> 请按照任务书的步骤，一步步执行：
>
> 1. 先初始化项目结构和 Makefile。
> 2. 配置好 Tailwind 和 Web Server。
> 3. 编写 UI 组件。
> 4. 最后实现 `/design-system` 页面。
>
> **注意**：每完成一个 Step，请运行测试或检查命令，确保没有错误后再继续。遇到 CSS 问题请参考 `UI_Design_System.md` 的严格定义。现在开始 Step 1。"
>


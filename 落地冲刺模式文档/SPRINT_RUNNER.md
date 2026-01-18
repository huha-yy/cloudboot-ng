CloudBoot NG 落地冲刺自动执行指令 (The Ralph-Loop)
1. 核心运行逻辑 (Loop Logic)
Claude, 请严格按照以下循环逻辑执行开发，直到 MISSION_CONTROL.md 中的所有“生产级必修项”标记为完成：
Read: 读取 MISSION_CONTROL.md 中优先级最高且未完成的任务。
Plan: 基于 ARCH.md 和 API.md 构思最快速、最稳健的实现方案（Ralph-Wiggum 模式：结果导向）。
Execute: 编写/修改代码。
Verify: 运行 go test ./... 确保核心逻辑不被破坏。
Sync: 运行 scripts/sync.sh 完成自动化版本固化。
Update: 更新 MISSION_CONTROL.md 中的任务状态及任何新发现的“技术债”。
Loop: 自动开始下一个任务。
2. 第一阶段：确定性与原子性 (The Idempotent Chain)
目标：杜绝硬件操作的“半截子”状态。
Provider 强制闭环：修改 internal/provider 接口，强制实现 Plan() -> Apply() 的原子序列。
状态检查机制：在执行硬件修改前，必须先调用 probe 确认当前状态。若当前状态已达标，立即跳过执行并标记为 Success（幂等性）。
超时熔断：所有底层调用必须封装在 context.WithTimeout 中，防止因硬件卡死导致的全局阻塞。
3. 第二阶段：全息日志流 (The Matrix View)
目标：实现极致透明的运维体验。
SSE 实时通道：在 Core 平台实现 /api/logs/stream 接口，利用 Server-Sent Events 将前线日志实时推送至前端。
HTMX 日志组件：利用 hx-ext="sse" 插件在 Web UI 实现全息滚动日志，支持自动滚动和关键词高亮（Error/Warning）。
状态机映射：将复杂的 Provider 输出日志映射为 UI 上的“关键进度步点”。
4. 第三阶段：安全与合规气闸 (Secure Logistics)
目标：通过银行合规性审计。
DRM 校验引擎：实现 internal/store 中的加密包指纹校验。
水印审计逻辑：在 Provider 运行时检测 watermark.json 与系统 License 的归属关系。
Session Key 加密：实现 Provider 发送过程中的动态二次加密，确保 Payload 仅在内存中解密。
5. 暴力开发准则 (Ralph-Wiggum Rules)
不要废话：减少长篇大论的解释，直接给出可运行的代码。
异常捕获：宁可报错停止，也不要带着不确定的状态继续运行。
上下文自压缩：每完成一个 commit，请检查上下文长度。如果感到吃力，自动更新 CONTEXT_SNAP.md 并“重启”你的逻辑记忆。
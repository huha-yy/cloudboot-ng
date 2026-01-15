这是一份定义 CloudBoot NG **核心资产（墨盒）** 生产、封装、交互与保护标准的最终技术规范。

它解决了我们商业模式中最关键的问题：**如何确保我们的驱动代码（知识）既能被机器理解，又能被 AI 批量生产，同时还绝不可能被竞争对手（云新）低成本窃取。**

---

# CloudBoot NG 全景产品设计白皮书
**第四卷：CSPM 标准与墨盒机制设计 (The Protocol)**

+ **项目代号**：Phoenix Protocol
+ **版本**：v1.0 (Final Spec)
+ **状态**：**已定稿 (Approved)**
+ **日期**：2026年1月
+ **密级**：核心机密 (Core Confidential)
+ **关联文档**：第二卷 (Core 平台)，第三卷 (BootOS)

---

## 1. 概述 (Overview)
### 1.1 设计愿景
**CSPM (CloudBoot Server Provider Mechanism)** 是物理机管理领域的通用语言。它定义了上层意图（Intent）如何转化为底层指令（Instruction）的标准协议。

我们的目标是构建一个**“硬件乐高体系”**：将物理机的复杂性拆解为标准化的原子组件，通过 AI 工厂批量生产，通过加密墨盒进行分发，通过声明式接口进行调用。

### 1.2 核心原则
1. **Decoupled (彻底解耦)**：业务逻辑（Provider）与底层工具（Adaptor）分离，配置数据（Config）与执行代码（Binary）分离。
2. **Secure (原生安全)**：墨盒默认为加密状态，仅在受信任的内存环境中解密运行，离线可审计。
3. **AI-First (AI 友好)**：协议设计完全基于 JSON Schema 和 标准 IO 流，极度利于 Claude Code 进行代码生成和逻辑验证。

---

## 2. 双层驱动架构 (The Bi-Layer Architecture)
为了应对信创硬件的“排列组合爆炸”，我们将驱动架构拆分为两层。

### 2.1 Layer 1: Provider (业务编排层)
这是**面向用户**的产品单元，也是我们 Store 中的售卖单元（SKU）。

+ **定义**：针对特定**“厂商+机型”**的逻辑封装。
+ **命名规范**：`provider-<vendor>-<model>` (例: `provider-huawei-taishan200`)。
+ **核心职责**：**“翻译与编排”**。
    - 它知道“华为 TaiShan 200”服务器由“LSI RAID卡”和“华为 BIOS”组成。
    - 它负责将用户的通用意图（“配置 RAID 10”）翻译成对底层 Adaptor 的具体调用参数。
    - 它处理机型特有的坑（Quirks），例如：“这款机器重启后需要等待 60秒才能识别网卡”。

### 2.2 Layer 2: Adaptor (原子执行层)
这是**面向芯片**的技术单元，也是我们技术壁垒的基石。

+ **定义**：针对底层芯片或管理协议的指令执行器。
+ **命名规范**：`adaptor-<category>-<chipset>` (例: `adaptor-raid-lsi3108`, `adaptor-bios-ami-aptio`).
+ **核心职责**：**“执行与屏蔽”**。
    - 它封装了厂商原始的二进制工具（如 `storcli`, `sas3ircu`, `ipmitool`）。
    - 它负责解析厂商工具那令人崩溃的非标文本输出，将其转化为标准 JSON。
+ **封装形态**：Adaptor 通常作为静态资源被编译进 Provider 二进制中，**对用户不可见**。

---

## 3. 交互协议设计 (The Interaction Protocol)
Provider 是一个独立的二进制程序，它与调用者（Agent）之间通过 **JSON over Stdin/Stdout** 进行通信。

### 3.1 核心动词 (Verbs)
Provider 必须支持以下命令行子命令：

1. `probe` (探测)
    - **输入**：无。
    - **逻辑**：扫描当前硬件环境，确认自己是否适配该机器。
    - **输出**：返回硬件指纹和组件状态。
2. `plan` (预演)
    - **输入**：`DesiredState` (期望配置) + `CurrentState` (当前状态)。
    - **逻辑**：计算 Diff。例如：当前是 RAID 5，期望是 RAID 10 -> 需要执行 `delete` 然后 `create`。
    - **输出**：`ExecutionPlan` (操作序列，用于向用户展示风险)。
3. `apply` (执行)
    - **输入**：`DesiredState`。
    - **逻辑**：真正调用 Adaptor 修改硬件。
    - **输出**：执行结果报告。

### 3.2 标准数据契约 (Data Schema)
**输入契约 (Agent -> Provider)**

```json
{
  "action": "apply",
  "resource": "raid",
  "params": {
    "level": "10",
    "drives": ["252:1", "252:2", "252:3", "252:4"],
    "strip_size": "64KB",
    "cache_policy": "WriteBack"
  },
  "context": {
    "timeout": 300,
    "debug": true
  },
  "overlay": {
    // 用户微调参数，覆盖默认逻辑
    "ignore_battery_error": true
  }
}
```

**输出契约 (Provider -> Agent)**

```json
{
  "status": "success",
  "changed": true,
  "data": {
    "volume_id": "0",
    "capacity_gb": 1200
  },
  "error": null
}
```

**日志流契约 (Provider -> Agent -> Server)**  
Provider 运行时的所有日志必须打印到 `Stderr`，格式为单行 JSON：

```json
{"ts": 1715000100, "level": "INFO", "msg": "正在初始化 RAID 控制器..."}
{"ts": 1715000105, "level": "WARN", "msg": "检测到电池电量低，根据 Overlay 设置跳过检查", "code": "W_BAT_LOW"}
```

---

## 4. 墨盒物理结构与 DRM 机制 (The Ink Cartridge)
这是我们防止云新“白嫖”的物理防线。

### 4.1 文件结构：`.cbp` (CloudBoot Package)
用户下载的 Provider 是一个经过特殊封装的 ZIP 包。

```latex
provider-huawei-taishan.cbp
├── meta/
│   ├── manifest.json       # 元数据（支持的硬件ID列表、版本）
│   └── watermark.json      # 【数字水印】下载者ID、下载时间、交易流水号
├── bin/
│   └── provider.enc        # 【核心载荷】AES-256 加密的二进制文件
└── signature.sig           # 官方私钥对全包的签名
```

### 4.2 离线 DRM 逻辑 (Offline DRM)
由于银行环境不联网，我们无法做实时在线验证，因此采用**“公钥/私钥 + 离线 License”**机制。

1. **加密**：官方 Store 使用**全局产品主钥 (Master Key)** 对 Provider 二进制进行加密。
2. **授权**：用户购买企业版后，获得一个 `.lic` 文件。该文件包含用户的**解密私钥**（该私钥是针对用户生成的），以及授权范围。
3. **运行**：
    - CloudBoot Core 导入 `.cbp` 包。
    - CloudBoot Core 使用 `.lic` 中的私钥，解密出 Provider。
    - **关键点**：Core **立即**使用一个随机生成的**临时会话密钥 (Session Key)** 重新加密 Provider，然后再发给 BootOS。
    - **结果**：即使有人在网络层截获了发给 BootOS 的包，没有内存中的 Session Key 也无法解密。

### 4.3 审计与追责
+ **被动触发**：如果 CloudBoot Core 检测到 `.cbp` 包中的水印 ID 与当前系统的 License ID 不一致。
+ **反应**：
    - 并不阻断运行（为了业务稳定性）。
    - 但在界面显眼位置显示红色横幅：**“非授权来源组件运行中：[来源用户ID]”**。
    - 写入不可删除的审计日志。
+ **杀伤力**：这使得云新的驻场人员不敢随意拷贝他们在其他项目下载的墨盒，否则一旦被客户发现，就是重大的合规事故。

---

## 5. 配置透明化与微调机制 (Transparency & Overlays)
我们不仅提供黑盒的驱动，还提供白盒的控制权。

### 5.1 Provider Schema (说明书)
每个 Provider 包内包含一个 `schema.json`，描述它支持的所有参数。

+ **用途**：
    - CloudBoot Core 根据它自动生成 Web 表单。
    - OS Designer 根据它进行参数校验。
    - AI 根据它生成推荐配置。

### 5.2 User Overlay (用户补丁)
为了解决“同一型号不同批次硬件差异”的问题，我们引入 Overlay 机制。

+ **机制**：用户可以在 CloudBoot Core 界面上，针对特定机型创建一个 **"Config Patch"**。
+ **逻辑**：
    - Standard Config (官方标准) + User Overlay (用户微调) = Effective Config (最终生效配置)。
+ **示例**：
    - 官方默认 `timeout = 30s`。
    - 用户 Overlay 设置 `timeout = 120s`。
    - Provider 运行时读取最终的 120s。
+ **价值**：
    - 用户不需要等我们发版就能解决现场的小问题。
    - 用户贡献的 Overlay 数据回传给我们，成为 AI 优化下一个版本的原料。

---

## 6. 研发 SOP：AI 驱动的生产线
为了快速填充 Store，我们必须依赖 AI。

### 6.1 知识工程 (Knowledge Engineering)
+ **输入**：硬件厂商 PDF 手册、CLI 帮助文档。
+ **AI 任务**：
    1. 提取所有 CLI 命令及其参数。
    2. 提取错误码列表及其含义。
    3. 生成 JSON 格式的 `HardwareKnowledgeBase`。

### 6.2 代码生成 (Code Generation)
+ **工具**：Claude Code。
+ **输入**：`HardwareKnowledgeBase` + `Provider Interface Definition`。
+ **输出**：符合 CSPM 标准的 Go 代码。
    - 自动处理 JSON 解析。
    - 自动生成 `exec.Command` 调用。
    - 自动生成正则表达式解析 CLI 输出。

### 6.3 强制 Mock (Mandatory Mocking)
+ **规则**：任何提交到代码库的 Provider，必须包含一个 `mock_test.go`。
+ **内容**：包含厂商工具的模拟输出（Stdout Sample）。
+ **目的**：确保在没有真机的 CI/CD 环境下，逻辑测试覆盖率达到 100%。

---

## 7. 结语
CSPM 协议是 CloudBoot NG 生态系统的**法律**。

它通过**标准化**解决了技术的碎片化，通过**墨盒机制**解决了商业的变现难，通过**Overlay**解决了现场的灵活性。它是我们将物理世界“代码化”的编译器规范。

基于此规范，我们的 AI 工厂可以日夜不停地生产墨盒，构建起一道云新无法逾越的资源壁垒。

---

_(文档结束)_


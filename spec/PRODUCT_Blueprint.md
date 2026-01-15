# 


这是一份关于 CloudBoot NG 重启的顶层战略蓝图。它确立了我们将要把这艘船开向何方，以及我们凭什么能赢。

---

# CloudBoot NG 全景产品设计白皮书

**第一卷：战略顶层与产品矩阵 (The Grand Strategy)**

*   **项目代号**：Phoenix (涅槃)
*   **版本**：v1.0
*   **日期**：2026年1月
*   **密级**：核心战略 (Core Strategy)
*   **起草**：CloudBoot 联合创始人 & 首席架构师

---

## 1. 战略背景：基础设施的“熵增”与秩序重构

### 1.1 时代的断裂：2016 vs 2026
CloudBoot 的重启，是基于对 IT 基础设施历史周期的深刻洞察。我们正处在一个“旧秩序崩塌，新秩序尚未建立”的混沌时刻。

*   **2016（CloudBoot 1.0 时代）—— 标准化的黄昏**
    *   **特征**：x86 架构（Intel/AMD）一统天下，硬件高度标准化。
    *   **痛点**：**“规模化”**。互联网公司需要一天装机 5000 台。
    *   **解法**：PXE 自动化脚本。CloudBoot 1.0 凭借极简的体验赢得了互联网极客的口碑。
    *   **局限**：在企业级市场，由于硬件过于标准，管理员靠手工也能应付，自动化工具被视为“锦上添花”。

*   **2026（CloudBoot 3.0 时代）—— 碎片化的黎明**
    *   **特征**：**“双重熵增”**。
        1.  **信创浪潮**：华为鲲鹏、中科海光、飞腾、申威、龙芯，加上五花八门的国产 RAID/网卡，硬件兼容性从“标配”变成了“盲盒”。
        2.  **AI 算力**：大模型训练要求数千张 GPU 卡协同，对底层 BIOS 配置、PCIe 带宽、固件版本的一致性要求达到了原子级。
    *   **痛点**：**“确定性”与“合规性”**。管理员（老王）不仅要面对装不上的机器，还要面对审计合规的压力。
    *   **解法**：**“人肉”失效，“代码”上位**。必须引入 **IaC（基础设施即代码）** 和 **AI 知识工厂** 来对抗熵增。

### 1.2 竞争格局：对“云新模式”的降维打击
当前市场的隐形冠军“云新科技”，其护城河建立在**“人肉填坑”**之上。
*   **云新模式**：利用庞大的驻场团队，通过带外（OOB）手工修补信创硬件的缺陷。这是一种**“手工业”**模式，成本高、效率低、且存在巨大的**“黑盒安全隐患”**。
*   **CloudBoot 模式**：**“AI 工业化”**。我们利用 AI 消化海量硬件手册，生成标准化的驱动代码（Provider），并通过自动化工具（PE/Core）进行交付。
*   **核心打击点**：我们用硅基算力（无限供给、零边际成本）去打击云新的碳基算力（有限供给、高边际成本）。

---

## 2. 核心定义与价值主张 (Core Definitions)

我们不只是在做一个软件，我们在定义一套标准。

### 2.1 品牌定义
*   **主品牌**：**CloudBoot**。
    *   继承 10 年前的开源品牌资产，唤醒老用户的记忆与情怀。
*   **核心 Slogan**：**“以代码重构基础设施秩序” (Reconstruct Infrastructure Order with Code)**。

### 2.2 双重身份定位
1.  **物理机领域的 Terraform**：
    *   这是给**极客（Dev）** 看的。
    *   意味着：标准化、可编程、GitOps、声明式 API。我们填补了 Terraform 在物理机领域的全球空白。
2.  **信创数据中心的“数字签证官”**：
    *   这是给**企业（Ops/Manager）** 看的。
    *   意味着：**准入（Admission）**与**验收（Validation）**。所有硬件必须经过 CloudBoot 的体检和清洗，获得“签证”后方可入网。我们解决了“供应链黑盒”的焦虑。

---

## 3. 产品矩阵：六大模块 (Product Matrix)

为了实现“PLG（产品驱动增长）”与“农村包围城市”的战略，我们构建了一个分层、联动的产品矩阵。

### 3.1 流量入口：CloudBoot.co (官网 SaaS)
*   **定义**：全网流量枢纽与用户运营中心。
*   **核心功能**：
    *   **Public Store**：展示生态繁荣度，提供驱动查询。
    *   **Developer Hub**：文档中心、社区论坛、老用户认证通道（OG 计划）。
    *   **SaaS 工具链**：在线版的 OS Profile Designer（配置生成器）。
*   **战略价值**：连接用户与我们，连接社区与厂商。

### 3.2 尖刀产品：CloudBoot PE (原 hw-init)
*   **定义**：面向一线运维（Ops）的**单机版应急与探针系统**。
*   **交付形态**：**可引导 ISO 镜像 (<200MB)**。支持 U 盘或 BMC 虚拟光驱启动。
*   **核心场景**：
    *   **救火**：机器起不来？插上 U 盘，一键配置 RAID，一键修复引导。
    *   **探针**：不装系统，直接扫描硬件指纹，生成《硬件体检报告》。
*   **战略价值**：**特洛伊木马**。它免费、好用、无依赖，能迅速占领运维人员的工具箱，并为我们采集真实的硬件指纹数据。

### 3.3 核心平台：CloudBoot Core
*   **定义**：面向企业 IT 部门的**私有化 IaC 管理平台**。
*   **交付形态**：**Golang 单体二进制文件**。支持离线部署。
*   **核心功能**：
    *   **批量生命周期管理**：发现、配置、安装、回收。
    *   **混合调度引擎**：既支持标准化的 **Provider**，也兼容用户存量的 Shell/Python 脚本（Legacy Bridge）。
    *   **Private Store**：内网环境下的墨盒管理与分发中心。
*   **开源策略**：**有限开源**。二进制免费下载（社区版），源码向认证老用户开放，高级企业功能（审计/多租户）通过 License 解锁。

### 3.4 商业中心：CloudBoot Store
*   **定义**：连接“打印机（平台）”与“墨盒（驱动）”的分发引擎。
*   **核心逻辑**：
    *   展示所有支持的硬件型号。
    *   处理**订阅鉴权**与**离线包下载**。
    *   发布**“信创硬件红黑榜”**，确立话语权。

### 3.5 生态桥梁：CloudBoot DevKit
*   **定义**：面向开发者（Dev）的扩展工具包。
*   **核心功能**：
    *   **Provider 微调器**：允许用户通过 Overlay 机制修改官方驱动参数。
    *   **BootOS DIY**：允许用户注入私有驱动，重新打包 BootOS ISO。
*   **战略价值**：**激活“带路党”**。让技术骨干参与进来，成为我们的免费布道师。

### 3.6 核心资产：Provider & Adaptor (墨盒)
这是我们技术壁垒的实体化。
*   **Provider (上层/业务)**：面向“华为 TaiShan 200”这种机型。用户在 Store 里看到的是 Provider。
*   **Adaptor (下层/原子)**：面向“LSI 3108”这种芯片。封装在 Provider 内部，用户不可见。

---

## 4. 商业模式与闭环逻辑 (The Business Model)

我们构建的是一个**“流量置换 + 权益倒逼”**的生态闭环，彻底摒弃传统的软件销售模式。

### 4.1 To Ops (运维): 免费占领心智
*   **策略**：CloudBoot PE 和 Core 社区版**永久免费**。
*   **目的**：利用“极致好用”的工具属性，最大化铺开安装量（Install Base）。只要老王习惯了用 PE 救火，他就会推动公司部署 Core。

### 4.2 To Dev (开发者): 权益激活老兵
*   **策略**：通过官网认证“Cloudboot 老用户”身份，赠送 DevKit 高级权限和源码访问权。
*   **目的**：将 2016 年的品牌资产转化为 2026 年的**种子推手**。

### 4.3 To Vendors (厂商): 认证即免费 (Certification as a Service)
**这是最具攻击性的商业创新。**
*   **逻辑**：
    1.  用户（Dev/Ops）想要使用某款新机型的 Provider。
    2.  如果该机型未认证，用户需付费订阅 Enterprise 才能下载。
    3.  用户（作为甲方）倒逼硬件厂商：“你们去过一下 CloudBoot 认证吧，认证了我就能免费用。”
    4.  **厂商（华为/海光）为了卖货，向 CloudBoot 支付认证费。**
    5.  **结果**：厂商付费，用户免费。我们获得了收入，且丰富了生态。

### 4.4 To Business (企业): 订阅即保险 (Subscription as Insurance)
*   **策略**：针对银行、国企等对合规和离线有刚需的客户，提供 **Enterprise Subscription**。
*   **卖点**：
    *   **离线墨盒包**：解决内网隔离问题。
    *   **合规审计报表**：解决“免责”和“过审”问题。
    *   **兜底服务**：针对未认证的老旧机器提供官方适配。

---

## 5. 总结：必胜的路径

CloudBoot 3.0 的顶层设计，不再是单纯的“卖软件”，而是**“做局”**。

1.  我们用 **CloudBoot PE** 和 **Core** 作为诱饵，捕获海量的**用户（Ops/Dev）**。
2.  我们利用**用户的话语权**，去撬动**厂商（Vendors）**的资源和预算。
3.  我们利用**厂商的认证**和**企业的合规焦虑**，实现**多重变现（To V + To B）**。

在这个局里，云新（CloudSino）被完全隔离在外。他们没有工具触达用户，没有标准号令厂商，没有数据驱动智能。**这是一场维度的碾压。**

---
**（第一卷 完）**

这是一份关于 CloudBoot NG 大脑中枢的详细架构设计文档。

它定义了**CloudBoot Core** 如何作为企业内网的指挥官，管理资产、编排任务、分发墨盒，并提供极致的用户体验。这份文档将指导研发团队构建一个**“单体交付、逻辑严密、体验惊艳”**的现代化平台。

---

# CloudBoot NG 全景产品设计白皮书

**第二卷：CloudBoot Core 平台架构详设 (The Brain)**

*   **项目代号**：Phoenix Core
*   **版本**：v1.0
*   **日期**：2026年1月
*   **密级**：核心机密 (Core Confidential)
*   **关联文档**：第一卷 (战略)，第三卷 (BootOS)

---

## 1. 架构总述：单体与现代化的融合

### 1.1 核心定义
**CloudBoot Core** 是部署在客户私有环境（企业内网/数据中心）的管控平台。它是整个体系的**大脑**。
*   **对下**：指挥 BootOS 进行物理操作。
*   **对上**：提供 API 给 Terraform 和蓝鲸。
*   **对内**：作为“私有商店”管理和分发加密的 Provider 墨盒。

### 1.2 物理架构：极致的“绿色软件”
为了适应银行严苛的部署环境（无外网、严禁安装依赖、审批流程长），我们采用 **Go Single Binary (单体二进制)** 架构。

*   **交付物**：一个约 50MB 的可执行文件 `cloudboot-core`。
*   **零依赖**：
    *   **Web Server**：内置 Gin/Echo，不依赖 Nginx/Apache。
    *   **Database**：内置 SQLite (WAL模式) 或支持外接 MySQL，不强制依赖外部 DB。
    *   **Frontend**：HTML/CSS/JS 资源通过 `embed.FS` 编译进二进制，无静态文件目录。
    *   **Services**：内置 TFTP、DHCP、HTTP、Syslog 服务。
*   **部署口号**：**“上传，chmod +x，运行。10秒内上线。”**

### 1.3 技术栈：GOTH 栈 (Go + HTMX)
我们放弃沉重的 React/Vue 前后端分离架构，回归 Web 本质，追求极致的开发效率和页面响应速度。
*   **后端渲染 (SSR)**：Go `html/template` (或 Templ)。
*   **动态交互**：**HTMX**。实现“点击按钮 -> 后端渲染 HTML 片段 -> 前端局部替换”的 SPA 级体验。
*   **微交互**：**Alpine.js**。处理 Dropdown、Modal、Tabs 等纯前端逻辑。
*   **样式**：**Tailwind CSS**。实现“深色工业风”视觉体系。

---

## 2. Private Store (私有商店)：核心枢纽设计

Private Store 是连接 CloudBoot.co (公网 SaaS) 和企业内网的**气闸 (Airgap)**，也是商业模式落地的关键。

### 2.1 逻辑架构：双向阀门
Private Store 必须解决**“如何把公网的墨盒安全地运进内网”**的问题。

*   **进货端 (Ingress)**：
    *   **Mode A: 在线同步** (适合互联网客户)：配置官网 API Key，Core 自动拉取已购买的 Provider 元数据和二进制包。
    *   **Mode B: 离线导入** (银行刚需)：管理员在官网下载 `.cbp` (CloudBoot Package) 或 `offline-bundle.zip`，通过浏览器上传到 Core。
*   **出货端 (Egress)**：
    *   面向 BootOS 的高性能文件服务。支持断点续传。
    *   **动态 DRM**：文件在磁盘上是静态加密的。下载时，Core 根据 License 权限，动态生成 Session Key，将“墨盒 + 钥匙”一起发给 BootOS。

### 2.2 验钞机机制 (Verification Logic)
这是防止盗版和保障供应链安全的核心。导入 `.cbp` 包时，Core 执行以下检查：

1.  **完整性验签**：验证包内的 `signature.sig` 是否由 CloudBoot 官方私钥签署。防止墨盒被篡改（植入后门）。
2.  **水印审计**：读取包内的 `watermark.json`。
    *   如果水印归属（如“中国建设银行”）与当前 Core 的 License 归属（如“中国农业银行”）不一致。
    *   **策略**：**不阻断**（保障业务连续性），但**记录高危审计日志**，并在界面显著位置显示红色警告条：**“非授权来源组件运行中”**。这借甲方之手打击了云新的违规操作。

---

## 3. 核心业务逻辑 (The Core Logic)

### 3.1 动态资产库 (Live Asset DB)
传统的 CMDB 是静态的“Excel 表格”，CloudBoot Core 的资产库是**活的**。

*   **影子资产 (Shadow Assets)**：
    *   当一台未录入的机器 PXE 启动 BootOS 时，它会自动向 Core 注册。
    *   Core 将其标记为 **“发现 (Discovered)”** 状态，并在 Dashboard 上弹窗提示。
*   **指纹比对与防篡改**：
    *   **基线快照**：机器第一次入库时，记录全量硬件指纹（CPU Stepping、内存条码、RAID卡固件）。
    *   **每次装机前检查**：再次扫描指纹并比对。
    *   **告警**：如果发现内存少了，或者网卡被换了，系统**拒绝装机**并触发“资产完整性告警”。

### 3.2 任务编排引擎 (The Orchestrator)
这是将“意图”转化为“行动”的调度中心。

*   **Pipeline 设计**：一个标准的装机任务被拆解为原子步骤序列：
    1.  `WOL/IPMI Power On` (开机)
    2.  `BootOS Handshake` (握手)
    3.  `Inventory Check` (资产核验)
    4.  `CSPM: RAID Config` (配置存储)
    5.  `CSPM: BIOS Config` (配置固件)
    6.  `OS Provisioning` (分发镜像)
    7.  `Post-Install Scripts` (后置脚本)
*   **断点续传**：
    *   引擎记录每一步的 State。
    *   如果 RAID 配置成功但 OS 安装失败，管理员修复网络后点击“重试”，系统直接从 OS 安装步骤开始，**不会**重新破坏 RAID 数据。

---

## 4. OS Designer (配置设计器)：体验层核心

这是给“老王”用的杀手级工具，彻底终结手写 Kickstart 的痛苦。

### 4.1 数据模型
*   **Profile**：一个 JSON 结构体，描述了 OS 的所有配置（分区、网络、用户、包）。
*   **Template**：系统内置的“黄金模板”（CentOS 7/8, Ubuntu, SUSE, Kylin, UOS）。

### 4.2 可视化编辑器 (Visual Editor)
我们不提供文本框，我们提供**GUI 向导**。

*   **分区编辑器 (Partition Editor)**：
    *   **交互**：一个横向的条形图代表硬盘。用户点击“+”增加分区，拖拽滑块调整大小。
    *   **逻辑可视化**：如果选择了 LVM，界面上会出现嵌套框：`PV -> VG -> LV`。用户能直观看到逻辑卷是在卷组里面的。
    *   **实时校验**：如果用户忘记创建 `/boot` 分区，或者 swap 设置得太小，界面直接红框报错。
*   **网络配置器**：
    *   图形化配置 Bond 模式（Mode 0/1/4）。自动生成复杂的 `ifcfg` 文件或 Netplan 配置。

### 4.3 实时编译 (Live Compiler)
*   **双栏视图**：左侧是表单，右侧是代码预览。
*   **动作**：当用户在左侧修改了时区，右侧 Kickstart 代码里的 `timezone Asia/Shanghai` 立即高亮闪烁并更新。
*   **后端**：Go Template 引擎负责将 JSON 渲染为 `ks.cfg` 或 `autoinst.xml`。

---

## 5. Web UI 与交互体验设计

### 5.1 视觉风格：Dark Console
*   **色调**：Slate-900 背景，Emerald-500 高亮。避免大面积白色刺眼，适合运维中心大屏展示。
*   **字体**：**JetBrains Mono** 用于所有 ID、IP、日志和代码。强化“工具属性”。

### 5.2 杀手级交互：The Matrix View (全息日志)
当用户点击正在安装的机器时，不会只看到一个进度条。

*   **界面**：弹出一个仿终端的黑色窗口。
*   **技术**：**SSE (Server-Sent Events)**。BootOS 发出的日志流，经过 Core 转发，实时推送到浏览器。
*   **内容**：
    *   `[KERNEL] Linux version 5.10.0-60.18.0.50.oe2203...`
    *   `[AGENT] Downloading provider-huawei-raid... 100%`
    *   `[RAID] Creating Virtual Drive 0 (RAID 10)... Success`
*   **价值**：让过程**透明化**。老王可以看到机器在“思考”和“工作”，这比云新的黑盒进度条让人放心一万倍。

---

## 6. API 与扩展性设计

为了履行“物理机 Terraform”的承诺，Core 必须 API First。

### 6.1 Terraform 友好接口
*   **资源定义**：
    *   `POST /api/v1/servers` (创建/纳管机器)
    *   `POST /api/v1/servers/{id}/provision` (触发装机)
*   **状态查询**：
    *   `GET /api/v1/servers/{id}` 返回包含详细硬件指纹的 JSON。
    *   这允许 Terraform 的 `data source` 读取物理机状态。

### 6.2 Legacy Bridge (老脚本兼容)
*   **Hook 机制**：
    *   在流水线的任意节点（如 Pre-Install, Post-Install），允许插入 `Shell` 或 `Python` 脚本。
    *   **执行环境**：这些脚本会被下发到 BootOS 环境中执行。
*   **价值**：老用户不需要立刻重写所有逻辑，可以直接复用他们积累了十年的 `raid_config.sh`，平滑迁移。

---

## 7. 总结

CloudBoot Core 的设计，是在**“极简架构”**与**“极繁业务”**之间走钢丝。

*   通过 **Go 单体** 和 **SQLite**，我们做到了部署的极简。
*   通过 **CSPM** 和 **OS Designer**，我们做到了业务的深度覆盖。
*   通过 **Private Store** 和 **审计机制**，我们做到了商业的闭环。

这就是那个能让老王**“敢用、想用、离不开”**的大脑中枢。

---
*(文档结束)*


这是一份关于 CloudBoot NG 执行底座的**最高级别技术规范文档**。

它详细定义了 BootOS v4 如何作为一个**“无状态的数字雇佣兵”**，潜入物理机内部，执行高难度的硬件治理任务。同时，它深入解构了 **CB-Kit（底层工具链）** 的内部实现，这是我们区别于传统 Shell 脚本流派的核心技术壁垒。

---

# CloudBoot NG 全景产品设计白皮书

**第三卷：BootOS v4 技术架构与 CB-Kit 工具链详设 (The Body)**

*   **项目代号**：Phoenix Core (OS Layer)
*   **版本**：v1.0
*   **日期**：2026年1月
*   **密级**：核心机密 (Core Confidential)
*   **关联文档**：第一卷 (战略)，第二卷 (Core 平台)

---

## 1. 概述 (Overview)

### 1.1 设计愿景
**BootOS v4** 是 CloudBoot NG 的**数据面（Data Plane）**。它不是一个为了运行业务应用而设计的通用操作系统，而是一个**专用的、瞬时的、驻留内存的硬件治理微内核**。

它的核心使命是在**不依赖本地硬盘、不修改客户现有环境**的前提下，为上层 Provider（墨盒）提供一个统一、稳定、兼容所有信创硬件（x86/ARM/LoongArch）的**沙箱运行环境**。

### 1.2 核心设计哲学
1.  **Immutable (不可变)**：镜像构建完成后即只读。运行时没有任何“配置漂移”。
2.  **Stateless (无状态)**：**重启即焚**。所有运行时数据存储在 RAM 中，断电后彻底消失，满足银行最高级别的安全合规要求。
3.  **Micro (微型化)**：摒弃通用 Linux 发行版的臃肿，目标体积 **< 150MB**，启动时间 **< 15秒**。

---

## 2. BootOS v4 操作系统架构 (OS Architecture)

我们采用 **“三明治”分层架构**，利用 Docker 容器化流水线进行构建。

### 2.1 Layer 0: 基础底座 (The Foundation)
*   **选型决策**：**OpenEuler 22.03 LTS SP3** (及后续版本)。
*   **选型逻辑**：
    *   **信创原生支持**：作为国产根社区，OpenEuler 拥有最完整的华为鲲鹏 (ARM64)、中科海光 (x86_64)、飞腾、龙芯的原生驱动支持。这解决了传统 CentOS “找不到网卡/RAID卡”的致命痛点。
    *   **内核兼容性**：Kernel 5.10+，原生支持 NVMe-oF、最新的 RAID 控制器及国密算法。
*   **构建方式**：**From Scratch**。不使用官方 ISO 安装，而是通过 `dnf --installroot` 仅提取内核 (`vmlinuz`)、引导加载程序 (`grub2/shim`) 和最小根文件系统。

### 2.2 Layer 1: 用户空间 (The Minimalist Userland)
为了极致压缩体积，我们对用户空间进行了手术刀式的裁剪。

*   **保留组件**：
    *   **Init System**: `Systemd` (精简版)。保留 Systemd 是为了获得强大的服务依赖管理和日志接管能力（Journald）。
    *   **Network Stack**: `NetworkManager` + `iproute2`。支持复杂的企业级网络环境（LACP Bonding, VLAN Tagging, PPPoE）。
    *   **Hardware Tools**: `pciutils` (lspci), `dmidecode`, `ethtool`, `nvme-cli`, `lsscsi`。
    *   **Crypto**: `openssl` (支持国密 SM2/3/4)。
*   **剔除组件**：
    *   **Shell**: 仅保留 `bash` 用于紧急调试，但在生产模式下不提供 TTY 登录。
    *   **Interpreters**: **彻底移除 Python, Perl, Ruby**。所有上层逻辑由 Go 语言实现，消除解释器依赖带来的体积膨胀和漏洞风险。
    *   **SSH Server**: 默认移除。BootOS 仅通过反向 HTTP 长连接与 Server 通信，杜绝端口扫描风险。

### 2.3 Layer 2: 运行时空间 (The Runtime Sandbox)
这是 BootOS 唯一的“可写”区域，也是“墨盒”运行的场所。

*   **介质**：**Tmpfs (内存文件系统)**。
*   **挂载点**：`/opt/cloudboot/runtime`。
*   **目录结构**：
    ```text
    /opt/cloudboot/runtime/
    ├── bin/          # 存放下载并解密后的 Provider 二进制 (墨盒)
    ├── lib/          # 存放动态链接库 (如有)
    ├── configs/      # 存放渲染后的配置文件 (ks.cfg, overlay.json)
    ├── logs/         # 硬件工具的原始输出缓冲区
    └── pipe/         # 进程间通信管道
    ```

---

## 3. CB-Kit 底层工具集详设 (The Toolkit)

为了贯彻 **Unix 哲学** 和 **单体交付** 策略，我们将所有逻辑封装在一个静态编译的 Go 二进制文件 **`cb-tools`** 中。运行时通过软链接分身为四个微工具。

### 3.1 cb-agent (总控 / The Brain)
**角色**：BootOS 的 PID 1 守护进程。

*   **核心职责**：
    1.  **握手与保活**：启动时解析 Kernel Cmdline (`cb_server=...`)，建立 WebSocket/HTTP2 长连接。
    2.  **配置融合 (Config Injection)**：
        *   接收 Server 下发的 **Task Spec**（标准配置）。
        *   接收 **User Overlay**（用户微调）。
        *   **逻辑**：将两者 Deep Merge，生成最终的 `effective_config.json` 传递给执行器。
    3.  **看门狗**：监控子进程资源使用，防止 OOM。

### 3.2 cb-probe (探针 / The Eyes)
**角色**：硬件指纹采集器。

*   **核心职责**：
    1.  **深度扫描**：
        *   调用 `lspci -nnvv` 获取 PCI 树。
        *   调用 `dmidecode` 获取 SMBIOS。
        *   **透传扫描**：尝试通过 `smartctl` 穿透 RAID 卡读取物理磁盘 SN。
    2.  **数据清洗 (Normalization)**：
        *   将杂乱的厂商字符串（如 "Huawei Technologies Co., Ltd."）标准化为枚举值（"Huawei"）。
        *   输出标准 JSON 指纹数据，供 Server 端资产库比对。

### 3.3 cb-fetch (物流 / The Hands)
**角色**：安全下载与 DRM 执行器。

*   **核心职责**：
    1.  **流式下载**：支持断点续传，不将大文件读入内存变量，而是直接通过 `io.Copy` 写入 Tmpfs 文件句柄。
    2.  **动态解密 (DRM)**：
        *   Provider 在 Store 中是加密存储的。
        *   `cb-fetch` 握手获取 **Session Key**。
        *   边下载、边解密、边写入。**密钥只在内存中存在，文件只在 Tmpfs 中存在。**
    3.  **完整性校验**：计算 SHA256 签名，防止中间人篡改。

### 3.4 cb-exec (执行 / The Fist)
**角色**：沙箱运行器。

*   **核心职责**：
    1.  **环境隔离**：设置纯净的 `ENV`，切换工作目录。
    2.  **IO 捕获与流化**：
        *   启动 Provider 子进程。
        *   实时读取 `Stdout` 和 `Stderr` 管道。
        *   **日志结构化**：将原始文本封装为 JSON (`{ts, level, source, msg}`)。
        *   通过 HTTP SSE 通道实时推送到 Server，实现 Web 端“黑客帝国”式的滚动日志效果。
    3.  **资源限制**：利用 `cgroups` 限制 Provider 的 CPU 和 内存配额。

---

## 4. 关键技术突破 (Key Technical Breakthroughs)

### 4.1 ZRAM 内存倍增技术
针对老旧服务器内存小（<8GB）及大文件传输场景的防线。

*   **机制**：BootOS 启动时加载 `zram` 模块，将 50% 物理内存划分为压缩块设备并挂载为 Swap。
*   **算法**：使用 `zstd` (Zstandard)，提供极高的压缩比和极低的 CPU 开销。
*   **效果**：物理 8GB 内存可承载逻辑 12GB+ 数据，有效防止解压大型固件包时 OOM。

### 4.2 双模引导加载器 (Hybrid Bootloader)
为了制作一个“通吃”的 ISO 镜像。

*   **Legacy BIOS**: 使用 `Isolinux/Pxelinux`。
*   **UEFI**: 使用 `Grub2-EFI`。
    *   **安全启动**：集成 Microsoft 签名的 `Shim` 加载器，确保能通过银行服务器开启了 Secure Boot 的验证。

### 4.3 离线与在线的统一逻辑
*   `cb-fetch` 支持多协议：
    *   `http://` -> 从 CloudBoot Core 下载。
    *   `file://` -> 从本地挂载点（U盘 `CLOUDBOOT_DATA` 分区）读取。
*   这使得 **CloudBoot PE (单机版)** 和 **CloudBoot Core (网络版)** 共用同一套代码逻辑，仅配置不同。

---

## 5. 接口定义 (Interface Specifications)

### 5.1 硬件指纹 Schema (cb-probe Output)
```json
{
  "schema_version": "v1.0",
  "basic": {
    "manufacturer": "Hygon",
    "product_name": "K100",
    "serial_number": "SN-12345678",
    "uuid": "..."
  },
  "cpu": { "arch": "x86_64", "model": "Hygon C86 7285", "cores": 32 },
  "memory": { "total_gb": 256, "dimms": [...] },
  "storage": {
    "controllers": [
      { "pci_id": "1000:005f", "vendor": "LSI", "model": "MegaRAID 3108" }
    ]
  },
  "network": [...]
}
```

### 5.2 任务执行 Schema (cb-exec Input)
```json
{
  "task_id": "job-8848",
  "action": "apply",
  "provider_path": "/opt/cloudboot/runtime/bin/provider-raid",
  "config": {
    // 标准配置 + 用户 Overlay 合并后的结果
    "raid_level": "10",
    "drives": "all",
    "quirks": { "init_timeout": 600 }
  }
}
```

---

## 6. 构建流水线 (Build Pipeline)

BootOS v4 的构建实现了 **Infrastructure as Code**。

1.  **Compile Tools**: 静态编译 `cb-tools` (CGO_ENABLED=0)。
2.  **Build Rootfs**: 使用 Docker 运行 `dnf --installroot` 构建基础环境。
3.  **Generate Initrd**: 在容器内运行 `dracut`，注入 `cb-tools` 和 `systemd unit`。
4.  **Pack ISO**: 使用 `xorriso` 生成支持双模引导的 ISO。
5.  **Test**: 自动在 QEMU 中启动 ISO，验证 Agent 是否能向 Mock Server 发送心跳。

---

## 7. 结语

BootOS v4 是 CloudBoot NG 体系中最坚硬的**“钻头”**。

它通过 **OpenEuler** 解决了信创兼容的广度，通过 **微内核架构** 解决了部署的灵活性，通过 **CB-Kit 工具链** 解决了商业模式的闭环。它是我们“以代码定义基础设施”愿景在物理世界的真实投射。

---
*(文档结束)*



这是一份定义 CloudBoot NG **核心资产（墨盒）** 生产、封装、交互与保护标准的最终技术规范。

它解决了我们商业模式中最关键的问题：**如何确保我们的驱动代码（知识）既能被机器理解，又能被 AI 批量生产，同时还绝不可能被竞争对手（云新）低成本窃取。**

---

# CloudBoot NG 全景产品设计白皮书

**第四卷：CSPM 标准与墨盒机制设计 (The Protocol)**

*   **项目代号**：Phoenix Protocol
*   **版本**：v1.0 (Final Spec)
*   **状态**：**已定稿 (Approved)**
*   **日期**：2026年1月
*   **密级**：核心机密 (Core Confidential)
*   **关联文档**：第二卷 (Core 平台)，第三卷 (BootOS)

---

## 1. 概述 (Overview)

### 1.1 设计愿景
**CSPM (CloudBoot Server Provider Mechanism)** 是物理机管理领域的通用语言。它定义了上层意图（Intent）如何转化为底层指令（Instruction）的标准协议。

我们的目标是构建一个**“硬件乐高体系”**：将物理机的复杂性拆解为标准化的原子组件，通过 AI 工厂批量生产，通过加密墨盒进行分发，通过声明式接口进行调用。

### 1.2 核心原则
1.  **Decoupled (彻底解耦)**：业务逻辑（Provider）与底层工具（Adaptor）分离，配置数据（Config）与执行代码（Binary）分离。
2.  **Secure (原生安全)**：墨盒默认为加密状态，仅在受信任的内存环境中解密运行，离线可审计。
3.  **AI-First (AI 友好)**：协议设计完全基于 JSON Schema 和 标准 IO 流，极度利于 Claude Code 进行代码生成和逻辑验证。

---

## 2. 双层驱动架构 (The Bi-Layer Architecture)

为了应对信创硬件的“排列组合爆炸”，我们将驱动架构拆分为两层。

### 2.1 Layer 1: Provider (业务编排层)
这是**面向用户**的产品单元，也是我们 Store 中的售卖单元（SKU）。

*   **定义**：针对特定**“厂商+机型”**的逻辑封装。
*   **命名规范**：`provider-<vendor>-<model>` (例: `provider-huawei-taishan200`)。
*   **核心职责**：**“翻译与编排”**。
    *   它知道“华为 TaiShan 200”服务器由“LSI RAID卡”和“华为 BIOS”组成。
    *   它负责将用户的通用意图（“配置 RAID 10”）翻译成对底层 Adaptor 的具体调用参数。
    *   它处理机型特有的坑（Quirks），例如：“这款机器重启后需要等待 60秒才能识别网卡”。

### 2.2 Layer 2: Adaptor (原子执行层)
这是**面向芯片**的技术单元，也是我们技术壁垒的基石。

*   **定义**：针对底层芯片或管理协议的指令执行器。
*   **命名规范**：`adaptor-<category>-<chipset>` (例: `adaptor-raid-lsi3108`, `adaptor-bios-ami-aptio`).
*   **核心职责**：**“执行与屏蔽”**。
    *   它封装了厂商原始的二进制工具（如 `storcli`, `sas3ircu`, `ipmitool`）。
    *   它负责解析厂商工具那令人崩溃的非标文本输出，将其转化为标准 JSON。
*   **封装形态**：Adaptor 通常作为静态资源被编译进 Provider 二进制中，**对用户不可见**。

---

## 3. 交互协议设计 (The Interaction Protocol)

Provider 是一个独立的二进制程序，它与调用者（Agent）之间通过 **JSON over Stdin/Stdout** 进行通信。

### 3.1 核心动词 (Verbs)
Provider 必须支持以下命令行子命令：

1.  **`probe`** (探测)
    *   **输入**：无。
    *   **逻辑**：扫描当前硬件环境，确认自己是否适配该机器。
    *   **输出**：返回硬件指纹和组件状态。
2.  **`plan`** (预演)
    *   **输入**：`DesiredState` (期望配置) + `CurrentState` (当前状态)。
    *   **逻辑**：计算 Diff。例如：当前是 RAID 5，期望是 RAID 10 -> 需要执行 `delete` 然后 `create`。
    *   **输出**：`ExecutionPlan` (操作序列，用于向用户展示风险)。
3.  **`apply`** (执行)
    *   **输入**：`DesiredState`。
    *   **逻辑**：真正调用 Adaptor 修改硬件。
    *   **输出**：执行结果报告。

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

```text
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

1.  **加密**：官方 Store 使用**全局产品主钥 (Master Key)** 对 Provider 二进制进行加密。
2.  **授权**：用户购买企业版后，获得一个 `.lic` 文件。该文件包含用户的**解密私钥**（该私钥是针对用户生成的），以及授权范围。
3.  **运行**：
    *   CloudBoot Core 导入 `.cbp` 包。
    *   CloudBoot Core 使用 `.lic` 中的私钥，解密出 Provider。
    *   **关键点**：Core **立即**使用一个随机生成的**临时会话密钥 (Session Key)** 重新加密 Provider，然后再发给 BootOS。
    *   **结果**：即使有人在网络层截获了发给 BootOS 的包，没有内存中的 Session Key 也无法解密。

### 4.3 审计与追责
*   **被动触发**：如果 CloudBoot Core 检测到 `.cbp` 包中的水印 ID 与当前系统的 License ID 不一致。
*   **反应**：
    *   并不阻断运行（为了业务稳定性）。
    *   但在界面显眼位置显示红色横幅：**“非授权来源组件运行中：[来源用户ID]”**。
    *   写入不可删除的审计日志。
*   **杀伤力**：这使得云新的驻场人员不敢随意拷贝他们在其他项目下载的墨盒，否则一旦被客户发现，就是重大的合规事故。

---

## 5. 配置透明化与微调机制 (Transparency & Overlays)

我们不仅提供黑盒的驱动，还提供白盒的控制权。

### 5.1 Provider Schema (说明书)
每个 Provider 包内包含一个 `schema.json`，描述它支持的所有参数。
*   **用途**：
    *   CloudBoot Core 根据它自动生成 Web 表单。
    *   OS Designer 根据它进行参数校验。
    *   AI 根据它生成推荐配置。

### 5.2 User Overlay (用户补丁)
为了解决“同一型号不同批次硬件差异”的问题，我们引入 Overlay 机制。

*   **机制**：用户可以在 CloudBoot Core 界面上，针对特定机型创建一个 **"Config Patch"**。
*   **逻辑**：
    *   Standard Config (官方标准) + User Overlay (用户微调) = Effective Config (最终生效配置)。
*   **示例**：
    *   官方默认 `timeout = 30s`。
    *   用户 Overlay 设置 `timeout = 120s`。
    *   Provider 运行时读取最终的 120s。
*   **价值**：
    *   用户不需要等我们发版就能解决现场的小问题。
    *   用户贡献的 Overlay 数据回传给我们，成为 AI 优化下一个版本的原料。

---

## 6. 研发 SOP：AI 驱动的生产线

为了快速填充 Store，我们必须依赖 AI。

### 6.1 知识工程 (Knowledge Engineering)
*   **输入**：硬件厂商 PDF 手册、CLI 帮助文档。
*   **AI 任务**：
    1.  提取所有 CLI 命令及其参数。
    2.  提取错误码列表及其含义。
    3.  生成 JSON 格式的 `HardwareKnowledgeBase`。

### 6.2 代码生成 (Code Generation)
*   **工具**：Claude Code。
*   **输入**：`HardwareKnowledgeBase` + `Provider Interface Definition`。
*   **输出**：符合 CSPM 标准的 Go 代码。
    *   自动处理 JSON 解析。
    *   自动生成 `exec.Command` 调用。
    *   自动生成正则表达式解析 CLI 输出。

### 6.3 强制 Mock (Mandatory Mocking)
*   **规则**：任何提交到代码库的 Provider，必须包含一个 `mock_test.go`。
*   **内容**：包含厂商工具的模拟输出（Stdout Sample）。
*   **目的**：确保在没有真机的 CI/CD 环境下，逻辑测试覆盖率达到 100%。

---

## 7. 结语

CSPM 协议是 CloudBoot NG 生态系统的**法律**。

它通过**标准化**解决了技术的碎片化，通过**墨盒机制**解决了商业的变现难，通过**Overlay**解决了现场的灵活性。它是我们将物理世界“代码化”的编译器规范。

基于此规范，我们的 AI 工厂可以日夜不停地生产墨盒，构建起一道云新无法逾越的资源壁垒。

---
*(文档结束)*



这是一份指导研发团队如何将逻辑落地为**“像素级完美界面”**和**“工业级代码”**的执行手册。

它确立了我们独特的技术栈选择（GOTH Stack）和革命性的研发模式（AI-Native），确保我们能以**“1 人 + AI”**的配置，在**30 天**内交付出媲美百人团队的成熟产品。

---

# CloudBoot NG 全景产品设计白皮书

**第五卷：UI/UX 规范与 AI 原生研发 SOP (The Execution)**

*   **项目代号**：Phoenix Execution
*   **版本**：v1.0 (Final Standard)
*   **状态**：**已定稿 (Approved)**
*   **日期**：2026年1月
*   **密级**：核心机密 (Core Confidential)
*   **关联文档**：第一至四卷

---

## 1. 概述 (Overview)

### 1.1 设计哲学：极简与极客的统一
CloudBoot NG 的前端工程不追求大厂流行的“重型前后端分离（React/Vue + API）”，而是回归 Web 的本质，追求**“单体交付、服务端渲染、超低延迟”**。

我们致力于打造一种**“深色工业风 (Dark Industrial)”**的视觉体验，让用户在使用时感受到如同操作精密仪器般的**确定性**与**掌控感**。

### 1.2 研发哲学：代码是耗材，Spec 是资产
在 Claude Code + Opus 4.5 的赋能下，我们不再手写每一行代码。我们通过编写高质量的**规格说明书 (Spec)** 和 **上下文文档 (Context)**，指挥 AI 军团完成代码的生成、测试与重构。

---

## 2. GOTH 技术栈规范 (The GOTH Stack)

为了实现**“单个二进制文件交付 (Single Binary)”**且不牺牲现代交互体验，我们选定以下技术栈：

*   **Go (Golang)**: 后端逻辑、HTTP 服务、模板渲染。
*   **O** (Operating System agnostic): 跨平台编译，无依赖。
*   **Templ / HTML**: 类型安全的 Go 模板语言，生成 HTML。
*   **HTMX**: 负责宏观交互（页面切换、表单提交、局部刷新）。
*   **Tailwind CSS**: 负责原子化样式。
*   **Alpine.js**: 负责微观交互（Dropdown、Modal、前端状态逻辑）。

### 2.1 架构优势
1.  **部署零摩擦**：没有 `node_modules`，没有 Nginx 配置，没有 CORS 问题。客户拿到一个 Binary，运行即上线。
2.  **开发极速**：HTMX 让后端工程师直接控制前端行为，无需编写大量 JSON API 胶水代码。
3.  **AI 友好**：Go 的强类型结构体与 HTML 模板在同一代码库中，Claude Code 可以一次性生成前后端逻辑，上下文无断层。

---

## 3. UI/UX 视觉设计规范 (Visual Design System)

本规范需同步应用于 **CloudBoot Core (产品)** 和 **CloudBoot.co (官网)**，保持品牌一致性。

### 3.1 色彩体系：Dark Mode Only
*   **背景 (Canvas)**: `Slate-950` (#020617) —— 深邃的蓝黑，非纯黑。
*   **表面 (Surface)**: `Slate-900` —— 卡片与侧边栏背景。
*   **边框 (Border)**: `Slate-800` —— 极其微弱的边界感。
*   **主色 (Primary)**: `Emerald-500` (#10b981) —— 代表“通过、在线、健康”。
    *   *用法*：主按钮、进度条、成功状态徽章。
*   **强调色 (Accent)**: `Violet-500` (#8b5cf6) —— 代表“AI、智能”。
    *   *用法*：AI 建议、自动生成内容的标记。
*   **警告/错误**: `Amber-500` / `Rose-500`。

### 3.2 排版与字体
*   **界面字体**: `Inter` 或 `System UI`。紧凑、现代。
*   **数据字体**: **`JetBrains Mono`**。
    *   *强制规则*：所有的 IP 地址、MAC 地址、版本号、日志内容、代码块，**必须**使用 JetBrains Mono。这是营造“极客感”的关键。

### 3.3 核心组件交互规范
1.  **卡片 (Cards)**:
    *   使用“玻璃拟态 (Glassmorphism)”：背景半透明，带有背景模糊 (Backdrop Blur)。
    *   Hover 效果：边框颜色微亮，产生轻微上浮感。
2.  **日志终端 (The Terminal)**:
    *   纯黑背景，绿色/白色文字。
    *   必须支持 **Auto-Scroll** (有新日志时自动滚动到底部)。
3.  **按钮 (Buttons)**:
    *   Primary: 翡翠绿背景，白色文字，微弱内阴影（模拟实体按键）。
    *   Secondary: 透明背景，Slate 边框。

---

## 4. AI 原生研发 SOP (The AI-Native Workflow)

这是 1 人抵 100 人的执行秘籍。我们通过**“三文档定律”**来约束和指挥 Claude Code。

### 4.1 上下文锚点 (Context Anchors)
在项目根目录下，必须始终维护三个核心 Markdown 文件，作为 AI 的“长期记忆”：

1.  **`ARCH.md` (架构宪法)**
    *   定义目录结构 (`cmd/`, `internal/`, `web/`)。
    *   定义技术栈选型（强制 Go+HTMX，禁止引入 React/Vue）。
    *   定义错误处理与日志规范。
2.  **`API.md` (数据契约)**
    *   定义核心 Struct：`Machine`, `Task`, `Provider`, `License`。
    *   定义 CSPM 协议的 JSON Schema。
    *   **原则**：AI 写代码前，必须先更新此文档。
3.  **`UI_KIT.md` (视觉标准)**
    *   包含 Tailwind 的 config 配置。
    *   包含核心组件（Card, Button, Badge）的 HTML/CSS 模板代码片段。
    *   **原则**：AI 生成新页面时，必须复用此文档中的组件代码。

### 4.2 开发循环 (The Loop)

**Step 1: 定义 (Define)**
*   人类动作：编写或更新 `SPEC` 文档，描述本轮迭代的需求（例如：“实现 OS Profile Designer 的分区编辑器”）。

**Step 2: 生成与测试 (Generate & Test)**
*   指令：*"Claude, read `ARCH.md` and `UI_KIT.md`. Implement the Partition Editor based on the new SPEC. Write unit tests first (TDD)."*
*   AI 动作：
    1.  生成 Go Struct。
    2.  编写 `_test.go`。
    3.  编写 HTML 模板（使用 Alpine.js 处理拖拽逻辑）。
    4.  运行测试 -> 失败 -> 修正 -> 成功。

**Step 3: 视觉验收 (Visual Audit)**
*   人类动作：运行程序，打开浏览器。
*   反馈：*"分区进度条颜色不对，应该用 Emerald-500。另外，拖拽时的动画有点卡顿，请优化 Alpine.js 逻辑。" (附上截图或描述)*

**Step 4: 固化 (Commit)**
*   指令：*"Update `UI_KIT.md` with the new Partition Component. Commit changes."*

---

## 5. 关键功能实现指引 (Key Implementation Guides)

### 5.1 实时日志流 (The Matrix View)
如何在 GOTH 栈中实现酷炫的实时日志？
*   **后端**：使用 Go 的 `net/http` 支持 **Server-Sent Events (SSE)**。
    *   创建一个 `Broker` 结构，接收 Agent 发来的日志，广播给所有连接的 HTTP 客户端。
*   **前端**：使用 HTMX 的 SSE 扩展。
    ```html
    <div hx-ext="sse" sse-connect="/api/logs/stream?job_id=123">
      <div sse-swap="message" hx-swap="beforeend">
        <!-- 新日志将自动追加到这里 -->
      </div>
    </div>
    ```
*   **效果**：零 JS 代码，实现高性能实时日志滚动。

### 5.2 动态表单 (OS Designer)
如何在没有 React 的情况下实现复杂的动态表单？
*   **后端**：Go 渲染初始 HTML。
*   **前端**：使用 **Alpine.js** 维护本地状态（State）。
    ```html
    <div x-data="{ partitions: [...] }">
      <template x-for="(p, index) in partitions">
        <!-- 渲染分区行 -->
      </template>
      <button @click="partitions.push({mountPoint: '/new'})">Add</button>
      <!-- 隐藏域，用于提交给后端 -->
      <input type="hidden" name="partitions_json" :value="JSON.stringify(partitions)">
    </div>
    ```

### 5.3 离线包导入 (Private Store)
*   **后端**：实现分块上传（Chunked Upload）接口，避免大文件（4GB+）占用过多内存。
*   **流式处理**：上传过程中，一边接收数据流，一边计算 SHA256 校验和，一边写入磁盘，**不读入 RAM**。

---

## 6. 构建与发布规范 (Build & Release)

### 6.1 嵌入式构建
使用 Go 1.16+ 的 `embed` 功能。
```go
//go:embed web/static web/templates
var content embed.FS
```
这确保了 `cloudboot-core` 是真正的单体文件，不依赖任何外部路径。

### 6.2 编译指令 (Makefile)
```makefile
build-prod:
    # 1. 编译 Tailwind (压缩)
    npm run build:css
    # 2. 注入版本信息，去除符号表(减小体积)，启用 CGO (SQLite)
    go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/cloudboot-core cmd/server/main.go
```

---

## 7. 结语

本卷文档完成了 CloudBoot NG 从**“战略构想”**到**“代码落地”**的最后一次翻译。

通过 **GOTH 技术栈**，我们规避了繁琐的前端工程化陷阱；
通过 **AI 原生 SOP**，我们将编码成本压缩到了极致。

现在，我们拥有了：
1.  清晰的**战略蓝图** (卷一)。
2.  严密的**平台逻辑** (卷二)。
3.  坚实的**底层底座** (卷三)。
4.  标准的**驱动协议** (卷四)。
5.  高效的**执行手册** (卷五)。

**所有图纸已绘就，只待开工。**
**Let's code the infrastructure order.**

---
*(全套白皮书结束)*
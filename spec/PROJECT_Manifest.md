# PROJECT_Manifest.md

## 1. Project Identity (产品定义)
- **Name**: CloudBoot NG (Next Generation)
- **Core Definition**: The "Terraform" for Bare Metal & The "Digital Visa Officer" for Infrastructure.
- **Mission**: Reconstruct infrastructure order with code. Replace manual, black-box operations ("CloudSino Mode") with automated, transparent, code-defined workflows.
- **Business Model**: 
  - **The Printer (Free)**: The Open Source Engine (CloudBoot Core + BootOS).
  - **The Ink (Paid)**: The Certified Hardware Providers (Registry/Store).

## 2. Supreme Principles (最高原则)
All code and architecture decisions MUST adhere to these principles. NO EXCEPTIONS.

### 2.1 Single Binary (单体交付)
- The entire CloudBoot Core platform MUST compile into a **single, static Go binary**.
- **FORBIDDEN**: External runtime dependencies (Node.js, Python, Java, Nginx, Systemd services).
- **REQUIRED**: Use `//go:embed` for all static assets (HTML/CSS/SQL).

### 2.2 Zero Dependency (零依赖)
- The binary must run on any modern Linux system (x86/ARM) by simply executing `chmod +x` and running.
- **Database**: Use embedded SQLite (with WAL mode). No external MySQL/PostgreSQL required for the base version.

### 2.3 Stateless Execution (无状态运行)
- **BootOS**: Runs entirely in RAM (Tmpfs). Reboots to a clean slate.
- **Agent**: Does not persist data locally on the managed node. All states are pushed to the Core.

### 2.4 Decoupled Architecture (软硬解耦)
- **CSPM**: The Core knows NOTHING about hardware specifics. It delegates all physical operations to external **Providers** via the CSPM protocol.
- **Core** handles Intent (What). **Provider** handles Instruction (How).

## 3. AI Agent Rules of Engagement (AI 行为准则)
When generating code, you (Claude) must follow these rules:

1.  **TDD (Test-Driven Development)**: 
    - You MUST write the test case (`_test.go`) *before* or *simultaneously* with the implementation.
    - Tests must cover logic, not just syntax.
2.  **Mock Everything**: 
    - Since we don't have physical servers in the dev environment, you MUST create **Mocks** for every hardware interaction (RAID, IPMI, PXE).
3.  **No Hallucinations**: 
    - Do not invent non-existent Go libraries. Use standard library or widely used, stable packages (e.g., `gin`, `gorm`, `cobra`).
4.  **Code Style**: 
    - Strict Go idioms.
    - Error handling must be explicit (wrap errors, don't just return them).
    - Comments are required for all exported functions.
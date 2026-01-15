# ARCH_Stack.md

## 1. Technology Stack (The GOTH Stack)
We use the **GOTH** stack to achieve modern UI with backend simplicity.

| Component | Choice | Justification |
| :--- | :--- | :--- |
| **Language** | **Go (Golang) 1.22+** | Static typing, concurrency, single binary build. |
| **Web Framework** | **Echo (v4)** | Fast, minimalist, robust middleware support. |
| **Database** | **SQLite3** | Embedded via `mattn/go-sqlite3`. Enable CGO. |
| **ORM** | **Gorm** | Rapid development, easy schema migration. |
| **Templating** | **html/template** | Standard Go templates. Simple and effective for SSR. |
| **Styling** | **Tailwind CSS** | Utility-first. Generated via CLI during build, embedded in binary. |
| **Macro-Interaction** | **HTMX** | Server-driven UI interactions (AJAX replacement). |
| **Micro-Interaction** | **Alpine.js** | Client-side reactivity (Dropdowns, Modals) without build steps. |

## 2. Directory Structure (Standard Layout)
Follow the strict Go project layout:
 
cloudboot-ng/
├── cmd/
│   ├── server/           # Main entry point for CloudBoot Core
│   ├── agent/            # Source for the BootOS Agent
│   ├── provider-mock/    # The Mock Hardware Provider (for testing)
│   └── tools/            # Helper utilities (e.g., config generator)
├── internal/
│   ├── core/             # Core business logic (Service Layer)
│   │   ├── machine/      # Machine lifecycle management
│   │   ├── job/          # Task scheduler & orchestration
│   │   └── cspm/         # Plugin loading & execution engine
│   ├── models/           # Gorm data models
│   ├── api/              # HTTP Handlers (HTMX & JSON API)
│   └── pkg/              # Shared utilities (Logger, Crypto)
├── web/
│   ├── static/           # Raw assets (JS libs, Images)
│   │   ├── css/          # Tailwind output
│   │   └── js/           # htmx.min.js, alpine.min.js
│   └── templates/        # HTML templates
│       ├── components/   # Reusable UI parts (Cards, Buttons)
│       ├── layouts/      # Base layouts
│       └── views/        # Page specific templates
├── scripts/              # Build scripts (Makefile, Dockerfile)
├── go.mod
└── README.md


## 3. Communication Protocols (通信协议)
### 3.1 HTTP / HTMX
+ **SSR**: Handlers return `c.Render(...)` for full pages.
+ **HTMX**: Handlers return HTML fragments (partials) for dynamic updates.
+ **API**: `/api/v1/...` returns JSON for external integrations (Terraform).

### 3.2 Real-time Logs (SSE)
+ Use **Server-Sent Events (SSE)** for streaming installation logs from Core to Browser.
+ Endpoint: `/stream/logs?job_id=xyz`
+ Format: `data: {"ts": "...", "level": "INFO", "msg": "..."}\n\n`

### 3.3 CSPM (Provider Protocol)
+ **Mechanism**: JSON over Stdin/Stdout.
+ **Security**: Providers are encrypted at rest. Decrypted to `tmpfs` only during execution.

## 4. Build & CI Specifications
### 4.1 Makefile Targets
+ `make dev`: Run Tailwind watch + Air (live reload) for Go.
+ `make build`: Build the production binary.
    - MUST run `tailwind build --minify` first.
    - MUST use `-ldflags="-s -w"` to strip symbols.
+ `make test`: Run unit tests (with Mock Providers).

### 4.2 Embedding Strategy
+ The `web/` directory MUST be embedded using `//go:embed`.
+ The application MUST verify asset integrity on startup.

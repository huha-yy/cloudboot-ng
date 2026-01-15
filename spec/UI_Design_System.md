# UI_Design_System.md

## 1. Design Philosophy (设计哲学)
- **Theme**: **"Dark Industrial"**. Imagine a cockpit of a sci-fi spaceship or a high-end server room at night.
- **Keywords**: Precision, Depth, Glow, Monospace.
- **Goal**: Make "Old Wang" feel safe (it looks professional) and "Geeks" feel cool (it looks like a hacker tool).

## 2. Color Palette (Tailwind Config)
All UI elements must strictly use these colors. Do not introduce random hex codes.

| Semantic | Tailwind Class | Hex | Usage |
| :--- | :--- | :--- | :--- |
| **Canvas** | `bg-slate-950` | `#020617` | Global background. The "Void". |
| **Surface** | `bg-slate-900` | `#0f172a` | Cards, Sidebar, Modals. |
| **Border** | `border-slate-800` | `#1e293b` | Subtle separation lines. |
| **Primary** | `emerald-500` | `#10b981` | Primary actions, "Success", "Online", Active states. |
| **Glow** | `shadow-emerald-500/20` | - | Subtle glow for primary elements. |
| **Accent** | `violet-500` | `#8b5cf6` | AI features, "Magic" buttons. |
| **Destructive** | `rose-500` | `#f43f5e` | Delete, Stop, Error. |
| **Text Main** | `text-slate-200` | `#e2e8f0` | High readability content. |
| **Text Muted** | `text-slate-400` | `#94a3b8` | Labels, meta info. |

## 3. Typography (字体策略)
- **UI Font**: `font-sans` (Inter or System UI). For headings, labels, buttons.
- **Data Font**: `font-mono` (**JetBrains Mono**).
  - **MANDATORY**: MUST be used for IDs (UUID), IP Addresses, MAC, Logs, Configuration Code, and Version Numbers.
  - *Reasoning*: Monospace fonts imply "Engineering Precision".

## 4. Atomic Components (组件代码库)
Claude MUST use these snippets when building pages.

### 4.1 The "Glass" Card (容器)
A card with subtle transparency and border.

<div class="bg-slate-900/50 backdrop-blur-sm border border-slate-800 rounded-lg p-6 shadow-sm">
    <h3 class="text-lg font-medium text-white mb-2">{{ .Title }}</h3>
    <div class="text-slate-400">
        {{ .Content }}
    </div>
</div>

### 4.2 Primary Button (行动点)
Tactile, with a subtle inner shadow and glow.


<button class="inline-flex items-center justify-center px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium rounded-md shadow-lg shadow-emerald-900/20 transition-all duration-200 active:translate-y-[1px]">
    <!-- Icon (Optional) -->
    <svg class="w-4 h-4 mr-2" ...></svg>
    {{ .Label }}
</button>



### 4.3 The "Matrix" Terminal (日志窗口)
The signature component for displaying logs.


<div class="w-full bg-black rounded-lg border border-slate-800 overflow-hidden font-mono text-sm shadow-inner">
    <!-- Header -->
    <div class="flex items-center px-4 py-2 bg-slate-900 border-b border-slate-800">
        <div class="flex space-x-2">
            <div class="w-3 h-3 rounded-full bg-rose-500/20"></div>
            <div class="w-3 h-3 rounded-full bg-amber-500/20"></div>
            <div class="w-3 h-3 rounded-full bg-emerald-500"></div> <!-- Active Green Dot -->
        </div>
        <div class="ml-4 text-xs text-slate-500">root@cloudboot-core: ~</div>
    </div>
    <!-- Log Stream Area (HTMX SSE Target) -->
    <div id="log-container" class="p-4 h-64 overflow-y-auto scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent">
        <div class="text-emerald-500/90">> Initializing hardware probe...</div>
        <div class="text-slate-300">> Found RAID Controller: LSI 3108</div>
        <!-- New logs append here -->
    </div>
</div>



### 4.4 Status Badge (状态指示器)
Pulsing dot for "Live" status.


<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
    <span class="relative flex h-2 w-2 mr-2">
      <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
      <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
    </span>
    Online
</span>



### 4.5 Form Input (表单)
Deep focus ring.


<div>
    <label class="block text-sm font-medium text-slate-400 mb-1">{{ .Label }}</label>
    <input type="text" class="block w-full bg-slate-950 border border-slate-700 rounded-md py-2 px-3 text-slate-200 placeholder-slate-600 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 sm:text-sm transition-colors" placeholder="...">
</div>



## 5. Interaction Patterns (交互模式)
### 5.1 HTMX Patterns (Macro)
+ **Lazy Loading**: Use `hx-trigger="load"` to fetch heavy data (like Asset Lists) after the page renders.
+ **Active Search**: Use `hx-trigger="keyup changed delay:500ms"` on search inputs to filter tables.
+ **Dialogs**: Use `hx-target="#modal-container"` to swap in modal HTML from the server.

### 5.2 Alpine.js Patterns (Micro)
+ **Toggle**: `<div x-data="{ open: false }">` for dropdowns.
+ **Tabs**: `<div x-data="{ tab: 'overview' }">` for switching content without server trips.
+ **Flash Messages**: `<div x-data="{ show: true }" x-init="setTimeout(() => show = false, 3000)">` for toast notifications.

## 6. Layout Structure (布局规范)
+ **Sidebar**: Fixed width (64px collapsed / 240px expanded). Darker than main content (`bg-slate-950`).
+ **Topbar**: Sticky, Glassmorphism (`backdrop-blur-md`).
+ **Main Content**: `max-w-7xl mx-auto p-6`.

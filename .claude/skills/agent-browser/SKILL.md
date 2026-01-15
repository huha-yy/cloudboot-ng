---
name: agent-browser
description: Browser automation CLI for AI agents using Playwright. Use when Claude needs to: (1) Navigate and interact with web pages (click, fill forms, type), (2) Take screenshots of websites, (3) Extract page content or accessibility trees, (4) Automate web workflows, (5) Scrape web data, (6) Test web applications. Triggers on requests like "open this URL", "click the button", "fill the form", "take a screenshot", "get page content".
---

# Agent Browser

## Overview

Agent-browser is a headless browser automation CLI designed for AI agents. It provides deterministic element selection via reference IDs (`@e1`, `@e2`) from accessibility tree snapshots, enabling reliable web automation.

## Prerequisites

```bash
# Install globally
npm install -g agent-browser

# Download Chromium (required)
agent-browser install

# Linux: install system dependencies
agent-browser install --with-deps
```

## AI Workflow (Recommended Pattern)

1. **Navigate**: `agent-browser open <url>`
2. **Snapshot**: `agent-browser snapshot -i --json` to get interactive elements with refs
3. **Act**: Use refs for actions (e.g., `agent-browser click @e2`)
4. **Re-snapshot**: After page changes, snapshot again to get updated refs

## Quick Reference

### Navigation & Interaction

```bash
# Navigate
agent-browser open "https://example.com"

# Click element
agent-browser click @e2              # By ref (preferred)
agent-browser click "#submit-btn"    # By CSS selector

# Fill form field
agent-browser fill @e5 "user@example.com"

# Type text (character by character)
agent-browser type @e5 "search query"

# Press keyboard key
agent-browser press Enter
agent-browser press "Control+a"

# Screenshot
agent-browser screenshot              # To stdout (base64)
agent-browser screenshot ./page.png   # To file

# Close browser
agent-browser close
```

### Information Retrieval

```bash
# Get accessibility tree with element refs
agent-browser snapshot                # Full tree
agent-browser snapshot -i             # Interactive elements only
agent-browser snapshot -i --json      # JSON format (best for parsing)
agent-browser snapshot -d 3           # Limit depth to 3

# Get element data
agent-browser get text @e2            # Text content
agent-browser get html @e2            # Inner HTML
agent-browser get value @e2           # Input value
agent-browser get attr @e2 href       # Attribute value

# Get page info
agent-browser get title
agent-browser get url
```

### State Checking

```bash
agent-browser is visible @e2
agent-browser is enabled @e2
agent-browser is checked @e2
```

### Waiting

```bash
agent-browser wait @e2                # Wait for element
agent-browser wait 2000               # Wait 2 seconds
agent-browser wait "networkidle"      # Wait for network idle
```

## Selectors

| Type | Syntax | Example |
|------|--------|---------|
| Ref (preferred) | `@e<n>` | `@e1`, `@e2` |
| CSS | Standard CSS | `#id`, `.class`, `div > button` |
| Text | `text=<value>` | `text=Submit` |
| Role | `role <type>` | `role button` |
| Label | `label <text>` | `label Email` |
| Placeholder | `placeholder <text>` | `placeholder Search` |
| XPath | `xpath=<path>` | `xpath=//button` |

**Always prefer refs (`@e1`)** - they are deterministic and come from snapshots.

## Global Flags

```bash
--session <name>     # Isolated browser session
--headed             # Show browser window (for debugging)
--json               # Machine-readable output
--debug              # Enable debug output
```

## Detailed Reference

For complete command documentation, see [references/commands.md](references/commands.md).

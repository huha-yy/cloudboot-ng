# Agent Browser Command Reference

## Table of Contents

- [Navigation Commands](#navigation-commands)
- [Interaction Commands](#interaction-commands)
- [Information Commands](#information-commands)
- [State Commands](#state-commands)
- [Advanced Commands](#advanced-commands)
- [Session Management](#session-management)

---

## Navigation Commands

### open

Navigate to a URL.

```bash
agent-browser open <url>
agent-browser open "https://example.com"
agent-browser open "https://example.com" --headed  # Show browser
```

### close

Close the browser instance.

```bash
agent-browser close
agent-browser close --session mySession  # Close specific session
```

---

## Interaction Commands

### click

Click an element.

```bash
agent-browser click <selector>
agent-browser click @e2                  # By ref (preferred)
agent-browser click "#submit"            # By CSS
agent-browser click "text=Login"         # By text
```

### fill

Fill a form field (clears existing content first).

```bash
agent-browser fill <selector> <text>
agent-browser fill @e5 "user@example.com"
agent-browser fill "#email" "test@test.com"
```

### type

Type text character by character (does not clear existing content).

```bash
agent-browser type <selector> <text>
agent-browser type @e5 "search query"
```

### press

Press a keyboard key or key combination.

```bash
agent-browser press <key>
agent-browser press Enter
agent-browser press Tab
agent-browser press "Control+a"
agent-browser press "Shift+ArrowDown"
```

### select

Select option(s) from a dropdown.

```bash
agent-browser select <selector> <value>
agent-browser select @e3 "option1"
agent-browser select "#country" "US"
```

### hover

Hover over an element.

```bash
agent-browser hover <selector>
agent-browser hover @e2
```

### scroll

Scroll the page or element.

```bash
agent-browser scroll down
agent-browser scroll up
agent-browser scroll @e2  # Scroll element into view
```

---

## Information Commands

### snapshot

Get accessibility tree with element references. **This is the most important command for AI workflows.**

```bash
agent-browser snapshot                    # Full tree
agent-browser snapshot -i                 # Interactive elements only
agent-browser snapshot -i --json          # JSON format (best for parsing)
agent-browser snapshot -c                 # Compact (remove empty elements)
agent-browser snapshot -d 3               # Limit depth to 3
agent-browser snapshot -s "#main"         # Scope to CSS selector
```

**Flags:**
- `-i, --interactive` - Show only interactive elements
- `-c, --compact` - Remove empty elements
- `-d, --depth <n>` - Limit tree depth
- `-s, --selector <sel>` - Scope to CSS selector
- `--json` - Output as JSON

### get

Retrieve element or page data.

```bash
# Element data
agent-browser get text @e2               # Text content
agent-browser get html @e2               # Inner HTML
agent-browser get value @e2              # Input value
agent-browser get attr @e2 href          # Attribute value
agent-browser get attr @e2 class         # Class attribute

# Page data
agent-browser get title                  # Page title
agent-browser get url                    # Current URL
```

### screenshot

Capture page screenshot.

```bash
agent-browser screenshot                 # Output base64 to stdout
agent-browser screenshot ./page.png      # Save to file
agent-browser screenshot --full-page     # Full page screenshot
```

---

## State Commands

### is

Check element state. Returns boolean.

```bash
agent-browser is visible @e2
agent-browser is enabled @e2
agent-browser is checked @e2
agent-browser is editable @e2
```

### wait

Wait for conditions.

```bash
agent-browser wait @e2                   # Wait for element
agent-browser wait 2000                  # Wait 2 seconds
agent-browser wait "networkidle"         # Wait for network idle
agent-browser wait "load"                # Wait for page load
agent-browser wait "domcontentloaded"    # Wait for DOM ready
```

---

## Advanced Commands

### find

Find elements by semantic locators.

```bash
agent-browser find role button click     # Find button and click
agent-browser find text "Submit" click   # Find by text and click
agent-browser find label "Email" fill "test@test.com"
```

### cookies

Manage browser cookies.

```bash
agent-browser cookies                    # List all cookies
agent-browser cookies --json             # JSON format
agent-browser cookies set <name> <value>
agent-browser cookies delete <name>
```

### storage

Access browser storage.

```bash
agent-browser storage local              # List localStorage
agent-browser storage session            # List sessionStorage
agent-browser storage local get <key>
agent-browser storage local set <key> <value>
```

### network

Intercept and mock network requests.

```bash
agent-browser network route "**/*.png" block    # Block images
agent-browser network route "**/api/*" mock '{"data": []}'
```

### tab / window

Manage browser tabs and windows.

```bash
agent-browser tab list                   # List open tabs
agent-browser tab new "https://example.com"
agent-browser tab switch <id>
agent-browser tab close <id>

agent-browser window list
agent-browser window new
```

---

## Session Management

Sessions provide isolated browser instances with separate cookies, storage, and history.

```bash
# Use named session
agent-browser open "https://example.com" --session user1
agent-browser click @e2 --session user1

# Multiple sessions
agent-browser open "https://example.com" --session admin
agent-browser open "https://example.com" --session guest

# Close specific session
agent-browser close --session user1
```

---

## Global Flags Reference

| Flag | Description |
|------|-------------|
| `--session <name>` | Use isolated browser session |
| `--headed` | Show browser window |
| `--json` | Machine-readable JSON output |
| `--debug` | Enable debug output |
| `--headers <json>` | HTTP headers (scoped to origin) |
| `--executable-path <path>` | Custom browser binary |
| `--cdp <port>` | Connect via Chrome DevTools Protocol |

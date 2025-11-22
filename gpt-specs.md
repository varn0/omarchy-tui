# Omarchy TUI – Technical Specification (Proof of Concept)

## 1. Overview

This document describes the technical specification for **Omarchy TUI**, a minimal terminal-based user interface written in **Go**, using the `tview` library.

The goal is to build a simple TUI that loads application categories and commands from a YAML configuration file and allows users to navigate and launch programs using only keyboard input.

This is a **proof of concept**, so the scope is intentionally limited.

---

## 2. Technology Stack

- **Language:** Go
- **UI Library:** `tview` (github.com/rivo/tview)
- **Configuration Format:** YAML
- **Configuration Location:**  
  `~/.config/omarchy.conf.yaml`
- **Persistence:** No database (no SQLite)

---

## 3. YAML Configuration Format

### 3.1 File Location

```
~/.config/omarchy.conf.yaml
```

### 3.2 Structure

The configuration file defines:

- A list of application **categories**
- A list of **applications**

### 3.3 Categories in PoC

This PoC will include exactly **three categories**:

- `editor`
- `browser`
- `terminal`

### 3.4 Example Configuration

```yaml
categories:
  - id: editor
    name: Text Editor

  - id: browser
    name: Web Browser

  - id: terminal
    name: Terminal Applications

apps_inventory:
  - name: NeoVim
    package_name: nvim
    keybinding: "Ctrl Shift Alt V"
    category: editor
    config_file: ~/.config/nvim/init.lua

  - name: Nano
    package_name: nano
    keybinding: "Ctrl Shift Alt N"
    category: editor

  - name: Firefox
    package_name: firefox
    keybinding: "Ctrl Shift Alt F"
    category: browser

  - name: Brave
    package_name: brave
    keybinding: default
    category: browser

  - name: Alacritty
    package_name: alacritty
    keybinding: default
    category: terminal

  - name: Kitty
    package_name: kitty
    keybinding: "Ctrl Shift Alt K"
    category: terminal
```

---

## 4. Domain Model

### 4.1 Entities

#### `Category`
- `id` (string)
- `name` (string)

#### `Application`
- `name` (string)
- `package_name` (string)
- `keybinding` (string)
- `category` (string, references `Category.id`)
- `config_file` (optional)
- `custom_config` (optional map)

#### `OmarchyConfig`
Root object containing:

- `categories []Category`
- `apps_inventory []Application`

---

## 5. Application Features

### 5.1 Minimum Features

1. Load configuration from YAML file.
2. Render a full-screen TUI using `tview`.
3. Display:
   - category list
   - app list
   - information/configuration panel
4. Keyboard-only navigation.
5. Selecting a category shows its apps.
6. Selecting an app allows:
   - launching it
   - marking it as default
   - configuring it (editing text)
7. Selecting a category allows:
   - choosing a default app

---

## 6. UI Layout

### 6.1 Overview

The UI uses a three-panel layout:

```
+------------------+-----------------------------+
| Categories list  |  App list for selected cat  |
| (left panel)     |  (right panel)              |
+------------------------------------------------+
| Bottom Information / Configuration Panel       |
+------------------------------------------------+
```

### 6.2 Panels

#### **Left Panel – Categories**
- Shows the category list.
- Arrow keys navigate up/down.
- Pressing Enter:
  - Moves focus to the app list pane.
- Bottom panel displays:
  - Category details:
    - Default app
    - Other apps in same category
  - Or a configuration UI when in “Set default app” mode

#### **Right Panel – Apps**
- Shows apps filtered by the selected category.
- Arrow keys navigate up/down.
- Pressing Enter opens a small action panel:
  - Launch app  
  - Mark as default for this category  
  - Configure (edit configuration file or custom values)

Bottom panel displays:
- App details:
  - Package name
  - Config file
  - Whether it is the default app for its category  
- Or a configuration text area when editing.

#### **Bottom Panel – Contextual Info / Config Editor**
Always visible.

**Modes:**

1. **Information Mode**
   - Shows contextual info about:
     - The currently highlighted category  
     - The currently highlighted app

2. **Configuration Mode**
   - Becomes a writable text area
   - Allows modifying:
     - category default app
     - custom config for an app

### 6.3 Input Behavior

- Navigation:
  | Key   | Action                   |
  |-------|--------------------------|
  | ↑ / ↓ | Move selection           |
  | ← / → | Switch panes             |
  | Enter | Select / confirm         |
  | Esc   | Cancel action and return |
  | q     | Quit                     |

- For categories:
  - Enter opens option to select a default app

- For apps:
  - Enter opens options:
    - Launch"
    - Configure
    - Set as default app for its category

### 6.4 Panel Implementation Notes

- Implemented with `tview.Flex`
- Bottom panel implemented as:
  - A `tview.TextView` for info mode
  - A `tview.InputField` or `tview.TextArea` (if using extended widgets) in configuration mode

---

## 7. Go Project Structure

```
omarchy-tui/
├─ cmd/
│  └─ omarchy/
│     └─ main.go
├─ internal/
│  ├─ config/
│  │  ├─ loader.go
│  │  └─ model.go
│  ├─ tui/
│  │  ├─ app.go
│  │  ├─ categories_view.go
│  │  ├─ apps_view.go
│  │  ├─ bottom_panel.go
│  │  └─ controller.go
│  └─ exec/
│     └─ runner.go
└─ go.mod
```

---

## 8. Responsibilities

### 8.1 `config/loader.go`
- Loads `~/.config/omarchy.conf.yaml`
- Parses YAML into structs
- Validates required structure

### 8.2 `tui/app.go`
- Creates the root `tview.Application`
- Composes the three panels
- Handles global exit events

### 8.3 `tui/categories_view.go`
- Renders category list
- Handles selection changes
- Sends selection events to controller

### 8.4 `tui/apps_view.go`
- Renders apps for active category
- Triggers app actions

### 8.5 `tui/bottom_panel.go`
- Displays contextual information
- Switches to editor mode when needed

### 8.6 `exec/runner.go`
- Launches external system program using `os/exec`

---

## 9. Keyboard Navigation

| Key   | Action                     |
|-------|----------------------------|
| ↑ / ↓ | Move cursor                |
| ← / → | Move focus between panels  |
| Enter | Confirm / open action menu |
| Esc   | Cancel edits or action     |
| q     | Quit application           |

---

## 10. Error Handling

Minimal but user-friendly:

- Missing config → show dialog and exit
- Invalid YAML → show error and exit
- Launch failure → show modal with system error

---

## 11. Future Enhancements

(Not in PoC)

- Search
- Command palette
- Pane resizing
- Theme management
- SQLite runtime analytics
- Plugin system

---

## 12. Deliverables

- Builds and runs in Go
- Loads YAML config
- TUI navigation fully keyboard-driven
- 3-panel layout working
- Launch and configure apps through UI


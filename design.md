# Linux Launcher (Raycast Alternative) - Design Doc

## 1. Overview
- **Goal:** Build a Raycast-style command launcher for Linux environments.
- **Target Platform:** Linux.
- **Language:** Go (Golang).

## 2. Technology Stack & Architecture
- **GUI Framework:** Gio UI (`gioui.org`) — chosen for immediate-mode GPU rendering, instant startup, low RAM/CPU footprint, and zero web-engine overhead.
- **Application Indexing:** Standard Go library (`os`, `path/filepath`) scanning `.desktop` files in `/usr/share/applications` and `~/.local/share/applications` to extract `Name`, `Exec`, and `Icon` metadata.

## 3. Core Window & Interaction Behavior
- **Window Style:** Floating, borderless/frameless command palette window centered on screen.
- **Visibility Lifecycle:** Appears and steals focus immediately upon global hotkey trigger; hides on blur (`Deactivated` event).
- **Search & Filtering:** Text input using Gio `widget.Editor` to filter indexed applications as the user types.
- **Keyboard Navigation:** Handle `key.Event` for `Up`/`Down` arrow keys to navigate results and `Return` to launch the selected application (via `exec.Command`).

## 4. Open Decisions (To Be Settled)
- **Global Hotkey Implementation:** Exact mechanism/library for registering global shortcuts across X11 and Wayland.
- **System Tray Integration:** Whether to include a system tray icon and which library to use.

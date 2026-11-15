# quicklaunch

Raycast-style command launcher for Linux. Targets **KDE Plasma on Wayland** first, built with [Gio UI](https://gioui.org) in Go.

- Global hotkey (default `Ctrl+Shift+Space`) summons a centered, frameless command palette
- Fuzzy search over installed applications (`.desktop` files, Flatpak included)
- App icons via the XDG icon theme (PNG + SVG)
- Hide on focus loss, tray icon, single-instance toggle

Status: early development — see [Todo.md](Todo.md) for the granular milestone plan and [design.md](design.md) for decisions.

## Build

Requires Go and the X11/Wayland/xkbcommon dev packages.

```sh
go build -o quicklaunch .
```

## License

To be decided.

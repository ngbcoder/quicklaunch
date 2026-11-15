# quicklaunch — Todo

Raycast-style command launcher for Linux (KDE Plasma / Wayland first). Design decisions live in `design.md`; working rules below.

## Working rules
- One commit **and push** per checkbox item; the `Todo.md` check-off is included in that item's commit.
- Ask, don't assume — open questions get asked before implementation.
- Granular milestones; never batch unrelated work into one commit.

## Legend
- `[ ]` not started · `[~]` in progress · `[x]` done · `[!]` blocked (note why)

---

## M0 — Repo & tooling
- [x] Create Todo.md milestone checklist (this file)
- [x] Add .gitignore and README.md
- [x] Init git repo (`main` branch)
- [x] Install `github-cli` (needs `sudo pacman -S github-cli`)
- [x] `gh auth login` (user, interactive)
- [x] Set global git identity from GitHub account name + noreply email
- [x] Create public GitHub repo (`ngbcoder/quicklaunch`) and push
- [x] Add minimal `main` package so CI has a build target
- [ ] Add GitHub Actions CI (vet, test, build with X11/Wayland dev packages)

## M1 — Hotkey spike (de-risk, highest-risk item first)
- [ ] Research KGlobalAccel D-Bus API; write `docs/kglobalaccel-notes.md`
- [ ] Spike: register Ctrl+Shift+Space via D-Bus from Go, verify invocation signal fires
- [ ] Record decision: raw KGlobalAccel vs. KDE custom shortcut + `--toggle` fallback

## M2 — Desktop entry parser
- [ ] Define `Entry` struct + grammar types
- [ ] Parse `[Desktop Entry]` sections/keys with tests + fixtures
- [ ] Localized keys (`Name[xx]`, best-effort locale) with tests
- [ ] Exec field-code tokenizer (`%f %F %u %U %i %c %k %%`) with tests
- [ ] Boolean/enum keys (`NoDisplay`, `Hidden`, `Terminal`, `Type`) with tests
- [ ] `OnlyShowIn`/`NotShowIn`/`TryExec` handling with tests
- [ ] `[Desktop Action *]` sub-entries with tests

## M3 — App index
- [ ] Scanner for `/usr/share/applications` + `~/.local/share/applications` (fixtures + tests)
- [ ] Flatpak export dirs (`~/.local/share/flatpak/exports/...`, `/var/lib/flatpak/exports/...`)
- [ ] Precedence/dedupe by desktop-file-id (user dir wins)
- [ ] Filter `NoDisplay`/`Hidden`/`Type != Application`
- [ ] Index store with thread-safe access + reload

## M4 — Launcher
- [ ] Detached `exec.Command` (setsid, working dir `$HOME`) with tests
- [ ] Field-code substitution incl. icon/name args with tests
- [ ] `Terminal=true` → `konsole -e` with tests
- [ ] Launch failures surfaced (log + notification path)

## M5 — Config
- [ ] XDG path helpers + defaults with tests
- [ ] JSON load/save, validation, forward-compatible with tests
- [ ] Hotkey string parsing (config ↔ KGlobalAccel key format) with tests

## M6 — Logging & app skeleton
- [ ] `slog` to `~/.local/state/quicklaunch/quicklaunch.log`
- [ ] `main` with flags: run (default), `--toggle`, `--install`, `--uninstall`
- [ ] Wire flags to no-op stubs; verify log file creation

## M7 — Icons
- [ ] XDG icon-theme index parsing + theme inheritance with tests
- [ ] Lookup across themes/sizes/extensions (svg, png; skip xpm) with tests
- [ ] SVG rasterization (oksvg + rasterx) with tests
- [ ] Image cache + async decode with tests
- [ ] Absolute-path `Icon=` and generic fallback with tests

## M8 — Gio UI shell
- [ ] Frameless window, size, theme/colors
- [ ] Query editor + result list layout
- [ ] Up/Down navigation, Enter launch, Esc hide
- [ ] Focus-loss hide
- [ ] Manual QA on KDE Wayland: show, type, launch, hide-on-blur

## M9 — Fuzzy search
- [ ] Scorer: exact / prefix / substring / subsequence with tests
- [ ] Match highlight ranges with tests
- [ ] Integrate: empty query shows all, ranked results in UI

## M10 — Single instance
- [ ] Socket lock in `XDG_RUNTIME_DIR` + stale-lock cleanup with tests
- [ ] `--toggle` IPC to running instance with tests
- [ ] Wire toggle → show/hide

## M11 — Tray
- [ ] `fyne.io/systray` icon + menu (Show, Quit)
- [ ] Menu actions routed to app via channels
- [ ] Tray icon asset

## M12 — Hotkey (full implementation)
- [ ] KGlobalAccel component/action registration + invocation handling
- [ ] Re-register on config hotkey change; unregister on exit
- [ ] Fallback path: KDE custom shortcut → `--toggle` (document in README)

## M13 — Window placement & lifecycle
- [ ] Test Gio `Center`/`Raise` under KWin Wayland; record results
- [ ] Fullscreen transparent overlay fallback if actions ignored
- [ ] Focus-stealing prevention workaround (activation token) if needed
- [ ] Manual QA: hotkey → instant focused show; blur → hide; no stale windows

## M14 — Packaging
- [ ] `--install`: binary → `~/.local/bin`, launcher `.desktop`, XDG autostart entry
- [ ] `--uninstall` reverses all of the above
- [ ] README: build, install, hotkey setup, config reference

## M15 — Hardening & release
- [ ] `go vet` + `-race` test run clean
- [ ] Full manual QA checklist pass
- [ ] Tag v0.1.0 release build

# KDE Plasma global shortcuts over D-Bus (KGlobalAccel)

Notes gathered while prototyping the quicklaunch hotkey on KDE Plasma 6.8.2,
Wayland session. Everything below was verified against a live session; the
KGlobalAccel API is not publicly documented, so this file is the reference.

## The pieces

| Thing | Value |
|---|---|
| Bus service | `org.kde.kglobalaccel` (on Plasma 6 the daemon lives in `kwin_wayland`) |
| Control object | `/kglobalaccel`, interface `org.kde.KGlobalAccel` |
| Component object | `/component/<component_unique>` (dots and dashes become `_`), interface `org.kde.kglobalaccel.Component` |

The daemon hosts the per-component objects itself; a client does **not** export
any D-Bus object. Clients subscribe to the component object's signals.

Inspect them at runtime with:

```sh
busctl --user tree org.kde.kglobalaccel
busctl --user introspect org.kde.kglobalaccel /kglobalaccel
```

## Registering a shortcut

Four calls, in this order:

1. `doRegister(as actionId)` — registers the action. `actionId` is a list of
   four strings: `[componentUnique, actionUnique, componentFriendly, actionFriendly]`.
2. `setShortcut(as actionId, ai keys, u flags)` — assigns the key.
3. `getComponent(s componentUnique) -> o path` — returns the component object
   path to listen on.
4. Subscribe to signals on that path.

The shortcut is written to `~/.config/kglobalshortcutsrc` and shows up in
**System Settings → Shortcuts**, where the user can rebind it.

### The `SetPresent` flag is mandatory

`flags` is a bit field (see `KGlobalAccelD::SetShortcutFlag`):

| Flag | Value | Meaning |
|---|---|---|
| `SetPresent` | 2 | Mark the shortcut as live. Without it the daemon stores the key but never grabs it. |
| `NoAutoloading` | 4 | Do not write the shortcut back to the user configuration. |
| `IsDefault` | 8 | Store as the action's default binding instead of the active one. |

`flags = 0` registers the key and persists it, but the daemon does **not** grab
it, the component reports `isActive() == false`, and no activation is ever
delivered. This was the first bug hit.

## Key encoding

Shortcuts are Qt key combination values — the same integers as
`QKeyCombination::toCombined()`, with KDE's modifier names:

| Modifier | Value |
|---|---|
| Shift | `0x02000000` |
| Ctrl | `0x04000000` |
| Alt | `0x08000000` |
| Meta | `0x10000000` |

| Shortcut | Value |
|---|---|
| `Ctrl+Shift+Space` | `0x06000020` |
| `Meta+Space` | `0x10000020` |
| `Ctrl+Shift+S` | `0x06000053` |

Special keys live at `0x01000000` + offset (`Key_Left` = `0x01000012`,
`Key_Right` = `0x01000014`, `Key_Escape` = `0x01000000`, `F1` = `0x01000030`, …).

Verified against real shortcuts on this machine: KWin's "Switch One Desktop to
the Right" is stored as `0x15000014` (`Meta+Ctrl+Right`) and "Window Operations
Menu" as `0x19000032` (`Alt+Meta+F3`).

## Activation signals

The daemon emits on the component object path:

| Signal | Body | Notes |
|---|---|---|
| `pressed` | `s actionId`, `a{sv} options` | Modern style. `options` can carry an `xdgActivationToken`. |
| `globalShortcutPressed` | `s componentUnique`, `s actionUnique`, `x timestamp` | Legacy style. |
| `globalShortcutRepeated` | same as above | Key repeat. |
| `globalShortcutReleased` | same as above | Key release. |

**Plasma 6.8 emits the legacy `globalShortcutPressed` only** — the modern
`pressed` signal, which carries the xdg activation token, was not observed.
Subscribe to both and prefer `pressed` when it arrives.

Without an activation token a Wayland client cannot ask the compositor for
focus, which matters for showing the palette window (see the window placement
milestone).

## godbus v5 gotchas

* `Conn.Call` takes the arguments only — the signature is derived from them.
  Passing a `dbus.Signature` as an argument sends it as a body value and the
  daemon rejects the call.
* `dbus.SignatureOf` cannot express nested arrays: `[][]int32` becomes `aai`,
  but the daemon's v2 API expects `a(ai)`. Use the v1 `setShortcut` method,
  which takes a flat `ai`.
* `Signal.Name` is `"<interface>.<member>"`, not the bare member name. Strip
  the prefix before comparing.
* `Store()` needs a target for methods that return a value, otherwise it
  fails with "length mismatch".
* The daemon's `Component.invokeShortcut` does not emit an activation, so it
  is useless for testing the callback path — a real key press is needed.

## Verifying by hand

```sh
# is the component live?
busctl --user call org.kde.kglobalaccel /component/<component> \
    org.kde.kglobalaccel.Component isActive

# which component owns a key?
busctl --user call org.kde.kglobalaccel /kglobalaccel \
    org.kde.KGlobalAccel getGlobalShortcutsByKey i <key>

# watch the raw traffic
dbus-monitor --session "type='signal',path='/component/<component>'"
```

The `cmd/kga-spike` tool in this repository performs the same registration and
prints activations; it was used to develop this file.

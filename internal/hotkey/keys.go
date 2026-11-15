// Package hotkey implements global shortcut registration for KDE Plasma
// using the org.kde.KGlobalAccel D-Bus service.
//
// Shortcuts are encoded the way KDE encodes them: as Qt key combination
// values, identical to QKeyCombination::toCombined() (for example
// "Ctrl+Shift+Space" is 0x06000020).
package hotkey

import (
	"fmt"
	"strconv"
	"strings"
)

// Qt key combination modifier flags, as used by QKeyCombination.
const (
	ModShift   int32 = 0x02000000
	ModControl int32 = 0x04000000
	ModAlt     int32 = 0x08000000
	ModMeta    int32 = 0x10000000
)

// Special key base. Qt encodes keys that have no ASCII representation
// (arrows, function keys, Escape, ...) at 0x01000000 plus an offset.
const specialKeyBase int32 = 0x01000000

// keyValues maps the lower case key names used by KDE's textual shortcut
// format ("Ctrl+Shift+Space") to their Qt key values.
var keyValues = map[string]int32{
	"space":      0x20,
	"escape":     specialKeyBase,
	"tab":        specialKeyBase + 1,
	"backtab":    specialKeyBase + 2,
	"backspace":  specialKeyBase + 3,
	"return":     specialKeyBase + 4,
	"enter":      specialKeyBase + 5,
	"insert":     specialKeyBase + 6,
	"delete":     specialKeyBase + 7,
	"pause":      specialKeyBase + 8,
	"print":      specialKeyBase + 9,
	"home":       specialKeyBase + 0x10,
	"end":        specialKeyBase + 0x11,
	"left":       specialKeyBase + 0x12,
	"up":         specialKeyBase + 0x13,
	"right":      specialKeyBase + 0x14,
	"down":       specialKeyBase + 0x15,
	"pageup":     specialKeyBase + 0x16,
	"pagedown":   specialKeyBase + 0x17,
	"altgr":      specialKeyBase + 0x24,
	"capslock":   specialKeyBase + 0x25,
	"numlock":    specialKeyBase + 0x26,
	"scrolllock": specialKeyBase + 0x27,
}

// keyNames maps special key values to their canonical display names.
var keyNames = map[int32]string{
	0x20:                  "Space",
	specialKeyBase:        "Escape",
	specialKeyBase + 1:    "Tab",
	specialKeyBase + 2:    "Backtab",
	specialKeyBase + 3:    "Backspace",
	specialKeyBase + 4:    "Return",
	specialKeyBase + 5:    "Enter",
	specialKeyBase + 6:    "Insert",
	specialKeyBase + 7:    "Delete",
	specialKeyBase + 8:    "Pause",
	specialKeyBase + 9:    "Print",
	specialKeyBase + 0x10: "Home",
	specialKeyBase + 0x11: "End",
	specialKeyBase + 0x12: "Left",
	specialKeyBase + 0x13: "Up",
	specialKeyBase + 0x14: "Right",
	specialKeyBase + 0x15: "Down",
	specialKeyBase + 0x16: "PageUp",
	specialKeyBase + 0x17: "PageDown",
	specialKeyBase + 0x24: "AltGr",
	specialKeyBase + 0x25: "CapsLock",
	specialKeyBase + 0x26: "NumLock",
	specialKeyBase + 0x27: "ScrollLock",
}

// namedModifiers maps modifier names to their combination flags.
var namedModifiers = map[string]int32{
	"shift":   ModShift,
	"ctrl":    ModControl,
	"control": ModControl,
	"alt":     ModAlt,
	"meta":    ModMeta,
	"super":   ModMeta,
}

// plusPlaceholder stands in for a literal plus key while the shortcut
// string is split on "+".
const plusPlaceholder = "\x00plus\x00"

// ParseSequence converts a KDE-style textual shortcut such as
// "Ctrl+Shift+Space" into the Qt key combination value used by
// KGlobalAccel. An empty string yields 0, which means "no shortcut".
func ParseSequence(seq string) (int32, error) {
	if strings.TrimSpace(seq) == "" {
		return 0, fmt.Errorf("hotkey: empty shortcut")
	}

	// "++" is the literal plus key, so it cannot survive a plain split.
	seq = strings.ReplaceAll(seq, "++", "+"+plusPlaceholder)

	var mods int32
	var key int32
	var haveKey bool

	for _, raw := range strings.Split(seq, "+") {
		field := strings.TrimSpace(raw)
		if field == "" {
			continue
		}
		lower := strings.ToLower(field)

		if mod, ok := namedModifiers[lower]; ok {
			mods |= mod
			continue
		}

		if haveKey {
			return 0, fmt.Errorf("hotkey: %q contains more than one key", seq)
		}

		val, err := keyValue(lower)
		if err != nil {
			return 0, fmt.Errorf("hotkey: %w", err)
		}
		key = val
		haveKey = true
	}

	if !haveKey {
		return 0, fmt.Errorf("hotkey: %q has no key", seq)
	}
	return mods | key, nil
}

// keyValue resolves a single (non-modifier) key name to its Qt key value.
//
// Qt normalises letters to upper case, because that is the character the
// key produces on a US layout.
func keyValue(name string) (int32, error) {
	if name == plusPlaceholder {
		return '+', nil
	}

	if v, ok := keyValues[name]; ok {
		return v, nil
	}

	// Function keys: f1 through f35.
	if len(name) >= 2 && name[0] == 'f' {
		if n, err := strconv.Atoi(name[1:]); err == nil {
			if n >= 1 && n <= 35 {
				return specialKeyBase + 0x30 + int32(n) - 1, nil
			}
		}
	}

	// Single printable character.
	if runes := []rune(name); len(runes) == 1 {
		r := runes[0]
		if r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		return int32(r), nil
	}

	return 0, fmt.Errorf("unknown key %q", name)
}

// SequenceString renders a key combination back into KDE's textual format.
// Modifier order follows Qt: Meta, Ctrl, Shift, Alt.
func SequenceString(value int32) string {
	parts := make([]string, 0, 5)
	if value&ModMeta != 0 {
		parts = append(parts, "Meta")
	}
	if value&ModControl != 0 {
		parts = append(parts, "Ctrl")
	}
	if value&ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if value&ModAlt != 0 {
		parts = append(parts, "Alt")
	}

	key := value & 0x01ffffff
	switch {
	case keyNames[key] != "":
		parts = append(parts, keyNames[key])
	case key >= 0x21 && key <= 0x7e:
		parts = append(parts, string(rune(key)))
	default:
		parts = append(parts, strconv.Itoa(int(key)))
	}
	return strings.Join(parts, "+")
}

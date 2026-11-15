package hotkey

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

// D-Bus names used by the KDE global accelerator service.
const (
	// Service is the well-known name of the global accelerator daemon.
	Service = "org.kde.kglobalaccel"
	// Path is the object exposing the org.kde.KGlobalAccel interface.
	Path = dbus.ObjectPath("/kglobalaccel")
	// Interface is the daemon's control interface.
	Interface = "org.kde.KGlobalAccel"
	// ComponentInterface is implemented by per-component objects that
	// deliver shortcut activation notifications.
	ComponentInterface = "org.kde.kglobalaccel.Component"
)

// SetShortcutFlag values accepted by the daemon's setShortcut call.
const (
	// FlagSetPresent marks the shortcut as present, which is what makes
	// the daemon grab the key and deliver activations. Without it the
	// shortcut is only stored in the configuration.
	FlagSetPresent uint32 = 2
	// FlagNoAutoloading keeps the daemon from writing the shortcut back
	// to the user configuration.
	FlagNoAutoloading uint32 = 4
	// FlagIsDefault stores the shortcut as the action's default instead
	// of its active binding.
	FlagIsDefault uint32 = 8
)

// Action identifies a shortcut to the global accelerator daemon.
type Action struct {
	// ComponentUnique is the component's unique name. It must be stable
	// across runs because the daemon stores shortcuts against it.
	ComponentUnique string
	// ActionUnique is the action's unique name within the component.
	ActionUnique string
	// ComponentFriendly and ActionFriendly are the human readable names
	// shown in the system settings shortcut editor.
	ComponentFriendly string
	ActionFriendly    string
}

// id is the four element action id used by every daemon call.
func (a Action) id() []string {
	return []string{a.ComponentUnique, a.ActionUnique, a.ComponentFriendly, a.ActionFriendly}
}

// Event describes a shortcut activation.
type Event struct {
	// ActionID is the unique name of the activated action.
	ActionID string
	// Options carries extra information from the daemon. On Wayland it
	// includes "xdgActivationToken", which can be used to raise a window
	// without the compositor suppressing the focus request.
	Options map[string]dbus.Variant
}

// Registration is an active shortcut registration.
type Registration struct {
	conn   *dbus.Conn
	action Action
	seq    int32
	path   dbus.ObjectPath
	events chan Event
}

// Register registers a global shortcut with the KDE global accelerator
// daemon and returns a registration that reports activations.
//
// The shortcut is registered with autoloading enabled, so KDE stores it in
// the user's configuration and it remains visible and editable in System
// Settings.
func Register(conn *dbus.Conn, action Action, seq int32) (*Registration, error) {
	if action.ComponentUnique == "" || action.ActionUnique == "" {
		return nil, fmt.Errorf("hotkey: component and action names are required")
	}
	if seq == 0 {
		return nil, fmt.Errorf("hotkey: refusing to register an empty shortcut")
	}

	daemon := conn.Object(Service, Path)
	id := action.id()

	if err := daemon.Call(Interface+".doRegister", 0, id).Store(); err != nil {
		return nil, fmt.Errorf("hotkey: doRegister: %w", err)
	}

	keys := []int32{seq}
	var applied []int32
	if err := daemon.Call(Interface+".setShortcut", 0, id, keys, FlagSetPresent).Store(&applied); err != nil {
		return nil, fmt.Errorf("hotkey: setShortcut: %w", err)
	}

	var path dbus.ObjectPath
	if err := daemon.Call(Interface+".getComponent", 0, action.ComponentUnique).Store(&path); err != nil {
		return nil, fmt.Errorf("hotkey: getComponent: %w", err)
	}

	if err := conn.AddMatchSignal(
		dbus.WithMatchObjectPath(path),
		dbus.WithMatchInterface(ComponentInterface),
	); err != nil {
		return nil, fmt.Errorf("hotkey: add match rule: %w", err)
	}

	raw := make(chan *dbus.Signal, 16)
	conn.Signal(raw)

	reg := &Registration{
		conn:   conn,
		action: action,
		seq:    seq,
		path:   path,
		events: make(chan Event, 16),
	}

	go reg.forward(raw)

	return reg, nil
}

// forward translates raw D-Bus signals into Events.
//
// The daemon may use either of two signalling styles. The modern "pressed"
// signal carries an options map that includes an xdg activation token,
// while the legacy "globalShortcutPressed" signal only carries the
// component, action and timestamp.
func (r *Registration) forward(raw <-chan *dbus.Signal) {
	for sig := range raw {
		var event Event

		// godbus names signals "<interface>.<member>".
		member := sig.Name
		if idx := strings.LastIndex(member, "."); idx >= 0 {
			member = member[idx+1:]
		}

		switch member {
		case "pressed":
			if len(sig.Body) < 2 {
				continue
			}
			actionID, ok := sig.Body[0].(string)
			if !ok {
				continue
			}
			options, _ := sig.Body[1].(map[string]dbus.Variant)
			event = Event{ActionID: actionID, Options: options}
		case "globalShortcutPressed", "globalShortcutRepeated":
			if len(sig.Body) < 2 {
				continue
			}
			actionID, ok := sig.Body[1].(string)
			if !ok {
				continue
			}
			event = Event{ActionID: actionID, Options: map[string]dbus.Variant{}}
		default:
			continue
		}

		select {
		case r.events <- event:
		default:
		}
	}
}

// Events returns the channel on which shortcut activations are delivered.
func (r *Registration) Events() <-chan Event {
	return r.events
}

// Shortcut returns the registered key combination.
func (r *Registration) Shortcut() int32 {
	return r.seq
}

// ComponentPath returns the D-Bus object path of the component.
func (r *Registration) ComponentPath() dbus.ObjectPath {
	return r.path
}

// Unregister removes the shortcut from the daemon.
func (r *Registration) Unregister() error {
	daemon := r.conn.Object(Service, Path)
	if err := daemon.Call(Interface+".unregister", 0,
		r.action.ComponentUnique, r.action.ActionUnique).Store(); err != nil {
		return fmt.Errorf("hotkey: unregister: %w", err)
	}
	return nil
}

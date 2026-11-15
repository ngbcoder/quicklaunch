// Command kga-spike is a diagnostic tool for the KDE global accelerator
// service. It registers a shortcut and reports activations.
//
// It exists to validate the D-Bus protocol against a live Plasma session;
// see docs/kglobalaccel-notes.md.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/godbus/dbus/v5"

	"github.com/ngbcoder/quicklaunch/internal/hotkey"
)

func main() {
	shortcut := flag.String("shortcut", "Ctrl+Shift+Space", "shortcut in KDE textual form")
	component := flag.String("component", "org.ngbcoder.quicklaunch", "component unique name")
	action := flag.String("action", "toggle", "action unique name")
	invoke := flag.Bool("invoke", false, "invoke the action through the daemon instead of waiting for a key press")
	flag.Parse()

	seq, err := hotkey.ParseSequence(*shortcut)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: session bus:", err)
		os.Exit(1)
	}
	defer conn.Close()

	reply, err := conn.RequestName(*component, dbus.NameFlagDoNotQueue)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: request name:", err)
		os.Exit(1)
	}
	fmt.Printf("bus name %q -> %v\n", *component, reply)

	reg, err := hotkey.Register(conn, hotkey.Action{
		ComponentUnique:   *component,
		ActionUnique:      *action,
		ComponentFriendly: "Quicklaunch",
		ActionFriendly:    "Toggle Quicklaunch",
	}, seq)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: register:", err)
		os.Exit(1)
	}
	fmt.Printf("registered %s = 0x%08x at %s\n", *shortcut, seq, reg.ComponentPath())

	if *invoke {
		comp := conn.Object(hotkey.Service, reg.ComponentPath())
		go func() {
			if err := comp.Call(hotkey.ComponentInterface+".invokeShortcut", 0, *action).Store(); err != nil {
				fmt.Fprintln(os.Stderr, "error: invokeShortcut:", err)
			}
		}()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case ev := <-reg.Events():
			fmt.Printf("activated action=%q options=%v\n", ev.ActionID, ev.Options)
		case s := <-sig:
			fmt.Println("signal:", s)
			return
		}
	}
}

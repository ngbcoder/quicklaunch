// Package apps indexes the applications installed on the system by
// reading their desktop entry files.
package apps

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ngbcoder/quicklaunch/internal/desktop"
)

// App is an application the launcher can start.
type App struct {
	// ID is the desktop file id, the file name relative to the
	// applications directory, for example "firefox.desktop".
	ID string
	// Name is the localised display name.
	Name string
	// GenericName is the localised generic description.
	GenericName string
	// Comment is the localised descriptive comment.
	Comment string
	// Exec is the raw command line of the entry.
	Exec string
	// Icon is the icon path or icon name of the entry.
	Icon string
	// Terminal asks for the command to run in a terminal.
	Terminal bool
	// Keywords and Categories come from the entry.
	Keywords   []string
	Categories []string
	// Source is the path of the desktop file.
	Source string
	// Dir is the directory the desktop file lives in. It is used as the
	// working directory of the command when the entry asks for one.
	Dir string
}

// Options control how applications are scanned.
type Options struct {
	// Locales is the chain of locales to prefer, most specific first.
	Locales []string
	// Desktop is the desktop environment the entries are filtered for.
	// An empty value accepts entries that do not restrict themselves.
	Desktop string
}

// Scan reads the applications of the given directories.
//
// Directories are read in order and the first entry of an id wins, which
// matches the precedence rules of the desktop entry specification.
func Scan(dirs []string, opts Options) ([]App, error) {
	seen := map[string]bool{}
	var apps []App

	for _, dir := range dirs {
		found, err := scanDir(dir, opts, seen)
		if err != nil {
			return nil, err
		}
		apps = append(apps, found...)
	}

	sort.SliceStable(apps, func(i, j int) bool {
		return strings.ToLower(apps[i].Name) < strings.ToLower(apps[j].Name)
	})

	return apps, nil
}

// scanDir reads one applications directory.
func scanDir(dir string, opts Options, seen map[string]bool) ([]App, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("apps: read %s: %w", dir, err)
	}

	var apps []App
	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dir, name)

		if entry.IsDir() {
			// Sub directories hold additional entries, for example
			// "kde4/foo.desktop". Their ids keep the sub directory.
			sub, err := scanSubDir(path, dir, opts, seen)
			if err != nil {
				return nil, err
			}
			apps = append(apps, sub...)
			continue
		}

		if !strings.HasSuffix(name, ".desktop") {
			continue
		}

		app, ok, err := loadApp(path, name, dir, opts, seen)
		if err != nil {
			return nil, err
		}
		if ok {
			apps = append(apps, app)
		}
	}

	return apps, nil
}

// scanSubDir reads the entries of a sub directory of an applications
// directory.
func scanSubDir(path, root string, opts Options, seen map[string]bool) ([]App, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("apps: read %s: %w", path, err)
	}

	var apps []App
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".desktop") {
			continue
		}

		id := filepath.Join(filepath.Base(path), entry.Name())
		app, ok, err := loadApp(filepath.Join(path, entry.Name()), id, root, opts, seen)
		if err != nil {
			return nil, err
		}
		if ok {
			apps = append(apps, app)
		}
	}
	return apps, nil
}

// loadApp parses a single desktop file and turns it into an application.
// It reports false when the entry is skipped.
func loadApp(path, id, root string, opts Options, seen map[string]bool) (App, bool, error) {
	if seen[id] {
		return App{}, false, nil
	}
	seen[id] = true

	entry, err := desktop.ParseFile(path)
	if err != nil {
		// A broken entry must not stop the scan.
		return App{}, false, nil
	}

	if entry.Type != "Application" {
		return App{}, false, nil
	}
	if !entry.Visible() {
		return App{}, false, nil
	}
	if !entry.TryExecAvailable() {
		return App{}, false, nil
	}
	if opts.Desktop != "" && !entry.ShowsIn(opts.Desktop) {
		return App{}, false, nil
	}

	return App{
		ID:          id,
		Name:        entry.NameFor(opts.Locales),
		GenericName: entry.GenericNameFor(opts.Locales),
		Comment:     entry.CommentFor(opts.Locales),
		Exec:        entry.Exec,
		Icon:        entry.Icon,
		Terminal:    entry.Terminal,
		Keywords:    entry.Keywords,
		Categories:  entry.Categories,
		Source:      path,
		Dir:         root,
	}, true, nil
}

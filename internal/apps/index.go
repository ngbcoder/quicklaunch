package apps

import (
	"os"
	"strings"
	"sync"
)

// Index holds the applications of the system and hands out safe copies of
// them while it is being reloaded.
type Index struct {
	mu   sync.RWMutex
	apps []App
}

// New returns an empty index.
func New() *Index {
	return &Index{}
}

// Load replaces the contents of the index.
func (i *Index) Load(apps []App) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.apps = apps
}

// Reload rescans the given directories and replaces the contents of the
// index.
func (i *Index) Reload(dirs []string, opts Options) error {
	apps, err := Scan(dirs, opts)
	if err != nil {
		return err
	}
	i.Load(apps)
	return nil
}

// Apps returns a copy of the indexed applications.
func (i *Index) Apps() []App {
	i.mu.RLock()
	defer i.mu.RUnlock()

	out := make([]App, len(i.apps))
	copy(out, i.apps)
	return out
}

// Len returns the number of indexed applications.
func (i *Index) Len() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return len(i.apps)
}

// Locales returns the chain of locales to prefer, most specific first.
// It is derived from the LC_ALL, LC_MESSAGES and LANG environment
// variables.
func Locales() []string {
	value := firstEnv("LC_ALL", "LC_MESSAGES", "LANG")
	if value == "" {
		return nil
	}

	// Strip the encoding, for example "de_DE.UTF-8".
	if i := strings.IndexByte(value, '.'); i >= 0 {
		value = value[:i]
	}
	if value == "" || value == "C" || value == "POSIX" {
		return nil
	}

	locales := []string{value}
	if i := strings.IndexByte(value, '_'); i >= 0 {
		locales = append(locales, value[:i])
	}
	return locales
}

// firstEnv returns the first environment variable that is set.
func firstEnv(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

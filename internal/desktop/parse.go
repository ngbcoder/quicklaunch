// Package desktop parses freedesktop.org desktop entry files.
//
// The implementation follows the Desktop Entry Specification: sections,
// key/value pairs, localised keys, escapes and list values.
package desktop

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	// MainGroup is the name of the group holding the entry's own keys.
	MainGroup = "Desktop Entry"
	// ActionGroupPrefix is the prefix of the groups holding named actions.
	ActionGroupPrefix = "Desktop Action "
)

// Entry is a parsed desktop entry.
type Entry struct {
	// Type is the entry type, usually "Application".
	Type string
	// Name is the unlocalised display name.
	Name string
	// GenericName is the unlocalised generic description.
	GenericName string
	// Comment is the unlocalised descriptive comment.
	Comment string
	// Exec is the command line to run.
	Exec string
	// TryExec is a command that must exist for the entry to be shown.
	TryExec string
	// Icon is either an absolute path or an icon name.
	Icon string
	// Terminal asks for the command to be run in a terminal.
	Terminal bool
	// NoDisplay hides the entry from menus.
	NoDisplay bool
	// Hidden marks the entry as deleted.
	Hidden bool
	// StartupNotify asks the launcher to show a busy notification.
	StartupNotify bool
	// StartupWMClass is the window class used for icon matching.
	StartupWMClass string
	// Path is the working directory for the command.
	Path string
	// URL is the target of a "Link" type entry.
	URL string
	// Categories, Keywords and MimeType are list values.
	Categories []string
	Keywords   []string
	MimeType   []string
	// OnlyShowIn and NotShowIn restrict the desktop environments the
	// entry appears in.
	OnlyShowIn []string
	NotShowIn  []string
	// Actions holds the named actions of the entry.
	Actions []Action

	// Source is the path of the file the entry was read from.
	Source string

	// localised holds the localised variants of string keys.
	localised map[string]map[string]string
}

// Action is a named sub-action of an entry, taken from a
// "[Desktop Action ...]" group.
type Action struct {
	// Name is the action's identifier.
	Name string
	// Exec is the command line of the action.
	Exec string
	// Icon is the action's icon.
	Icon string
	// Label is the unlocalised display name of the action.
	Label string
}

// group is a single section of a desktop file.
type group struct {
	name      string
	keys      map[string]string
	localised map[string]map[string]string
}

// newGroup returns an empty group with the given name.
func newGroup(name string) *group {
	return &group{
		name:      name,
		keys:      map[string]string{},
		localised: map[string]map[string]string{},
	}
}

// set stores a key. Keys with a locale suffix, such as "Name[de_DE]",
// are stored separately from their unlocalised counterpart.
//
// Values are stored as written; escape sequences are resolved when the
// value is read, so that list values keep their escaped separators.
func (g *group) set(key, value string) {
	if name, locale, ok := splitLocale(key); ok {
		if g.localised[name] == nil {
			g.localised[name] = map[string]string{}
		}
		g.localised[name][locale] = value
		return
	}
	g.keys[key] = value
}

// get returns an unlocalised key with its escape sequences resolved.
func (g *group) get(key string) string {
	return unescape(g.keys[key])
}

// getRaw returns an unlocalised key as written, without resolving escape
// sequences. List values need this, because their separators may be
// escaped.
func (g *group) getRaw(key string) string {
	return g.keys[key]
}

// lookup returns the best value for a key, preferring the given locales in
// order and falling back to the unlocalised value.
func (g *group) lookup(key string, locales []string) string {
	for _, locale := range locales {
		if values, ok := g.localised[key]; ok {
			if value, ok := values[locale]; ok {
				return unescape(value)
			}
		}
	}
	return g.get(key)
}

// splitLocale splits "Name[de_DE]" into "Name" and "de_DE". It reports
// false for keys without a locale suffix.
func splitLocale(key string) (name, locale string, ok bool) {
	start := strings.IndexByte(key, '[')
	if start < 0 || !strings.HasSuffix(key, "]") {
		return key, "", false
	}
	return key[:start], key[start+1 : len(key)-1], true
}

// ParseFile reads and parses a desktop entry file.
func ParseFile(path string) (*Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entry, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	entry.Source = path
	return entry, nil
}

// ParseBytes parses a desktop entry from a byte slice.
func ParseBytes(data []byte) (*Entry, error) {
	return Parse(strings.NewReader(string(data)))
}

// Parse reads a desktop entry.
//
// Only the "[Desktop Entry]" group and "[Desktop Action ...]" groups are
// interpreted; any other group is ignored, as the specification requires.
func Parse(r io.Reader) (*Entry, error) {
	groups, err := parseGroups(r)
	if err != nil {
		return nil, err
	}

	main := newGroup(MainGroup)
	for _, g := range groups {
		if g.name == MainGroup {
			main = g
			break
		}
	}
	if len(main.keys) == 0 && len(main.localised) == 0 {
		return nil, fmt.Errorf("desktop: missing %q group", MainGroup)
	}

	entry := &Entry{
		Type:           main.get("Type"),
		Name:           main.get("Name"),
		GenericName:    main.get("GenericName"),
		Comment:        main.get("Comment"),
		Exec:           main.get("Exec"),
		TryExec:        main.get("TryExec"),
		Icon:           main.get("Icon"),
		Terminal:       boolValue(main.get("Terminal")),
		NoDisplay:      boolValue(main.get("NoDisplay")),
		Hidden:         boolValue(main.get("Hidden")),
		StartupNotify:  boolValue(main.get("StartupNotify")),
		StartupWMClass: main.get("StartupWMClass"),
		Path:           main.get("Path"),
		URL:            main.get("URL"),
		Categories:     listValue(main.getRaw("Categories")),
		Keywords:       listValue(main.getRaw("Keywords")),
		MimeType:       listValue(main.getRaw("MimeType")),
		OnlyShowIn:     listValue(main.getRaw("OnlyShowIn")),
		NotShowIn:      listValue(main.getRaw("NotShowIn")),
		localised:      main.localised,
	}

	for _, g := range groups {
		if !strings.HasPrefix(g.name, ActionGroupPrefix) {
			continue
		}
		entry.Actions = append(entry.Actions, Action{
			Name:  strings.TrimPrefix(g.name, ActionGroupPrefix),
			Exec:  g.get("Exec"),
			Icon:  g.get("Icon"),
			Label: g.get("Name"),
		})
	}

	return entry, nil
}

// NameFor returns the display name for the given locale. The locale chain
// is tried in order, for example "de_DE.UTF-8", "de_DE", "de", before
// falling back to the unlocalised name.
func (e *Entry) NameFor(locales []string) string {
	if value, ok := e.localised["Name"]; ok {
		for _, locale := range locales {
			if v, ok := value[locale]; ok {
				return unescape(v)
			}
		}
	}
	return e.Name
}

// GenericNameFor returns the localised generic name, if any.
func (e *Entry) GenericNameFor(locales []string) string {
	if value, ok := e.localised["GenericName"]; ok {
		for _, locale := range locales {
			if v, ok := value[locale]; ok {
				return unescape(v)
			}
		}
	}
	return e.GenericName
}

// CommentFor returns the localised comment, if any.
func (e *Entry) CommentFor(locales []string) string {
	if value, ok := e.localised["Comment"]; ok {
		for _, locale := range locales {
			if v, ok := value[locale]; ok {
				return unescape(v)
			}
		}
	}
	return e.Comment
}

// ShowsIn reports whether the entry should appear in the given desktop
// environment, according to OnlyShowIn and NotShowIn.
func (e *Entry) ShowsIn(desktop string) bool {
	if len(e.OnlyShowIn) > 0 {
		found := false
		for _, d := range e.OnlyShowIn {
			if strings.EqualFold(d, desktop) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for _, d := range e.NotShowIn {
		if strings.EqualFold(d, desktop) {
			return false
		}
	}
	return true
}

// Visible reports whether the entry should be offered to the user.
func (e *Entry) Visible() bool {
	return !e.NoDisplay && !e.Hidden
}

// parseGroups splits a desktop file into its groups.
func parseGroups(r io.Reader) ([]*group, error) {
	var groups []*group
	var current *group

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return nil, fmt.Errorf("desktop: malformed group header on line %d", lineNo)
			}
			name := strings.TrimSpace(line[1 : len(line)-1])
			if name == "" {
				return nil, fmt.Errorf("desktop: empty group name on line %d", lineNo)
			}
			current = newGroup(name)
			groups = append(groups, current)
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("desktop: malformed key on line %d", lineNo)
		}
		if current == nil {
			// Keys before any group header belong to the main group.
			current = newGroup(MainGroup)
			groups = append(groups, current)
		}
		current.set(strings.TrimSpace(key), strings.TrimSpace(value))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

// unescape resolves the escape sequences the specification defines.
func unescape(value string) string {
	if !strings.ContainsRune(value, '\\') {
		return value
	}

	var b strings.Builder
	b.Grow(len(value))
	for i := 0; i < len(value); i++ {
		if value[i] != '\\' || i+1 >= len(value) {
			b.WriteByte(value[i])
			continue
		}
		i++
		switch value[i] {
		case 's':
			b.WriteByte(' ')
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '\\':
			b.WriteByte('\\')
		case ';':
			b.WriteByte(';')
		default:
			// Unknown escapes are kept verbatim.
			b.WriteByte('\\')
			b.WriteByte(value[i])
		}
	}
	return b.String()
}

// boolValue interprets a boolean key. Anything other than "true" is false,
// as the specification requires.
func boolValue(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true")
}

// listValue splits a list value on semicolons, dropping empty entries.
// Separators that are escaped stay part of the value.
func listValue(value string) []string {
	if value == "" {
		return nil
	}
	var items []string
	for _, part := range splitEscaped(value, ';') {
		item := strings.TrimSpace(unescape(part))
		if item == "" {
			continue
		}
		items = append(items, item)
	}
	return items
}

// splitEscaped splits a value on a separator, ignoring separators that are
// escaped with a backslash.
func splitEscaped(value string, separator byte) []string {
	var parts []string
	var current strings.Builder

	for i := 0; i < len(value); i++ {
		if value[i] == '\\' && i+1 < len(value) {
			current.WriteByte(value[i])
			i++
			current.WriteByte(value[i])
			continue
		}
		if value[i] == separator {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(value[i])
	}
	parts = append(parts, current.String())
	return parts
}

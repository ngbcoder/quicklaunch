package desktop

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseBasic(t *testing.T) {
	entry, err := ParseFile("testdata/basic.desktop")
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if entry.Type != "Application" {
		t.Errorf("Type = %q, want %q", entry.Type, "Application")
	}
	if entry.Name != "Firefox" {
		t.Errorf("Name = %q, want %q", entry.Name, "Firefox")
	}
	if entry.GenericName != "Web Browser" {
		t.Errorf("GenericName = %q, want %q", entry.GenericName, "Web Browser")
	}
	if entry.Exec != "firefox %u" {
		t.Errorf("Exec = %q, want %q", entry.Exec, "firefox %u")
	}
	if entry.Icon != "firefox" {
		t.Errorf("Icon = %q, want %q", entry.Icon, "firefox")
	}
	if entry.Terminal {
		t.Error("Terminal = true, want false")
	}
	if !entry.StartupNotify {
		t.Error("StartupNotify = false, want true")
	}
	if entry.StartupWMClass != "firefox" {
		t.Errorf("StartupWMClass = %q, want %q", entry.StartupWMClass, "firefox")
	}
	if want := []string{"Network", "WebBrowser", "Browser"}; !reflect.DeepEqual(entry.Categories, want) {
		t.Errorf("Categories = %v, want %v", entry.Categories, want)
	}
	if want := []string{"web", "internet"}; !reflect.DeepEqual(entry.Keywords, want) {
		t.Errorf("Keywords = %v, want %v", entry.Keywords, want)
	}
	if want := []string{"text/html", "application/xhtml+xml"}; !reflect.DeepEqual(entry.MimeType, want) {
		t.Errorf("MimeType = %v, want %v", entry.MimeType, want)
	}
	if want := []string{"GNOME"}; !reflect.DeepEqual(entry.NotShowIn, want) {
		t.Errorf("NotShowIn = %v, want %v", entry.NotShowIn, want)
	}
	if !strings.HasSuffix(entry.Source, "basic.desktop") {
		t.Errorf("Source = %q, want it to end with basic.desktop", entry.Source)
	}
}

func TestParseLocalizedNames(t *testing.T) {
	entry, err := ParseFile("testdata/basic.desktop")
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	tests := []struct {
		name    string
		locales []string
		want    string
	}{
		{"exact locale", []string{"de_DE"}, "Mozilla Feuerfuchs"},
		{"language only", []string{"de"}, "Feuerfuchs"},
		{"falls back to unlocalised", []string{"it"}, "Firefox"},
		{"first matching locale wins", []string{"fr", "de"}, "Renard de feu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := entry.NameFor(tt.locales); got != tt.want {
				t.Errorf("NameFor(%v) = %q, want %q", tt.locales, got, tt.want)
			}
		})
	}

	if got := entry.GenericNameFor([]string{"de"}); got != "Webbrowser" {
		t.Errorf("GenericNameFor(de) = %q, want %q", got, "Webbrowser")
	}
}

func TestParseActions(t *testing.T) {
	entry, err := ParseFile("testdata/actions.desktop")
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if len(entry.Actions) != 2 {
		t.Fatalf("got %d actions, want 2", len(entry.Actions))
	}

	first := entry.Actions[0]
	if first.Name != "new-window" {
		t.Errorf("action name = %q, want %q", first.Name, "new-window")
	}
	if first.Exec != "editor --new-window %U" {
		t.Errorf("action Exec = %q, want %q", first.Exec, "editor --new-window %U")
	}
	if first.Icon != "document-new" {
		t.Errorf("action Icon = %q, want %q", first.Icon, "document-new")
	}
	if first.Label != "New Window" {
		t.Errorf("action Label = %q, want %q", first.Label, "New Window")
	}
}

func TestParseEscapes(t *testing.T) {
	entry, err := ParseFile("testdata/escapes.desktop")
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if entry.Name != "Escaped Value" {
		t.Errorf("Name = %q, want %q", entry.Name, "Escaped Value")
	}
	if entry.Comment != "First Line\nSecond" {
		t.Errorf("Comment = %q, want %q", entry.Comment, "First Line\nSecond")
	}
	if want := []string{"One", "Two;Three"}; !reflect.DeepEqual(entry.Categories, want) {
		t.Errorf("Categories = %v, want %v", entry.Categories, want)
	}
}

func TestParseHiddenAndNoDisplay(t *testing.T) {
	entry, err := ParseFile("testdata/hidden.desktop")
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if !entry.NoDisplay {
		t.Error("NoDisplay = false, want true")
	}
	if !entry.Hidden {
		t.Error("Hidden = false, want true")
	}
	if entry.Visible() {
		t.Error("Visible() = true, want false")
	}
}

func TestShowsIn(t *testing.T) {
	onlyKDE, err := ParseFile("testdata/onlykde.desktop")
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if !onlyKDE.ShowsIn("KDE") {
		t.Error("ShowsIn(KDE) = false, want true")
	}
	if onlyKDE.ShowsIn("GNOME") {
		t.Error("ShowsIn(GNOME) = true, want false")
	}

	notGNOME, err := ParseFile("testdata/basic.desktop")
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if notGNOME.ShowsIn("GNOME") {
		t.Error("ShowsIn(GNOME) = true, want false")
	}
	if !notGNOME.ShowsIn("KDE") {
		t.Error("ShowsIn(KDE) = false, want true")
	}
}

func TestParseLinkType(t *testing.T) {
	entry, err := ParseFile("testdata/link.desktop")
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if entry.Type != "Link" {
		t.Errorf("Type = %q, want %q", entry.Type, "Link")
	}
	if entry.URL != "file:///home/user" {
		t.Errorf("URL = %q, want %q", entry.URL, "file:///home/user")
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := ParseFile("testdata/malformed.desktop"); err == nil {
		t.Error("expected an error for a malformed file")
	} else if !strings.Contains(err.Error(), "line") {
		t.Errorf("error = %v, want it to mention the line number", err)
	}

	if _, err := ParseFile("testdata/does-not-exist.desktop"); err == nil {
		t.Error("expected an error for a missing file")
	}
}

func TestParseIgnoresUnknownGroups(t *testing.T) {
	entry, err := ParseFile("testdata/actions.desktop")
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if entry.Name != "Editor" {
		t.Errorf("Name = %q, want %q", entry.Name, "Editor")
	}
}

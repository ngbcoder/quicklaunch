package apps

import (
	"os"
	"path/filepath"
	"testing"
)

// writeEntry writes a desktop file into a directory.
func writeEntry(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

// ids returns the ids of the given applications.
func ids(apps []App) []string {
	out := make([]string, len(apps))
	for i, a := range apps {
		out[i] = a.ID
	}
	return out
}

func TestScanFindsEntries(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "firefox.desktop", `[Desktop Entry]
Type=Application
Name=Firefox
Exec=firefox %u
Icon=firefox
`)

	apps, err := Scan([]string{dir}, Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("got %d apps, want 1", len(apps))
	}

	app := apps[0]
	if app.ID != "firefox.desktop" {
		t.Errorf("ID = %q, want %q", app.ID, "firefox.desktop")
	}
	if app.Name != "Firefox" {
		t.Errorf("Name = %q, want %q", app.Name, "Firefox")
	}
	if app.Exec != "firefox %u" {
		t.Errorf("Exec = %q, want %q", app.Exec, "firefox %u")
	}
	if app.Icon != "firefox" {
		t.Errorf("Icon = %q, want %q", app.Icon, "firefox")
	}
}

func TestScanFiltersEntries(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "visible.desktop", "[Desktop Entry]\nType=Application\nName=Visible\nExec=visible\n")
	writeEntry(t, dir, "nodisplay.desktop", "[Desktop Entry]\nType=Application\nName=Hidden\nExec=hidden\nNoDisplay=true\n")
	writeEntry(t, dir, "hidden.desktop", "[Desktop Entry]\nType=Application\nName=Deleted\nExec=deleted\nHidden=true\n")
	writeEntry(t, dir, "link.desktop", "[Desktop Entry]\nType=Link\nName=Link\nURL=https://example.org\n")
	writeEntry(t, dir, "notryexec.desktop", "[Desktop Entry]\nType=Application\nName=NoTryExec\nExec=app\nTryExec=definitely-not-a-program-xyz\n")

	apps, err := Scan([]string{dir}, Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("got %d apps (%v), want 1", len(apps), ids(apps))
	}
	if apps[0].ID != "visible.desktop" {
		t.Errorf("ID = %q, want %q", apps[0].ID, "visible.desktop")
	}
}

func TestScanRespectsDesktopEnvironment(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "kde.desktop", "[Desktop Entry]\nType=Application\nName=KDE Only\nExec=kde\nOnlyShowIn=KDE;\n")
	writeEntry(t, dir, "gnome.desktop", "[Desktop Entry]\nType=Application\nName=GNOME Only\nExec=gnome\nOnlyShowIn=GNOME;\n")
	writeEntry(t, dir, "all.desktop", "[Desktop Entry]\nType=Application\nName=All\nExec=all\n")

	apps, err := Scan([]string{dir}, Options{Desktop: "KDE"})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("got %d apps (%v), want 2", len(apps), ids(apps))
	}
	for _, app := range apps {
		if app.ID == "gnome.desktop" {
			t.Error("GNOME only entry should not appear on KDE")
		}
	}
}

func TestScanDeduplicatesById(t *testing.T) {
	low := t.TempDir()
	high := t.TempDir()
	writeEntry(t, low, "app.desktop", "[Desktop Entry]\nType=Application\nName=System Version\nExec=system\n")
	writeEntry(t, high, "app.desktop", "[Desktop Entry]\nType=Application\nName=User Version\nExec=user\n")

	// The first directory has priority.
	apps, err := Scan([]string{high, low}, Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("got %d apps, want 1", len(apps))
	}
	if apps[0].Name != "User Version" {
		t.Errorf("Name = %q, want %q", apps[0].Name, "User Version")
	}
}

func TestScanSubDirectories(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, filepath.Join(dir, "kde4"), "app.desktop", "[Desktop Entry]\nType=Application\nName=Sub\nExec=sub\n")

	apps, err := Scan([]string{dir}, Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("got %d apps, want 1", len(apps))
	}
	if apps[0].ID != filepath.Join("kde4", "app.desktop") {
		t.Errorf("ID = %q, want %q", apps[0].ID, filepath.Join("kde4", "app.desktop"))
	}
}

func TestScanMissingDirectory(t *testing.T) {
	apps, err := Scan([]string{filepath.Join(t.TempDir(), "missing")}, Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("got %d apps, want 0", len(apps))
	}
}

func TestScanLocalizesNames(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "app.desktop", `[Desktop Entry]
Type=Application
Name=Application
Name[de]=Anwendung
Exec=app
`)

	apps, err := Scan([]string{dir}, Options{Locales: []string{"de"}})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("got %d apps, want 1", len(apps))
	}
	if apps[0].Name != "Anwendung" {
		t.Errorf("Name = %q, want %q", apps[0].Name, "Anwendung")
	}
}

func TestScanSortsByName(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "c.desktop", "[Desktop Entry]\nType=Application\nName=Charlie\nExec=c\n")
	writeEntry(t, dir, "a.desktop", "[Desktop Entry]\nType=Application\nName=alpha\nExec=a\n")
	writeEntry(t, dir, "b.desktop", "[Desktop Entry]\nType=Application\nName=Bravo\nExec=b\n")

	apps, err := Scan([]string{dir}, Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	want := []string{"alpha", "Bravo", "Charlie"}
	for i, app := range apps {
		if app.Name != want[i] {
			t.Errorf("app %d = %q, want %q", i, app.Name, want[i])
		}
	}
}

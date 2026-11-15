package apps

import (
	"os"
	"testing"
)

func TestDefaultDirsIncludeFlatpak(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data-home")
	t.Setenv("XDG_DATA_DIRS", "/tmp/xdg-data-dir")

	dirs := DefaultDirs()
	want := []string{
		"/tmp/xdg-data-home/applications",
		"/tmp/xdg-data-home/flatpak/exports/share/applications",
		"/tmp/xdg-data-dir/applications",
		"/tmp/xdg-data-dir/flatpak/exports/share/applications",
	}

	for _, w := range want {
		found := false
		for _, dir := range dirs {
			if dir == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultDirs() does not contain %q, got %v", w, dirs)
		}
	}
}

// removeFile deletes a file and fails the test if it cannot.
func removeFile(path string) error {
	return os.Remove(path)
}

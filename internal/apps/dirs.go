package apps

import (
	"os"
	"path/filepath"
)

// DefaultDirs returns the directories that are scanned, highest priority
// first. Later directories never override entries of earlier ones.
func DefaultDirs() []string {
	dirs := make([]string, 0, 8)

	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(homeDir(), ".local", "share")
	}
	dirs = append(dirs, filepath.Join(dataHome, "applications"))
	dirs = append(dirs, flatpakExports(dataHome))

	for _, dir := range dataDirs() {
		dirs = append(dirs, filepath.Join(dir, "applications"))
		dirs = append(dirs, flatpakExports(dir))
	}

	return dirs
}

// flatpakExports returns the directory Flatpak exports its applications to.
func flatpakExports(dataDir string) string {
	return filepath.Join(dataDir, "flatpak", "exports", "share", "applications")
}

// dataDirs returns the XDG data directories.
func dataDirs() []string {
	value := os.Getenv("XDG_DATA_DIRS")
	if value == "" {
		return []string{"/usr/local/share", "/usr/share"}
	}
	return filepath.SplitList(value)
}

// homeDir returns the user's home directory.
func homeDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return os.Getenv("HOME")
}

package desktop

import (
	"os"
	"path/filepath"
	"strings"
)

// TryExecAvailable reports whether the entry's TryExec command exists.
//
// An empty TryExec means the entry is always available. Otherwise the
// value is either a full path or the name of a program in PATH.
func (e *Entry) TryExecAvailable() bool {
	if e.TryExec == "" {
		return true
	}

	if strings.ContainsRune(e.TryExec, os.PathSeparator) {
		return isExecutable(e.TryExec)
	}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		if isExecutable(filepath.Join(dir, e.TryExec)) {
			return true
		}
	}
	return false
}

// isExecutable reports whether path is a regular, executable file.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode().Perm()&0o111 != 0
}

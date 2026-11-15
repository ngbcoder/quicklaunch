package apps

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestIndexReload(t *testing.T) {
	dir := t.TempDir()
	writeEntry(t, dir, "app.desktop", "[Desktop Entry]\nType=Application\nName=App\nExec=app\n")

	index := New()
	if index.Len() != 0 {
		t.Errorf("Len() = %d, want 0", index.Len())
	}

	if err := index.Reload([]string{dir}, Options{}); err != nil {
		t.Fatalf("Reload returned error: %v", err)
	}
	if index.Len() != 1 {
		t.Errorf("Len() = %d, want 1", index.Len())
	}

	apps := index.Apps()
	if len(apps) != 1 || apps[0].Name != "App" {
		t.Fatalf("Apps() = %v, want a single App", apps)
	}

	// Removing the file empties the index on the next reload.
	if err := removeFile(filepath.Join(dir, "app.desktop")); err != nil {
		t.Fatalf("remove file: %v", err)
	}
	if err := index.Reload([]string{dir}, Options{}); err != nil {
		t.Fatalf("Reload returned error: %v", err)
	}
	if index.Len() != 0 {
		t.Errorf("Len() = %d, want 0", index.Len())
	}
}

func TestIndexAppsReturnsACopy(t *testing.T) {
	index := New()
	index.Load([]App{{ID: "app.desktop", Name: "App"}})

	apps := index.Apps()
	apps[0].Name = "Changed"

	if again := index.Apps(); again[0].Name != "App" {
		t.Errorf("index was modified through the returned slice: %q", again[0].Name)
	}
}

func TestIndexIsSafeForConcurrentUse(t *testing.T) {
	index := New()
	index.Load([]App{{ID: "app.desktop", Name: "App"}})

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			index.Load([]App{{ID: "app.desktop", Name: "App"}})
		}()
		go func() {
			defer wg.Done()
			_ = index.Apps()
			_ = index.Len()
		}()
	}
	wg.Wait()
}

package desktop

import "testing"

func TestTryExecAvailable(t *testing.T) {
	tests := []struct {
		name    string
		tryExec string
		want    bool
	}{
		{"empty means available", "", true},
		{"absolute path that exists", "/bin/sh", true},
		{"absolute path that does not exist", "/bin/definitely-not-here", false},
		{"program in path", "sh", true},
		{"program that is not in path", "definitely-not-a-program-xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &Entry{TryExec: tt.tryExec}
			if got := entry.TryExecAvailable(); got != tt.want {
				t.Errorf("TryExecAvailable(%q) = %v, want %v", tt.tryExec, got, tt.want)
			}
		})
	}
}

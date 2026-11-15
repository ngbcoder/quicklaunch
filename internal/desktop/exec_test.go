package desktop

import (
	"reflect"
	"testing"
)

func TestExpand(t *testing.T) {
	opts := ExpandOptions{
		Name:     "Firefox",
		Icon:     "/usr/share/icons/firefox.png",
		Files:    []string{"/tmp/one.pdf", "/tmp/two.pdf"},
		URLs:     []string{"https://example.org"},
		Location: "file:///tmp/one.pdf",
	}

	tests := []struct {
		name string
		exec string
		opts ExpandOptions
		want []string
	}{
		{
			name: "plain command",
			exec: "firefox",
			want: []string{"firefox"},
		},
		{
			name: "single url",
			exec: "firefox %u",
			want: []string{"firefox", "https://example.org"},
		},
		{
			name: "url list",
			exec: "firefox %U",
			want: []string{"firefox", "https://example.org"},
		},
		{
			name: "file list expands to several arguments",
			exec: "viewer %F",
			want: []string{"viewer", "/tmp/one.pdf", "/tmp/two.pdf"},
		},
		{
			name: "single file",
			exec: "viewer %f",
			want: []string{"viewer", "/tmp/one.pdf"},
		},
		{
			name: "icon expands to two arguments",
			exec: "app %i",
			want: []string{"app", "--icon", "/usr/share/icons/firefox.png"},
		},
		{
			name: "name code",
			exec: "app %c",
			want: []string{"app", "Firefox"},
		},
		{
			name: "location code",
			exec: "app %k",
			want: []string{"app", "file:///tmp/one.pdf"},
		},
		{
			name: "double percent is a literal",
			exec: "app %%",
			want: []string{"app", "%"},
		},
		{
			name: "quoted argument keeps its spaces",
			exec: `app "argument with spaces" --flag=value`,
			want: []string{"app", "argument with spaces", "--flag=value"},
		},
		{
			name: "code inside a larger argument",
			exec: "app --file=%f",
			want: []string{"app", "--file=/tmp/one.pdf"},
		},
		{
			name: "deprecated code behaves like its replacement",
			exec: "app %d",
			want: []string{"app", "/tmp/one.pdf"},
		},
		{
			name: "unknown code is dropped",
			exec: "app %z",
			want: []string{"app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			use := tt.opts
			if use.Name == "" && use.Icon == "" && use.Files == nil && use.URLs == nil {
				use = opts
			}

			got, err := Expand(tt.exec, use)
			if err != nil {
				t.Fatalf("Expand(%q) returned error: %v", tt.exec, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expand(%q) = %v, want %v", tt.exec, got, tt.want)
			}
		})
	}
}

func TestExpandRemovesUnsatisfiableCodes(t *testing.T) {
	empty := ExpandOptions{}

	tests := []struct {
		name string
		exec string
		want []string
	}{
		{"file code without files", "app %f", []string{"app"}},
		{"file list without files", "app %F", []string{"app"}},
		{"url code without urls", "app %u", []string{"app"}},
		{"icon code without icon", "app %i", []string{"app"}},
		{"name code without name", "app %c", []string{"app"}},
		{"several codes", "app %f %u %F", []string{"app"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Expand(tt.exec, empty)
			if err != nil {
				t.Fatalf("Expand(%q) returned error: %v", tt.exec, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expand(%q) = %v, want %v", tt.exec, got, tt.want)
			}
		})
	}
}

func TestExpandErrors(t *testing.T) {
	for _, exec := range []string{"app %", `app "unbalanced`, "app trailing\\"} {
		if _, err := Expand(exec, ExpandOptions{}); err == nil {
			t.Errorf("Expand(%q) expected an error", exec)
		}
	}
}

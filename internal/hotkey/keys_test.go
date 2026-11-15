package hotkey

import "testing"

func TestParseSequence(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int32
	}{
		{"ctrl shift space", "Ctrl+Shift+Space", 0x06000020},
		{"meta space", "Meta+Space", 0x10000020},
		{"ctrl shift s", "Ctrl+Shift+S", 0x06000053},
		{"ctrl plus", "Ctrl++", 0x0400002b},
		{"meta ctrl right", "Meta+Ctrl+Right", 0x15000014},
		{"alt meta f3", "Alt+Meta+F3", 0x19000032},
		{"lower case modifiers", "ctrl+shift+space", 0x06000020},
		{"spaces around parts", " Ctrl + Shift + Space ", 0x06000020},
		{"function key", "Ctrl+F5", 0x05000034},
		{"arrow key", "Alt+Left", 0x09000012},
		{"plus key", "Ctrl++", 0x0400002b},
		{"letter is upper cased", "Ctrl+Shift+s", 0x06000053},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSequence(tt.in)
			if err != nil {
				t.Fatalf("ParseSequence(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseSequence(%q) = 0x%08x, want 0x%08x", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseSequenceErrors(t *testing.T) {
	for _, in := range []string{"", "Ctrl", "Ctrl+Shift", "Ctrl+NotAKey", "Ctrl+Shift+Space+Extra"} {
		if _, err := ParseSequence(in); err == nil {
			t.Errorf("ParseSequence(%q) expected an error", in)
		}
	}
}

func TestSequenceStringRoundTrip(t *testing.T) {
	for _, in := range []string{"Ctrl+Shift+Space", "Meta+Space", "Ctrl+Shift+S", "Alt+Left", "Meta+Ctrl+Right"} {
		value, err := ParseSequence(in)
		if err != nil {
			t.Fatalf("ParseSequence(%q) returned error: %v", in, err)
		}
		if got := SequenceString(value); got != in {
			t.Errorf("SequenceString(0x%08x) = %q, want %q", value, got, in)
		}
	}
}

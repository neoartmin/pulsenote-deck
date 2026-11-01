package main

import "testing"

func TestParseShortcutValid(t *testing.T) {
	cases := []struct {
		raw  string
		want []byte
	}{
		{"CTRL+SHIFT+S", []byte{0x11, 0x10, 'S'}},
		{"alt+tab", []byte{0x12, 0x09}},
		{"WIN+F24", []byte{0x5b, 0x87}},
		{"ENTER", []byte{0x0d}},
	}
	for _, tc := range cases {
		got, err := parseShortcut(tc.raw)
		if err != nil || len(got) != len(tc.want) {
			t.Fatalf("%q: %v %v", tc.raw, got, err)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("%q: got %v want %v", tc.raw, got, tc.want)
			}
		}
	}
}
func TestParseShortcutNeverInjectsInvalidPartialCombo(t *testing.T) {
	for _, raw := range []string{"CTRL+NOT_A_KEY", "CTRL+", "+A", "CTRL++S", "CTRL+CONTROL+S", "CTRL+SHIFT+ALT+WIN+A+B", "F25", ""} {
		got, err := parseShortcut(raw)
		if err == nil || got != nil {
			t.Errorf("invalid %q: %v %v", raw, got, err)
		}
	}
}

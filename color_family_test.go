package semstyle

import "testing"

func TestColorFamily(t *testing.T) {
	tests := []struct {
		in     string
		family string
		ok     bool
	}{
		{"magenta", "magenta", true},
		{"Bright-Magenta", "magenta", true},
		{"13", "magenta", true},
		{"black", "", true},
		{"bright-black", "", true},
		{"white", "", true},
		{"bright-white", "", true},
		{"#ff8000", "yellow", true},
		{"#e01010", "red", true},
		{"#808080", "", true},
		{"#101010", "", true},
		{"-", "", false},
		{"", "", false},
		{"base08", "", false},
	}
	for _, tt := range tests {
		family, ok := ColorFamily(tt.in)
		if family != tt.family || ok != tt.ok {
			t.Errorf("ColorFamily(%q) = %q, %v; want %q, %v", tt.in, family, ok, tt.family, tt.ok)
		}
	}
}

func TestBasicColor(t *testing.T) {
	tests := []struct {
		in    string
		basic string
		ok    bool
	}{
		{"magenta", "magenta", true},
		{"bright-black", "black", true},
		{"bright-white", "white", true},
		{"15", "white", true},
		{"#ff8000", "yellow", true},
		{"#808080", "white", true},
		{"#404040", "black", true},
		{"-", "", false},
		{"base05", "", false},
	}
	for _, tt := range tests {
		basic, ok := BasicColor(tt.in)
		if basic != tt.basic || ok != tt.ok {
			t.Errorf("BasicColor(%q) = %q, %v; want %q, %v", tt.in, basic, ok, tt.basic, tt.ok)
		}
	}
}

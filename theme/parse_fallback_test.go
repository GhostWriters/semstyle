package semtheme

import (
	"testing"

	"github.com/GhostWriters/semstyle"
)

func TestResolveColorsHonorsRegisteredFallback(t *testing.T) {
	semstyle.ClearFallbacks()
	defer semstyle.ClearFallbacks()
	semstyle.RegisterFallback("PanelBorder", true, "Helpline")

	tf := ThemeFile{
		Styles: map[string]string{
			"Helpline":   "{{[black:cyan:-]}}",
			"PanelTitle": "{{|PanelBorder|}}",
			// PanelBorder deliberately omitted -- relies on the registered
			// fallback rule, same as a theme file that comments it out.
		},
	}
	resolved, err := ResolveColors(tf)
	if err != nil {
		t.Fatalf("ResolveColors failed: %v", err)
	}
	want := "black:cyan:"
	if got := resolved["PanelTitle"]; got != want {
		t.Errorf("PanelTitle = %q, want %q", got, want)
	}
}

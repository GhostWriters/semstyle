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

// TestResolveColorsHardResetExpandsPerField verifies a bare "~" (semstyle's
// "defer entirely to the real terminal" hard-reset marker, mirroring "-"'s
// existing per-field expansion) resolves to "~:~:" rather than "~::" -- the
// bg field must itself carry the "~" marker so CodeToStyle actually clears
// it, instead of being left empty (which CodeToStyle treats as "field not
// specified," silently leaving whatever background was already active).
func TestResolveColorsHardResetExpandsPerField(t *testing.T) {
	tf := ThemeFile{
		Styles: map[string]string{
			"Item": "{{[~]}}",
		},
	}
	resolved, err := ResolveColors(tf)
	if err != nil {
		t.Fatalf("ResolveColors failed: %v", err)
	}
	want := "~:~:"
	if got := resolved["Item"]; got != want {
		t.Errorf("Item = %q, want %q", got, want)
	}
}

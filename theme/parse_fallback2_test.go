package semtheme

import (
	"testing"

	"github.com/GhostWriters/semstyle"
)

func TestResolveColorsHonorsTwoHopFallback(t *testing.T) {
	semstyle.ClearFallbacks()
	defer semstyle.ClearFallbacks()
	semstyle.RegisterFallback("PanelBorder", true, "Helpline")
	semstyle.RegisterFallback("Helpline", true, "StatusBar")

	tf := ThemeFile{
		Styles: map[string]string{
			"StatusBar":  "{{[white:black:-]}}",
			"PanelTitle": "{{|PanelBorder|}}",
			// Helpline and PanelBorder both omitted -- two-hop chain.
		},
	}
	resolved, err := ResolveColors(tf)
	if err != nil {
		t.Fatalf("ResolveColors failed: %v", err)
	}
	want := "white:black:"
	if got := resolved["PanelTitle"]; got != want {
		t.Errorf("PanelTitle = %q, want %q", got, want)
	}
}

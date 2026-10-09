package semlg_test

import (
	"testing"

	"charm.land/lipgloss/v2"
	semlg "github.com/GhostWriters/semstyle/lg"
)

// High intensity brightens only the colors a style sets; an unset one stays
// unset rather than becoming a brightened black.
func TestHighIntensityLeavesUnsetColors(t *testing.T) {
	fgOnly := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))

	for name, got := range map[string]lipgloss.Style{
		"StyleFlags.Apply": semlg.StyleFlags{HighIntensity: true}.Apply(fgOnly),
		"CodeToStyle ::H":  semlg.CodeToStyle("::H", fgOnly, lipgloss.NewStyle()),
	} {
		if _, ok := got.GetBackground().(lipgloss.NoColor); !ok {
			t.Errorf("%s: background = %v; want unset", name, got.GetBackground())
		}
		if _, ok := got.GetForeground().(lipgloss.NoColor); ok {
			t.Errorf("%s: foreground unset; want brightened #808080", name)
		}
	}
}

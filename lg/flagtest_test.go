package semlg_test

import (
	"fmt"
	"testing"

	semstyle "github.com/GhostWriters/semstyle"
	semtheme "github.com/GhostWriters/semstyle/theme"
	semlg "github.com/GhostWriters/semstyle/lg"
	"charm.land/lipgloss/v2"
)

func TestDashDReset(t *testing.T) {
	// Direct CodeToStyle test: ::-D on a bold+reverse base (fg:bg:flags = 2 colons)
	base := lipgloss.NewStyle().Bold(true).Reverse(true)
	result := semlg.CodeToStyle("::-D", base, lipgloss.NewStyle())
	fmt.Printf("CodeToStyle ::-D   Bold=%v Reverse=%v Faint=%v\n", result.GetBold(), result.GetReverse(), result.GetFaint())
	if result.GetBold() {
		t.Error("Bold should be cleared by -D")
	}
	if result.GetReverse() {
		t.Error("Reverse should be cleared by -D")
	}
	if !result.GetFaint() {
		t.Error("Faint should be set by D after reset")
	}

	// ToStyle test via theme/parse resolution path.
	// The theme resolver pre-resolves references, so titledisabled gets stored
	// as the resolved raw code from titlemenu ("::BR") with the -D modifier merged.
	// We simulate that by using ResolveValue directly.
	rawValues := map[string]string{
		"titlemenu":     "::BR",
		"titledisabled": "{{|titlemenu:::-D|}}",
	}
	resolved, err := semtheme.ResolveValue("{{|titledisabled|}}", rawValues, map[string]bool{}, "{{|", "|}}", "{{[", "]}}")
	fmt.Printf("ResolveValue: %q err=%v\n", resolved, err)
	st := semstyle.New()
	st.SetThemeMap(map[string]string{
		"titledisabled": resolved,
	})
	expanded2 := st.ToTags("{{|titledisabled|}}")
	fmt.Printf("ToTags of resolved: %q\n", expanded2)
	direct := semlg.CodeToStyle("::D", lipgloss.NewStyle(), lipgloss.NewStyle())
	fmt.Printf("Direct ::D Faint=%v\n", direct.GetFaint())
	s := semlg.ToStyle(st, "{{|titledisabled|}}", lipgloss.NewStyle(), lipgloss.NewStyle())
	fmt.Printf("ToStyle resolved   Bold=%v Reverse=%v Faint=%v\n", s.GetBold(), s.GetReverse(), s.GetFaint())
	if s.GetBold() {
		t.Error("ToStyle: Bold should be cleared by -D modifier")
	}
	if s.GetReverse() {
		t.Error("ToStyle: Reverse should be cleared by -D modifier")
	}
	if !s.GetFaint() {
		t.Error("ToStyle: Faint should be set by D after reset")
	}
}

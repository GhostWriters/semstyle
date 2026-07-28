package semlg

import (
	"strings"
	"testing"

	"github.com/GhostWriters/semstyle"
	"charm.land/lipgloss/v2"
)

func TestToANSIOnBackgroundHardResetWithStyledContent(t *testing.T) {
	semstyle.RegisterConsoleTag("Error", "{{[red::B]}}")
	blank := lipgloss.NewStyle()

	// Content that already carries its own leading ANSI codes, matching a
	// real log line like a colored "[ERROR]" tag -- the scenario that slips
	// past a naive "already starts with an escape code" check.
	styledContent := "{{|Error|}}[ERROR]{{[-]}} something failed"

	got := ToANSIOnBackground(styledContent, blank)
	if !strings.HasPrefix(got, semstyle.CodeReset) {
		t.Errorf("ToANSIOnBackground with a blank bg and pre-styled content = %q, want it to start with an active reset code %q", got, semstyle.CodeReset)
	}
}

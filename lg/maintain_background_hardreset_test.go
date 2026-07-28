package semlg

import (
	"strings"
	"testing"

	"github.com/GhostWriters/semstyle"
	"charm.land/lipgloss/v2"
)

func TestMaintainBackgroundHardResetActivelyClears(t *testing.T) {
	blank := lipgloss.NewStyle()
	got := MaintainBackground("plain text", blank)
	if !strings.HasPrefix(got, semstyle.CodeReset) {
		t.Errorf("MaintainBackground with a blank style = %q, want it to start with an active reset code %q", got, semstyle.CodeReset)
	}
	if got == "plain text" {
		t.Errorf("MaintainBackground with a blank style returned the text unchanged -- it silently maintained whatever was already painted instead of actively resetting")
	}
}

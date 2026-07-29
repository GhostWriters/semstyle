package semlg

import (
	"strings"
	"testing"

	"github.com/GhostWriters/semstyle"
	"charm.land/lipgloss/v2"
)

// TestCodeToStyleHardResetActivelyEmitsBytes verifies that a style built
// from a hard-reset code ("~", whole or per-field) actually emits ANSI
// bytes when rendered, instead of the zero bytes a lipgloss.Style with
// nothing "set" would otherwise produce. Zero bytes lets whatever color
// preceded this text on the same output stream bleed through -- exactly
// the wrong outcome for a marker whose whole point is to defer to the
// terminal's own colors, not to silently inherit unrelated prior state.
func TestCodeToStyleHardResetActivelyEmitsBytes(t *testing.T) {
	blank := lipgloss.NewStyle()

	t.Run("whole code", func(t *testing.T) {
		got := CodeToStyle("~", blank, blank).Render("text")
		if got == "text" {
			t.Fatalf("CodeToStyle(%q) rendered with zero ANSI bytes -- prior terminal state would bleed through", "~")
		}
		if !strings.Contains(got, semstyle.CodeReset) {
			t.Errorf("CodeToStyle(%q).Render = %q, want it to contain an active reset %q", "~", got, semstyle.CodeReset)
		}
	})

	t.Run("per-field, both channels", func(t *testing.T) {
		got := CodeToStyle("~:~:", blank, blank).Render("text")
		if got == "text" {
			t.Fatalf("CodeToStyle(%q) rendered with zero ANSI bytes -- prior terminal state would bleed through", "~:~:")
		}
		if !strings.Contains(got, "\x1b[39m") || !strings.Contains(got, "\x1b[49m") {
			t.Errorf("CodeToStyle(%q).Render = %q, want it to contain both default-fg (39) and default-bg (49) resets", "~:~:", got)
		}
	})

	t.Run("per-field, foreground only", func(t *testing.T) {
		got := CodeToStyle("~:cyan:", blank, blank).Render("text")
		if !strings.Contains(got, "\x1b[39m") {
			t.Errorf("CodeToStyle(%q).Render = %q, want it to contain the default-fg reset %q", "~:cyan:", got, "\x1b[39m")
		}
	})

	t.Run("no hard reset, unaffected", func(t *testing.T) {
		got := CodeToStyle("white:cyan:", blank, blank).Render("text")
		if strings.Contains(got, "\x1b[39m") || strings.Contains(got, "\x1b[49m") {
			t.Errorf("CodeToStyle(%q).Render = %q, unexpectedly contains a default-color reset", "white:cyan:", got)
		}
	})
}

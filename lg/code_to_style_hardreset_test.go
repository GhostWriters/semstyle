package semlg

import (
	"strings"
	"testing"

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
			t.Fatalf("CodeToStyle(%q) rendered with zero bytes -- prior terminal state would bleed through", "~")
		}
		if !strings.Contains(got, hardResetFull) {
			t.Errorf("CodeToStyle(%q).Render = %q, want it to contain %q", "~", got, hardResetFull)
		}
	})

	t.Run("per-field, both channels", func(t *testing.T) {
		got := CodeToStyle("~:~:", blank, blank).Render("text")
		if got == "text" {
			t.Fatalf("CodeToStyle(%q) rendered with zero bytes -- prior terminal state would bleed through", "~:~:")
		}
		if !strings.Contains(got, hardResetFG) || !strings.Contains(got, hardResetBG) {
			t.Errorf("CodeToStyle(%q).Render = %q, want it to contain both %q and %q", "~:~:", got, hardResetFG, hardResetBG)
		}
	})

	t.Run("per-field, foreground only", func(t *testing.T) {
		got := CodeToStyle("~:cyan:", blank, blank).Render("text")
		if !strings.Contains(got, hardResetFG) {
			t.Errorf("CodeToStyle(%q).Render = %q, want it to contain %q", "~:cyan:", got, hardResetFG)
		}
	})

	t.Run("no hard reset, unaffected", func(t *testing.T) {
		got := CodeToStyle("white:cyan:", blank, blank).Render("text")
		if strings.Contains(got, hardResetFG) || strings.Contains(got, hardResetBG) {
			t.Errorf("CodeToStyle(%q).Render = %q, unexpectedly contains a hard-reset sequence", "white:cyan:", got)
		}
	})
}

// TestHardResetSurvivesWidthWrap verifies a hard-reset style's output
// survives lipgloss's own Width()-based word-wrap unmangled -- it's real,
// standards-compliant ANSI (unlike a custom out-of-band marker), so the
// same ANSI-aware logic that already skips ordinary escape codes during
// wrapping handles it correctly with no special-casing needed anywhere.
func TestHardResetSurvivesWidthWrap(t *testing.T) {
	blank := lipgloss.NewStyle()
	text := CodeToStyle("~:~:", blank, blank).Render("Update DockSTARTer2")

	wrapped := lipgloss.NewStyle().Width(10).Render(text)

	// The escape sequence's own literal characters ("39", "49", "[", "m",
	// etc.) must never appear as part of the WRAPPED VISIBLE TEXT -- i.e.
	// split across a line break by the wrapper misjudging where the escape
	// ends. Checking against the plain (tag/ANSI-stripped) rendering is the
	// reliable way to verify that, rather than guessing at exact wrap points.
	plain := StripANSI(wrapped)
	if plain != "Update    \nDockSTARTe\nr2        " {
		t.Fatalf("width-wrap did not cleanly wrap the visible text: plain = %q", plain)
	}
}

// TestHardResetPassesThroughMaintainBackgroundUntouched is the actual
// scenario that broke before: Item's rendered text (built from an explicit
// "~") gets folded into a line and reprocessed by MaintainBackground using
// the surrounding dialog's own (unrelated, e.g. accent-colored) background
// -- once directly, and again after being folded into a still-larger line,
// simulating the outer per-line box-rendering pass every dialog goes
// through. The hard reset must survive both passes without the dialog's
// color ever getting appended after it.
func TestHardResetPassesThroughMaintainBackgroundUntouched(t *testing.T) {
	blank := lipgloss.NewStyle()
	itemText := CodeToStyle("~:~:", blank, blank).Render("Configuration")

	dialog := lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("4"))
	dialogCode := dialog.Render("_")
	dialogCode = strings.Split(dialogCode, "_")[0]

	firstPass := MaintainBackground(itemText, dialog)
	secondPass := MaintainBackground(firstPass+"   ", dialog)

	if !strings.Contains(secondPass, hardResetFG) || !strings.Contains(secondPass, hardResetBG) {
		t.Fatalf("hard-reset sequence did not survive two nested MaintainBackground passes: %q", secondPass)
	}
	if idx := strings.Index(secondPass, hardResetFG); idx != -1 {
		after := secondPass[idx+len(hardResetFG):]
		if strings.HasPrefix(after, dialogCode) {
			t.Errorf("dialog's own color was appended right after the fg hard reset: %q", secondPass)
		}
	}
	if idx := strings.Index(secondPass, hardResetBG); idx != -1 {
		after := secondPass[idx+len(hardResetBG):]
		if strings.HasPrefix(after, dialogCode) {
			t.Errorf("dialog's own color was appended right after the bg hard reset: %q", secondPass)
		}
	}
}

// TestMaintainBackgroundStillMaintainsRoutineResets is the control case:
// verifies the hard-reset change didn't break MaintainBackground's actual,
// original purpose -- a routine "-" (soft reset, a bare single-parameter
// escape, not the combined hard-reset form) inside otherwise-plain text
// must still pick the ambient style's own colors back up.
func TestMaintainBackgroundStillMaintainsRoutineResets(t *testing.T) {
	ambient := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Background(lipgloss.Color("4"))
	ambientCode := ambient.Render("_")
	ambientCode = strings.Split(ambientCode, "_")[0]

	got := MaintainBackground("\x1b[0mplain text", ambient)
	if !strings.Contains(got, ambientCode) {
		t.Errorf("MaintainBackground did not re-apply the ambient style after a routine reset: %q", got)
	}
}

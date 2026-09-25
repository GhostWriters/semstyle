package semstyle

import (
	"strings"
	"testing"
)

// newTwoThemeStyler registers one theme unprefixed and a second under the
// "ct-web_" namespace, the way a host app keeps several complete themes
// registered side by side.
func newTwoThemeStyler() *Styler {
	st := New()
	st.RegisterThemeTagRaw("Title", "black:-:U")
	st.RegisterThemeTagRaw("Accent", "cyan:-:-")
	st.RegisterThemeTagRaw("ct-web_Title", "red:-:B")
	return st
}

func TestActiveThemePrefixRoutesUnprefixedLookups(t *testing.T) {
	st := newTwoThemeStyler()
	st.SetActiveThemePrefix("ct-web_")

	if got, want := st.GetRawTagCode("Title"), "red:-:B"; got != want {
		t.Errorf("GetRawTagCode(Title) = %q, want %q", got, want)
	}
	if got := st.GetColorDefinition("Title"); !strings.Contains(got, "red:-:B") {
		t.Errorf("GetColorDefinition(Title) = %q, want it to carry %q", got, "red:-:B")
	}
	if got := st.ToTags("{{|Title|}}x", ""); !strings.Contains(got, "red:-:B") {
		t.Errorf("ToTags(theme mode) = %q, want it to carry %q", got, "red:-:B")
	}

	st.SetActiveThemePrefix("")
	if got, want := st.GetRawTagCode("Title"), "black:-:U"; got != want {
		t.Errorf("after clearing, GetRawTagCode(Title) = %q, want %q", got, want)
	}
}

// TestActiveThemePrefixIsIsolated verifies a tag the active namespace
// doesn't define never borrows the unprefixed theme's value -- it goes
// straight to the console tier instead.
func TestActiveThemePrefixIsIsolated(t *testing.T) {
	st := newTwoThemeStyler()
	st.RegisterConsoleTagRaw("Accent", "green:-:-")
	st.SetActiveThemePrefix("ct-web_")

	if got, want := st.GetRawTagCode("Accent"), "green:-:-"; got != want {
		t.Errorf("GetRawTagCode(Accent) = %q, want console value %q (not the unprefixed theme's %q)", got, want, "cyan:-:-")
	}

	st.SetAutoConsoleFallback(false)
	if got := st.GetRawTagCode("Accent"); got != "" {
		t.Errorf("with console fallback off, GetRawTagCode(Accent) = %q, want \"\"", got)
	}
}

// TestExplicitPrefixStillOverlays verifies the existing overlay behavior of
// an explicitly passed prefix (a theme preview) is unchanged: an undefined
// tag falls through to the unprefixed theme.
func TestExplicitPrefixStillOverlays(t *testing.T) {
	st := newTwoThemeStyler()
	if got, want := st.GetRawTagCodeWithPrefix("Accent", "preview_"), "cyan:-:-"; got != want {
		t.Errorf("GetRawTagCodeWithPrefix(Accent, preview_) = %q, want overlay value %q", got, want)
	}
}

// TestExplicitPrefixWinsOverActive verifies an explicitly passed prefix is
// used as-is (overlay) even while an active prefix is set.
func TestExplicitPrefixWinsOverActive(t *testing.T) {
	st := newTwoThemeStyler()
	st.RegisterThemeTagRaw("preview_Title", "blue:-:-")
	st.SetActiveThemePrefix("ct-web_")

	if got, want := st.GetRawTagCodeWithPrefix("Title", "preview_"), "blue:-:-"; got != want {
		t.Errorf("GetRawTagCodeWithPrefix(Title, preview_) = %q, want %q", got, want)
	}
	if got, want := st.GetRawTagCodeWithPrefix("Accent", "preview_"), "cyan:-:-"; got != want {
		t.Errorf("GetRawTagCodeWithPrefix(Accent, preview_) = %q, want overlay value %q", got, want)
	}
}

// TestActiveThemePrefixFallbackRulesScoped verifies a fallback rule
// resolves within the active namespace, not the unprefixed theme.
func TestActiveThemePrefixFallbackRulesScoped(t *testing.T) {
	st := newTwoThemeStyler()
	st.RegisterFallback("TitleWarn", true, "Title")
	st.SetActiveThemePrefix("ct-web_")

	if got, want := st.GetRawTagCode("TitleWarn"), "red:-:B"; got != want {
		t.Errorf("GetRawTagCode(TitleWarn) = %q, want the active namespace's Title %q", got, want)
	}
}

// TestActiveThemePrefixConsoleOnlyUnaffected verifies console-mode
// expansion (ToTags with no prefix argument) ignores the active prefix.
func TestActiveThemePrefixConsoleOnlyUnaffected(t *testing.T) {
	st := newTwoThemeStyler()
	st.RegisterConsoleTagRaw("Notice", "green:-:-")
	st.SetActiveThemePrefix("ct-web_")

	if got := st.ToTags("{{|Notice|}}x"); !strings.Contains(got, "green:-:-") {
		t.Errorf("ToTags(console mode) = %q, want it to carry %q", got, "green:-:-")
	}
}

func TestRunWithRenderScopeRestores(t *testing.T) {
	prevTint, prevTheme := ActiveTintKey(), ActiveThemePrefix()
	t.Cleanup(func() {
		SetActiveTint(prevTint)
		SetActiveThemePrefix(prevTheme)
	})
	SetActiveTint("outer-tint")
	SetActiveThemePrefix("outer_")

	RunWithRenderScope("inner-tint", "ct-web_", func() {
		if got := ActiveTintKey(); got != "inner-tint" {
			t.Errorf("inside scope, ActiveTintKey() = %q, want %q", got, "inner-tint")
		}
		if got := ActiveThemePrefix(); got != "ct-web_" {
			t.Errorf("inside scope, ActiveThemePrefix() = %q, want %q", got, "ct-web_")
		}
	})

	if got := ActiveTintKey(); got != "outer-tint" {
		t.Errorf("after scope, ActiveTintKey() = %q, want %q", got, "outer-tint")
	}
	if got := ActiveThemePrefix(); got != "outer_" {
		t.Errorf("after scope, ActiveThemePrefix() = %q, want %q", got, "outer_")
	}
}

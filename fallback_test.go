package semstyle

import (
	"testing"
)

// TestRegisterFallbackGetRawTagCode verifies GetRawTagCode consults a
// registered fallback when the tag itself isn't defined.
func TestRegisterFallbackGetRawTagCode(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Title", "black:-:U")
	st.RegisterFallback("TitleWarn", "Title")

	got := st.GetRawTagCode("TitleWarn")
	want := "black:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(TitleWarn) = %q, want %q", got, want)
	}
}

// TestRegisterFallbackDoesNotOverrideDefined verifies a tag's own value wins
// over its fallback when both are defined.
func TestRegisterFallbackDoesNotOverrideDefined(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Title", "black:-:U")
	st.RegisterThemeTagRaw("TitleWarn", "yellow:-:U")
	st.RegisterFallback("TitleWarn", "Title")

	got := st.GetRawTagCode("TitleWarn")
	want := "yellow:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(TitleWarn) = %q, want %q (own value should win over fallback)", got, want)
	}
}

// TestRegisterFallbackChain verifies a fallback chain (A -> B -> C) resolves
// through multiple hops when only the final tag is actually defined.
func TestRegisterFallbackChain(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Title", "black:-:U")
	st.RegisterFallback("TitleFocused", "Title")
	st.RegisterFallback("TitleSubMenuFocused", "TitleFocused")

	got := st.GetRawTagCode("TitleSubMenuFocused")
	want := "black:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(TitleSubMenuFocused) = %q, want %q", got, want)
	}
}

// TestRegisterFallbackCycleGuard verifies a fallback cycle resolves to
// unset rather than hanging.
func TestRegisterFallbackCycleGuard(t *testing.T) {
	st := New()
	st.RegisterFallback("A", "B")
	st.RegisterFallback("B", "A")

	got := st.GetRawTagCode("A")
	if got != "" {
		t.Errorf("GetRawTagCode(A) = %q, want empty (cycle should not resolve)", got)
	}
}

// TestRegisterFallbackInlineTagExpansion verifies a fallback applies to
// inline "{{|name|}}" text expansion, not just GetRawTagCode -- the whole
// point being that a fallback rule applies wherever a tag can appear.
func TestRegisterFallbackInlineTagExpansion(t *testing.T) {
	st := New()
	st.RegisterThemeTag("Title", "{{[black::U]}}")
	st.RegisterFallback("TitleWarn", "Title")

	got := st.ToTags("{{|TitleWarn|}}", "")
	want := "{{[black::U]}}"
	if got != want {
		t.Errorf("ToTags(TitleWarn) = %q, want %q", got, want)
	}
}

// TestClearThemeMapPreservesFallbacks verifies ClearThemeMap does NOT clear
// registered fallback rules -- they're a structural relationship in the
// caller's tag-naming scheme, not per-theme state, so they should survive a
// theme reload/switch. A newly loaded theme's own value for Title is picked
// up automatically since the fallback is resolved lazily.
func TestClearThemeMapPreservesFallbacks(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Title", "black:-:U")
	st.RegisterFallback("TitleWarn", "Title")
	st.ClearThemeMap()

	if got := st.GetRawTagCode("TitleWarn"); got != "" {
		t.Errorf("GetRawTagCode(TitleWarn) after ClearThemeMap = %q, want empty (Title itself was cleared, so its fallback resolves to nothing until a new theme defines it)", got)
	}

	// Simulate the next theme load defining Title again -- the fallback
	// rule registered before either load should still apply.
	st.RegisterThemeTagRaw("Title", "yellow:-:U")
	if got, want := st.GetRawTagCode("TitleWarn"), "yellow:-:U"; got != want {
		t.Errorf("GetRawTagCode(TitleWarn) after reloading Title = %q, want %q (fallback rule should have survived ClearThemeMap)", got, want)
	}
}

// TestClearFallbacks verifies ClearFallbacks removes rules without touching
// the theme map itself.
func TestClearFallbacks(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Title", "black:-:U")
	st.RegisterFallback("TitleWarn", "Title")
	st.ClearFallbacks()

	if got := st.GetRawTagCode("TitleWarn"); got != "" {
		t.Errorf("GetRawTagCode(TitleWarn) after ClearFallbacks = %q, want empty", got)
	}
	if got := st.GetRawTagCode("Title"); got != "black:-:U" {
		t.Errorf("GetRawTagCode(Title) after ClearFallbacks = %q, want %q (theme map should be untouched)", got, "black:-:U")
	}
}

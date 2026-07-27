package semstyle

import (
	"testing"
)

// TestRegisterFallbackGetRawTagCode verifies GetRawTagCode consults a
// registered fallback when the tag itself isn't defined.
func TestRegisterFallbackGetRawTagCode(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Title", "black:-:U")
	st.RegisterFallback("TitleWarn", true, "Title")

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
	st.RegisterFallback("TitleWarn", true, "Title")

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
	st.RegisterFallback("TitleFocused", true, "Title")
	st.RegisterFallback("TitleSubMenuFocused", true, "TitleFocused")

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
	st.RegisterFallback("A", true, "B")
	st.RegisterFallback("B", true, "A")

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
	st.RegisterFallback("TitleWarn", true, "Title")

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
	st.RegisterFallback("TitleWarn", true, "Title")
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
	st.RegisterFallback("TitleWarn", true, "Title")
	st.ClearFallbacks()

	if got := st.GetRawTagCode("TitleWarn"); got != "" {
		t.Errorf("GetRawTagCode(TitleWarn) after ClearFallbacks = %q, want empty", got)
	}
	if got := st.GetRawTagCode("Title"); got != "black:-:U" {
		t.Errorf("GetRawTagCode(Title) after ClearFallbacks = %q, want %q (theme map should be untouched)", got, "black:-:U")
	}
}

// TestRegisterFallbackMultipleTriesInOrder verifies multiple candidates are
// tried in order, using the first one that resolves.
func TestRegisterFallbackMultipleTriesInOrder(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("B", "yellow:-:U")
	st.RegisterThemeTagRaw("C", "green:-:U")
	st.RegisterFallback("X", false, "A", "B", "C")

	got := st.GetRawTagCode("X")
	want := "yellow:-:U" // A undefined, B is the first that resolves
	if got != want {
		t.Errorf("GetRawTagCode(X) = %q, want %q", got, want)
	}
}

// TestRegisterFallbackNoChain verifies that with followChains=false, a
// candidate's own separate fallback rule is NOT consulted -- only its
// direct theme/console value counts.
func TestRegisterFallbackNoChain(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Root", "cyan:-:U")
	st.RegisterFallback("A", true, "Root") // A's own independent fallback, unrelated to X's list
	st.RegisterFallback("X", false, "A", "B")

	got := st.GetRawTagCode("X")
	if got != "" {
		t.Errorf("GetRawTagCode(X) = %q, want empty (A isn't directly defined; its own fallback shouldn't be consulted from X's list)", got)
	}

	// Confirm A does resolve on its own -- proves the fallback rule itself
	// works, just isn't reached via X's list.
	if got := st.GetRawTagCode("A"); got != "cyan:-:U" {
		t.Errorf("GetRawTagCode(A) directly = %q, want %q", got, "cyan:-:U")
	}
}

// TestRegisterFallbackMultipleFollowChains verifies that with
// followChains=true and multiple candidates, a candidate's own fallback
// rule IS consulted.
func TestRegisterFallbackMultipleFollowChains(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Root", "cyan:-:U")
	st.RegisterFallback("A", true, "Root")
	st.RegisterFallback("X", true, "A", "B")

	got := st.GetRawTagCode("X")
	want := "cyan:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(X) = %q, want %q (A isn't directly defined, but its own fallback to Root should be followed)", got, want)
	}
}

// TestRegisterFallbackReplacesPreviousRule verifies registering again for
// the same name replaces its previous rule entirely.
func TestRegisterFallbackReplacesPreviousRule(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Old", "red:-:U")
	st.RegisterThemeTagRaw("New", "blue:-:U")
	st.RegisterFallback("X", true, "Old")
	st.RegisterFallback("X", false, "New")

	got := st.GetRawTagCode("X")
	want := "blue:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(X) = %q, want %q (second RegisterFallback call should have replaced the first)", got, want)
	}
}

// TestRegisterFallbackDirectLiteral verifies a candidate written with
// direct-tag syntax ("{{[fg:bg:flags]}}") is used as a literal raw code,
// with no lookup.
func TestRegisterFallbackDirectLiteral(t *testing.T) {
	st := New()
	st.RegisterFallback("TitleWarn", true, "{{[white:black:B]}}")

	got := st.GetRawTagCode("TitleWarn")
	want := "white:black:B"
	if got != want {
		t.Errorf("GetRawTagCode(TitleWarn) = %q, want %q", got, want)
	}
}

// TestRegisterFallbackDirectLiteralIgnoresOwnDefinition verifies a literal
// candidate is used verbatim even if a tag happens to be registered under
// that same raw text as a name -- it's never looked up as a name at all.
func TestRegisterFallbackDirectLiteralIgnoresOwnDefinition(t *testing.T) {
	st := New()
	// Deliberately register something under the literal's own text as a
	// name, to prove it's never consulted.
	st.RegisterThemeTagRaw("white:black:B", "red:-:U")
	st.RegisterFallback("TitleWarn", true, "{{[white:black:B]}}")

	got := st.GetRawTagCode("TitleWarn")
	want := "white:black:B"
	if got != want {
		t.Errorf("GetRawTagCode(TitleWarn) = %q, want %q (literal candidate should never be looked up as a name)", got, want)
	}
}

// TestRegisterFallbackSemanticTagSyntaxEquivalent verifies a candidate
// written with semantic-tag syntax ("{{|Name|}}") behaves identically to
// passing the bare name.
func TestRegisterFallbackSemanticTagSyntaxEquivalent(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Title", "black:-:U")
	st.RegisterFallback("TitleWarn", true, "{{|Title|}}")

	got := st.GetRawTagCode("TitleWarn")
	want := "black:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(TitleWarn) = %q, want %q", got, want)
	}
}

// TestRegisterFallbackLiteralAmongNamedCandidates verifies a literal
// candidate can act as a final catch-all at the end of an ordered list of
// named candidates.
func TestRegisterFallbackLiteralAmongNamedCandidates(t *testing.T) {
	st := New()
	st.RegisterFallback("X", false, "A", "B", "{{[gray::D]}}")

	got := st.GetRawTagCode("X")
	want := "gray::D"
	if got != want {
		t.Errorf("GetRawTagCode(X) = %q, want %q (A and B undefined, should fall through to the literal)", got, want)
	}
}

// TestAutoConsoleFallbackDefaultEnabled verifies console fallback is
// automatic by default -- the library's original behavior, unchanged.
func TestAutoConsoleFallbackDefaultEnabled(t *testing.T) {
	st := New()
	if !st.AutoConsoleFallback() {
		t.Error("AutoConsoleFallback() = false on a new Styler, want true (default)")
	}
	st.RegisterConsoleTagRaw("title", "cyan:-:U")

	got := st.GetRawTagCode("Title") // undefined in theme map
	want := "cyan:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(Title) = %q, want %q (should fall through to console by default)", got, want)
	}
}

// TestSetAutoConsoleFallbackDisabled verifies disabling the toggle stops
// theme-mode lookups from silently picking up an unrelated console-map
// default when the theme itself doesn't define the tag.
func TestSetAutoConsoleFallbackDisabled(t *testing.T) {
	st := New()
	st.SetAutoConsoleFallback(false)
	if st.AutoConsoleFallback() {
		t.Error("AutoConsoleFallback() = true after SetAutoConsoleFallback(false)")
	}
	st.RegisterConsoleTagRaw("title", "cyan:-:U")

	got := st.GetRawTagCode("Title") // undefined in theme map
	if got != "" {
		t.Errorf("GetRawTagCode(Title) = %q, want empty (auto console fallback disabled)", got)
	}
}

// TestSetAutoConsoleFallbackDisabledDoesNotBreakConsoleModeCalls verifies
// disabling the toggle has no effect on true console-mode calls (ToANSI/
// ToTags with no prefix), which target the console map directly and were
// never routed through the automatic secondary tier in the first place.
func TestSetAutoConsoleFallbackDisabledDoesNotBreakConsoleModeCalls(t *testing.T) {
	st := New()
	st.SetAutoConsoleFallback(false)
	st.RegisterConsoleTag("Notice", "{{[cyan::B]}}")

	got := st.ToTags("{{|Notice|}}") // no prefix = console-mode
	want := "{{[cyan::B]}}"
	if got != want {
		t.Errorf("ToTags(Notice) console-mode = %q, want %q", got, want)
	}
}

// TestConsoleTagExplicitOptIn verifies ConsoleTag lets a specific tag fall
// back to the console map even after SetAutoConsoleFallback(false).
func TestConsoleTagExplicitOptIn(t *testing.T) {
	st := New()
	st.SetAutoConsoleFallback(false)
	st.RegisterConsoleTagRaw("title", "cyan:-:U")
	st.RegisterFallback("TitleWarn", true, ConsoleTag("Title"))

	got := st.GetRawTagCode("TitleWarn")
	want := "cyan:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(TitleWarn) = %q, want %q (ConsoleTag should reach the console map despite the tier being disabled)", got, want)
	}
}

// TestConsoleTagIgnoresThemeValue verifies a ConsoleTag candidate only
// consults the console map, even if the same name also has a theme-map
// value -- it's a deliberate console-only lookup, not theme-first.
func TestConsoleTagIgnoresThemeValue(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("title", "red:-:U")
	st.RegisterConsoleTagRaw("title", "cyan:-:U")
	st.RegisterFallback("TitleWarn", true, ConsoleTag("Title"))

	got := st.GetRawTagCode("TitleWarn")
	want := "cyan:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(TitleWarn) = %q, want %q (ConsoleTag should use the console value, not the theme value)", got, want)
	}
}

// TestConsoleTagNoAmbiguityWithModifierSyntax verifies the ConsoleTag
// marker can't be confused with a plain "console:Name"-style string --
// registering a literal string that happens to look like that should NOT
// be treated as a console-only candidate.
func TestConsoleTagNoAmbiguityWithModifierSyntax(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("console:title", "green:-:U")
	st.RegisterFallback("X", false, "console:Title")

	got := st.GetRawTagCode("X")
	want := "green:-:U"
	if got != want {
		t.Errorf("GetRawTagCode(X) = %q, want %q (a plain string candidate, even one that looks like \"console:Title\", should be treated as a literal tag name, not the ConsoleTag marker)", got, want)
	}
}

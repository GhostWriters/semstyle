package semstyle

import (
	"strings"
	"testing"
)

// TestRegisterConsoleTagMultiPart verifies that a multi-tag value registered via
// RegisterConsoleTag (e.g. "{{[-]}}{{[gray::D]}}") round-trips correctly through
// ToTags without being mangled into a single re-wrapped direct tag.
func TestRegisterConsoleTagMultiPart(t *testing.T) {
	st := New()
	st.RegisterConsoleTag("timestamp", "{{[-]}}{{[gray::D]}}")

	got := st.ToTags("{{|timestamp|}}")
	want := "{{[-]}}{{[gray::D]}}"
	if got != want {
		t.Errorf("ToTags multi-part console tag: got %q, want %q", got, want)
	}
}

// TestRegisterConsoleTagSinglePart verifies that a single-tag value still strips
// and re-wraps correctly (the normal path is unchanged).
func TestRegisterConsoleTagSinglePart(t *testing.T) {
	st := New()
	st.RegisterConsoleTag("notice", "{{[green]}}")

	got := st.ToTags("{{|notice|}}")
	want := "{{[green]}}"
	if got != want {
		t.Errorf("ToTags single-part console tag: got %q, want %q", got, want)
	}
}

// TestRegisterThemeTagMultiPart verifies the same round-trip for the theme map.
func TestRegisterThemeTagMultiPart(t *testing.T) {
	st := New()
	st.RegisterThemeTag("timestamp", "{{[-]}}{{[gray::D]}}")

	got := st.ToTags("{{|timestamp|}}", "")
	want := "{{[-]}}{{[gray::D]}}"
	if got != want {
		t.Errorf("ToTags multi-part theme tag: got %q, want %q", got, want)
	}
}

// TestToTagsEmptyPrefixConsoleFallback verifies that ToTags with an empty prefix
// (theme map + console fallback) correctly falls back to the console map when the
// tag is not in the theme map.
func TestToTagsEmptyPrefixConsoleFallback(t *testing.T) {
	st := New()
	st.RegisterConsoleTag("timestamp", "{{[-]}}{{[gray::D]}}")

	// No prefix: console map only — must work.
	got := st.ToTags("{{|timestamp|}}")
	want := "{{[-]}}{{[gray::D]}}"
	if got != want {
		t.Errorf("ToTags no prefix: got %q, want %q", got, want)
	}

	// Empty prefix: theme map with console fallback — must also find it.
	got2 := st.ToTags("{{|timestamp|}}", "")
	if got2 != want {
		t.Errorf("ToTags empty prefix console fallback: got %q, want %q", got2, want)
	}
}

// TestRegisterConsoleTagMultiPartWithFlags verifies a multi-part value that includes
// flag modifiers (bold, dim) also round-trips correctly.
func TestRegisterConsoleTagMultiPartWithFlags(t *testing.T) {
	st := New()
	st.RegisterConsoleTag("error", "{{[-]}}{{[red::B]}}")

	got := st.ToTags("{{|error|}}")
	want := "{{[-]}}{{[red::B]}}"
	if got != want {
		t.Errorf("ToTags multi-part with flags: got %q, want %q", got, want)
	}
}

// TestRegisterConsoleTagMultiPartFgBg verifies multi-part values with both fg and bg.
func TestRegisterConsoleTagMultiPartFgBg(t *testing.T) {
	st := New()
	st.RegisterConsoleTag("fatal", "{{[-]}}{{[white:red]}}")

	got := st.ToTags("{{|fatal|}}")
	want := "{{[-]}}{{[white:red]}}"
	if got != want {
		t.Errorf("ToTags multi-part fg+bg: got %q, want %q", got, want)
	}
}

// TestReplaceThemeTagsNilKeepClearsFirst verifies a nil keep removes every
// existing entry before populate runs, same as ClearThemeMap followed by
// registration.
func TestReplaceThemeTagsNilKeepClearsFirst(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("stale", "{{[red]}}")

	st.ReplaceThemeTags(nil, func(register func(name, rawValue string)) {
		register("fresh", "{{[green]}}")
	})

	if got := st.GetRawTagCode("stale"); got != "" {
		t.Errorf("stale tag survived a nil-keep ReplaceThemeTags: got %q, want empty", got)
	}
	if got := st.GetRawTagCode("fresh"); got != "{{[green]}}" {
		t.Errorf("fresh tag: got %q, want %q", got, "{{[green]}}")
	}
}

// TestReplaceThemeTagsKeepPredicate verifies keep is consulted per existing
// key: entries it returns false for are removed, entries it returns true
// for survive untouched, and populate's registrations land alongside them.
func TestReplaceThemeTagsKeepPredicate(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("preview_old", "{{[red]}}")
	st.RegisterThemeTagRaw("other", "{{[blue]}}")

	st.ReplaceThemeTags(
		func(key string) bool { return !strings.HasPrefix(key, "preview_") },
		func(register func(name, rawValue string)) {
			register("preview_new", "{{[green]}}")
		},
	)

	if got := st.GetRawTagCode("preview_old"); got != "" {
		t.Errorf("preview_old should have been removed by keep: got %q", got)
	}
	if got := st.GetRawTagCode("other"); got != "{{[blue]}}" {
		t.Errorf("other should have survived keep: got %q, want %q", got, "{{[blue]}}")
	}
	if got := st.GetRawTagCode("preview_new"); got != "{{[green]}}" {
		t.Errorf("preview_new: got %q, want %q", got, "{{[green]}}")
	}
}

// TestReplaceThemeTagsWithPrefix verifies it matches the same keys
// UnregisterPrefix would (case-insensitive, "_"-terminated), leaving
// unrelated entries alone.
func TestReplaceThemeTagsWithPrefix(t *testing.T) {
	st := New()
	st.RegisterThemeTagRaw("Preview_Old", "{{[red]}}")
	st.RegisterThemeTagRaw("other", "{{[blue]}}")

	st.ReplaceThemeTagsWithPrefix("Preview", func(register func(name, rawValue string)) {
		register("Preview_New", "{{[green]}}")
	})

	if got := st.GetRawTagCode("preview_old"); got != "" {
		t.Errorf("preview_old should have been removed: got %q", got)
	}
	if got := st.GetRawTagCode("other"); got != "{{[blue]}}" {
		t.Errorf("other should have survived: got %q, want %q", got, "{{[blue]}}")
	}
	if got := st.GetRawTagCode("preview_new"); got != "{{[green]}}" {
		t.Errorf("preview_new: got %q, want %q", got, "{{[green]}}")
	}
}

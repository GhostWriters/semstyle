package semstyle

import (
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

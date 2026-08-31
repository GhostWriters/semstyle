package semstyle

import (
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

func TestHyperlinkModeInlineDefault(t *testing.T) {
	SetPreferredProfile(colorprofile.TrueColor)
	HyperlinkModeFunc = nil
	st := New()

	r := st.ToANSI(`{{|File::::https://dockstarter.com|}}dockstarter.com{{[-]}}`)
	if !strings.Contains(r, "\x1b]8;") {
		t.Errorf("expected OSC8 hyperlink, got %q", r)
	}
	if strings.Contains(r, "(https://dockstarter.com)") {
		t.Errorf("inline mode should not show the URL, got %q", r)
	}
}

func TestHyperlinkModeOff(t *testing.T) {
	SetPreferredProfile(colorprofile.TrueColor)
	HyperlinkModeFunc = func() HyperlinkMode { return HyperlinkModeOff }
	defer func() { HyperlinkModeFunc = nil }()
	st := New()

	r := st.ToANSI(`{{|File::::https://dockstarter.com|}}dockstarter.com{{[-]}}`)
	if strings.Contains(r, "\x1b]8;") {
		t.Errorf("off mode should emit no OSC8 escape, got %q", r)
	}
	if !strings.Contains(r, "dockstarter.com") {
		t.Errorf("off mode should still show the label text, got %q", r)
	}
}

func TestHyperlinkModeAuto(t *testing.T) {
	SetPreferredProfile(colorprofile.TrueColor)
	HyperlinkModeFunc = func() HyperlinkMode { return HyperlinkModeAuto }
	defer func() { HyperlinkModeFunc = nil }()
	st := New()

	r := st.ToANSI(`{{|File::::https://dockstarter.com|}}DockSTARTer{{[-]}}`)
	if !strings.Contains(r, "DockSTARTer") {
		t.Errorf("auto mode should show the label, got %q", r)
	}
	if !strings.Contains(r, "https://dockstarter.com") {
		t.Errorf("auto mode should show the URL, got %q", r)
	}
	if !strings.Contains(r, " (") || !strings.Contains(r, ")") {
		t.Errorf("auto mode should wrap the URL in plain parens, got %q", r)
	}
	openParen := strings.Index(r, " (")
	linkOpen := strings.Index(r, "\x1b]8;")
	closeParen := strings.LastIndex(r, ")")
	linkClose := strings.LastIndex(r, "\x1b]8;;\a")
	if openParen < 0 || linkOpen < openParen {
		t.Errorf("the opening paren should come before the OSC8 link, got %q", r)
	}
	if closeParen < 0 || linkClose < 0 || closeParen < linkClose {
		t.Errorf("the closing paren should come after the OSC8 link closes, got %q", r)
	}
	if strings.Count(r, "\x1b]8;") != 2 {
		// Exactly one OSC8 open marker and one close marker: only the parenthesized URL
		// is clickable, the label is plain styled text (not a second, redundant link).
		t.Errorf("auto mode should OSC8-wrap only the URL, not the label, got %q", r)
	}
	labelStart := strings.Index(r, "DockSTARTer")
	linkStart := strings.Index(r, "\x1b]8;")
	if labelStart < 0 || linkStart < 0 || labelStart > linkStart {
		t.Errorf("auto mode's label should appear before its OSC8 link, got %q", r)
	}
}

func TestHyperlinkModeSetOnStylerOverridesGlobalFunc(t *testing.T) {
	SetPreferredProfile(colorprofile.TrueColor)
	HyperlinkModeFunc = func() HyperlinkMode { return HyperlinkModeInline }
	defer func() { HyperlinkModeFunc = nil }()

	st := New()
	st.SetHyperlinkMode(HyperlinkModeAuto)

	r := st.ToANSI(`{{|File::::https://dockstarter.com|}}DockSTARTer{{[-]}}`)
	if !strings.Contains(r, " (") || !strings.Contains(r, "https://dockstarter.com") {
		t.Errorf("per-Styler override should win over the global HyperlinkModeFunc, got %q", r)
	}

	// The Default styler (and any other Styler) must be unaffected by st's override.
	other := New()
	r2 := other.ToANSI(`{{|File::::https://dockstarter.com|}}DockSTARTer{{[-]}}`)
	if strings.Contains(r2, " (") {
		t.Errorf("a Styler without SetHyperlinkMode should still follow HyperlinkModeFunc, got %q", r2)
	}

	st.ClearHyperlinkMode()
	r3 := st.ToANSI(`{{|File::::https://dockstarter.com|}}DockSTARTer{{[-]}}`)
	if strings.Contains(r3, " (") {
		t.Errorf("ClearHyperlinkMode should revert to HyperlinkModeFunc (Inline here), got %q", r3)
	}
}

func TestHyperlinkModeAutoLocationOnlyFlagDowngradesToInline(t *testing.T) {
	SetPreferredProfile(colorprofile.TrueColor)
	HyperlinkModeFunc = func() HyperlinkMode { return HyperlinkModeAuto }
	defer func() { HyperlinkModeFunc = nil }()
	st := New()

	// A path-segment-style tag: label is just a piece of the location ("clhatch"), not
	// separate descriptive text -- flags field carries the "N" locationOnly marker.
	r := st.ToANSI(`{{|Folder:::N:file:///home/clhatch|}}clhatch{{[-]}}`)
	if strings.Contains(r, " (") {
		t.Errorf("locationOnly flag should suppress auto's url annotation, got %q", r)
	}
	if !strings.Contains(r, "clhatch") || !strings.Contains(r, "\x1b]8;") {
		t.Errorf("locationOnly flag should still render as a normal Inline hyperlink, got %q", r)
	}

	// Off must still win over the locationOnly flag.
	HyperlinkModeFunc = func() HyperlinkMode { return HyperlinkModeOff }
	r2 := st.ToANSI(`{{|Folder:::N:file:///home/clhatch|}}clhatch{{[-]}}`)
	if strings.Contains(r2, "\x1b]8;") {
		t.Errorf("off mode should still suppress the link even with locationOnly set, got %q", r2)
	}
}

func TestHyperlinkModeOffOnRegisteredTag(t *testing.T) {
	SetPreferredProfile(colorprofile.TrueColor)
	HyperlinkModeFunc = func() HyperlinkMode { return HyperlinkModeOff }
	defer func() { HyperlinkModeFunc = nil }()
	st := New()
	st.RegisterConsoleTag("mylink", "cyan::U")
	st.RegisterHyperlinkTag("mylink")

	r := st.ToANSI(`{{|mylink|}}https://dockstarter.com{{[-]}}`)
	if strings.Contains(r, "\x1b]8;") {
		t.Errorf("off mode should emit no OSC8 escape for registered tags either, got %q", r)
	}
	if !strings.Contains(r, "https://dockstarter.com") {
		t.Errorf("off mode should still show the content text, got %q", r)
	}
}

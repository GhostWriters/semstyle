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
	if !strings.Contains(r, "(https://dockstarter.com)") {
		t.Errorf("auto mode should show the parenthesized URL, got %q", r)
	}
	if strings.Count(r, "\x1b]8;") < 4 {
		// Two OSC8 open markers (label + url) and two close markers expected.
		t.Errorf("auto mode should OSC8-wrap both the label and the URL, got %q", r)
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

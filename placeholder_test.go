package semstyle

import (
	"image/color"
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// withDeferredTints runs fn with deferred tints on and active tint key key.
func withDeferredTints(t *testing.T, key string, fn func()) {
	t.Helper()
	SetDeferTints(true)
	defer SetDeferTints(false)
	RunWithTint(key, fn)
}

func TestDeferredTintsUntintedResolvesToANSI(t *testing.T) {
	var out string
	withDeferredTints(t, "", func() {
		out = ToANSI("{{[cyan:blue]}}x{{[-]}}")
	})
	if !strings.Contains(out, ";2;1;") {
		t.Fatalf("expected a placeholder, got %q", out)
	}
	got := ResolveTints(out, "")
	if !strings.Contains(got, "\x1b[36m") || !strings.Contains(got, "\x1b[44m") {
		t.Errorf("untinted placeholders should resolve to plain ANSI codes, got %q", got)
	}
}

func TestDeferredTintsFollowCurrentTint(t *testing.T) {
	const key = "test-deferred-follow"
	defer UnregisterTint(key)
	var out string
	withDeferredTints(t, key, func() {
		out = ToANSI("{{[cyan]}}x{{[-]}}")
	})

	if got := ResolveTints(out, ""); !strings.Contains(got, "\x1b[36m") {
		t.Errorf("before the tint is registered, want plain cyan, got %q", got)
	}
	RegisterTint(key, Palette{Cyan: "#102030"})
	if got := ResolveTints(out, ""); !strings.Contains(got, "38;2;16;32;48") {
		t.Errorf("after registering, want the tint's cyan, got %q", got)
	}
	RegisterTint(key, Palette{Cyan: "#405060"})
	if got := ResolveTints(out, ""); !strings.Contains(got, "38;2;64;80;96") {
		t.Errorf("after changing the tint, want its new cyan, got %q", got)
	}
}

func TestDeferredTintsDefaultKey(t *testing.T) {
	const key = "test-deferred-default"
	RegisterTint(key, Palette{Red: "#aa0000"})
	defer UnregisterTint(key)
	var out string
	withDeferredTints(t, "", func() {
		out = ToANSI("{{[red]}}x{{[-]}}")
	})
	if got := ResolveTints(out, key); !strings.Contains(got, "38;2;170;0;0") {
		t.Errorf("palette ID 0 should resolve with the default key, got %q", got)
	}
}

func TestDeferredTintsMixedPalettes(t *testing.T) {
	RegisterTint("test-mix-a", Palette{Green: "#00aa00"})
	RegisterTint("test-mix-b", Palette{Green: "#00bb00"})
	defer UnregisterTint("test-mix-a")
	defer UnregisterTint("test-mix-b")
	var a, b string
	withDeferredTints(t, "test-mix-a", func() { a = ToANSI("{{[green]}}a{{[-]}}") })
	withDeferredTints(t, "test-mix-b", func() { b = ToANSI("{{[green]}}b{{[-]}}") })
	got := ResolveTints(a+b, "")
	if !strings.Contains(got, "38;2;0;170;0") || !strings.Contains(got, "38;2;0;187;0") {
		t.Errorf("each region should resolve with its own tint, got %q", got)
	}
}

func TestDeferredTintsExtraSlotFallback(t *testing.T) {
	var out string
	withDeferredTints(t, "", func() {
		out = ToANSI("{{[base01:base02]}}x{{[-]}}")
	})
	got := ResolveTints(out, "")
	if !strings.Contains(got, "\x1b[90m") || !strings.Contains(got, "\x1b[100m") {
		t.Errorf("untinted base01/base02 should resolve to bright black, got %q", got)
	}
}

func TestResolveTintsKeepsOtherParams(t *testing.T) {
	const key = "test-deferred-params"
	RegisterTint(key, Palette{Blue: "#0000cc"})
	defer UnregisterTint(key)
	hex, ok := placeholderHex(key, "blue")
	if !ok {
		t.Fatal("blue should have a placeholder")
	}
	r, g, b, _ := parseHex(hex)
	in := "\x1b[1;48;2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b) + ";4mx\x1b[0m\x1b[2J"
	want := "\x1b[1;48;2;0;0;204;4mx\x1b[0m\x1b[2J"
	if got := ResolveTints(in, ""); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDeferredTintsLiteralColorNudged(t *testing.T) {
	var c color.Color
	withDeferredTints(t, "", func() {
		c = ToColor("#01c0e5")
	})
	if got := ToColorStr(c); got != "#02c0e5" {
		t.Errorf("a literal color in the placeholder range should be nudged, got %s", got)
	}
}

func TestNearBlackIsNotAPlaceholder(t *testing.T) {
	for _, in := range []string{"\x1b[38;2;1;0;0mx", "\x1b[48;2;1;1;1mx", "\x1b[38;2;1;2;3mx"} {
		if got := ResolveTints(in, ""); got != in {
			t.Errorf("near-black truecolor should pass through, %q became %q", in, got)
		}
	}
}

func TestDeferredTintsToColorAndResolveColor(t *testing.T) {
	const key = "test-deferred-color"
	RegisterTint(key, Palette{Magenta: "#880088"})
	defer UnregisterTint(key)
	var c, resolved color.Color
	withDeferredTints(t, key, func() {
		c = ToColor("magenta")
		resolved = ResolveColor(c)
	})
	if _, _, ok := decodeColor(c); !ok {
		t.Fatalf("ToColor should return a placeholder, got %s", ToColorStr(c))
	}
	if got := ToColorStr(resolved); got != "#880088" {
		t.Errorf("ResolveColor should return the tint's color, got %s", got)
	}
	if got := ToColorStr(ResolveColor(lipgloss.Color("#123456"))); got != "#123456" {
		t.Errorf("ResolveColor should leave other colors alone, got %s", got)
	}
}

func TestTintSnapshotKeepsOnePalette(t *testing.T) {
	const key = "test-deferred-snapshot"
	RegisterTint(key, Palette{Blue: "#000011"})
	defer UnregisterTint(key)
	id, _ := paletteID(key)
	snap := newTintSnapshot("")
	if got := snap.slot(id, 4); got != "#000011" {
		t.Fatalf("want the registered blue, got %q", got)
	}
	RegisterTint(key, Palette{Blue: "#000022"})
	if got := snap.slot(id, 4); got != "#000011" {
		t.Errorf("a snapshot should keep the palette it first saw, got %q", got)
	}
}

func TestResolveColorWithDefault(t *testing.T) {
	const key = "test-deferred-colordefault"
	RegisterTint(key, Palette{Red: "#aa1111"})
	defer UnregisterTint(key)
	var c color.Color
	withDeferredTints(t, "", func() { c = ToColor("red") })
	if got := ToColorStr(ResolveColorWithDefault(c, key)); got != "#aa1111" {
		t.Errorf("a keyless placeholder should resolve with the default key, got %s", got)
	}
	if got, want := ToColorStr(ResolveColorWithDefault(c, "")), ToColorStr(lipgloss.Color("1")); got != want {
		t.Errorf("with no default key it should resolve to plain red, got %s", got)
	}
}

func TestBrightPlaceholder(t *testing.T) {
	var c color.Color
	withDeferredTints(t, "", func() { c = ToColor("yellow") })
	bright, ok := BrightPlaceholder(c)
	if !ok {
		t.Fatal("yellow's placeholder should be recognized")
	}
	if _, slot, _ := decodeColor(bright); slot != 11 {
		t.Errorf("want bright yellow's slot 11, got %d", slot)
	}
	if _, ok := BrightPlaceholder(lipgloss.Color("#123456")); ok {
		t.Error("a plain color isn't a placeholder")
	}
}

func TestDeferTintsOffUnchanged(t *testing.T) {
	if got := ToANSI("{{[cyan]}}x{{[-]}}"); strings.Contains(got, ";2;1;") {
		t.Errorf("deferral off should render no placeholders, got %q", got)
	}
}

func decodeColor(c color.Color) (id, slot int, ok bool) {
	r, g, b, _ := c.RGBA()
	return decodePlaceholder(int(r>>8), int(g>>8), int(b>>8))
}

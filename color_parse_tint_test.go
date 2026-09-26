package semstyle

import (
	"context"
	"testing"
)

func TestToColorCtxTintsEverySlot(t *testing.T) {
	const key = "color-parse-tint-test"
	RegisterTint(key, Palette{
		Black: "#101010", White: "#e0e0e0", BrightBlack: "#606060",
		Base04: "#a0a0a0",
	})
	defer UnregisterTint(key)
	ctx := WithTint(context.Background(), key)

	tests := []struct{ name, want string }{
		{"white", "#e0e0e0"},  // ANSI name
		{"base05", "#e0e0e0"}, // its slot alias
		{"base04", "#a0a0a0"}, // no ANSI index, set by the tint
		{"base02", "#444444"}, // no ANSI index, blended from Black and BrightBlack
	}
	for _, tt := range tests {
		if got := ToColorStr(ToColorCtx(ctx, tt.name)); got != tt.want {
			t.Errorf("ToColorCtx(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}

	// Without a tint, names resolve untinted.
	if got := ToColorStr(ToColorCtx(context.Background(), "white")); got == "#e0e0e0" {
		t.Errorf("untinted white took the tint's value")
	}
}

func TestToANSISlotNameBackground(t *testing.T) {
	// Without a tint, a slot name as the background gets the same code as
	// its ANSI color name.
	for _, pair := range [][2]string{{"base00:base0d", "black:blue"}, {"base05:base03", "white:bright-black"}} {
		got, want := ToANSI("{{["+pair[0]+"]}}x"), ToANSI("{{["+pair[1]+"]}}x")
		if got != want {
			t.Errorf("ToANSI(%s) = %q, want %q (as %s)", pair[0], got, want, pair[1])
		}
	}
}

func TestUntintedSlotsWithoutANSIAssignment(t *testing.T) {
	// Without a tint, the slots with no ANSI assignment use their stand-in.
	for slot, name := range slotANSIFallback {
		if got, want := ToANSI("{{["+slot+":"+slot+"]}}x"), ToANSI("{{["+name+":"+name+"]}}x"); got != want {
			t.Errorf("ToANSI(%s) = %q, want %q (as %s)", slot, got, want, name)
		}
		if got, want := ToColorStr(ToColor(slot)), ToColorStr(ToColor(name)); got != want {
			t.Errorf("ToColor(%s) = %q, want %q (as %s)", slot, got, want, name)
		}
	}
}

func TestDarkerBackgroundsFollowVariant(t *testing.T) {
	dark := Palette{Black: "#202020", White: "#d0d0d0"}
	light := Palette{Black: "#e0e0e0", White: "#303030"}
	tests := []struct {
		name string
		p    Palette
		want string // base10's direction from Black
	}{
		{"inferred dark", dark, "darker"},
		{"inferred light", light, "lighter"},
		{"declared light", Palette{Black: "#202020", White: "#d0d0d0", Variant: "light"}, "lighter"},
		{"declared dark", Palette{Black: "#e0e0e0", White: "#303030", Variant: "dark"}, "darker"},
	}
	for _, tt := range tests {
		for _, slot := range []string{"base10", "base11"} {
			got := luminance(tt.p.slot(slot))
			bg := luminance(tt.p.Black)
			if (tt.want == "darker") != (got < bg) {
				t.Errorf("%s: %s = %s, want %s than %s", tt.name, slot, tt.p.slot(slot), tt.want, tt.p.Black)
			}
		}
		if b10, b11 := luminance(tt.p.slot("base10")), luminance(tt.p.slot("base11")); (tt.want == "darker") != (b11 < b10) {
			t.Errorf("%s: base11 should be further from the background than base10", tt.name)
		}
	}
}

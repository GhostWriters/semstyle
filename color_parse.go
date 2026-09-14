package semstyle

import (
	"context"
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Standard ANSI Color Reference (for tcell/lipgloss mapping):
// black=0, red=1, green=2, yellow=3, blue=4, magenta=5, cyan=6, white=7
// bright variants: +8
//
// Each of the 16 also has a base16/base24 slot name alias (see
// tinted-theming/base24's styling.md, the canonical source for which base0X
// maps to which ANSI color) resolving to the exact same index, so a theme
// author can write either vocabulary interchangeably -- "red" and "base08"
// are the same color, not two separate ones. The 8 base24 slots with no
// ANSI terminal assignment (base01/02/04/06/09/0F/10/11 -- backgrounds and
// the "orange"/"brown" categories) have no alias here since nothing consumes
// them as a color name yet.
var ansiColorIndex = map[string]string{
	"black":          "0",
	"red":            "1",
	"green":          "2",
	"yellow":         "3",
	"blue":           "4",
	"magenta":        "5",
	"cyan":           "6",
	"white":          "7",
	"bright-black":   "8",
	"bright-red":     "9",
	"bright-green":   "10",
	"bright-yellow":  "11",
	"bright-blue":    "12",
	"bright-magenta": "13",
	"bright-cyan":    "14",
	"bright-white":   "15",

	"base00": "0",
	"base08": "1",
	"base0b": "2",
	"base0a": "3",
	"base0d": "4",
	"base0e": "5",
	"base0c": "6",
	"base05": "7",
	"base03": "8",
	"base12": "9",
	"base14": "10",
	"base13": "11",
	"base16": "12",
	"base17": "13",
	"base15": "14",
	"base07": "15",
}

// ToColor converts a color name or hex string to a color.Color, with no
// tint applied (equivalent to ToColorCtx(context.Background(), c)).
func ToColor(c string) color.Color {
	return ToColorCtx(context.Background(), c)
}

// ToColorCtx is ToColor, but for one of the 16 standard ANSI color names,
// checks ctx's registered tint (see WithTint) first -- substituting its
// literal hex value in place of the plain ANSI-index color those names
// otherwise resolve to. Falls through to the untinted behavior when ctx
// carries no tint, or the tint doesn't set that particular slot.
func ToColorCtx(ctx context.Context, c string) color.Color {
	c = strings.ToLower(strings.TrimSpace(c))

	if strings.HasPrefix(c, "#") {
		return lipgloss.Color(c)
	}
	if _, ok := ansiColorIndex[c]; ok {
		if hex, ok := tintColorForCtx(ctx, c); ok {
			return lipgloss.Color(hex)
		}
	}
	if idx, ok := ansiColorIndex[c]; ok {
		return lipgloss.Color(idx)
	}
	if hexVal := GetHexForColor(c); hexVal != "" {
		return lipgloss.Color(hexVal)
	}
	return lipgloss.Color(c)
}

// ToColorStr extracts the string representation (hex or ANSI index) from a color.Color.
// Package-level function — not Styler-specific as color formatting is stateless.
func ToColorStr(c color.Color) string {
	if c == nil {
		return ""
	}
	if s, ok := c.(fmt.Stringer); ok {
		str := s.String()
		if len(str) > 0 && str[0] >= '0' && str[0] <= '9' {
			return str
		}
		if strings.HasPrefix(str, "#") {
			return strings.ToLower(str)
		}
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

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

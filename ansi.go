package semstyle

import (
	"context"
	"fmt"
	"image/color"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	tcellColor "github.com/gdamore/tcell/v3/color"
)

// bareReset is the trailing reset lipgloss actually appends when rendering
// an empty string with a style set -- confirmed via
// lipgloss.NewStyle().Foreground(...).Render(""): "\x1b[38;2;rr;gg;bbm\x1b[m",
// the short/bare form (no explicit "0"), not CodeReset ("\x1b[0m"). Trimming
// only CodeReset here silently left this suffix in place, which reset the
// color immediately before whatever content followed it -- e.g. an
// extended/hex color (or a semstyle tint substitution) rendering as
// invisible/default-colored text, since only a later, separately-appended
// SGR code (like a bold flag) would still show.
const bareReset = "\x1b[m"

// colorToFGSequence returns the ANSI opening sequence for a foreground color,
// using the lipgloss renderer (profile-aware via Bubble Tea or auto-detected).
func colorToFGSequence(c color.Color) string {
	rendered := lipgloss.NewStyle().Foreground(c).Render("")
	rendered = strings.TrimSuffix(rendered, CodeReset)
	return strings.TrimSuffix(rendered, bareReset)
}

// colorToBGSequence returns the ANSI opening sequence for a background color.
func colorToBGSequence(c color.Color) string {
	rendered := lipgloss.NewStyle().Background(c).Render("")
	rendered = strings.TrimSuffix(rendered, CodeReset)
	return strings.TrimSuffix(rendered, bareReset)
}

// parseStyleCodeToANSI parses fg:bg:flags format and returns ANSI codes,
// with no tint applied (equivalent to parseStyleCodeToANSICtx(context.
// Background(), content)).
func (st *Styler) parseStyleCodeToANSI(content string) string {
	return st.parseStyleCodeToANSICtx(context.Background(), content)
}

// parseStyleCodeToANSICtx is parseStyleCodeToANSI, but for each of the 16
// standard ANSI color names (fg and bg), checks ctx's registered tint (see
// WithTint) first -- substituting its literal hex value in place of the
// plain ANSI-index code those names otherwise resolve to. Falls through to
// the untinted behavior when ctx carries no tint, or the tint doesn't set
// that particular slot.
// Uses the lipgloss global renderer (set from colorprofile in profile.go).
func (st *Styler) parseStyleCodeToANSICtx(ctx context.Context, content string) string {
	if content == "-" {
		return CodeReset
	}
	if content == "~" {
		return CodeHardReset
	}

	st.ensureMaps()
	st.mu.RLock()
	defer st.mu.RUnlock()

	// Split by colons: fg:bg:flags
	parts := strings.Split(content, ":")
	var codes strings.Builder

	// Pre-emptive reset of flags ONLY if they start with '-'
	if len(parts) > 2 && strings.HasPrefix(parts[2], "-") {
		// 22:Bold/Dim off, 23:Italic off, 24:Underline off, 25:Blink off, 27:Reverse off, 29:Strikethrough off
		codes.WriteString("\x1b[22m\x1b[23m\x1b[24m\x1b[25m\x1b[27m\x1b[29m")
	}

	// Flags (peek for H early to affect colors)
	highIntensity := false
	if len(parts) > 2 {
		f := parts[2]
		if strings.Contains(f, "H") {
			highIntensity = true
		}
	}

	// Part 0: Foreground color
	if len(parts) > 0 && parts[0] == "~" {
		codes.WriteString(CodeHardFGReset)
	} else if len(parts) > 0 && parts[0] == "-" {
		codes.WriteString(CodeFGReset)
	} else if len(parts) > 0 && parts[0] != "" {
		colorName := strings.ToLower(parts[0])
		if highIntensity {
			if brightName, ok := st.getBrightVariant(colorName); ok {
				colorName = brightName
			}
		}

		// Check for non-color attributes (bold, etc.) first
		if code, ok := st.attributeMap[colorName]; ok {
			codes.WriteString(code)
			goto FoundFG
		}

		// A registered tint (see WithTint) overrides one of the 16 standard
		// ANSI color names with a literal truecolor value, bypassing the
		// plain ANSI-index code below -- checked before it so the tint wins.
		if hex, ok := tintColorForCtx(ctx, colorName); ok {
			codes.WriteString(colorToFGSequence(lipgloss.Color(hex)))
			goto FoundFG
		}

		// Check st.ansiMap for standard colors (direct ANSI codes, max compatibility)
		if code, ok := st.ansiMap[colorName]; ok {
			codes.WriteString(code)
			goto FoundFG
		}

		// Extended color: resolve via tcell → hex → lipgloss
		if strings.HasPrefix(colorName, "#") {
			codes.WriteString(colorToFGSequence(lipgloss.Color(colorName)))
		} else {
			tc := ResolveTcellColor(colorName)
			if tc != tcellColor.Default {
				hexVal := tc.Hex()
				if hexVal >= 0 {
					codes.WriteString(colorToFGSequence(lipgloss.Color(fmt.Sprintf("#%06x", hexVal))))
				}
			}
		}
	}
FoundFG:

	// Part 1: Background color
	if len(parts) > 1 && parts[1] == "~" {
		codes.WriteString(CodeHardBGReset)
	} else if len(parts) > 1 && parts[1] == "-" {
		codes.WriteString(CodeBGReset)
	} else if len(parts) > 1 && parts[1] != "" {
		colorName := strings.ToLower(parts[1])
		if highIntensity {
			if brightName, ok := st.getBrightVariant(colorName); ok {
				colorName = brightName
			}
		}

		if code, ok := st.attributeMap[colorName]; ok {
			codes.WriteString(code)
			goto FoundBG
		}

		if hex, ok := tintColorForCtx(ctx, colorName); ok {
			codes.WriteString(colorToBGSequence(lipgloss.Color(hex)))
			goto FoundBG
		}

		if code, ok := st.ansiMap[colorName+"bg"]; ok {
			codes.WriteString(code)
			goto FoundBG
		}

		if strings.HasPrefix(colorName, "#") {
			codes.WriteString(colorToBGSequence(lipgloss.Color(colorName)))
		} else {
			tc := ResolveTcellColor(colorName)
			if tc != tcellColor.Default {
				hexVal := tc.Hex()
				if hexVal >= 0 {
					codes.WriteString(colorToBGSequence(lipgloss.Color(fmt.Sprintf("#%06x", hexVal))))
				}
			}
		}
	}
FoundBG:

	// Part 2: Flags (each character is a flag: b=bold, u=underline, etc.)
	if len(parts) > 2 && parts[2] != "" {
		f := strings.TrimPrefix(parts[2], "-")
		for _, flag := range f {
			flagStr := string(flag)
			if code, ok := st.ansiMap[flagStr]; ok {
				codes.WriteString(code)
			}
		}
	}

	return codes.String()
}

var ansiRegex = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*((?:[a-zA-Z\\d]*(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]*)*)?\u0007|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-ntqry=><~]))")

// oscStRegex matches OSC sequences terminated by ST (\x1b\\) rather than BEL (\x07),
// which ansiRegex does not cover. Used to strip hyperlinks (OSC 8) and similar sequences.
var oscStRegex = regexp.MustCompile("\x1b][^\x1b]*\x1b\\\\")

// StripANSI removes all ANSI escape sequences from a string, including OSC sequences
// terminated by ST (\x1b\\) such as terminal hyperlinks (OSC 8).
func StripANSI(text string) string {
	text = oscStRegex.ReplaceAllString(text, "")
	return ansiRegex.ReplaceAllString(text, "")
}

// getBrightVariant attempts to get the bright variant of a color name.
// Callers must already hold st.mu (read or write); this method takes no
// lock itself.
func (st *Styler) getBrightVariant(name string) (string, bool) {
	if strings.HasPrefix(name, "bright-") {
		return name, true
	}
	if _, ok := st.ansiMap["bright-"+name]; ok {
		return "bright-" + name, true
	}
	return name, false
}

// --- package-level delegators to Default ---
func parseStyleCodeToANSI(content string) string {
	return Default.parseStyleCodeToANSI(content)
}

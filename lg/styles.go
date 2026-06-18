package semlg

import (
	"fmt"
	"image/color"
	"regexp"
	"strings"

	"github.com/GhostWriters/semstyle"
	"charm.land/lipgloss/v2"
)

// StyleFlags holds ANSI style modifier state parsed from a flags field.
type StyleFlags struct {
	Bold          bool
	Underline     bool
	Italic        bool
	Blink         bool
	Dim           bool
	Reverse       bool
	Strikethrough bool
	HighIntensity bool
}

// Apply applies all set flags to a lipgloss style.
func (f StyleFlags) Apply(s lipgloss.Style) lipgloss.Style {
	s = s.
		Bold(f.Bold).
		Underline(f.Underline).
		Italic(f.Italic).
		Blink(f.Blink).
		Faint(f.Dim).
		Reverse(f.Reverse).
		Strikethrough(f.Strikethrough)
	if f.HighIntensity {
		if fg := s.GetForeground(); fg != nil {
			s = s.Foreground(brightenColor(fg))
		}
		if bg := s.GetBackground(); bg != nil {
			s = s.Background(brightenColor(bg))
		}
	}
	return s
}

// ResetFlags clears all text attributes from a style.
func ResetFlags(s lipgloss.Style) lipgloss.Style {
	return StyleFlags{}.Apply(s)
}

// ToStyle resolves any semantic or direct tags in text and applies the resulting
// style to the provided lipgloss.Style, resetting to resetStyle on a reset tag.
func ToStyle(st *semstyle.Styler, text string, style lipgloss.Style, resetStyle lipgloss.Style) lipgloss.Style {
	translated := st.ToTags(text)
	re := st.GetDelimitedRegex()
	for _, subMatch := range re.FindAllStringSubmatch(translated, -1) {
		semantic := subMatch[1]
		direct := subMatch[2]
		if semantic != "" {
			tagName := strings.Trim(semantic, "_")
			def := st.GetColorDefinition(tagName)
			style = ToStyle(st, def, style, resetStyle)
		} else if direct != "" {
			if direct == "|" || direct == "-" {
				style = resetStyle
			} else {
				style = CodeToStyle(strings.Trim(direct, "|"), style, resetStyle)
			}
		}
	}
	return style
}

// CodeToFlags parses the flags field of a raw fg:bg:flags code into a StyleFlags struct.
func CodeToFlags(rawCode string) StyleFlags {
	parts := strings.Split(rawCode, ":")
	if len(parts) < 3 {
		return StyleFlags{}
	}
	s := strings.TrimPrefix(parts[2], "-")
	var f StyleFlags
	for _, char := range s {
		switch char {
		case 'B':
			f.Bold = true
		case 'b':
			f.Bold = false
		case 'U':
			f.Underline = true
		case 'u':
			f.Underline = false
		case 'I':
			f.Italic = true
		case 'i':
			f.Italic = false
		case 'D':
			f.Dim = true
		case 'd':
			f.Dim = false
		case 'L':
			f.Blink = true
		case 'l':
			f.Blink = false
		case 'R':
			f.Reverse = true
		case 'r':
			f.Reverse = false
		case 'S':
			f.Strikethrough = true
		case 's':
			f.Strikethrough = false
		}
	}
	return f
}

// CodeToStyle applies a raw fg:bg:flags code to a lipgloss.Style.
func CodeToStyle(styleCode string, style lipgloss.Style, resetStyle lipgloss.Style) lipgloss.Style {
	if styleCode == "~" {
		return lipgloss.NewStyle()
	}
	if styleCode == semstyle.CodeReset || styleCode == "-" {
		return resetStyle
	}

	parts := strings.Split(styleCode, ":")

	if len(parts) > 2 && strings.HasPrefix(parts[2], "-") {
		style = ResetFlags(style)
	}

	if len(parts) > 0 && parts[0] != "" {
		switch parts[0] {
		case "~":
			style = style.Foreground(lipgloss.Color(""))
		case "-":
			style = style.Foreground(resetStyle.GetForeground())
		default:
			if c := semstyle.ToColor(parts[0]); c != nil {
				style = style.Foreground(c)
			}
		}
	}

	if len(parts) > 1 && parts[1] != "" {
		switch parts[1] {
		case "~":
			style = style.Background(lipgloss.Color(""))
		case "-":
			style = style.Background(resetStyle.GetBackground())
		default:
			if c := semstyle.ToColor(parts[1]); c != nil {
				style = style.Background(c)
			}
		}
	}

	if len(parts) > 2 {
		s := strings.TrimPrefix(parts[2], "-")
		for _, char := range s {
			switch char {
			case 'B':
				style = style.Bold(true)
			case 'b':
				style = style.Bold(false)
			case 'U':
				style = style.Underline(true)
			case 'u':
				style = style.Underline(false)
			case 'I':
				style = style.Italic(true)
			case 'i':
				style = style.Italic(false)
			case 'D':
				style = style.Faint(true)
			case 'd':
				style = style.Faint(false)
			case 'L':
				style = style.Blink(true)
			case 'l':
				style = style.Blink(false)
			case 'R':
				style = style.Reverse(true)
			case 'r':
				style = style.Reverse(false)
			case 'S':
				style = style.Strikethrough(true)
			case 's':
				style = style.Strikethrough(false)
			case 'H':
				if fg := style.GetForeground(); fg != nil {
					style = style.Foreground(brightenColor(fg))
				}
				if bg := style.GetBackground(); bg != nil {
					style = style.Background(brightenColor(bg))
				}
			}
		}
	}

	return style
}

// ToANSIOnBackground renders tagged text to ANSI and ensures it displays correctly
// against the given parent background style. It prepends the parent's ANSI codes,
// appends a reset, then calls MaintainBackground so inner resets re-assert the
// parent colors rather than bleeding to the terminal default.
func ToANSIOnBackground(s string, bg lipgloss.Style, prefix ...string) string {
	rendered := semstyle.ToANSI(s, prefix...)
	getANSI := func(st lipgloss.Style) string {
		r := st.Render("_")
		return strings.Split(r, "_")[0]
	}
	full := getANSI(bg) + rendered + semstyle.CodeReset
	return MaintainBackground(full, bg)
}

// MaintainBackground replaces ANSI resets with the reset followed by the parent
// style's codes, preventing content-level resets from bleeding to the terminal
// default background. It also ensures the string starts with the parent's full
// ANSI code so unstyled/plain text inherits the background.
func MaintainBackground(text string, style lipgloss.Style) string {
	getANSI := func(s lipgloss.Style) string {
		rendered := s.Render("_")
		return strings.Split(rendered, "_")[0]
	}

	fullCode := getANSI(style)
	if fullCode == "" {
		return text
	}

	if !strings.HasPrefix(text, "\x1b[") {
		text = fullCode + text
	}

	fgCode := getANSI(lipgloss.NewStyle().Foreground(style.GetForeground()))
	bgCode := getANSI(lipgloss.NewStyle().Background(style.GetBackground()))

	re := regexp.MustCompile(`\x1b\[(?:0|39|49)?m`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		switch match {
		case "\x1b[39m":
			return match + fgCode
		case "\x1b[49m":
			return match + bgCode
		default:
			return match + fullCode
		}
	})
}

// brightenColor brightens a color by 30% of remaining headroom toward white.
func brightenColor(c color.Color) color.Color {
	if c == nil {
		return c
	}
	rr, gg, bb, _ := c.RGBA()
	r := int(rr >> 8)
	g := int(gg >> 8)
	b := int(bb >> 8)
	r = min(255, r+int(float64(255-r)*0.3))
	g = min(255, g+int(float64(255-g)*0.3))
	b = min(255, b+int(float64(255-b)*0.3))
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

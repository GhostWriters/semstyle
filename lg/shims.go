package semlg

// shims.go re-exports the semstyle public API so callers only need to import
// semstyle/lg rather than both semstyle and semstyle/lg.

import (
	"image/color"
	"regexp"

	"github.com/GhostWriters/semstyle"
	"github.com/charmbracelet/colorprofile"
)

// -- Types --

type Styler = semstyle.Styler

// -- Default instance --

var Default = semstyle.Default

// -- Constants: reset --

const (
	CodeReset      = semstyle.CodeReset
	CodeFGReset    = semstyle.CodeFGReset
	CodeBGReset    = semstyle.CodeBGReset
	CodeHardReset  = semstyle.CodeHardReset
	CodeHardFGReset = semstyle.CodeHardFGReset
	CodeHardBGReset = semstyle.CodeHardBGReset
)

// -- Constants: modifiers --

const (
	CodeBold          = semstyle.CodeBold
	CodeDim           = semstyle.CodeDim
	CodeUnderline     = semstyle.CodeUnderline
	CodeBlink         = semstyle.CodeBlink
	CodeReverse       = semstyle.CodeReverse
	CodeItalic        = semstyle.CodeItalic
	CodeStrikethrough = semstyle.CodeStrikethrough

	CodeBoldOff          = semstyle.CodeBoldOff
	CodeDimOff           = semstyle.CodeDimOff
	CodeUnderlineOff     = semstyle.CodeUnderlineOff
	CodeBlinkOff         = semstyle.CodeBlinkOff
	CodeReverseOff       = semstyle.CodeReverseOff
	CodeItalicOff        = semstyle.CodeItalicOff
	CodeStrikethroughOff = semstyle.CodeStrikethroughOff
)

// -- Constants: foreground colors --

const (
	CodeBlack   = semstyle.CodeBlack
	CodeRed     = semstyle.CodeRed
	CodeGreen   = semstyle.CodeGreen
	CodeYellow  = semstyle.CodeYellow
	CodeBlue    = semstyle.CodeBlue
	CodeMagenta = semstyle.CodeMagenta
	CodeCyan    = semstyle.CodeCyan
	CodeWhite   = semstyle.CodeWhite

	CodeBrightBlack   = semstyle.CodeBrightBlack
	CodeBrightRed     = semstyle.CodeBrightRed
	CodeBrightGreen   = semstyle.CodeBrightGreen
	CodeBrightYellow  = semstyle.CodeBrightYellow
	CodeBrightBlue    = semstyle.CodeBrightBlue
	CodeBrightMagenta = semstyle.CodeBrightMagenta
	CodeBrightCyan    = semstyle.CodeBrightCyan
	CodeBrightWhite   = semstyle.CodeBrightWhite
)

// -- Constants: background colors --

const (
	CodeBlackBg   = semstyle.CodeBlackBg
	CodeRedBg     = semstyle.CodeRedBg
	CodeGreenBg   = semstyle.CodeGreenBg
	CodeYellowBg  = semstyle.CodeYellowBg
	CodeBlueBg    = semstyle.CodeBlueBg
	CodeMagentaBg = semstyle.CodeMagentaBg
	CodeCyanBg    = semstyle.CodeCyanBg
	CodeWhiteBg   = semstyle.CodeWhiteBg
)

// -- Delimiter vars -- kept in sync with semstyle via SetDelimiters below.
// Read these for the current default delimiter values.

var (
	SemanticPrefix string
	SemanticSuffix string
	DirectPrefix   string
	DirectSuffix   string
)

func init() {
	SemanticPrefix = semstyle.SemanticPrefix
	SemanticSuffix = semstyle.SemanticSuffix
	DirectPrefix   = semstyle.DirectPrefix
	DirectSuffix   = semstyle.DirectSuffix
}

// -- Render policy -- this is a function var; assign semstyle.RenderPolicy directly.

func SetRenderPolicy(fn func() bool) { semstyle.Default.SetRenderPolicy(fn) }

// -- Core rendering --

func ToANSI(s string, prefix ...string) string          { return semstyle.ToANSI(s, prefix...) }
func ToTags(s string, prefix ...string) string          { return semstyle.ToTags(s, prefix...) }
func ToPlain(s string) string                           { return semstyle.ToPlain(s) }
func StripTags(s string) string                         { return semstyle.StripTags(s) }
func StripANSI(s string) string                         { return semstyle.StripANSI(s) }
func StripDelimiters(s string) string                   { return semstyle.StripDelimiters(s) }
func Sprintf(format string, a ...any) string            { return semstyle.Sprintf(format, a...) }
func ExpandTagsWithMap(text string, styleMap map[string]string, stripUnresolvable bool, prefix string) string {
	return semstyle.ExpandTagsWithMap(text, styleMap, stripUnresolvable, prefix)
}

// -- Tag wrapping --

func WrapSemantic(name string) string { return semstyle.WrapSemantic(name) }
func WrapDirect(code string) string   { return semstyle.WrapDirect(code) }

// -- Registration --

func New() *semstyle.Styler { return semstyle.New() }

func RegisterConsoleTag(name, taggedValue string)    { semstyle.RegisterConsoleTag(name, taggedValue) }
func RegisterConsoleTagRaw(name, rawValue string)    { semstyle.RegisterConsoleTagRaw(name, rawValue) }
func RegisterThemeTag(name, taggedValue string)      { semstyle.RegisterThemeTag(name, taggedValue) }
func RegisterThemeTagRaw(name, rawValue string)      { semstyle.RegisterThemeTagRaw(name, rawValue) }
func RegisterSemanticTag(name, taggedValue string)   { semstyle.RegisterSemanticTag(name, taggedValue) }
func RegisterSemanticTagRaw(name, rawValue string)   { semstyle.RegisterSemanticTagRaw(name, rawValue) }
func RegisterHyperlinkTag(name string)               { semstyle.RegisterHyperlinkTag(name) }
func RegisterColor(name, value string)               { semstyle.RegisterColor(name, value) }
func GetRawTagCode(name string) string               { return semstyle.GetRawTagCode(name) }
func GetColorDefinition(name string) string          { return semstyle.GetColorDefinition(name) }
func UnregisterColor(name string)                    { semstyle.UnregisterColor(name) }
func UnregisterPrefix(prefix string)                 { semstyle.UnregisterPrefix(prefix) }
func ClearThemeMap()                                 { semstyle.ClearThemeMap() }
func ResetCustomColors()                             { semstyle.ResetCustomColors() }
func BuildColorMap()                                 { semstyle.BuildColorMap() }

// -- Color utilities --

func ToColor(s string) color.Color    { return semstyle.ToColor(s) }
func ToColorStr(c color.Color) string { return semstyle.ToColorStr(c) }
func GetHexForColor(name string) string { return semstyle.GetHexForColor(name) }

// -- Profile --

func GetPreferredProfile() colorprofile.Profile      { return semstyle.GetPreferredProfile() }
func SetPreferredProfile(p colorprofile.Profile)     { semstyle.SetPreferredProfile(p) }

// -- Regex --

func GetDelimitedRegex() *regexp.Regexp { return semstyle.GetDelimitedRegex() }
func GetDirectRegex() *regexp.Regexp    { return semstyle.GetDirectRegex() }

// -- Delimiter config --

func SetDelimiters(semPre, semSuf, dirPre, dirSuf string) {
	semstyle.SetDelimiters(semPre, semSuf, dirPre, dirSuf)
	SemanticPrefix = semPre
	SemanticSuffix = semSuf
	DirectPrefix   = dirPre
	DirectSuffix   = dirSuf
}

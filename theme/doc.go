// Package semtheme is the optional theming layer for semstyle. It parses TOML
// theme files into resolved semantic style maps and registers them into a
// [semstyle.Styler].
//
// # Theme file format
//
// A theme is a TOML file with four sections:
//
//   - [metadata] — name, description, author.
//   - [palette] — reusable color variables referenced as $name in style values.
//   - [styles] — semantic name → style value. Values may use direct tags
//     ({{[fg:bg:flags]}}), reference palette variables ($name), or reference
//     other styles by name.
//   - [defaults] — opaque app-defined table passed through without
//     interpretation; the consuming application decodes it into its own
//     settings (e.g. border styles, panel config).
//
// An optional [syntax] section overrides the tag delimiters used in the file.
//
// # Usage
//
// Parse and register a theme into the default styler in one call:
//
//	data, _ := os.ReadFile("midnight.theme")
//	defaults, err := semtheme.RegisterInto(data, "")
//	// defaults is the theme's [defaults] table as map[string]any
//
// To register under a namespace prefix (e.g. for previews or per-surface
// themes), pass a non-empty prefix:
//
//	semtheme.RegisterInto(data, "preview")
//	// Tags become "preview_Title", "preview_Error", etc.
//
// To work with the parsed form directly:
//
//	tf, _     := semtheme.Parse(data)
//	styles, _ := semtheme.ResolveColors(tf)
//	s := semstyle.New()
//	s.SetThemeMap(styles)
//
// # Lipgloss style conversion
//
// When building TUI components with lipgloss rather than writing ANSI to a
// terminal, use the style conversion functions to apply semstyle tags to a
// [lipgloss.Style]:
//
//	// From a tagged string:
//	style := semtheme.ToStyle(semstyle.Default, "{{|Error|}}", base, base)
//
//	// From a raw fg:bg:flags code:
//	style := semtheme.CodeToStyle("red::B", base, base)
//
//	// Parse flags only:
//	flags := semtheme.CodeToFlags("red::B")
//	style = flags.Apply(style)
package semtheme

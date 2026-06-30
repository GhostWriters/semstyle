// Package semstyle provides a semantic terminal styling engine for Go.
//
// # Overview
//
// semstyle lets you write text with named, tag-based markup and resolve it to
// ANSI escape sequences at render time. Two tag types are supported:
//
//   - Semantic tags — {{|Name|}} — reference a named style resolved against a
//     style map. Change the map (e.g. load a theme) and every tag re-styles
//     without touching call sites.
//
//   - Direct tags — {{[fg:bg:flags]}} — inline ANSI styling, e.g.
//     {{[red:black:B]}} for bold red on black.
//
// # Tag flags
//
// The flags field of a direct tag uses single characters: B=Bold, D=Dim,
// U=Underline, I=Italic, L=Blink, R=Reverse, S=Strikethrough (uppercase=on,
// lowercase=off). H shifts the fg/bg color to its bright variant. A leading -
// resets all attributes before applying the remaining flags.
//
// # Style resolution
//
// Each [Styler] keeps two semantic maps: a console map (base/built-in tags)
// and a theme map (overrides loaded from a theme file). [ToANSI] without a
// prefix resolves against the console map; with a prefix it resolves
// theme-first with console fallback.
//
// # Package-level vs per-instance
//
// A process-wide [Default] styler backs the package-level functions. Most
// programs only need this:
//
//	semstyle.RegisterConsoleTag("Notice", "{{[cyan::B]}}")
//	fmt.Println(semstyle.ToANSI("{{|Notice|}}hello{{[-]}}"))
//
// For multiple independent style configurations (e.g. different themes for
// different surfaces), create a [Styler] directly:
//
//	s := semstyle.New()
//	s.RegisterThemeTag("Title", "{{[magenta::B]}}")
//	out := s.ToANSI("{{|Title|}}Report{{[-]}}")
//
// # Hyperlinks
//
// Terminal hyperlinks (OSC 8) can be produced two ways:
//
// Register a tag name — its enclosed content becomes both the URL and label:
//
//	semstyle.RegisterHyperlinkTag("URL")
//	semstyle.ToANSI("{{|URL|}}https://example.com{{[-]}}")
//
// Or use an explicit URL as the final colon-field of any tag. The enclosed
// content is the visible text; the field is the URL. An empty field uses the
// content as both:
//
//	// Direct tag: fg:bg:flags:url
//	semstyle.ToANSI("{{[cyan::U:https://dockstarter.com]}}DockSTARTer Website{{[-]}}")
//
//	// Semantic tag: name:fg:bg:flags:url (empty fields keep registered style)
//	semstyle.ToANSI("{{|mylink::::https://dockstarter.com|}}DockSTARTer Website{{[-]}}")
//
// # Render policy
//
// By default [ToANSI] always emits ANSI. Set a policy to suppress color
// conditionally (e.g. when output is redirected):
//
//	semstyle.SetRenderPolicy(func() bool { return isTerminal() })
//
// # Theming
//
// The semstyle/theme subpackage parses TOML theme files into style maps.
// Import it only when file-driven themes are needed:
//
//	import semtheme "github.com/GhostWriters/semstyle/theme"
//
//	data, _ := os.ReadFile("midnight.theme")
//	defaults, _ := semtheme.RegisterInto(data, "")
package semstyle

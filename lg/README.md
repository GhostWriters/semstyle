# semstyle/lg

The **lipgloss integration layer** for [`semstyle`](..): converts semstyle tags and raw
style codes into `lipgloss.Style` values, and maintains parent background colors across ANSI
resets when compositing TUI components.

It also **re-exports the full `semstyle` API**, so most TUI applications only need this one
import. The conventional alias is `semstyle`, matching the underlying package name:

```go
import semstyle "github.com/GhostWriters/semstyle/lg"

semstyle.RegisterConsoleTag("Notice", "{{[cyan::B]}}")
fmt.Println(semstyle.ToANSI("{{|Notice|}}hello{{[-]}}"))
style := semstyle.ToStyle(semstyle.Default, "{{|Error|}}", base, base)
safe  := semstyle.MaintainBackground(rendered, parentStyle)
```

## Lipgloss API

### Style conversion

Apply a raw `fg:bg:flags` code to a lipgloss style:

```go
style := semstyle.CodeToStyle("red::B", base, resetStyle)
```

Resolve any semstyle tags (semantic or direct) and apply the result:

```go
style := semstyle.ToStyle(semstyle.Default, "{{|Error|}}", base, resetStyle)
```

Parse only the flags portion into a `StyleFlags` struct:

```go
flags := semstyle.CodeToFlags("red::B")
style  = flags.Apply(style)
```

Clear all text attributes (bold, italic, etc.) from a style:

```go
style = semstyle.ResetFlags(style)
```

### Background maintenance

When nesting lipgloss-rendered content inside a parent with a background color, inner ANSI
resets bleed back to the terminal default instead of the parent's background.
`MaintainBackground` intercepts those resets and re-asserts the parent style's colors:

```go
safe := semstyle.MaintainBackground(innerRendered, parentStyle)
```

It handles three reset forms:

- `\x1b[0m` / `\x1b[m` — full reset: re-asserts fg + bg + all attributes
- `\x1b[39m` — FG reset: re-asserts fg only
- `\x1b[49m` — BG reset: re-asserts bg only

It also ensures the string starts with the parent's full ANSI code so unstyled text
inherits the background automatically.

## StyleFlags

`StyleFlags` holds the parsed on/off state for each text modifier. Its `.Apply(style)`
method applies all flags to a lipgloss style in one call:

| Field | Flag char |
| --- | --- |
| `Bold` | `B` / `b` |
| `Dim` | `D` / `d` |
| `Underline` | `U` / `u` |
| `Italic` | `I` / `i` |
| `Blink` | `L` / `l` |
| `Reverse` | `R` / `r` |
| `Strikethrough` | `S` / `s` |
| `HighIntensity` | `H` — brightens fg/bg color |

## Re-exported semstyle API

All `semstyle` package-level functions, constants, and variables are available directly:
`ToANSI`, `ToTags`, `ToPlain`, `StripTags`, `StripANSI`, `Sprintf`, `ToColor`,
`ToColorStr`, `WrapSemantic`, `WrapDirect`, `RegisterConsoleTag`, `RegisterThemeTag`,
`RegisterHyperlinkTag`, `BuildColorMap`, `ClearThemeMap`, `SetThemeMap`, `SetRenderPolicy`,
`SetDelimiters`, `GetPreferredProfile`, `SetPreferredProfile`, `Default`, `New`, all
`Code*` constants, and the `SemanticPrefix` / `SemanticSuffix` / `DirectPrefix` /
`DirectSuffix` delimiter vars.

See the [semstyle README](../README.md) for full documentation.

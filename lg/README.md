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

// Render tagged text against a parent container background in one call:
fmt.Println(semstyle.ToANSIOnBackground("{{|Notice|}}hello{{[-]}}", parentStyle))

// Or build a lipgloss style from tags:
style := semstyle.ToStyle(semstyle.Default, "{{|Error|}}", base, base)

// Or maintain background on already-rendered ANSI (e.g. from a third-party component):
safe := semstyle.MaintainBackground(alreadyRendered, parentStyle)
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

When nesting content inside a parent with a background color, inner ANSI resets bleed back
to the terminal default instead of the parent's background. Two functions address this:

**`ToANSIOnBackground`** — the common case: render tagged text and maintain the background
in one call. Equivalent to `ToANSI` + prefix injection + reset + `MaintainBackground`:

```go
out := semstyle.ToANSIOnBackground("{{|Error|}}oops{{[-]}}", parentStyle)
// optional prefix for theme map lookup:
out  = semstyle.ToANSIOnBackground("{{|Error|}}oops{{[-]}}", parentStyle, "preview")
```

**`MaintainBackground`** — for already-rendered ANSI strings you didn't produce yourself
(e.g. output from a third-party lipgloss component or a viewport):

```go
safe := semstyle.MaintainBackground(alreadyRendered, parentStyle)
```

Both handle three reset forms:

- `\x1b[0m` / `\x1b[m` — full reset: re-asserts fg + bg + all attributes
- `\x1b[39m` — FG reset: re-asserts fg only
- `\x1b[49m` — BG reset: re-asserts bg only

Both also ensure the string starts with the parent's full ANSI code so unstyled text
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

In addition to the lipgloss-specific functions above (`ToANSIOnBackground`,
`MaintainBackground`, `ToStyle`, `CodeToStyle`, `CodeToFlags`, `ResetFlags`, `StyleFlags`),
all `semstyle` package-level functions, constants, and variables are available directly:
`ToANSI`, `ToTags`, `ToPlain`, `StripTags`, `StripANSI`, `Sprintf`, `ToColor`,
`ToColorStr`, `WrapSemantic`, `WrapDirect`, `RegisterConsoleTag`, `RegisterThemeTag`,
`RegisterHyperlinkTag`, `BuildColorMap`, `ClearThemeMap`, `SetThemeMap`, `SetRenderPolicy`,
`SetDelimiters`, `GetPreferredProfile`, `SetPreferredProfile`, `Default`, `New`, all
`Code*` constants, and the `SemanticPrefix` / `SemanticSuffix` / `DirectPrefix` /
`DirectSuffix` delimiter vars.

See the [semstyle README](../README.md) for full documentation.

# semstyle

A small, dependency-light engine for **semantic terminal styling** in Go. Write text with
named, tag-based markup and resolve it to ANSI escape sequences at render time:

```go
semstyle.ToANSI("{{|Error|}}failed{{[-]}}: {{[red::B]}}retry{{[-]}}")
```

It depends only on `lipgloss`, `colorprofile`, and `tcell/color` for color resolution — no
application, TTY, or config coupling.

## Tag format

The tag format is based on the color tag syntax used by
[cview](https://code.rocket9labs.com/tslocum/cview) /
[tview](https://github.com/rivo/tview), extended with semantic names, additional flags, and
hyperlink support.

### Direct tags

Direct tags apply inline styling immediately: `{{[fg:bg:flags]}}…{{[-]}}`

```text
{{[red:black:B]}}   ← bold red on black
{{[::U]}}           ← underline only; fg and bg unchanged
{{[green]}}         ← change fg only; bg and flags unchanged
{{[:blue]}}         ← change bg only; fg and flags unchanged
{{[-:blue]}}        ← reset fg to default, set bg to blue
{{[red:-]}}         ← set fg to red, reset bg to default
{{[-]}}             ← reset all styling
```

**Color values** — any named ANSI color (`red`, `bright-blue`, …), hex (`#ff8800`), empty
to leave the current color unchanged, or `-` to reset that color to the terminal default.

**Flags** (each is a single character):

| Flag | Meaning |
| --- | --- |
| `B` / `b` | Bold on / off |
| `D` / `d` | Dim on / off |
| `U` / `u` | Underline on / off |
| `I` / `i` | Italic on / off |
| `L` / `l` | Blink on / off |
| `R` / `r` | Reverse on / off |
| `S` / `s` | Strikethrough on / off |
| `H` | High Intensity: shifts fg/bg to bright variant (`red` → `bright-red`) |
| `-` (leading) | Reset all attributes first, then apply remaining flags |

A leading `-` in the flags field resets all text attributes (bold, underline, etc.) without
touching colors. It can stand alone to reset attributes only, or be combined with flags to
reset-then-set in one step:

```text
{{[::-]}}    ← reset all attributes; fg and bg unchanged
{{[::-B]}}   ← reset all attributes, then set bold; fg and bg unchanged
{{[-]}}      ← reset everything (fg, bg, and all attributes)
```

### Semantic tags

Semantic tags reference a **named style** resolved at render time against a style map:
`{{|Name|}}…{{[-]}}`

```text
{{|Error|}}something went wrong{{[-]}}
{{|Title|}}My App{{[-]}}
```

Change the style map (load a theme) and every semantic tag re-styles — no call sites change.
Semantic tags are the primary use case; direct tags are what semantic tags resolve *to*.

Semantic tags also accept optional per-use overrides in the same `fg:bg:flags` format,
applied on top of the registered style:

```text
{{|Error:yellow|}}           ← Error style but fg overridden to yellow
{{|Error:yellow:black:BU|}}  ← fg, bg, and flags all overridden
{{|Error::black|}}           ← bg overridden; fg left unchanged (empty field)
{{|Error:-:black|}}          ← bg overridden; fg reset to terminal default
{{|Error:yellow:-|}}         ← fg overridden; bg reset to terminal default
{{|Error::-|}}               ← reset all attributes; fg and bg unchanged
{{|Error::-B|}}              ← reset all attributes, then set bold; fg and bg unchanged
```

## Layout

- **`semstyle`** (this package) — the styling engine. Import it alone if you only need
  tag-based styling and define styles in code.
- **`semstyle/theme`** — an optional layer that parses theme **files** (TOML) into style
  maps for the engine. Import it only if you want file-driven themes; it pulls in a TOML
  parser. See [theme/README.md](theme/README.md) and the [Theming](#theming) section below.

They're one module, two packages: styling-only consumers never compile the theme package or
its TOML dependency.

## Two ways to use it

### 1. Package-level (simple, one global config)

A process-wide `Default` styler backs the package functions. This is all most programs need:

```go
import "github.com/GhostWriters/semstyle"

semstyle.RegisterConsoleTag("Notice", "{{[cyan::B]}}")
fmt.Println(semstyle.ToANSI("{{|Notice|}}hello{{[-]}}"))
plain := semstyle.ToPlain(styled)   // remove all tags + ANSI
```

### 2. Per-instance `Styler` (multiple independent configs)

Each `Styler` owns its own tag/color maps, so you can run several independent style
configurations in one process (e.g. different themes for different surfaces):

```go
s := semstyle.New()
s.RegisterThemeTag("Title", "{{[magenta::B]}}")
out := s.ToANSI("{{|Title|}}Report{{[-]}}")
```

The package functions are thin delegators to `Default`, so the global API and the
per-instance API are identical in behavior.

## Style resolution

Each `Styler` keeps two semantic maps:

- **console map** — built-in / base tags (registered via `RegisterConsoleTag`).
- **theme map** — overrides loaded from a theme; takes precedence over the console map.

`ToANSI` without a prefix resolves against the console map; with a prefix it resolves
theme-first with console fallback. Supply a theme map with `SetThemeMap` (or the `semtheme`
companion package, which parses theme files into a map).

## Key API

| Function / method | Purpose |
| --- | --- |
| `ToANSI(s, prefix...)` | Expand tags → ANSI; console map by default, theme map when prefix given |
| `ToTags(s, prefix...)` | Expand semantic tags → direct tags; stops before ANSI conversion |
| `ToPlain(s)` | Remove all tags **and** ANSI escapes → plain text |
| `StripTags(s)` | Remove tags only, leaving any existing ANSI intact |
| `ToColor(s)` | Color name or hex string → `color.Color` |
| `ToColorStr(c)` | `color.Color` → hex or ANSI index string |
| `Sprintf(fmt, a...)` | Format string then apply `ToANSI` |
| `RegisterConsoleTag(name, val)` / `…Raw` | Define a base semantic tag |
| `RegisterThemeTag(name, val)` / `…Raw` | Define a theme semantic tag |
| `SetThemeMap(m)` | Replace the theme map wholesale |
| `SetRenderPolicy(fn)` | Gate rendering (return false → `ToPlain` instead of color) |
| `New()` | Create an independent `*Styler` |
| `(*Styler).SetDelimiters(…)` | Customize this Styler's tag delimiters |
| `(*Styler).RegisterHyperlinkTag(name)` | Make a tag render its content as a terminal hyperlink |

## Converting to lipgloss styles

`ToANSI` produces terminal escape sequences for plain output. When building a
**lipgloss** `Style` (e.g. for a TUI component), use `semtheme.ToStyle` instead — it
resolves any semantic or direct tags and applies the result directly to a lipgloss style:

```go
base := lipgloss.NewStyle()
style := semtheme.ToStyle(semstyle.Default, "{{|Error|}}", base, base)
```

Use `ToTags` directly when passing styled text to a TUI compositor that understands direct
tags natively rather than ANSI escapes.

## Delimiters

The default delimiters are `{{|`…`|}}` (semantic) and `{{[`…`]}}` (direct). Each `Styler`
gets these at construction and can override them independently:

```go
s := semstyle.New()
s.SetDelimiters("{|", "|}", "{[", "]}")   // this Styler only
```

`semstyle.SetDelimiters(...)` (package-level) changes the standard values and applies them
to `Default`. Delimiters are per-`Styler` state, so different stylers can use different
markup syntaxes in the same process.

## Hyperlinks

There are two ways to produce terminal hyperlinks (OSC 8).

### Registered hyperlink tags

Register a tag name — its enclosed content becomes the URL, displayed as-is:

```go
semstyle.RegisterHyperlinkTag("URL")
semstyle.ToANSI("{{|URL|}}https://example.com{{[-]}}") // URL is both destination and label
```

The registration lives on the `Styler` (independent of style maps), so it persists across
theme changes.

### Inline explicit hyperlinks

Add a label as the final colon-field of any tag. The enclosed content is the URL; the label
is the visible text. An empty label uses the URL as both:

**Direct tag** — label is the 4th field (`fg:bg:flags:label`):

```text
{{[cyan::U:DockSTARTer Website]}}https://dockstarter.com{{[-]}}
{{[cyan::U:]}}https://dockstarter.com{{[-]}}   ← empty label: URL shown as text
```

**Semantic tag** — label is the 5th field (`name:fg:bg:flags:label`); use empty fields to
keep the registered style with no color overrides:

```text
{{|mylink:red:black:B:DockSTARTer Website|}}https://dockstarter.com{{[-]}}
{{|mylink::::DockSTARTer Website|}}https://dockstarter.com{{[-]}}   ← no overrides
```

Both forms work whether or not the tag is also registered as a hyperlink tag — the explicit
label always takes precedence for the display text.

## Render policy

By default `ToANSI` always emits ANSI. A host that wants to suppress color when output is
redirected (or any other condition) sets a policy:

```go
semstyle.SetRenderPolicy(func() bool { return isTerminal() })
```

When the policy returns false, `ToANSI` strips instead of rendering.

## Theming

The `semstyle/theme` subpackage parses theme **files** (TOML) into a style map you can hand
to a styler. It's optional — import it only for file-driven themes.

```go
import (
    "github.com/GhostWriters/semstyle"
    semtheme "github.com/GhostWriters/semstyle/theme"
)

data, _ := os.ReadFile("midnight.theme")
defaults, _ := semtheme.RegisterInto(data, "") // parse + register into the Default styler
// `defaults` is the theme's opaque [defaults] table (map[string]any) for your app to interpret.
```

A theme file carries `[metadata]`, an optional `[palette]` (reusable `$vars`), `[styles]`
(semantic name → style, may reference palette vars or other styles), optional `[syntax]`
delimiter overrides, and an opaque `[defaults]` table that `semtheme` passes through without
interpreting (so any app can define whatever UI defaults it wants). Full details and the
theme-file format are in [theme/README.md](theme/README.md).

## Notes

- **Delimiters** and **hyperlink tags** are per-`Styler`; the package-level standard
  delimiter values seed each new `Styler`.
- The detected **color profile** is process-wide — it describes the output terminal, not an
  individual style configuration.
- Hard-reset constants (`CodeHardReset`, etc.) are multi-parameter SGR variants useful when
  a compositor intercepts single-parameter resets.

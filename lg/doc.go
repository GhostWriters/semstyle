// Package semlg provides lipgloss integration for semstyle.
//
// It bridges the pure-ANSI [semstyle] engine and [lipgloss] styles, letting you
// apply semstyle tag codes directly to lipgloss style values and maintain
// background colors across ANSI resets when compositing TUI components.
//
// # Style conversion
//
// Apply a raw fg:bg:flags code to a lipgloss style:
//
//	style := semlg.CodeToStyle("red::B", base, resetStyle)
//
// Resolve any semstyle tags (semantic or direct) and apply the result:
//
//	style := semlg.ToStyle(semstyle.Default, "{{|Error|}}", base, resetStyle)
//
// Parse only the flags portion into a [StyleFlags] struct:
//
//	flags := semlg.CodeToFlags("red::B")
//	style = flags.Apply(style)
//
// # Background maintenance
//
// When nesting lipgloss-rendered content inside a parent with a background color,
// inner ANSI resets bleed back to the terminal default. [MaintainBackground]
// intercepts those resets and re-asserts the parent style's colors:
//
//	safe := semlg.MaintainBackground(innerRendered, parentStyle)
package semlg

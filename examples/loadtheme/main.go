// Package main demonstrates loading a theme from an actual TOML file on disk
// (as opposed to an inline string) and using the styles it registers.
package main

import (
	"fmt"
	"os"

	"github.com/GhostWriters/semstyle"
	semtheme "github.com/GhostWriters/semstyle/theme"
)

func main() {
	// Register base console styles as a fallback for when no theme is loaded.
	semstyle.RegisterConsoleTag("Notice", "{{[cyan]}}")
	semstyle.RegisterConsoleTag("Warn", "{{[yellow]}}")
	semstyle.RegisterConsoleTag("Success", "{{[green]}}")
	semstyle.RegisterConsoleTag("Muted", "{{[white]}}")

	fmt.Println("--- Console defaults ---")
	fmt.Println(semstyle.ToANSI(
		"{{|Notice|}}info{{[-]}}  {{|Warn|}}warn{{[-]}}  {{|Success|}}ok{{[-]}}  {{|Muted|}}detail{{[-]}}",
	))

	// Load midnight.toml from disk and register its [styles] over the console map.
	data, err := os.ReadFile("midnight.toml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read theme:", err)
		os.Exit(1)
	}
	defaults, err := semtheme.RegisterInto(data, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse theme:", err)
		os.Exit(1)
	}

	fmt.Println("--- After loading midnight.toml ---")
	// Theme-registered styles live in the theme map, so resolving them requires
	// passing a prefix arg to ToANSI (empty string selects the unprefixed theme map).
	fmt.Println(semstyle.ToANSI(
		"{{|Notice|}}info{{[-]}}  {{|Warn|}}warn{{[-]}}  {{|Success|}}ok{{[-]}}  {{|Muted|}}detail{{[-]}}", "",
	))

	// Styles can reference other styles and layer modifier overrides on top
	// (e.g. "{{|Notice:::B|}}" reuses Notice's colors with bold added).
	fmt.Println(semstyle.ToANSI("{{|NoticeBold|}}bold notice{{[-]}}", ""))

	// [defaults] is an opaque, app-defined table the theme can carry alongside
	// its styles — e.g. border color, panel layout. The consuming app decodes
	// whatever keys it cares about.
	fmt.Printf("--- defaults table ---\n%#v\n", defaults)
}

// Package main demonstrates semstyle theming: loading a TOML theme file that
// overrides base console styles, and using per-instance Stylers for independent
// style surfaces.
package main

import (
	"fmt"
	"os"

	"github.com/GhostWriters/semstyle"
	semtheme "github.com/GhostWriters/semstyle/theme"
)

// theme.toml content used in this example:
//
//	[styles]
//	Notice  = "{{[#00CFFF]}}"
//	Warn    = "{{[#FFB347]}}"
//	Success = "{{[#90EE90::B]}}"
const sampleTOML = `
[styles]
Notice  = "{{[#00CFFF]}}"
Warn    = "{{[#FFB347]}}"
Success = "{{[#90EE90::B]}}"
`

func main() {
	// Register base console styles (fallback when no theme overrides).
	semstyle.RegisterConsoleTag("Notice", "{{[cyan]}}")
	semstyle.RegisterConsoleTag("Warn", "{{[yellow]}}")
	semstyle.RegisterConsoleTag("Success", "{{[green]}}")

	fmt.Println("--- Before theme ---")
	fmt.Println(semstyle.ToANSI("{{|Notice|}}info{{[-]}}  {{|Warn|}}warn{{[-]}}  {{|Success|}}ok{{[-]}}"))

	// Load a theme — styles in [styles] override the console map.
	defaults, err := semtheme.RegisterInto([]byte(sampleTOML), "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "theme error:", err)
		os.Exit(1)
	}
	_ = defaults // themes can carry app-specific [defaults] metadata

	fmt.Println("--- After theme ---")
	fmt.Println(semstyle.ToANSI("{{|Notice|}}info{{[-]}}  {{|Warn|}}warn{{[-]}}  {{|Success|}}ok{{[-]}}"))

	// Per-instance Styler: independent style surface, unaffected by Default.
	st := semstyle.New()
	st.RegisterConsoleTag("Notice", "{{[blue::B]}}")
	fmt.Println("--- Per-instance styler ---")
	fmt.Println(st.ToANSI("{{|Notice|}}my own blue notice{{[-]}}"))
}

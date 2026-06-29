// Package main demonstrates basic semstyle usage: registering semantic tags,
// writing styled text, and rendering to ANSI.
package main

import (
	"fmt"

	"github.com/GhostWriters/semstyle"
)

func main() {
	// Register named styles once at startup.
	semstyle.RegisterConsoleTag("Notice", "{{[cyan]}}")
	semstyle.RegisterConsoleTag("Warn", "{{[yellow]}}")
	semstyle.RegisterConsoleTag("Error", "{{[red::B]}}")
	semstyle.RegisterConsoleTag("Success", "{{[green]}}")

	// Use semantic tags in any string. The tag name is case-insensitive.
	fmt.Println(semstyle.ToANSI("{{|Notice|}}info:{{[-]}} everything is fine"))
	fmt.Println(semstyle.ToANSI("{{|Warn|}}warn:{{[-]}} disk usage above 80%"))
	fmt.Println(semstyle.ToANSI("{{|Error|}}error:{{[-]}} connection refused"))
	fmt.Println(semstyle.ToANSI("{{|Success|}}ok:{{[-]}} deployment complete"))

	// Direct tags inline a style without registration.
	fmt.Println(semstyle.ToANSI("{{[magenta::B]}}bold magenta{{[-]}} back to normal"))

	// Multi-part tags: reset first, then apply color (common pattern for log levels).
	semstyle.RegisterConsoleTag("Timestamp", "{{[-]}}{{[gray::D]}}")
	fmt.Println(semstyle.ToANSI("{{|Timestamp|}}2026-06-29 04:13:19{{[-]}} log message"))
}

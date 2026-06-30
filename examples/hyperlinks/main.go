// Package main demonstrates terminal hyperlinks (OSC 8) via semstyle.
// Hyperlinks only render in terminals that support OSC 8 (most modern ones do).
package main

import (
	"fmt"

	"github.com/GhostWriters/semstyle"
)

func main() {
	// Method 1: register a tag name as a hyperlink tag.
	// The enclosed content becomes the URL; the URL is also the visible label.
	semstyle.RegisterHyperlinkTag("URL")
	fmt.Println(semstyle.ToANSI("Visit {{|URL|}}https://github.com/GhostWriters/semstyle{{[-]}} for docs."))

	// Method 2: explicit URL as the last colon-field of a direct tag.
	// Format: {{[fg:bg:flags:url]}}Display Text{{[-]}}
	fmt.Println(semstyle.ToANSI(
		"{{[cyan::U:https://github.com/GhostWriters/semstyle]}}semstyle on GitHub{{[-]}}",
	))

	// Method 3: semantic tag with explicit URL.
	// Format: {{|name:fg:bg:flags:url|}}Display Text{{[-]}}
	// Empty fields inherit the registered style.
	semstyle.RegisterConsoleTag("Link", "{{[cyan::U]}}")
	fmt.Println(semstyle.ToANSI(
		"{{|Link::::https://github.com/GhostWriters/semstyle|}}semstyle on GitHub{{[-]}}",
	))

	// Empty URL field: content is used as both the link target and visible text.
	fmt.Println(semstyle.ToANSI(
		"{{[cyan::U:]}}https://github.com/GhostWriters/semstyle{{[-]}}",
	))
}

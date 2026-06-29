// Package main demonstrates the flags field in direct tags: bold, dim, underline,
// italic, blink, reverse, strikethrough, bright (H), and the reset prefix (-).
package main

import (
	"fmt"

	"github.com/GhostWriters/semstyle"
)

func main() {
	show := func(label, tag string) {
		fmt.Printf("%-30s %s\n", label, semstyle.ToANSI(tag+"demo{{[-]}}"))
	}

	show("Bold (B):",           "{{[::B]}}")
	show("Dim (D):",            "{{[::D]}}")
	show("Underline (U):",      "{{[::U]}}")
	show("Italic (I):",         "{{[::I]}}")
	show("Blink (L):",          "{{[::L]}}")
	show("Reverse (R):",        "{{[::R]}}")
	show("Strikethrough (S):",  "{{[::S]}}")
	show("Bright fg (H):",      "{{[red::H]}}")
	show("Bold+Dim:",           "{{[::BD]}}")
	show("Red bold:",           "{{[red::B]}}")
	show("White on red:",       "{{[white:red]}}")
	show("White on red bold:",  "{{[white:red:B]}}")

	// Leading - in flags resets all attributes then applies the new ones.
	// Useful after accumulated styles to start fresh.
	semstyle.RegisterConsoleTag("Disabled", "{{[::-D]}}")
	show("Reset+Dim (-D):",     "{{|Disabled|}}")

	// Chaining: bold blue text, then dim continuation.
	fmt.Println(semstyle.ToANSI("{{[blue::B]}}Title{{[-]}} {{[::D]}}(details){{[-]}}"))
}

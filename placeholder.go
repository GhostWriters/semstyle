package semstyle

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"charm.land/lipgloss/v2"
)

// Deferred tints: while deferral is on (see SetDeferTints), each color a tint
// can set renders as a placeholder truecolor value naming the tint key and
// the slot, instead of a final color. ResolveTints later swaps each
// placeholder for that key's current color, so text rendered once and kept
// (a cache, a scrollback buffer) always shows the tint that is current when
// it is drawn.
//
// A placeholder is #01PPSS: red 0x01 marks it, PP is placeholderIDBase plus
// the tint key's palette ID (see paletteID), and SS is placeholderSlotBase
// plus the slot -- 0-15 the ANSI colors, then the eight base24 slots with no
// ANSI index, in placeholderExtraSlots order. Palette ID 0 is "no key": it
// resolves to ResolveTints' default key. The high green and blue keep
// placeholders away from the near-black colors real output uses.

const (
	placeholderMarker   = 0x01
	placeholderIDBase   = 0xc0
	placeholderSlotBase = 0xe0
	placeholderMaxIDs   = 0x100 - placeholderIDBase
)

// placeholderExtraSlots are the slots after the 16 ANSI colors, in slot
// order (16, 17, ...).
var placeholderExtraSlots = [...]string{"base01", "base02", "base04", "base06", "base09", "base0f", "base10", "base11"}

var deferTints atomic.Bool

// SetDeferTints turns deferred tints on or off process-wide. Turn it on only
// around rendering whose output reaches the terminal through ResolveTints.
func SetDeferTints(on bool) {
	deferTints.Store(on)
}

// DeferTints reports whether deferred tints are on (see SetDeferTints).
func DeferTints() bool {
	return deferTints.Load()
}

var (
	paletteMu   sync.RWMutex
	paletteIDs  = map[string]int{"": 0}
	paletteKeys = []string{""}
)

// paletteID returns key's palette ID, assigning the next free one the first
// time key is seen. ok is false once all placeholderMaxIDs IDs are taken.
func paletteID(key string) (id int, ok bool) {
	paletteMu.RLock()
	id, ok = paletteIDs[key]
	paletteMu.RUnlock()
	if ok {
		return id, true
	}
	paletteMu.Lock()
	defer paletteMu.Unlock()
	if id, ok = paletteIDs[key]; ok {
		return id, true
	}
	if len(paletteKeys) >= placeholderMaxIDs {
		return 0, false
	}
	id = len(paletteKeys)
	paletteIDs[key] = id
	paletteKeys = append(paletteKeys, key)
	return id, true
}

// paletteKey returns the tint key with palette ID id.
func paletteKey(id int) (string, bool) {
	paletteMu.RLock()
	defer paletteMu.RUnlock()
	if id < 0 || id >= len(paletteKeys) {
		return "", false
	}
	return paletteKeys[id], true
}

// placeholderSlot returns the slot number for colorName (already lowercased),
// or false if no tint can set it.
func placeholderSlot(colorName string) (int, bool) {
	for i, name := range placeholderExtraSlots {
		if name == colorName {
			return 16 + i, true
		}
	}
	if idx, ok := ansiColorIndex[colorName]; ok {
		n, _ := strconv.Atoi(idx)
		return n, true
	}
	return 0, false
}

// slotName returns the color name for slot.
func slotName(slot int) (string, bool) {
	switch {
	case slot >= 0 && slot < 8:
		return BasicColors[slot], true
	case slot >= 8 && slot < 16:
		return "bright-" + BasicColors[slot-8], true
	case slot >= 16 && slot < 16+len(placeholderExtraSlots):
		return placeholderExtraSlots[slot-16], true
	}
	return "", false
}

// placeholderHex returns the placeholder for colorName (already lowercased)
// under key, or false if no tint can set colorName.
func placeholderHex(key, colorName string) (string, bool) {
	slot, ok := placeholderSlot(colorName)
	if !ok {
		return "", false
	}
	id, ok := paletteID(key)
	if !ok {
		return "", false
	}
	return encodePlaceholder(id, slot), true
}

// encodePlaceholder returns the "#rrggbb" placeholder for palette ID id and
// slot.
func encodePlaceholder(id, slot int) string {
	return fmt.Sprintf("#%02x%02x%02x", placeholderMarker, placeholderIDBase+id, placeholderSlotBase+slot)
}

// deferredColor returns the placeholder to render for colorName (already
// lowercased) under the applicable tint key, when deferral is on.
func deferredColor(key, colorName string) (string, bool) {
	if !DeferTints() {
		return "", false
	}
	return placeholderHex(key, colorName)
}

// avoidPlaceholder nudges a literal "#rrggbb" color in the placeholder range,
// so it can never be mistaken for a placeholder. Any other value is returned
// unchanged.
func avoidPlaceholder(hex string) string {
	if !DeferTints() {
		return hex
	}
	r, g, b, ok := parseHex(hex)
	if !ok || r != placeholderMarker || g < placeholderIDBase || b < placeholderSlotBase {
		return hex
	}
	return fmt.Sprintf("#%02x%02x%02x", r+1, g, b)
}

// decodePlaceholder reports whether r, g, b is a placeholder, and its
// palette ID and slot.
func decodePlaceholder(r, g, b int) (id, slot int, ok bool) {
	if r != placeholderMarker || g < placeholderIDBase || b < placeholderSlotBase {
		return 0, 0, false
	}
	id, slot = g-placeholderIDBase, b-placeholderSlotBase
	if _, named := slotName(slot); !named {
		return 0, 0, false
	}
	if _, known := paletteKey(id); !known {
		return 0, 0, false
	}
	return id, slot, true
}

// tintSnapshot resolves placeholders against one snapshot of each palette it
// looks up, so a tint re-registered partway through resolving one string
// can't leave that string mixing old and new colors.
type tintSnapshot struct {
	defaultKey string
	palettes   map[int]*Palette // nil entry: that palette ID is untinted
}

func newTintSnapshot(defaultKey string) *tintSnapshot {
	return &tintSnapshot{defaultKey: defaultKey, palettes: map[int]*Palette{}}
}

// slot returns the hex color the tint for palette ID id sets for slot, or ""
// when that palette is untinted or doesn't set it.
func (t *tintSnapshot) slot(id, slot int) string {
	p, seen := t.palettes[id]
	if !seen {
		key, _ := paletteKey(id)
		if id == 0 {
			key = t.defaultKey
		}
		if key != "" {
			if found, ok := getTint(key); ok {
				p = &found
			}
		}
		t.palettes[id] = p
	}
	if p == nil {
		return ""
	}
	name, _ := slotName(slot)
	return p.slot(name)
}

// untintedIndex returns the ANSI index slot renders as without a tint.
func untintedIndex(slot int) int {
	if slot < 16 {
		return slot
	}
	name, _ := slotName(slot)
	n, _ := strconv.Atoi(ansiColorIndex[slotANSIFallback[name]])
	return n
}

// ResolveColor is ResolveColorWithDefault with the active tint key (see
// SetActiveTint) as the default key. Inside a render pass that also passes
// that key to ResolveTints, both resolve a placeholder the same way.
func ResolveColor(c color.Color) color.Color {
	return ResolveColorWithDefault(c, getActiveTintKey())
}

// ResolveColorWithDefault returns c with a placeholder resolved to its
// tint's current color (or its ANSI color without one), for code that needs
// c's real RGB values; a placeholder rendered with no key resolves with
// defaultKey, as ResolveTints does. Any other color is returned unchanged.
func ResolveColorWithDefault(c color.Color, defaultKey string) color.Color {
	if c == nil {
		return c
	}
	r, g, b, a := c.RGBA()
	if a == 0 {
		return c
	}
	id, slot, ok := decodePlaceholder(int(r>>8), int(g>>8), int(b>>8))
	if !ok {
		return c
	}
	if hex := newTintSnapshot(defaultKey).slot(id, slot); hex != "" {
		return lipgloss.Color(hex)
	}
	return lipgloss.Color(strconv.Itoa(untintedIndex(slot)))
}

// BrightPlaceholder returns the bright form of c when c is a placeholder (see
// SetDeferTints): the bright slot for one of the 8 normal ANSI colors, else
// c itself. ok is false when c isn't a placeholder.
func BrightPlaceholder(c color.Color) (bright color.Color, ok bool) {
	if c == nil {
		return c, false
	}
	r, g, b, a := c.RGBA()
	if a == 0 {
		return c, false
	}
	id, slot, ok := decodePlaceholder(int(r>>8), int(g>>8), int(b>>8))
	if !ok {
		return c, false
	}
	if slot < 8 {
		slot += 8
	}
	return lipgloss.Color(encodePlaceholder(id, slot)), true
}

// ResolveTints returns s with every placeholder in its SGR sequences (see
// SetDeferTints) replaced by the current color of the tint it names, or by
// the plain ANSI color when that tint is unset or doesn't set the slot.
// Palette ID 0 resolves with defaultKey.
func ResolveTints(s, defaultKey string) string {
	if !strings.Contains(s, ";2;1;") {
		return s
	}
	snap := newTintSnapshot(defaultKey)
	var out strings.Builder
	out.Grow(len(s))
	for {
		start := strings.Index(s, "\x1b[")
		if start < 0 {
			out.WriteString(s)
			break
		}
		end := start + 2
		for end < len(s) && (s[end] >= '0' && s[end] <= '9' || s[end] == ';' || s[end] == ':') {
			end++
		}
		if end >= len(s) {
			out.WriteString(s)
			break
		}
		out.WriteString(s[:start])
		params := s[start+2 : end]
		if s[end] == 'm' && strings.Contains(params, ";2;1;") {
			out.WriteString("\x1b[")
			out.WriteString(resolveSGRParams(params, snap))
		} else {
			out.WriteString(s[start:end])
		}
		s = s[end:]
	}
	return out.String()
}

// resolveSGRParams resolves the placeholders in one SGR sequence's
// semicolon-separated parameters.
func resolveSGRParams(params string, snap *tintSnapshot) string {
	parts := strings.Split(params, ";")
	out := make([]string, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if (p == "38" || p == "48" || p == "58") && i+4 < len(parts) && parts[i+1] == "2" {
			r, errR := strconv.Atoi(parts[i+2])
			g, errG := strconv.Atoi(parts[i+3])
			b, errB := strconv.Atoi(parts[i+4])
			if errR == nil && errG == nil && errB == nil {
				if id, slot, ok := decodePlaceholder(r, g, b); ok {
					out = append(out, resolvedSGR(p, id, slot, snap))
					i += 4
					continue
				}
			}
		}
		out = append(out, p)
	}
	return strings.Join(out, ";")
}

// resolvedSGR returns the SGR parameters for a placeholder of kind 38 (fg),
// 48 (bg) or 58 (underline).
func resolvedSGR(kind string, id, slot int, snap *tintSnapshot) string {
	if hex := snap.slot(id, slot); hex != "" {
		if r, g, b, ok := parseHex(hex); ok {
			return fmt.Sprintf("%s;2;%d;%d;%d", kind, r, g, b)
		}
	}
	idx := untintedIndex(slot)
	switch kind {
	case "38":
		if idx < 8 {
			return strconv.Itoa(30 + idx)
		}
		return strconv.Itoa(90 + idx - 8)
	case "48":
		if idx < 8 {
			return strconv.Itoa(40 + idx)
		}
		return strconv.Itoa(100 + idx - 8)
	}
	return fmt.Sprintf("%s;5;%d", kind, idx)
}

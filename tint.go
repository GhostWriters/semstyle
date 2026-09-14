package semstyle

import (
	"context"
	"sync"
)

// Palette holds the 16 standard ANSI colors as hex values ("#rrggbb").
// An empty field falls through to normal (untinted) resolution for that
// slot -- a Palette is never all-or-nothing.
type Palette struct {
	Black, Red, Green, Yellow, Blue, Magenta, Cyan, White                                                 string
	BrightBlack, BrightRed, BrightGreen, BrightYellow, BrightBlue, BrightMagenta, BrightCyan, BrightWhite string
}

// slot returns the hex value registered for the given ANSI color name
// (already lowercased, e.g. "red", "bright-red"), or "" if this palette
// doesn't set that slot. Also accepts each name's base16/base24 slot alias
// (see ansiColorIndex's matching comment) so a tint substitutes correctly
// regardless of which vocabulary a tag definition used.
func (p Palette) slot(name string) string {
	switch name {
	case "black", "base00":
		return p.Black
	case "red", "base08":
		return p.Red
	case "green", "base0b":
		return p.Green
	case "yellow", "base0a":
		return p.Yellow
	case "blue", "base0d":
		return p.Blue
	case "magenta", "base0e":
		return p.Magenta
	case "cyan", "base0c":
		return p.Cyan
	case "white", "base05":
		return p.White
	case "bright-black", "base03":
		return p.BrightBlack
	case "bright-red", "base12":
		return p.BrightRed
	case "bright-green", "base14":
		return p.BrightGreen
	case "bright-yellow", "base13":
		return p.BrightYellow
	case "bright-blue", "base16":
		return p.BrightBlue
	case "bright-magenta", "base17":
		return p.BrightMagenta
	case "bright-cyan", "base15":
		return p.BrightCyan
	case "bright-white", "base07":
		return p.BrightWhite
	default:
		return ""
	}
}

var (
	tintsMu sync.RWMutex
	tints   = map[string]Palette{}
)

// RegisterTint registers a named Palette, so that rendering done with a
// context carrying that key (see WithTint) substitutes each of the 16
// standard ANSI colors it sets with a literal truecolor value, instead of
// the plain ANSI-index code those names otherwise resolve to. Registering
// under an already-used key replaces it.
func RegisterTint(key string, p Palette) {
	tintsMu.Lock()
	defer tintsMu.Unlock()
	tints[key] = p
}

// UnregisterTint removes a previously registered tint. A no-op if key
// isn't registered.
func UnregisterTint(key string) {
	tintsMu.Lock()
	defer tintsMu.Unlock()
	delete(tints, key)
}

func getTint(key string) (Palette, bool) {
	tintsMu.RLock()
	defer tintsMu.RUnlock()
	p, ok := tints[key]
	return p, ok
}

type tintCtxKey struct{}

// WithTint returns a context carrying key, selecting which registered
// tint (if any) applies to color resolution done with that context via
// ToANSICtx/SprintfCtx. An empty key or one with no matching
// RegisterTint call means no tint applies -- identical to not calling
// WithTint at all.
//
// Most callers don't need this: see SetActiveTint for the common
// single-active-tint case.
func WithTint(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, tintCtxKey{}, key)
}

var (
	activeTintMu  sync.RWMutex
	activeTintKey string
)

// SetActiveTint sets which registered tint (see RegisterTint) applies to
// every non-ctx call (ToANSI, ToColor, ToStyle, CodeToStyle, Sprintf, ...)
// process-wide, and to any ctx-aware call whose context carries no tint of
// its own (see WithTint) -- WithTint always wins when both are set. Pass ""
// to clear it (no tint applies to non-ctx calls).
//
// This is the simple option for a process that only ever wants one active
// tint at a time (e.g. a single local session); a process juggling several
// concurrent sessions with different tints should use WithTint per-session
// instead, since this setting is shared by every caller in the process.
func SetActiveTint(key string) {
	activeTintMu.Lock()
	defer activeTintMu.Unlock()
	activeTintKey = key
}

func getActiveTintKey() string {
	activeTintMu.RLock()
	defer activeTintMu.RUnlock()
	return activeTintKey
}

// ActiveTintKey returns the key currently set via SetActiveTint (or made
// active by an in-progress RunWithTint call), or "" if none. Useful as part
// of a cache key for any caller that memoizes semstyle's non-ctx rendering
// output -- that output depends on the active tint, so a cache keyed only
// on the input text would return stale results across a tint change or
// between two RunWithTint scopes using different keys.
func ActiveTintKey() string {
	return getActiveTintKey()
}

// runWithTintMu serializes RunWithTint calls process-wide. SetActiveTint
// itself only guards the variable, not a "set, render, restore" sequence --
// two goroutines calling SetActiveTint around their own render passes could
// still interleave and see each other's key. RunWithTint owns this lock so
// every caller gets that safety for free instead of each needing its own
// mutex (see its doc comment).
var runWithTintMu sync.Mutex

// RunWithTint makes key the active tint (see SetActiveTint) for the
// duration of fn, then restores whatever was active before, all while
// holding a package-level lock so concurrent callers can't interleave and
// see each other's key mid-render.
//
// This is for callers who need SetActiveTint's process-wide-default
// behavior but can't thread a context.Context down to the render call site
// to use WithTint/ToANSICtx instead (e.g. rendering through a fixed
// callback interface like Bubble Tea's View() string). It serializes those
// callers against each other -- fn from two goroutines never runs
// concurrently -- so pick this only when that render work is cheap enough
// to serialize; a caller that can thread ctx through instead should prefer
// WithTint, which has no such restriction.
func RunWithTint(key string, fn func()) {
	runWithTintMu.Lock()
	defer runWithTintMu.Unlock()
	prev := getActiveTintKey()
	SetActiveTint(key)
	defer SetActiveTint(prev)
	fn()
}

// BeginTint is RunWithTint split into a begin/restore pair instead of a
// closure, for a caller whose "duration" isn't a single function call it
// can wrap -- e.g. a long, multi-step startup sequence with early returns
// scattered through it, where wrapping everything in one closure would
// either miss code after an early return or require restructuring that
// control flow just to fit the closure shape. Call restore (typically via
// defer, right after Begin) exactly once when done.
//
// Holds the same lock as RunWithTint for the entire span between the call
// and restore -- fine for a single-threaded-with-respect-to-tint caller
// (e.g. one bare CLI invocation, which never needs a second, different
// tint active concurrently with itself), but never call this from
// something that must interleave with other RunWithTint/BeginTint callers
// during that span, or they'll block for the whole duration instead of
// just one render.
func BeginTint(key string) (restore func()) {
	runWithTintMu.Lock()
	prev := getActiveTintKey()
	SetActiveTint(key)
	return func() {
		SetActiveTint(prev)
		runWithTintMu.Unlock()
	}
}

// tintColorForCtx returns the literal hex value the applicable tint (ctx's,
// if it carries one via WithTint, else the process-wide active tint set via
// SetActiveTint) sets for colorName (already lowercased), and whether one
// was found. Safe to call with a nil ctx.
func tintColorForCtx(ctx context.Context, colorName string) (string, bool) {
	key := ""
	if ctx != nil {
		key, _ = ctx.Value(tintCtxKey{}).(string)
	}
	if key == "" {
		key = getActiveTintKey()
	}
	if key == "" {
		return "", false
	}
	p, ok := getTint(key)
	if !ok {
		return "", false
	}
	hex := p.slot(colorName)
	if hex == "" {
		return "", false
	}
	return hex, true
}

package semstyle

import (
	"strings"
)

// Per-Styler map state lives on the Styler struct (see styler.go). The maps are created in
// New(); ensureMaps is a defensive guard for zero-value access.

// ensureMaps ensures color maps are built if they were missed by init
func (st *Styler) ensureMaps() {
	if len(st.ansiMap) == 0 {
		st.BuildColorMap()
	}
}

// BuildColorMap initializes the ANSI code and attribute name mappings.
// Default semantic tag registrations are handled separately by RegisterBaseTags.
func (st *Styler) BuildColorMap() {
	if st.ansiMap == nil {
		st.ansiMap = make(map[string]string)
	}
	if st.consoleMap == nil {
		st.consoleMap = make(map[string]string)
	}
	if st.themeMap == nil {
		st.themeMap = make(map[string]string)
	}
	if st.attributeMap == nil {
		st.attributeMap = make(map[string]string)
	}

	// Standard ANSI mappings
	st.ansiMap["-"] = CodeReset
	st.ansiMap["reset"] = CodeReset
	st.ansiMap["B"] = CodeBold
	st.ansiMap["b"] = CodeBoldOff
	st.ansiMap["D"] = CodeDim
	st.ansiMap["d"] = CodeDimOff
	st.ansiMap["U"] = CodeUnderline
	st.ansiMap["u"] = CodeUnderlineOff
	st.ansiMap["L"] = CodeBlink
	st.ansiMap["l"] = CodeBlinkOff
	st.ansiMap["R"] = CodeReverse
	st.ansiMap["r"] = CodeReverseOff
	st.ansiMap["I"] = CodeItalic
	st.ansiMap["i"] = CodeItalicOff
	st.ansiMap["S"] = CodeStrikethrough
	st.ansiMap["s"] = CodeStrikethroughOff

	// Attribute mappings — reset only; per-attribute on/off is handled by ansiMap in the flags field
	st.attributeMap["reset"] = CodeReset
	st.attributeMap["-"] = CodeReset

	// Colors...
	st.ansiMap["black"] = CodeBlack
	st.ansiMap["red"] = CodeRed
	st.ansiMap["green"] = CodeGreen
	st.ansiMap["yellow"] = CodeYellow
	st.ansiMap["blue"] = CodeBlue
	st.ansiMap["magenta"] = CodeMagenta
	st.ansiMap["cyan"] = CodeCyan
	st.ansiMap["white"] = CodeWhite
	st.ansiMap["bright-black"] = CodeBrightBlack
	st.ansiMap["bright-red"] = CodeBrightRed
	st.ansiMap["bright-green"] = CodeBrightGreen
	st.ansiMap["bright-yellow"] = CodeBrightYellow
	st.ansiMap["bright-blue"] = CodeBrightBlue
	st.ansiMap["bright-magenta"] = CodeBrightMagenta
	st.ansiMap["bright-cyan"] = CodeBrightCyan
	st.ansiMap["bright-white"] = CodeBrightWhite

	st.ansiMap["blackbg"] = CodeBlackBg
	st.ansiMap["redbg"] = CodeRedBg
	st.ansiMap["greenbg"] = CodeGreenBg
	st.ansiMap["yellowbg"] = CodeYellowBg
	st.ansiMap["bluebg"] = CodeBlueBg
	st.ansiMap["magentabg"] = CodeMagentaBg
	st.ansiMap["cyanbg"] = CodeCyanBg
	st.ansiMap["whitebg"] = CodeWhiteBg
	st.ansiMap["bright-blackbg"] = CodeBrightBlackBg
	st.ansiMap["bright-redbg"] = CodeBrightRedBg
	st.ansiMap["bright-greenbg"] = CodeBrightGreenBg
	st.ansiMap["bright-yellowbg"] = CodeBrightYellowBg
	st.ansiMap["bright-bluebg"] = CodeBrightBlueBg
	st.ansiMap["bright-magentabg"] = CodeBrightMagentaBg
	st.ansiMap["bright-cyanbg"] = CodeBrightCyanBg
	st.ansiMap["bright-whitebg"] = CodeBrightWhiteBg

}

// RegisterConsoleTag registers a semantic tag with its standardized tag value in the console map.
func (st *Styler) RegisterConsoleTag(name, taggedValue string) {
	stripped := st.StripDelimiters(taggedValue)
	// If stripping produced a value that still contains delimiters the input was
	// multi-part (e.g. "{{[-]}}{{[gray::D]}}"); store the original tagged form so
	// ExpandTagsWithMap can return it intact rather than re-wrapping as one tag.
	if strings.Contains(stripped, st.dirPre) || strings.Contains(stripped, st.semPre) {
		st.RegisterConsoleTagRaw(name, taggedValue)
	} else {
		st.RegisterConsoleTagRaw(name, stripped)
	}
}

// RegisterConsoleTagRaw registers a semantic tag with a raw style code in the console map.
func (st *Styler) RegisterConsoleTagRaw(name, rawValue string) {
	st.ensureMaps()
	st.mu.Lock()
	st.consoleMap[strings.ToLower(name)] = rawValue
	st.mu.Unlock()
}

// RegisterThemeTag registers a semantic tag with its standardized tag value in the theme map.
func (st *Styler) RegisterThemeTag(name, taggedValue string) {
	stripped := st.StripDelimiters(taggedValue)
	if strings.Contains(stripped, st.dirPre) || strings.Contains(stripped, st.semPre) {
		st.RegisterThemeTagRaw(name, taggedValue)
	} else {
		st.RegisterThemeTagRaw(name, stripped)
	}
}

// RegisterThemeTagRaw registers a semantic tag with a raw style code in the theme map.
func (st *Styler) RegisterThemeTagRaw(name, rawValue string) {
	st.ensureMaps()
	st.mu.Lock()
	st.themeMap[strings.ToLower(name)] = rawValue
	st.mu.Unlock()
}

// maxFallbackDepth caps RegisterFallback chain-following so a mistaken or
// accidental cycle (A falls back to B falls back to A) can't hang resolution
// -- it just gives up and resolves to unset past this many hops.
const maxFallbackDepth = 8

// themeOnlyLookup resolves name against styleMap (nil means themeMap),
// with no console-map consultation at all. Callers must already hold
// st.mu (read or write); this method takes no lock itself.
func (st *Styler) themeOnlyLookup(styleMap map[string]string, prefix, name string) (string, bool) {
	m := styleMap
	if m == nil {
		m = st.themeMap
	}
	if prefix != "" {
		if raw, ok := m[prefix+name]; ok {
			return raw, true
		}
	}
	if raw, ok := m[name]; ok {
		return raw, true
	}
	return "", false
}

// directLookup resolves name via themeOnlyLookup, then -- if still
// unresolved -- the console map, unless disableAutoConsoleFallback is set
// (see SetAutoConsoleFallback). This is the three-tier lookup GetRawTagCode
// and ExpandTagsWithMap have always done for the primary name being
// resolved, and is also used to resolve an individual followChains=false
// fallback candidate. An explicit ConsoleTag candidate (consoleOnlyLookup)
// is unaffected by the toggle either way, since that's an intentional
// per-tag opt-in rather than this automatic tier. Callers must already
// hold st.mu (read or write); this method takes no lock itself.
func (st *Styler) directLookup(styleMap map[string]string, prefix, name string) (string, bool) {
	if raw, ok := st.themeOnlyLookup(styleMap, prefix, name); ok {
		return raw, true
	}
	if !st.disableAutoConsoleFallback {
		if raw, ok := st.consoleMap[name]; ok {
			return raw, true
		}
	}
	return "", false
}

// consoleOnlyLookup resolves name against the console map only, regardless
// of disableAutoConsoleFallback -- used for an explicit "console:name"
// fallback candidate, a deliberate per-tag opt-in rather than the automatic
// tier that setting controls. Callers must already hold st.mu.
func (st *Styler) consoleOnlyLookup(name string) (string, bool) {
	raw, ok := st.consoleMap[name]
	return raw, ok
}

// AutoConsoleFallback reports whether this Styler automatically checks the
// console map when a theme-mode lookup doesn't find name in the theme map.
// True by default (the library's original behavior). See
// SetAutoConsoleFallback.
func (st *Styler) AutoConsoleFallback() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return !st.disableAutoConsoleFallback
}

// SetAutoConsoleFallback enables or disables automatic console-map fallback
// for theme-mode lookups (GetRawTagCode, GetColorDefinition, and inline
// "{{|name|}}" text expansion when using a prefix/theme map). Enabled by
// default, matching the library's original behavior: an undefined
// theme-mode tag silently resolves to whatever base/console default
// happens to share its name -- which can surprise a theme author who
// simply forgot to define that tag, since the color that appears may have
// nothing to do with their theme. Disabling this makes an undefined theme
// tag resolve to nothing (unstyled) instead, so a missing definition is
// visible rather than silently masked by an unrelated color. A caller that
// wants specific tags to still fall back to console after disabling this
// can register that explicitly per tag with RegisterFallback's
// "console:name" candidate syntax.
func (st *Styler) SetAutoConsoleFallback(enabled bool) {
	st.mu.Lock()
	st.disableAutoConsoleFallback = !enabled
	st.mu.Unlock()
}

// fallbackCandidate is one entry in a fallbackRule's candidate list.
//   - literal: a pre-resolved raw style code (from a direct-tag-syntax
//     argument, e.g. "{{[white:black:B]}}"), used as-is with no lookup.
//   - consoleOnly: value is looked up in the console map only (from a
//     "console:name" argument), regardless of disableAutoConsoleFallback
//     and with no further chaining -- an explicit, deliberate opt-in.
//   - otherwise: value is a tag name, resolved the normal way.
type fallbackCandidate struct {
	literal     bool
	consoleOnly bool
	value       string // raw code if literal, lowercased tag name otherwise
}

// fallbackRule is one name's registered RegisterFallback rule: an ordered
// list of candidates (a single-element list for the common one-fallback
// case) plus whether each non-literal candidate follows its own separate
// rule in turn.
type fallbackRule struct {
	candidates   []fallbackCandidate
	followChains bool
}

// lookupRaw resolves name via themeOnlyLookup, then -- if still unresolved
// -- consults name's registered fallback rule (RegisterFallback) if any,
// checked BEFORE the automatic console-map tier: an explicit rule for name
// is a more specific, intentional statement than the blanket "any tag can
// silently pick up a same-named console default" tier, so it takes
// priority. Only once name has neither a theme value nor a registered rule
// does resolution fall through to the automatic console tier (unless
// disabled via SetAutoConsoleFallback) as the true last resort. A single
// non-literal candidate with followChains=true is walked in this same loop
// (so cycle detection covers the whole chain, e.g. A -> B -> C); anything
// else (a literal candidate, multiple candidates, or followChains=false)
// is resolved per-candidate instead (see the loop body for why). Callers
// must already hold st.mu (read or write); this method takes no lock
// itself.
func (st *Styler) lookupRaw(styleMap map[string]string, prefix, name string) (string, bool) {
	seen := make(map[string]bool, maxFallbackDepth)
	for range maxFallbackDepth {
		if seen[name] {
			return "", false // cycle guard
		}
		seen[name] = true

		if raw, ok := st.themeOnlyLookup(styleMap, prefix, name); ok {
			return raw, true
		}

		rule, hasRule := st.fallbackMap[name]
		if !hasRule {
			if !st.disableAutoConsoleFallback {
				if raw, ok := st.consoleMap[name]; ok {
					return raw, true
				}
			}
			return "", false
		}

		if len(rule.candidates) == 1 && rule.followChains &&
			!rule.candidates[0].literal && !rule.candidates[0].consoleOnly {
			// The common single-fallback-that-chains case: continue this
			// same loop (not a recursive call) so seen/maxFallbackDepth
			// cover the whole chain in one place, exactly as a multi-hop
			// A -> B -> C chain needs.
			name = rule.candidates[0].value
			continue
		}
		for _, candidate := range rule.candidates {
			if candidate.literal {
				return candidate.value, true
			}
			if candidate.consoleOnly {
				if raw, ok := st.consoleOnlyLookup(candidate.value); ok {
					return raw, true
				}
				continue
			}
			if rule.followChains {
				if raw, ok := st.lookupRaw(styleMap, prefix, candidate.value); ok {
					return raw, true
				}
				continue
			}
			if raw, ok := st.directLookup(styleMap, prefix, candidate.value); ok {
				return raw, true
			}
		}
		return "", false
	}
	return "", false
}

// ResolveFallbackViaLookup resolves name's registered fallback rule (if any)
// using lookup for non-literal, non-console-only candidates instead of this
// Styler's own theme map -- for a caller (e.g. a theme-file parser) that
// needs to honor fallback rules while resolving a reference against its own
// data (a theme's raw [styles] values) rather than an already-registered
// Styler map. Literal candidates are returned as-is; console-tagged
// candidates still resolve against this Styler's console map via
// consoleOnlyLookup. lookup should return (rawValue, true) if it has its own
// definition for the given candidate name, or ("", false) otherwise -- it is
// not expected to itself chase fallback rules; ResolveFallbackViaLookup
// handles that by recursing when a candidate has no lookup hit and the rule
// (or the candidate's own rule) says to follow chains. Returns ("", false)
// if name has no registered rule or nothing resolves.
func ResolveFallbackViaLookup(name string, lookup func(string) (string, bool)) (string, bool) {
	return Default.ResolveFallbackViaLookup(name, lookup)
}

// ResolveFallbackViaLookup is the Styler method behind the package-level
// ResolveFallbackViaLookup func. See its doc comment for behavior.
func (st *Styler) ResolveFallbackViaLookup(name string, lookup func(string) (string, bool)) (string, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.resolveFallbackViaLookup(strings.ToLower(name), lookup, make(map[string]bool, maxFallbackDepth))
}

func (st *Styler) resolveFallbackViaLookup(name string, lookup func(string) (string, bool), seen map[string]bool) (string, bool) {
	if seen[name] {
		return "", false // cycle guard
	}
	seen[name] = true

	rule, hasRule := st.fallbackMap[name]
	if !hasRule {
		return "", false
	}

	if len(rule.candidates) == 1 && rule.followChains &&
		!rule.candidates[0].literal && !rule.candidates[0].consoleOnly {
		cand := rule.candidates[0].value
		if raw, ok := lookup(cand); ok {
			return raw, true
		}
		return st.resolveFallbackViaLookup(cand, lookup, seen)
	}
	for _, candidate := range rule.candidates {
		if candidate.literal {
			return candidate.value, true
		}
		if candidate.consoleOnly {
			if raw, ok := st.consoleOnlyLookup(candidate.value); ok {
				return raw, true
			}
			continue
		}
		if raw, ok := lookup(candidate.value); ok {
			return raw, true
		}
		if rule.followChains {
			if raw, ok := st.resolveFallbackViaLookup(candidate.value, lookup, seen); ok {
				return raw, true
			}
		}
	}
	return "", false
}

// RegisterFallback declares one or more fallback candidates for name: when
// name isn't registered in the theme or console map, tag resolution tries
// each candidate in order and uses the first that resolves. Applies
// everywhere a tag name is resolved -- GetRawTagCode, GetColorDefinition,
// and inline "{{|name|}}" text expansion -- not just one call site, so a
// caller need not re-derive the fallback at each place it references the
// tag.
//
// Each candidate is normally a tag name (with or without this Styler's
// semantic delimiters, e.g. "Title" or "{{|Title|}}" are equivalent). A
// candidate written with direct-tag delimiters instead (e.g.
// "{{[white:black:B]}}") is a literal: used as a final raw style code with
// no lookup or further chaining, for a caller that wants an inline default
// rather than another named tag. A candidate produced by ConsoleTag(name)
// is looked up in the console map only, regardless of
// SetAutoConsoleFallback -- for a caller that has disabled the automatic
// console tier but still wants specific tags to fall back to it
// deliberately:
//
//	semstyle.RegisterFallback("TitleWarn", true, "{{|Title|}}")            // same as "Title"
//	semstyle.RegisterFallback("TitleWarn", true, "{{[white:black:B]}}")    // literal, no lookup
//	semstyle.RegisterFallback("TitleWarn", true, semstyle.ConsoleTag("Title")) // console map only
//
// followChains controls whether a non-literal candidate that isn't itself
// directly registered also follows its own separate RegisterFallback rule
// in turn (irrelevant for a literal candidate, which is never looked up):
//
//   - true (the common case, e.g. a single fallback that may itself have a
//     fallback): a candidate is resolved the same way GetRawTagCode would,
//     following its own rule too. A single candidate with followChains=true
//     is a plain single-fallback chain -- RegisterFallback("TitleWarn",
//     true, "Title") -- and chains transitively (up to maxFallbackDepth
//     hops, with a cycle guard) if "Title" itself has its own rule.
//   - false: only a candidate's own direct theme/console value counts. A
//     candidate used this way doesn't inherit whatever fallback it might
//     separately have when resolved on its own elsewhere -- useful for an
//     ordered list of candidates that should stay self-contained to name
//     rather than reaching into each candidate's independent wiring.
//
// Registering again for the same name replaces its previous rule entirely.
func (st *Styler) RegisterFallback(name string, followChains bool, candidates ...string) {
	st.ensureMaps()
	st.mu.Lock()
	if st.fallbackMap == nil {
		st.fallbackMap = make(map[string]fallbackRule)
	}
	parsed := make([]fallbackCandidate, len(candidates))
	for i, c := range candidates {
		parsed[i] = st.parseFallbackCandidate(c)
	}
	st.fallbackMap[strings.ToLower(name)] = fallbackRule{candidates: parsed, followChains: followChains}
	st.mu.Unlock()
}

// consoleTagMarker prefixes a ConsoleTag(name)-produced string. It starts
// with a NUL byte specifically so it can never collide with anything a
// human would type as a plain candidate string (e.g. a bare "console:Title"
// string would be ambiguous with the tag:fgColor modifier convention used
// elsewhere in this package's tag syntax -- is "console" a tag name and
// "Title" a color override, or a namespace? A control-character marker a
// caller can only produce via ConsoleTag has no such reading).
const consoleTagMarker = "\x00console:"

// ConsoleTag returns a RegisterFallback candidate value meaning "look name
// up in the console map only, regardless of SetAutoConsoleFallback" -- use
// as one of RegisterFallback's candidates to opt a specific tag back into
// console fallback after disabling the automatic tier for this Styler.
func ConsoleTag(name string) string {
	return consoleTagMarker + name
}

// parseFallbackCandidate classifies one RegisterFallback candidate argument:
// a ConsoleTag(name) marker (checked first) marks the remainder
// consoleOnly; direct-tag syntax (st.dirPre/dirSuf) becomes a literal raw
// code; anything else (a bare name, or semantic-tag syntax st.semPre/
// semSuf) becomes a lowercased tag name. Mirrors the direct-vs-semantic
// detection RegisterThemeTag/RegisterConsoleTag already do for a
// registered value.
func (st *Styler) parseFallbackCandidate(c string) fallbackCandidate {
	if rest, ok := strings.CutPrefix(c, consoleTagMarker); ok {
		fc := st.parseFallbackCandidate(rest)
		fc.consoleOnly = true
		return fc
	}
	if strings.HasPrefix(c, st.dirPre) && strings.HasSuffix(c, st.dirSuf) {
		return fallbackCandidate{literal: true, value: st.StripDelimiters(c)}
	}
	return fallbackCandidate{value: strings.ToLower(st.StripDelimiters(c))}
}

// ClearFallbacks removes all registered fallback rules. Unlike
// ClearThemeMap, callers must call this explicitly -- see ClearThemeMap's
// doc comment for why fallback rules aren't cleared automatically.
func (st *Styler) ClearFallbacks() {
	st.mu.Lock()
	st.fallbackMap = make(map[string]fallbackRule)
	st.mu.Unlock()
}

// GetRawTagCode returns the raw style code (fg:bg:flags) for the given tag name from the theme map.
// Returns "" if the tag is not registered and has no resolvable fallback.
func (st *Styler) GetRawTagCode(name string) string {
	st.ensureMaps()
	st.mu.RLock()
	defer st.mu.RUnlock()
	raw, _ := st.lookupRaw(nil, "", strings.ToLower(name))
	return raw
}

// RegisterSemanticTag registers a tag into BOTH the console (base) and theme maps — a
// convenience for defining a style that should resolve identically whether or not a theme
// is active. Prefer RegisterConsoleTag / RegisterThemeTag when you want to target one map.
func (st *Styler) RegisterSemanticTag(name, taggedValue string) {
	st.RegisterConsoleTag(name, taggedValue)
	st.RegisterThemeTag(name, taggedValue)
}

// RegisterSemanticTagRaw is the raw-value form of RegisterSemanticTag (registers to both maps).
func (st *Styler) RegisterSemanticTagRaw(name, rawValue string) {
	st.RegisterConsoleTagRaw(name, rawValue)
	st.RegisterThemeTagRaw(name, rawValue)
}

// GetColorDefinition returns the formatted tag value (with brackets) for a semantic tag.
// It searches the theme map first, then console map.
func (st *Styler) GetColorDefinition(name string) string {
	st.ensureMaps()
	name = strings.TrimPrefix(name, "_")
	name = strings.TrimSuffix(name, "_")
	content := strings.ToLower(name)

	st.mu.RLock()
	raw, ok := st.lookupRaw(nil, "", content)
	st.mu.RUnlock()

	if !ok || raw == "" {
		return ""
	}
	return st.WrapDirect(raw)
}

// UnregisterColor removes a semantic tag from both maps
func (st *Styler) UnregisterColor(name string) {
	st.ensureMaps()
	name = strings.TrimPrefix(name, "_")
	name = strings.TrimSuffix(name, "_")
	content := strings.ToLower(name)

	st.mu.Lock()
	delete(st.consoleMap, content)
	delete(st.themeMap, content)
	st.mu.Unlock()
}

// UnregisterPrefix removes all semantic tags that start with the given prefix from both maps
func (st *Styler) UnregisterPrefix(prefix string) {
	st.ensureMaps()
	searchPrefix := strings.ToLower(strings.TrimSuffix(prefix, "_") + "_")
	st.mu.Lock()
	for key := range st.consoleMap {
		if strings.HasPrefix(key, searchPrefix) {
			delete(st.consoleMap, key)
		}
	}
	for key := range st.themeMap {
		if strings.HasPrefix(key, searchPrefix) {
			delete(st.themeMap, key)
		}
	}
	st.mu.Unlock()
}

// ClearThemeMap removes all entries from the theme map. Fallback rules
// (RegisterFallback, RegisterFallbackList) are left untouched -- unlike theme tag values, a
// fallback rule is typically a structural relationship in the caller's own
// tag-naming scheme (e.g. "a Radio tag falls back to its Checkbox
// equivalent") that holds regardless of which theme is loaded, so callers
// registering such rules should do so once (e.g. at startup) rather than
// re-registering on every theme load. Call ClearFallbacks explicitly if a
// caller does want a clean slate.
func (st *Styler) ClearThemeMap() {
	st.mu.Lock()
	st.themeMap = make(map[string]string)
	st.mu.Unlock()
}

// ResetCustomColors clears all semantic tags and rebuilds from Colors struct
func (st *Styler) ResetCustomColors() {
	st.BuildColorMap()
}

// StripDelimiters removes any of this Styler's known delimiters from a style string to get
// the raw content, falling back to the library-standard delimiters if customised.
func (st *Styler) StripDelimiters(text string) string {
	if strings.HasPrefix(text, st.semPre) && strings.HasSuffix(text, st.semSuf) {
		return text[len(st.semPre) : len(text)-len(st.semSuf)]
	}
	if strings.HasPrefix(text, st.dirPre) && strings.HasSuffix(text, st.dirSuf) {
		return text[len(st.dirPre) : len(text)-len(st.dirSuf)]
	}
	// Fallback to standard delimiters if this Styler uses custom ones
	if st.semPre != "{{|" {
		if strings.HasPrefix(text, "{{|") && strings.HasSuffix(text, "|}}") {
			return text[3 : len(text)-3]
		}
	}
	if st.dirPre != "{{[" {
		if strings.HasPrefix(text, "{{[") && strings.HasSuffix(text, "]}}") {
			return text[3 : len(text)-3]
		}
	}
	return text
}

// --- package-level delegators to Default ---
func StripDelimiters(text string) string { return Default.StripDelimiters(text) }

func BuildColorMap() {
	Default.BuildColorMap()
}

func RegisterConsoleTag(name, taggedValue string) {
	Default.RegisterConsoleTag(name, taggedValue)
}

func RegisterConsoleTagRaw(name, rawValue string) {
	Default.RegisterConsoleTagRaw(name, rawValue)
}

func RegisterThemeTag(name, taggedValue string) {
	Default.RegisterThemeTag(name, taggedValue)
}

func RegisterThemeTagRaw(name, rawValue string) {
	Default.RegisterThemeTagRaw(name, rawValue)
}

func GetRawTagCode(name string) string {
	return Default.GetRawTagCode(name)
}

func RegisterFallback(name string, followChains bool, candidates ...string) {
	Default.RegisterFallback(name, followChains, candidates...)
}

func ClearFallbacks() {
	Default.ClearFallbacks()
}

func AutoConsoleFallback() bool {
	return Default.AutoConsoleFallback()
}

func SetAutoConsoleFallback(enabled bool) {
	Default.SetAutoConsoleFallback(enabled)
}

func RegisterSemanticTag(name, taggedValue string) {
	Default.RegisterSemanticTag(name, taggedValue)
}

func RegisterSemanticTagRaw(name, rawValue string) {
	Default.RegisterSemanticTagRaw(name, rawValue)
}

func GetColorDefinition(name string) string {
	return Default.GetColorDefinition(name)
}

func UnregisterColor(name string) {
	Default.UnregisterColor(name)
}

func UnregisterPrefix(prefix string) {
	Default.UnregisterPrefix(prefix)
}

func ClearThemeMap() {
	Default.ClearThemeMap()
}

func ResetCustomColors() {
	Default.ResetCustomColors()
}

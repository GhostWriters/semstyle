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

// directLookup resolves name against styleMap (nil means themeMap) with
// console-map fallback -- the three-tier lookup GetRawTagCode and
// ExpandTagsWithMap have always done, with no fallback-chain consultation.
// Callers must already hold st.mu (read or write); this method takes no
// lock itself.
func (st *Styler) directLookup(styleMap map[string]string, prefix, name string) (string, bool) {
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
	if raw, ok := st.consoleMap[name]; ok {
		return raw, true
	}
	return "", false
}

// fallbackRule is one name's registered RegisterFallback rule: an ordered
// list of candidates (a single-element list for the common one-fallback
// case) plus whether each candidate follows its own separate rule in turn.
type fallbackRule struct {
	candidates   []string
	followChains bool
}

// lookupRaw resolves name via directLookup, then -- if still unresolved --
// consults name's registered fallback rule (RegisterFallback) if any. A
// single-candidate, chain-following rule is walked in this same loop (so
// cycle detection covers the whole chain, e.g. A -> B -> C); a multi-
// candidate list, or a rule with followChains=false, is resolved
// per-candidate instead (see the loop body for why). Callers must already
// hold st.mu (read or write); this method takes no lock itself.
func (st *Styler) lookupRaw(styleMap map[string]string, prefix, name string) (string, bool) {
	seen := make(map[string]bool, maxFallbackDepth)
	for range maxFallbackDepth {
		if seen[name] {
			return "", false // cycle guard
		}
		seen[name] = true

		if raw, ok := st.directLookup(styleMap, prefix, name); ok {
			return raw, true
		}

		rule, hasRule := st.fallbackMap[name]
		if !hasRule {
			return "", false
		}

		if len(rule.candidates) == 1 && rule.followChains {
			// The common single-fallback-that-chains case: continue this
			// same loop (not a recursive call) so seen/maxFallbackDepth
			// cover the whole chain in one place, exactly as a multi-hop
			// A -> B -> C chain needs.
			name = rule.candidates[0]
			continue
		}
		for _, candidate := range rule.candidates {
			if rule.followChains {
				if raw, ok := st.lookupRaw(styleMap, prefix, candidate); ok {
					return raw, true
				}
				continue
			}
			if raw, ok := st.directLookup(styleMap, prefix, candidate); ok {
				return raw, true
			}
		}
		return "", false
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
// followChains controls whether a candidate that isn't itself directly
// registered also follows its own separate RegisterFallback rule in turn:
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
	lowered := make([]string, len(candidates))
	for i, c := range candidates {
		lowered[i] = strings.ToLower(c)
	}
	st.fallbackMap[strings.ToLower(name)] = fallbackRule{candidates: lowered, followChains: followChains}
	st.mu.Unlock()
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

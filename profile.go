package semstyle

import (
	"os"
	"sync"

	"github.com/charmbracelet/colorprofile"
)

// preferredProfile stores the detected or forced color profile used when
// rendering extended/hex colors. Detected from stderr at init; can be
// overridden via SetPreferredProfile, or temporarily for one render via
// RunWithProfile.
//
// This is a single process-wide value, detected once from the process's
// own stderr/environment at init -- correct for a process that only ever
// serves the terminal it was launched from (a bare CLI invocation, or a
// local TUI session), but wrong for a long-running process serving
// multiple sessions with genuinely different terminal capabilities (e.g. a
// daemon accepting several SSH/web connections): its own launch-time
// stderr says nothing about what any of those connecting clients can
// render. Such a caller should compute each session's own profile (e.g.
// via colorprofile.Env with that session's negotiated TERM/COLORTERM,
// since there's no local file descriptor to check isatty against) and use
// RunWithProfile to apply it just for that session's own renders.
var (
	profileMu        sync.RWMutex
	preferredProfile colorprofile.Profile
)

func init() {
	preferredProfile = colorprofile.Detect(os.Stderr, os.Environ())
}

// GetPreferredProfile returns the detected or forced color profile.
func GetPreferredProfile() colorprofile.Profile {
	profileMu.RLock()
	defer profileMu.RUnlock()
	return preferredProfile
}

// SetPreferredProfile explicitly sets the color profile (useful for
// testing, or a process that only ever serves one known terminal).
func SetPreferredProfile(p colorprofile.Profile) {
	profileMu.Lock()
	defer profileMu.Unlock()
	preferredProfile = p
}

// runWithProfileMu serializes RunWithProfile calls process-wide, same
// reason as RunWithTint's own lock (see its doc comment): a bare
// SetPreferredProfile only guards the variable, not a "set, render,
// restore" sequence.
var runWithProfileMu sync.Mutex

// RunWithProfile makes p the preferred color profile (see
// SetPreferredProfile) for the duration of fn, then restores whatever was
// set before, while holding a package-level lock so concurrent callers
// can't interleave and see each other's profile mid-render. See
// preferredProfile's doc comment for when this is needed over just calling
// SetPreferredProfile once at startup.
func RunWithProfile(p colorprofile.Profile, fn func()) {
	runWithProfileMu.Lock()
	defer runWithProfileMu.Unlock()
	prev := GetPreferredProfile()
	SetPreferredProfile(p)
	defer SetPreferredProfile(prev)
	fn()
}

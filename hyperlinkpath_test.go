package semstyle

import (
	"path/filepath"
	"strings"
	"testing"
)

// withHyperlinkEligible temporarily sets HyperlinkEligibleFunc to a fixed value for the
// duration of fn, restoring the previous func afterward.
func withHyperlinkEligible(t *testing.T, eligible bool, fn func()) {
	t.Helper()
	orig := HyperlinkEligibleFunc
	HyperlinkEligibleFunc = func() bool { return eligible }
	defer func() { HyperlinkEligibleFunc = orig }()
	fn()
}

// nativePath joins segments using the OS-native separator, mirroring how a real path
// would look on whichever OS the test happens to run on.
func nativePath(segments ...string) string {
	return string(filepath.Separator) + filepath.Join(segments...)
}

func TestFormatFilePath(t *testing.T) {
	sep := string(filepath.Separator)
	path := nativePath("home", "clhatch", ".config", "compose", ".env")

	withHyperlinkEligible(t, true, func() {
		got := FormatFilePath(path)
		for _, want := range []string{
			"{{|Folder:::N:file:///home/|}}home{{[-]}}",
			"{{|Folder:::N:file:///home/clhatch/|}}clhatch{{[-]}}",
			"{{|Folder:::N:file:///home/clhatch/.config/|}}.config{{[-]}}",
			"{{|Folder:::N:file:///home/clhatch/.config/compose/|}}compose{{[-]}}",
			"{{|File:::N:file:///home/clhatch/.config/compose/.env|}}.env{{[-]}}",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("FormatFilePath(%q) missing segment %q, got %q", path, want, got)
			}
		}
		// Separators must be styled (matching the segment they precede) but never
		// wrapped in a hyperlink of their own.
		if !strings.Contains(got, "{{|Folder|}}"+sep+"{{[-]}}") {
			t.Errorf("FormatFilePath(%q) should style '%s' separators as plain Folder tags, got %q", path, sep, got)
		}
		if !strings.Contains(got, "{{|File|}}"+sep+"{{[-]}}") {
			t.Errorf("FormatFilePath(%q) should style the separator before the filename as a plain File tag, got %q", path, got)
		}
	})

	withHyperlinkEligible(t, false, func() {
		got := FormatFilePath(path)
		want := "{{|Folder|}}" + sep + "{{[-]}}{{|Folder|}}home{{[-]}}" +
			"{{|Folder|}}" + sep + "{{[-]}}{{|Folder|}}clhatch{{[-]}}" +
			"{{|Folder|}}" + sep + "{{[-]}}{{|Folder|}}.config{{[-]}}" +
			"{{|Folder|}}" + sep + "{{[-]}}{{|Folder|}}compose{{[-]}}" +
			"{{|File|}}" + sep + "{{[-]}}{{|File|}}.env{{[-]}}"
		if got != want {
			t.Errorf("ineligible FormatFilePath(%q) = %q, want %q (no URL param)", path, got, want)
		}
	})
}

func TestFormatFolderPath(t *testing.T) {
	sep := string(filepath.Separator)
	path := nativePath("home", "clhatch", ".config", "appdata")

	withHyperlinkEligible(t, true, func() {
		got := FormatFolderPath(path)
		for _, want := range []string{
			"{{|Folder:::N:file:///home/|}}home{{[-]}}",
			"{{|Folder:::N:file:///home/clhatch/|}}clhatch{{[-]}}",
			"{{|Folder:::N:file:///home/clhatch/.config/|}}.config{{[-]}}",
			"{{|Folder:::N:file:///home/clhatch/.config/appdata/|}}appdata{{[-]}}",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("FormatFolderPath(%q) missing segment %q, got %q", path, want, got)
			}
		}
		if strings.Contains(got, "{{|File") {
			t.Errorf("FormatFolderPath(%q) should never emit a File tag, got %q", path, got)
		}
	})

	withHyperlinkEligible(t, false, func() {
		got := FormatFolderPath(path)
		want := "{{|Folder|}}" + sep + "{{[-]}}{{|Folder|}}home{{[-]}}" +
			"{{|Folder|}}" + sep + "{{[-]}}{{|Folder|}}clhatch{{[-]}}" +
			"{{|Folder|}}" + sep + "{{[-]}}{{|Folder|}}.config{{[-]}}" +
			"{{|Folder|}}" + sep + "{{[-]}}{{|Folder|}}appdata{{[-]}}"
		if got != want {
			t.Errorf("ineligible FormatFolderPath(%q) = %q, want %q (no URL param)", path, got, want)
		}
	})
}

func TestFormatFileName(t *testing.T) {
	path := nativePath("tmp", "ds2.global.abc123.tmp")

	withHyperlinkEligible(t, true, func() {
		got := FormatFileName(".env", path)
		want := "{{|File::::" + FileURL(path) + "|}}.env{{[-]}}"
		if got != want {
			t.Errorf("FormatFileName(...) = %q, want %q", got, want)
		}
	})

	withHyperlinkEligible(t, false, func() {
		got := FormatFileName(".env", path)
		want := "{{|File|}}.env{{[-]}}"
		if got != want {
			t.Errorf("ineligible FormatFileName(...) = %q, want %q", got, want)
		}
	})

	// An empty path means no real location is known -- style only, no link.
	got := FormatFileName(".env", "")
	want := "{{|File|}}.env{{[-]}}"
	if got != want {
		t.Errorf("FormatFileName(name, \"\") = %q, want %q", got, want)
	}
}

func TestFormatFolderName(t *testing.T) {
	path := nativePath("home", "clhatch", "appdata")

	withHyperlinkEligible(t, true, func() {
		got := FormatFolderName("appdata", path)
		want := "{{|Folder::::" + FileURL(path+"/") + "|}}appdata{{[-]}}"
		if got != want {
			t.Errorf("FormatFolderName(...) = %q, want %q", got, want)
		}
	})

	got := FormatFolderName("appdata", "")
	want := "{{|Folder|}}appdata{{[-]}}"
	if got != want {
		t.Errorf("FormatFolderName(name, \"\") = %q, want %q", got, want)
	}
}

func TestFormatUserFolderPath(t *testing.T) {
	sep := string(filepath.Separator)
	baseDir := nativePath("home", "clhatch", ".config", "user")
	fullPath := filepath.Join(baseDir, "themes", "mytheme")

	withHyperlinkEligible(t, true, func() {
		got := FormatUserFolderPath(baseDir, fullPath)
		if !strings.HasPrefix(got, "{{|Folder:::N:"+FileURL(ensureTrailingSlash(baseDir))+"|}}user{{[-]}}{{|Folder|}}:{{[-]}}") {
			t.Errorf("FormatUserFolderPath(...) = %q, want a leading hyperlinked \"user:\" prefix", got)
		}
		if !strings.Contains(got, "{{|Folder:::N:"+FileURL(ensureTrailingSlash(fullPath))+"|}}mytheme{{[-]}}") {
			t.Errorf("FormatUserFolderPath(...) = %q, want the final segment hyperlinked to its own cumulative path", got)
		}
		if !strings.Contains(got, "{{|Folder|}}"+sep+"{{[-]}}") {
			t.Errorf("FormatUserFolderPath(...) = %q, want styled-but-unlinked separators", got)
		}
	})

	withHyperlinkEligible(t, false, func() {
		got := FormatUserFolderPath(baseDir, fullPath)
		want := "{{|Folder|}}user{{[-]}}{{|Folder|}}:{{[-]}}" +
			"{{|Folder|}}themes{{[-]}}{{|Folder|}}" + sep + "{{[-]}}{{|Folder|}}mytheme{{[-]}}"
		if got != want {
			t.Errorf("ineligible FormatUserFolderPath(...) = %q, want %q", got, want)
		}
	})

	// baseDir itself displays as bare "user:".
	withHyperlinkEligible(t, true, func() {
		got := FormatUserFolderPath(baseDir, baseDir)
		want := "{{|Folder:::N:" + FileURL(ensureTrailingSlash(baseDir)) + "|}}user{{[-]}}{{|Folder|}}:{{[-]}}"
		if got != want {
			t.Errorf("FormatUserFolderPath(baseDir, baseDir) = %q, want %q", got, want)
		}
	})

	// fullPath outside baseDir falls back to plain FormatFolderPath.
	withHyperlinkEligible(t, true, func() {
		outside := nativePath("var", "log")
		got := FormatUserFolderPath(baseDir, outside)
		want := FormatFolderPath(outside)
		if got != want {
			t.Errorf("FormatUserFolderPath(outside baseDir) = %q, want fallback %q", got, want)
		}
	})
}

func TestFormatUserFilePath(t *testing.T) {
	baseDir := nativePath("home", "clhatch", ".config", "user")
	fullPath := filepath.Join(baseDir, "themes", "mytheme.ds2theme")

	withHyperlinkEligible(t, true, func() {
		got := FormatUserFilePath(baseDir, fullPath)
		if !strings.Contains(got, "{{|File:::N:"+FileURL(fullPath)+"|}}mytheme.ds2theme{{[-]}}") {
			t.Errorf("FormatUserFilePath(...) = %q, want the final segment tagged File (no trailing slash)", got)
		}
	})
}

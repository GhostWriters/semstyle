package semstyle

import (
	"net/url"
	"path/filepath"
	"strings"
)

// HyperlinkEligibleFunc, when set, is consulted by FormatFilePath, FormatFolderPath,
// FormatFile, FormatFolder, FormatUserFilePath, and FormatUserFolderPath before including
// a file:// URL in the tag markup they build. The host app sets this to encode its own
// eligibility policy (e.g. "is the rendering terminal on the same machine as this
// process?") -- semstyle has no way to know that on its own. Nil means always eligible.
var HyperlinkEligibleFunc func() bool

func hyperlinkEligible() bool {
	return HyperlinkEligibleFunc == nil || HyperlinkEligibleFunc()
}

// FileURL converts a filesystem path to a file:// URL, handling both POSIX and Windows
// path separators/drive letters via the stdlib rather than hand-rolling OS detection.
func FileURL(path string) string {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p}).String()
}

// ensureTrailingSlash appends "/" if not already present. Used for folder targets: a
// trailing slash forces path resolution to require a directory (POSIX open/execve fail
// with ENOTDIR otherwise), which rules out a same-named executable being run instead of
// the folder being opened.
func ensureTrailingSlash(path string) string {
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\") {
		return path
	}
	return path + "/"
}

// FormatFilePath returns raw (unresolved) semstyle tag markup for a file reference --
// unquoted (add quotes yourself if the surrounding message needs them). Suitable for a
// logger.Notice/Info/etc. message string, resolved later by each destination's own
// renderer (console/file/TUI viewport). Each path segment (directory component, plus
// filename) gets its own hyperlink to that segment's own path, so a folder segment opens
// that folder without needing the final file; separators are colored to match but never
// wrapped in a hyperlink. The final segment is tagged {{|File|}}; earlier segments are
// {{|Folder|}} with a trailing "/" on their target URL (see ensureTrailingSlash) so they
// can only resolve to a directory, never execute a same-named file. Tags carry an
// explicit file:// URL unless HyperlinkEligibleFunc says the caller shouldn't include one
// right now (e.g. the rendering terminal is known to be on a different machine).
func FormatFilePath(path string) string {
	return formatPathSegments(path, true)
}

// FormatFolderPath is FormatFilePath's counterpart for referencing a directory rather
// than a single file -- every segment, including the last, is tagged {{|Folder|}}.
func FormatFolderPath(path string) string {
	return formatPathSegments(path, false)
}

// FormatFileName returns raw (unresolved), unquoted semstyle tag markup for a display
// name (e.g. a short label like ".env" rather than the actual on-disk name) that should
// link to path. Pass an empty path if no real path is known: the name is still styled,
// just without a hyperlink, since linking to "" would otherwise point at the process's
// working/root directory. Call FormatFilePath directly when the full path is the thing
// to display.
func FormatFileName(name, path string) string {
	return FormatFile("File", path, name)
}

// FormatFolderName is FormatFileName's {{|Folder|}}-styled counterpart.
func FormatFolderName(name, path string) string {
	return FormatFolder("Folder", path, name)
}

// FormatFile returns raw (unresolved) semstyle tag markup for path, styled with tag (the
// caller's choice -- lets a path be hyperlinked under any semantic style, not just
// "File") and hyperlinked to path itself when HyperlinkEligibleFunc permits it. Displays
// path verbatim as the visible text unless a different label is given via the optional
// name -- most callers have nothing shorter to show than the real path, so this avoids
// having to pass path twice. Never forces a trailing slash, since path is understood to
// reference a single file, not a directory.
func FormatFile(tag, path string, name ...string) string {
	return formatPathTag(tag, pathLabel(path, name), path, false)
}

// FormatFolder is FormatFile's directory counterpart: same tag flexibility and
// optional-name default, but always forces a trailing slash on the hyperlink target (via
// ensureTrailingSlash) regardless of what tag is used, so it can only ever resolve to a
// directory -- this is keyed off the caller's explicit choice of FormatFile vs
// FormatFolder, not by string-matching tag, so it stays correct even when tag isn't
// literally "Folder".
func FormatFolder(tag, path string, name ...string) string {
	return formatPathTag(tag, pathLabel(path, name), path, true)
}

// pathLabel returns name[0] if given and non-empty, else falls back to path itself -- the
// default-argument pattern for FormatFile/FormatFolder's optional display label.
func pathLabel(path string, name []string) string {
	if len(name) > 0 && name[0] != "" {
		return name[0]
	}
	return path
}

func formatPathTag(tag, name, path string, isFolder bool) string {
	if path == "" || !hyperlinkEligible() {
		return "{{|" + tag + "|}}" + name + "{{[-]}}"
	}
	isDefaultLabel := name == path
	if isFolder {
		path = ensureTrailingSlash(path)
	}
	target := FileURL(path)
	// ":::N:" (empty fg/bg, "N" flag) marks the tag as location-only for
	// HyperlinkModeAuto -- only when name is just path displayed verbatim (the
	// FormatFile/FormatFolder default), not a genuinely different caller-supplied label
	// (FormatFileName/FormatFolderName): a distinct label doesn't reveal the destination
	// the way the bare path does, so auto still has something worth adding there. See
	// locationOnlyFlag.
	if isDefaultLabel {
		return "{{|" + tag + ":::" + string(locationOnlyFlag) + ":" + target + "|}}" + name + "{{[-]}}"
	}
	return "{{|" + tag + "::::" + target + "|}}" + name + "{{[-]}}"
}

// FormatUserFolderPath returns raw semstyle markup for fullPath expressed as
// "user:<relative path>" instead of the raw absolute filesystem path -- e.g. a user apps
// folder override at .../user/apps/media.d/plex displays as "user:media.d/plex", and
// baseDir itself displays as bare "user:". baseDir is hyperlinked as just the word "user"
// (the ":" is styled but not part of the link, same as the "/" separators below); each
// remaining path segment gets its own hyperlink to its own real cumulative path, exactly
// like FormatFolderPath does for a plain absolute path -- so the whole "user:..." reads
// as one continuous clickable path, not just a label. Falls back to plain
// FormatFolderPath(fullPath) if fullPath isn't actually under baseDir.
func FormatUserFolderPath(baseDir, fullPath string) string {
	return formatUserPathSegments(baseDir, fullPath, false)
}

// FormatUserFilePath is FormatUserFolderPath's counterpart for referencing a file rather
// than a directory -- the final segment is tagged {{|File|}}, matching FormatFilePath.
func FormatUserFilePath(baseDir, fullPath string) string {
	return formatUserPathSegments(baseDir, fullPath, true)
}

func formatUserPathSegments(baseDir, fullPath string, lastIsFile bool) string {
	rel, err := filepath.Rel(baseDir, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		if lastIsFile {
			return FormatFilePath(fullPath)
		}
		return FormatFolderPath(fullPath)
	}
	eligible := hyperlinkEligible()

	var b strings.Builder
	if !eligible {
		b.WriteString("{{|Folder|}}user{{[-]}}")
	} else {
		b.WriteString("{{|Folder:::" + string(locationOnlyFlag) + ":" + FileURL(ensureTrailingSlash(baseDir)) + "|}}user{{[-]}}")
	}
	// The ":" is styled to match but never wrapped in the hyperlink -- same convention as
	// the "/" separators below, only the segment itself (here, the word "user") is
	// clickable.
	b.WriteString("{{|Folder|}}:{{[-]}}")
	if rel == "." {
		return b.String()
	}

	segments := strings.Split(filepath.ToSlash(rel), "/")
	sep := string(filepath.Separator)
	lastIdx := -1
	for i, s := range segments {
		if s != "" {
			lastIdx = i
		}
	}
	for i, seg := range segments {
		if seg == "" {
			continue
		}
		tag := "Folder"
		if lastIsFile && i == lastIdx {
			tag = "File"
		}
		if i > 0 {
			b.WriteString("{{|" + tag + "|}}" + sep + "{{[-]}}")
		}
		if !eligible {
			b.WriteString("{{|" + tag + "|}}" + seg + "{{[-]}}")
		} else {
			cumulative := filepath.Join(baseDir, filepath.Join(segments[:i+1]...))
			if tag == "Folder" {
				cumulative = ensureTrailingSlash(cumulative)
			}
			b.WriteString("{{|" + tag + ":::" + string(locationOnlyFlag) + ":" + FileURL(cumulative) + "|}}" + seg + "{{[-]}}")
		}
	}
	return b.String()
}

func formatPathSegments(path string, lastIsFile bool) string {
	// A path may use "\" or "/" as its separator regardless of the host OS; normalize to
	// "/" (via the stdlib, rather than hand-rolling OS-separator detection) so
	// segment-splitting is uniform. The separator actually displayed uses the OS-native
	// character (via filepath.Separator) so a Windows path still reads with "\" rather
	// than switching to "/".
	segments := strings.Split(filepath.ToSlash(path), "/")
	eligible := hyperlinkEligible()
	sep := string(filepath.Separator)

	lastIdx := -1
	for i, s := range segments {
		if s != "" {
			lastIdx = i
		}
	}

	var b strings.Builder
	for i, seg := range segments {
		tag := "Folder"
		if lastIsFile && i == lastIdx {
			tag = "File"
		}
		if i > 0 {
			// The separator is styled to match the segment it precedes so the path reads
			// as one continuous colored span, but it's never wrapped in a hyperlink --
			// only whole segments are clickable.
			b.WriteString("{{|" + tag + "|}}" + sep + "{{[-]}}")
		}
		if seg == "" {
			continue
		}
		if !eligible {
			b.WriteString("{{|" + tag + "|}}" + seg + "{{[-]}}")
		} else {
			cumulative := strings.Join(segments[:i+1], "/")
			if tag == "Folder" {
				cumulative = ensureTrailingSlash(cumulative)
			}
			b.WriteString("{{|" + tag + ":::" + string(locationOnlyFlag) + ":" + FileURL(cumulative) + "|}}" + seg + "{{[-]}}")
		}
	}
	return b.String()
}

// FormatLink returns raw (unresolved), unquoted semstyle tag markup for a label wrapped
// in a hyperlink tag pointing to url -- e.g. a version number linking to its GitHub
// release page, or a service name linking to its docs page. Unlike FormatFilePath, this
// has no eligibility check: an http(s) URL is equally reachable from a local terminal, an
// SSH client, or a browser, so there's no "different machine" concern the way there is
// for file:// links. An empty url renders label as plain styled text with no link at all.
func FormatLink(tag, label, url string) string {
	if url == "" {
		return "{{|" + tag + "|}}" + label + "{{[-]}}"
	}
	return "{{|" + tag + "::::" + url + "|}}" + label + "{{[-]}}"
}

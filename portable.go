package termtheme

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// FormatVersion is the current portable .theme format version, written as the
// `format` key and the `# termtheme vN` header. Readers parse any version
// best-effort: a higher version warns but never hard-fails (forward compat).
const FormatVersion = 1

// MarshalOptions parameterizes a portable export.
type MarshalOptions struct {
	App        string // producer app name, recorded in the header (informational)
	AppVersion string // producer version, recorded in the header (informational)
	// Roles is the set of roles the producing app renders, in display order.
	// They are dumped first; any other role the config carries is re-emitted
	// after them (cross-app passthrough), so a foreign role survives a round
	// trip through an app that does not render it. When nil, Roles() is used.
	Roles []Role
}

// Marshal serializes a theme to the portable .theme format: a versioned header,
// a `format` key, an optional `theme = <base>` line, then a full inline dump of
// every role's effective spec — so the receiver never has to reconstruct the
// producer's base palette. base is the producer's builtin palette the config was
// resolved against; roles the config did not override are dumped with base's
// effective code.
//
// The output is byte-droppable straight into an app's live theme.conf: the
// `# termtheme` / `# source` lines are comments, `format` is tolerated-and-warned
// by any binary that predates it, and `theme`/role lines are understood by all.
func Marshal(cfg ThemeConfig, base Theme, opts MarshalOptions) []byte {
	resolved := cfg.Resolve(base)
	roles := opts.Roles
	if roles == nil {
		roles = Roles()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# termtheme v%d\n", FormatVersion)
	if app := strings.TrimSpace(opts.App); app != "" {
		b.WriteString("# source = ")
		b.WriteString(app)
		if v := strings.TrimSpace(opts.AppVersion); v != "" {
			b.WriteByte(' ')
			b.WriteString(v)
		}
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	fmt.Fprintf(&b, "format = %d\n", FormatVersion)
	if bn := writableBaseName(cfg.BaseName); bn != "" {
		b.WriteString("theme = ")
		b.WriteString(bn)
		b.WriteByte('\n')
	}

	seen := make(map[Role]bool, len(roles))
	writeRole := func(role Role) {
		spec := effectiveSpec(cfg, resolved, role)
		if spec == "" {
			return
		}
		b.WriteString(string(role))
		b.WriteString(" = ")
		b.WriteString(spec)
		b.WriteByte('\n')
		seen[role] = true
	}
	for _, role := range roles {
		if seen[role] {
			continue
		}
		writeRole(role)
	}
	// Passthrough: roles the config carries that the producing app does not
	// render, emitted verbatim after the known roles, sorted for stable output.
	var extra []Role
	for role := range cfg.Specs {
		if !seen[role] && !roleIn(roles, role) {
			extra = append(extra, role)
		}
	}
	sort.Slice(extra, func(i, j int) bool { return extra[i] < extra[j] })
	for _, role := range extra {
		writeRole(role)
	}
	return []byte(b.String())
}

// Meta is the header/version information recovered from a portable file.
type Meta struct {
	Format     int    // the file's declared format version (0 if absent)
	App        string // producing app, from `# source = <app> <version>`
	AppVersion string
	// Warnings carries forward-compat and tolerate-and-warn diagnostics
	// (unknown roles, a newer-than-supported format).
	Warnings []string
}

// Unmarshal parses a portable .theme file into a ThemeConfig plus its Meta. It
// understands the portable additions (the `format` key and the header comments)
// and otherwise delegates to ParseThemeConfig, so the role/base/unknown-key
// semantics are identical to a plain theme.conf. A format newer than
// FormatVersion is read best-effort with a warning rather than rejected.
func Unmarshal(data []byte) (ThemeConfig, Meta, error) {
	var meta Meta
	var body strings.Builder
	for _, raw := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if rest, ok := strings.CutPrefix(comment, "source ="); ok {
				fields := strings.Fields(strings.TrimSpace(rest))
				if len(fields) > 0 {
					meta.App = fields[0]
				}
				if len(fields) > 1 {
					meta.AppVersion = fields[1]
				}
			}
			body.WriteString(raw)
			body.WriteByte('\n')
			continue
		}
		if key, value, ok := cutAssignment(trimmed); ok && normalizeKey(key) == "format" {
			if n, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
				meta.Format = n
			}
			continue // pulled out so ParseThemeConfig does not warn on it
		}
		body.WriteString(raw)
		body.WriteByte('\n')
	}

	cfg, err := ParseThemeConfig([]byte(body.String()))
	if err != nil {
		return ThemeConfig{}, meta, err
	}
	meta.Warnings = append(meta.Warnings, cfg.Warnings...)
	if meta.Format > FormatVersion {
		meta.Warnings = append(meta.Warnings,
			fmt.Sprintf("pack format %d is newer than supported %d; reading best-effort", meta.Format, FormatVersion))
	}
	return cfg, meta, nil
}

// ConfigOptions parameterizes serialization to the live theme.conf grammar
// (the import-write path), as opposed to the portable export of Marshal.
type ConfigOptions struct {
	// Header lines are written as leading "# " comments (no leading hash).
	Header []string
	// Roles is the emit order; when nil, Roles() is used. Roles the config
	// carries outside this list are appended (passthrough), sorted.
	Roles []Role
}

// FormatThemeConfig serializes a ThemeConfig to the live theme.conf grammar: an
// optional header comment block, a `theme = <base>` line for a non-default base,
// and one line per role that has an explicit spec (roles left to inherit the
// base are omitted to keep the file lean). This is the writer an app's theme
// editor and import command use to persist the live config.
func FormatThemeConfig(cfg ThemeConfig, opts ConfigOptions) []byte {
	var b strings.Builder
	for _, line := range opts.Header {
		b.WriteString("# ")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if len(opts.Header) > 0 {
		b.WriteByte('\n')
	}
	if bn := writableBaseName(cfg.BaseName); bn != "" {
		b.WriteString("theme = ")
		b.WriteString(bn)
		b.WriteString("\n\n")
	}
	roles := opts.Roles
	if roles == nil {
		roles = Roles()
	}
	seen := make(map[Role]bool, len(roles))
	writeRole := func(role Role) {
		spec := strings.TrimSpace(cfg.Specs[role])
		if spec == "" {
			return
		}
		b.WriteString(string(role))
		b.WriteString(" = ")
		b.WriteString(spec)
		b.WriteByte('\n')
		seen[role] = true
	}
	for _, role := range roles {
		if seen[role] {
			continue
		}
		writeRole(role)
	}
	var extra []Role
	for role := range cfg.Specs {
		if !seen[role] && !roleIn(roles, role) {
			extra = append(extra, role)
		}
	}
	sort.Slice(extra, func(i, j int) bool { return extra[i] < extra[j] })
	for _, role := range extra {
		writeRole(role)
	}
	return []byte(b.String())
}

// effectiveSpec returns the most faithful inline spec to dump for role: the
// human spec if the config customized it (readable), otherwise the role's
// resolved SGR code (so the dump is always complete, never an inheritance gap).
func effectiveSpec(cfg ThemeConfig, resolved Theme, role Role) string {
	if spec := strings.TrimSpace(cfg.Specs[role]); spec != "" {
		return spec
	}
	return strings.TrimSpace(resolved.Codes[role])
}

// writableBaseName returns the base name to serialize, or "" when the base is
// the implicit terminal default (kept out of files to stay lean). The terminal
// aliases are universal naming conventions, not palette data, so recognizing
// them here stays palette-independent.
func writableBaseName(name string) string {
	trimmed := strings.TrimSpace(name)
	switch strings.ToLower(trimmed) {
	case "", "terminal", "default", "auto":
		return ""
	default:
		return trimmed
	}
}

func roleIn(roles []Role, role Role) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

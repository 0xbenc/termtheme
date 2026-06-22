// Package termtheme is the shared theming engine for terminal TUIs. It owns the
// cross-compat-critical, palette-independent layer that sibling apps (passage,
// ssherpa, and future TUIs) had each been carrying a near-identical copy of: the
// semantic role registry, the theme.conf grammar, the style-spec interpreter,
// the SGR + grapheme-cluster render helpers, and a portable, import/export
// friendly .theme file format.
//
// What termtheme deliberately does NOT own is each app's builtin palettes
// (TerminalTheme/VividTheme genuinely diverge per app) or its environment and
// config-path resolution. The palette is always supplied by the caller — see
// ThemeConfig.Resolve — so unifying the engine never silently restyles an app.
package termtheme

// Role is a semantic styling slot. Roles name what a span of text MEANS
// (a danger, a selection, a border) rather than a concrete color, so a theme
// can be retargeted across terminals and apps. The set below is the universal
// superset shared across apps; an individual app renders some subset of it
// (its "app roles") but may carry, preserve, and re-export the rest verbatim.
type Role string

const (
	RoleTitle       Role = "title"
	RolePrimary     Role = "primary"
	RoleSecondary   Role = "secondary"
	RoleAccent      Role = "accent"
	RoleMuted       Role = "muted"
	RoleSubtle      Role = "subtle"
	RoleForeground  Role = "foreground"
	RoleSelected    Role = "selected"
	RoleSelectedBar Role = "selected_bar"
	RoleBorder      Role = "border"
	RoleSuccess     Role = "success"
	RoleWarning     Role = "warning"
	RoleDanger      Role = "danger"
	RoleInfo        Role = "info"
	RoleSearch      Role = "search"
	RolePill        Role = "pill"
)

// Roles returns the universal role superset in canonical display order. Adding a
// role here is the single source of truth: older binaries that predate it hit
// the tolerate-and-warn path in ParseThemeConfig and never hard-fail.
func Roles() []Role {
	return []Role{
		RoleTitle,
		RolePrimary,
		RoleSecondary,
		RoleAccent,
		RoleMuted,
		RoleSubtle,
		RoleForeground,
		RoleSelected,
		RoleSelectedBar,
		RoleBorder,
		RoleSuccess,
		RoleWarning,
		RoleDanger,
		RoleInfo,
		RoleSearch,
		RolePill,
	}
}

// roleAliases maps every accepted config key (normalized: lowercased, '-' -> '_')
// to its role. It is the union of the sibling apps' alias maps, so a theme file
// written by any app resolves the same role on any other.
var roleAliases = map[string]Role{
	"title":         RoleTitle,
	"primary":       RolePrimary,
	"secondary":     RoleSecondary,
	"accent":        RoleAccent,
	"muted":         RoleMuted,
	"subtle":        RoleSubtle,
	"dim":           RoleSubtle,
	"foreground":    RoleForeground,
	"fg":            RoleForeground,
	"text":          RoleForeground,
	"selected":      RoleSelected,
	"selection":     RoleSelected,
	"selected_bar":  RoleSelectedBar,
	"selection_bar": RoleSelectedBar,
	"bar":           RoleSelectedBar,
	"border":        RoleBorder,
	"rule":          RoleBorder,
	"success":       RoleSuccess,
	"warning":       RoleWarning,
	"danger":        RoleDanger,
	"error":         RoleDanger,
	"info":          RoleInfo,
	"search":        RoleSearch,
	"pill":          RolePill,
}

// RoleForKey resolves a config key to its role. The key is normalized first, so
// "Selected-Bar", "selection_bar", and "bar" all resolve to RoleSelectedBar.
func RoleForKey(key string) (Role, bool) {
	role, ok := roleAliases[normalizeKey(key)]
	return role, ok
}

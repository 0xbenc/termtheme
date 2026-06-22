package termtheme

import (
	"fmt"
	"strings"
)

// ThemeConfig is the parsed contents of a theme file: an optional base palette
// name, the per-role overrides as both normalized SGR Codes and the original
// human Specs (preserved so a file can be re-edited and re-serialized without
// loss), and any non-fatal parse Warnings.
//
// Codes/Specs carry EVERY recognized universal role the file mentions, not just
// the ones the host app renders. A role an app does not paint (a foreign role
// from a sibling app) is still parked here so it survives a round-trip — see
// Marshal's passthrough handling.
type ThemeConfig struct {
	BaseName string
	Codes    map[Role]string
	Specs    map[Role]string
	// Warnings collects non-fatal diagnostics, such as role keys this binary
	// does not know about. Unknown keys are tolerated so a theme file written
	// by a newer app never hard-fails an older binary; callers decide where to
	// surface them.
	Warnings []string
}

// ParseThemeConfig parses the theme.conf grammar: '#' line comments, blank
// lines, and "key = value" / "key : value" assignments. The keys "theme"/"base"
// set the base palette name; every other key must resolve to a universal role
// (via RoleForKey) whose value is a style spec. Unknown role keys are collected
// as Warnings and skipped (forward compatibility); a malformed line or an
// invalid spec for a KNOWN role is a hard error.
func ParseThemeConfig(data []byte) (ThemeConfig, error) {
	cfg := ThemeConfig{
		Codes: make(map[Role]string),
		Specs: make(map[Role]string),
	}
	lines := strings.Split(string(data), "\n")
	for index, raw := range lines {
		line := stripComment(raw)
		if line == "" {
			continue
		}
		key, value, ok := cutAssignment(line)
		if !ok {
			return ThemeConfig{}, fmt.Errorf("line %d: expected key=value", index+1)
		}
		key = normalizeKey(key)
		value = strings.TrimSpace(value)
		switch key {
		case "theme", "base":
			if value == "" {
				return ThemeConfig{}, fmt.Errorf("line %d: theme cannot be empty", index+1)
			}
			cfg.BaseName = value
		default:
			role, ok := roleAliases[key]
			if !ok {
				cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("line %d: unknown theme role %q ignored", index+1, key))
				continue
			}
			code, err := ParseStyleSpec(value)
			if err != nil {
				return ThemeConfig{}, fmt.Errorf("line %d: %w", index+1, err)
			}
			cfg.Codes[role] = code
			cfg.Specs[role] = value
		}
	}
	return cfg, nil
}

func stripComment(line string) string {
	if index := strings.IndexByte(line, '#'); index >= 0 {
		line = line[:index]
	}
	return strings.TrimSpace(line)
}

func cutAssignment(line string) (string, string, bool) {
	if key, value, ok := strings.Cut(line, "="); ok {
		return strings.TrimSpace(key), strings.TrimSpace(value), true
	}
	if key, value, ok := strings.Cut(line, ":"); ok {
		return strings.TrimSpace(key), strings.TrimSpace(value), true
	}
	return "", "", false
}

func normalizeKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "_")
	return key
}

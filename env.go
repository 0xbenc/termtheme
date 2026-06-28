package termtheme

import (
	"os"
	"path/filepath"
	"strings"
)

// Environment and theme-config-path helpers. These are pure (no global state)
// and app-parameterized: each app passes its own name, which selects the
// "<APP>_THEME_FILE" / "<APP>_NO_COLOR" variables and the
// "<configDir>/<app>/theme.conf" default. They are shared so every sibling app
// reads the environment and resolves its config path identically; the fail-open
// theme *resolution* (base palette + overrides + normalization) stays in each
// app, which is why those palettes are not shipped here.

// EnvMap turns an environment slice ("KEY=value") into a lookup map. A nil slice
// falls back to the current process environment.
func EnvMap(env []string) map[string]string {
	if env == nil {
		env = os.Environ()
	}
	values := make(map[string]string, len(env))
	for _, item := range env {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			values[key] = value
		}
	}
	return values
}

// ExpandPath expands a leading "~" or "~/" to the user's home directory.
func ExpandPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// EnvTruthy reports whether an environment value reads as "on". Empty and the
// usual falsy spellings are false; anything else is true.
func EnvTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// ResolveThemeFile picks an app's theme-config path. Precedence: an explicit
// file argument, then "<APP>_THEME_FILE", then (unless skipDefault) the default
// "<configDir>/<app>/theme.conf". The bool reports whether the path was
// explicitly requested, so callers can treat a missing explicit file as an
// error while tolerating a missing default. env may be nil.
func ResolveThemeFile(app, file string, env []string, skipDefault bool) (string, bool) {
	if strings.TrimSpace(file) != "" {
		return ExpandPath(file), true
	}
	values := EnvMap(env)
	if value := strings.TrimSpace(values[strings.ToUpper(app)+"_THEME_FILE"]); value != "" {
		return ExpandPath(value), true
	}
	if skipDefault {
		return "", false
	}
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		return "", false
	}
	return filepath.Join(configDir, app, "theme.conf"), false
}

// EnvNoColor reports whether color should be disabled for an app: an explicit
// option, the app's own "<APP>_NO_COLOR" (truthy), or the cross-tool NO_COLOR
// (any non-empty value) all force it. env may be nil.
func EnvNoColor(app string, env []string, optNoColor bool) bool {
	values := EnvMap(env)
	return optNoColor || EnvTruthy(values[strings.ToUpper(app)+"_NO_COLOR"]) || values["NO_COLOR"] != ""
}

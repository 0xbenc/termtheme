package termtheme

// Theme is a concrete, fully-resolved palette: a role->SGR-code map plus a
// NoColor switch. termtheme ships no Theme values of its own — the builtin
// palettes live in each app because they genuinely diverge — so a Theme is
// produced either by an app constructing one directly or by ThemeConfig.Resolve
// overlaying a config onto an app-supplied base.
type Theme struct {
	Name    string
	NoColor bool
	Codes   map[Role]string
}

// Style wraps value in the SGR code for role. A role with no code (absent, or
// explicitly cleared to inherit) renders plain. NoColor renders everything
// plain. Because Style is a direct lookup with no implicit palette fill, the
// caller is responsible for handing Style a Theme whose Codes are already
// complete — which ThemeConfig.Resolve guarantees.
func (t Theme) Style(role Role, value string) string {
	return Apply(t.NoColor, t.Codes[role], value)
}

// WithNoColor returns a copy of the theme with NoColor set.
func (t Theme) WithNoColor(noColor bool) Theme {
	out := t.Clone()
	out.NoColor = noColor
	return out
}

// Clone returns a deep copy so mutating the result never aliases the original's
// Codes map.
func (t Theme) Clone() Theme {
	return Theme{
		Name:    t.Name,
		NoColor: t.NoColor,
		Codes:   copyRoleCodes(t.Codes),
	}
}

// IsZero reports whether the theme carries no information.
func (t Theme) IsZero() bool {
	return t.Name == "" && len(t.Codes) == 0 && !t.NoColor
}

// Resolve overlays the config's role codes onto base and returns a complete
// theme. base is the caller's chosen builtin palette (terminal/vivid/…);
// termtheme intentionally ships no palettes, so the base is always supplied by
// the host app — this is the seam that keeps per-app palette divergence intact.
//
// Roles the config does not set keep base's code (fail-open, never errors).
// Roles the config carries that are absent from base — including foreign roles
// an app does not itself render — are still included, so a resolved theme can
// be re-serialized without dropping them.
func (cfg ThemeConfig) Resolve(base Theme) Theme {
	out := base.Clone()
	if out.Codes == nil {
		out.Codes = make(map[Role]string)
	}
	if len(cfg.Codes) > 0 {
		out.Name = "custom"
		for role, code := range cfg.Codes {
			out.Codes[role] = code
		}
	}
	return out
}

func copyRoleCodes(codes map[Role]string) map[Role]string {
	if codes == nil {
		return nil
	}
	copied := make(map[Role]string, len(codes))
	for role, code := range codes {
		copied[role] = code
	}
	return copied
}

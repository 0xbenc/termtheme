package termtheme

import (
	"strings"
	"testing"
)

func TestThemeStyle(t *testing.T) {
	th := Theme{Codes: map[Role]string{RoleDanger: "31"}}
	if got := th.Style(RoleDanger, "x"); got != "\x1b[31mx\x1b[0m" {
		t.Errorf("Style(danger) = %q", got)
	}
	// A role with no code renders plain (no implicit palette fill).
	if got := th.Style(RolePrimary, "x"); got != "x" {
		t.Errorf("Style(unset) = %q, want plain", got)
	}
	if got := th.WithNoColor(true).Style(RoleDanger, "x"); got != "x" {
		t.Errorf("Style(noColor) = %q, want plain", got)
	}
}

func TestResolveOverlaysOntoBase(t *testing.T) {
	base := Theme{Name: "terminal", Codes: map[Role]string{
		RolePrimary: "36",
		RoleDanger:  "31",
		RoleTitle:   "32",
	}}
	cfg := ThemeConfig{
		Codes: map[Role]string{RolePrimary: "35"}, // override primary -> magenta
		Specs: map[Role]string{RolePrimary: "magenta"},
	}
	got := cfg.Resolve(base)
	if got.Name != "custom" {
		t.Errorf("Name = %q, want custom (overrides present)", got.Name)
	}
	if got.Codes[RolePrimary] != "35" {
		t.Errorf("primary = %q, want overridden 35", got.Codes[RolePrimary])
	}
	if got.Codes[RoleDanger] != "31" || got.Codes[RoleTitle] != "32" {
		t.Errorf("non-overridden roles not inherited from base: %v", got.Codes)
	}
	// Resolve must not mutate the base map.
	if base.Codes[RolePrimary] != "36" {
		t.Errorf("base mutated: primary = %q", base.Codes[RolePrimary])
	}
}

func TestResolveCarriesForeignRoles(t *testing.T) {
	// A base that lacks selected_bar still ends up carrying it when the config
	// has it (the passthrough case: ssherpa-style base + a passage role).
	base := Theme{Name: "terminal", Codes: map[Role]string{RolePrimary: "36"}}
	cfg := ThemeConfig{Codes: map[Role]string{RoleSelectedBar: "100"}}
	got := cfg.Resolve(base)
	if got.Codes[RoleSelectedBar] != "100" {
		t.Errorf("foreign role dropped: %v", got.Codes)
	}
}

// terminalLike and vividLike stand in for the apps' divergent builtin palettes
// in tests; the real palettes stay in each app.
func terminalLike() Theme {
	codes := map[Role]string{}
	for _, r := range Roles() {
		codes[r] = "39"
	}
	codes[RolePrimary] = "36"
	codes[RoleDanger] = "31"
	return Theme{Name: "terminal", Codes: codes}
}

func vividLike() Theme {
	codes := map[Role]string{}
	for _, r := range Roles() {
		codes[r] = "38;2;200;200;200"
	}
	codes[RolePrimary] = "1;38;2;96;221;255"
	return Theme{Name: "vivid", Codes: codes}
}

func TestParseStyleSpecStableAcrossRoundTrip(t *testing.T) {
	// Whatever ParseThemeConfig accepts, FormatThemeConfig + reparse must
	// preserve the spec text.
	cfg, _ := ParseThemeConfig([]byte("primary = bold red\nborder = 90\n"))
	out := FormatThemeConfig(cfg, ConfigOptions{Header: []string{"x"}})
	back, _ := ParseThemeConfig(out)
	if back.Specs[RolePrimary] != "bold red" {
		t.Errorf("primary spec = %q after round-trip", back.Specs[RolePrimary])
	}
	if back.Specs[RoleBorder] != "90" {
		t.Errorf("border spec = %q after round-trip", back.Specs[RoleBorder])
	}
	if strings.Contains(string(out), "theme =") {
		t.Errorf("no base set, but theme= emitted:\n%s", out)
	}
}

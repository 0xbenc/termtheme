package termtheme

import (
	"strings"
	"testing"
)

func TestMarshalDumpsEveryRole(t *testing.T) {
	// A vivid export with no user overrides: every rendered role is dumped
	// inline (full-role dump, no inheritance gaps), plus the base line.
	cfg := ThemeConfig{BaseName: "vivid"}
	out := Marshal(cfg, vividLike(), MarshalOptions{App: "passage", AppVersion: "0.6.0"})
	text := string(out)

	for _, want := range []string{"# termtheme v1", "# source = passage 0.6.0", "format = 1", "theme = vivid"} {
		if !strings.Contains(text, want) {
			t.Errorf("export missing %q:\n%s", want, text)
		}
	}
	for _, role := range Roles() {
		if !strings.Contains(text, string(role)+" = ") {
			t.Errorf("export missing role %q (full-role dump required):\n%s", role, text)
		}
	}
}

func TestMarshalOmitsTerminalBase(t *testing.T) {
	out := string(Marshal(ThemeConfig{BaseName: "terminal"}, terminalLike(), MarshalOptions{App: "ssherpa"}))
	if strings.Contains(out, "theme =") {
		t.Errorf("terminal base should be implicit:\n%s", out)
	}
}

func TestUnmarshalRoundTrip(t *testing.T) {
	cfg := ThemeConfig{
		BaseName: "vivid",
		Codes:    map[Role]string{RolePrimary: "31"},
		Specs:    map[Role]string{RolePrimary: "red"},
	}
	out := Marshal(cfg, vividLike(), MarshalOptions{App: "passage", AppVersion: "1.0.0"})

	back, meta, err := Unmarshal(out)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if meta.Format != 1 {
		t.Errorf("meta.Format = %d, want 1", meta.Format)
	}
	if meta.App != "passage" || meta.AppVersion != "1.0.0" {
		t.Errorf("meta source = %q %q, want passage 1.0.0", meta.App, meta.AppVersion)
	}
	if back.BaseName != "vivid" {
		t.Errorf("BaseName = %q, want vivid", back.BaseName)
	}
	// The user override survives as its readable spec.
	if back.Specs[RolePrimary] != "red" {
		t.Errorf("primary spec = %q, want red", back.Specs[RolePrimary])
	}
	if len(meta.Warnings) != 0 {
		t.Errorf("warnings = %v, want none (format key must not warn)", meta.Warnings)
	}
}

// TestCrossAppPassthrough is the headline interchange contract: a passage theme
// carrying selected_bar, exported, then re-exported by an app that does NOT
// render selected_bar (ssherpa, whose AppRoles exclude it), must preserve the
// role verbatim so passage -> ssherpa -> passage is lossless.
func TestCrossAppPassthrough(t *testing.T) {
	ssherpaRoles := rolesExcept(RoleSelectedBar)

	// passage exports a theme that customizes selected_bar.
	passageCfg := ThemeConfig{
		BaseName: "vivid",
		Codes:    map[Role]string{RoleSelectedBar: "48;2;45;55;72", RolePrimary: "31"},
		Specs:    map[Role]string{RoleSelectedBar: "48;2;45;55;72", RolePrimary: "red"},
	}
	passageExport := Marshal(passageCfg, vividLike(), MarshalOptions{App: "passage", Roles: Roles()})

	// ssherpa imports it: parses fine, parks selected_bar even though it never
	// paints it (no warning — the shared alias map recognizes the role).
	ssherpaCfg, _, err := Unmarshal(passageExport)
	if err != nil {
		t.Fatalf("ssherpa import error: %v", err)
	}
	if ssherpaCfg.Specs[RoleSelectedBar] == "" {
		t.Fatal("ssherpa dropped selected_bar on import")
	}

	// ssherpa re-exports using ITS role set (no selected_bar). Passthrough must
	// still re-emit the foreign role.
	ssherpaExport := Marshal(ssherpaCfg, vividLike(), MarshalOptions{App: "ssherpa", Roles: ssherpaRoles})
	if !strings.Contains(string(ssherpaExport), "selected_bar = ") {
		t.Fatalf("ssherpa re-export dropped selected_bar (passthrough failed):\n%s", ssherpaExport)
	}

	// passage re-imports: the bar color is intact end to end.
	final, _, err := Unmarshal(ssherpaExport)
	if err != nil {
		t.Fatalf("passage re-import error: %v", err)
	}
	if final.Specs[RoleSelectedBar] != "48;2;45;55;72" {
		t.Errorf("round-trip lost selected_bar: %q", final.Specs[RoleSelectedBar])
	}
	if final.Specs[RolePrimary] != "red" {
		t.Errorf("round-trip lost primary: %q", final.Specs[RolePrimary])
	}
}

func TestUnmarshalForwardCompatWarnsNotFails(t *testing.T) {
	// A future format version is read best-effort with a warning, never an error.
	data := []byte("# termtheme v99\nformat = 99\nprimary = red\nsparkle = bold\n")
	cfg, meta, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal of newer format errored: %v", err)
	}
	if cfg.Specs[RolePrimary] != "red" {
		t.Errorf("known role not parsed from newer format")
	}
	var sawFuture, sawUnknown bool
	for _, w := range meta.Warnings {
		if strings.Contains(w, "newer than supported") {
			sawFuture = true
		}
		if strings.Contains(w, "sparkle") {
			sawUnknown = true
		}
	}
	if !sawFuture {
		t.Errorf("missing forward-compat warning: %v", meta.Warnings)
	}
	if !sawUnknown {
		t.Errorf("missing unknown-role warning: %v", meta.Warnings)
	}
}

// TestExportDropsIntoLiveConfig verifies a .theme export parses cleanly through
// the plain ParseThemeConfig path (an old binary or a direct drop into
// theme.conf), with the portable-only `format` key surfacing as a tolerated
// warning rather than an error.
func TestExportDropsIntoLiveConfig(t *testing.T) {
	out := Marshal(ThemeConfig{BaseName: "vivid"}, vividLike(), MarshalOptions{App: "passage"})
	cfg, err := ParseThemeConfig(out)
	if err != nil {
		t.Fatalf("a .theme file must parse as plain theme.conf, got: %v", err)
	}
	if cfg.BaseName != "vivid" {
		t.Errorf("BaseName = %q, want vivid", cfg.BaseName)
	}
	var sawFormat bool
	for _, w := range cfg.Warnings {
		if strings.Contains(w, "format") {
			sawFormat = true
		}
	}
	if !sawFormat {
		t.Errorf("expected `format` to be a tolerated warning on the plain path: %v", cfg.Warnings)
	}
}

func rolesExcept(skip Role) []Role {
	var out []Role
	for _, r := range Roles() {
		if r != skip {
			out = append(out, r)
		}
	}
	return out
}

package termtheme

import (
	"strings"
	"testing"
)

func TestParseStyleSpec(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"none", ""},
		{"plain", ""},
		{"red", "31"},
		{"bold", "1"},
		{"bold red", "1;31"},
		{"bold cyan, underline", "1;36;4"},
		{"bright-white reverse", "97;7"},
		{"dim", "2"},
		{"bg-blue", "44"},
		{"bg-bright-blue", "104"},
		{"fg-magenta", "35"},
		{"1;31", "1;31"},
		{"1;38;2;255;0;0", "1;38;2;255;0;0"},
		{"BOLD  RED", "1;31"}, // case-insensitive + extra spaces
	}
	for _, c := range cases {
		got, err := ParseStyleSpec(c.in)
		if err != nil {
			t.Errorf("ParseStyleSpec(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseStyleSpec(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseStyleSpecRejectsUnknownToken(t *testing.T) {
	if _, err := ParseStyleSpec("imaginary"); err == nil {
		t.Fatal("expected error for unknown token")
	}
}

func TestParseThemeConfig(t *testing.T) {
	data := []byte(`
# a comment
theme = vivid
primary = magenta
selected_bar : 100
danger = 1;31
`)
	cfg, err := ParseThemeConfig(data)
	if err != nil {
		t.Fatalf("ParseThemeConfig error: %v", err)
	}
	if cfg.BaseName != "vivid" {
		t.Errorf("BaseName = %q, want vivid", cfg.BaseName)
	}
	if cfg.Codes[RolePrimary] != "35" {
		t.Errorf("primary code = %q, want 35", cfg.Codes[RolePrimary])
	}
	if cfg.Specs[RolePrimary] != "magenta" {
		t.Errorf("primary spec = %q, want magenta", cfg.Specs[RolePrimary])
	}
	if cfg.Codes[RoleSelectedBar] != "100" {
		t.Errorf("selected_bar code = %q, want 100 (':' assignment)", cfg.Codes[RoleSelectedBar])
	}
	if len(cfg.Warnings) != 0 {
		t.Errorf("warnings = %v, want none", cfg.Warnings)
	}
}

func TestParseThemeConfigToleratesUnknownRoles(t *testing.T) {
	cfg, err := ParseThemeConfig([]byte("primary = red\nhyperlink = blue\n"))
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if cfg.Codes[RolePrimary] != "31" {
		t.Errorf("primary not parsed")
	}
	if len(cfg.Warnings) != 1 || !strings.Contains(cfg.Warnings[0], "hyperlink") {
		t.Errorf("warnings = %v, want one mentioning hyperlink", cfg.Warnings)
	}
}

func TestParseThemeConfigRejectsMalformed(t *testing.T) {
	for _, in := range []string{"primary red\n", "primary = imaginary\n", "[section]\n"} {
		if _, err := ParseThemeConfig([]byte(in)); err == nil {
			t.Errorf("ParseThemeConfig(%q) = nil error, want failure", in)
		}
	}
}

func TestRoleForKeyNormalizes(t *testing.T) {
	cases := map[string]Role{
		"selected-bar":  RoleSelectedBar,
		"Selection_Bar": RoleSelectedBar,
		"bar":           RoleSelectedBar,
		"fg":            RoleForeground,
		"error":         RoleDanger,
		"dim":           RoleSubtle,
	}
	for key, want := range cases {
		got, ok := RoleForKey(key)
		if !ok || got != want {
			t.Errorf("RoleForKey(%q) = %q,%v want %q", key, got, ok, want)
		}
	}
	if _, ok := RoleForKey("nope"); ok {
		t.Error("RoleForKey(nope) = ok, want not ok")
	}
}

package termtheme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvMap(t *testing.T) {
	m := EnvMap([]string{"A=1", "B=two=three", "noeq", "C="})
	if m["A"] != "1" {
		t.Errorf("A = %q, want 1", m["A"])
	}
	if m["B"] != "two=three" {
		t.Errorf("B = %q, want two=three (split on first =)", m["B"])
	}
	if _, ok := m["noeq"]; ok {
		t.Errorf("entry without = should be skipped")
	}
	if v, ok := m["C"]; !ok || v != "" {
		t.Errorf("C = %q ok=%v, want empty present", v, ok)
	}
}

func TestEnvTruthy(t *testing.T) {
	for _, s := range []string{"", "0", "false", "no", "off", " OFF ", "False"} {
		if EnvTruthy(s) {
			t.Errorf("EnvTruthy(%q) = true, want false", s)
		}
	}
	for _, s := range []string{"1", "true", "yes", "on", "anything"} {
		if !EnvTruthy(s) {
			t.Errorf("EnvTruthy(%q) = false, want true", s)
		}
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	if got := ExpandPath("~"); got != home {
		t.Errorf("ExpandPath(~) = %q, want %q", got, home)
	}
	if got := ExpandPath("~/x/y"); got != filepath.Join(home, "x/y") {
		t.Errorf("ExpandPath(~/x/y) = %q, want %q", got, filepath.Join(home, "x/y"))
	}
	if got := ExpandPath("/abs/path"); got != "/abs/path" {
		t.Errorf("ExpandPath(/abs/path) = %q, want unchanged", got)
	}
}

func TestResolveThemeFile(t *testing.T) {
	// explicit file wins and is marked explicit
	if path, explicit := ResolveThemeFile("passage", "/x/theme.conf", nil, false); path != "/x/theme.conf" || !explicit {
		t.Errorf("explicit: got (%q,%v), want (/x/theme.conf,true)", path, explicit)
	}
	// env var, app-scoped + uppercased
	if path, explicit := ResolveThemeFile("passage", "", []string{"PASSAGE_THEME_FILE=/e/t.conf"}, false); path != "/e/t.conf" || !explicit {
		t.Errorf("env: got (%q,%v), want (/e/t.conf,true)", path, explicit)
	}
	// skipDefault returns no path, not explicit
	if path, explicit := ResolveThemeFile("ssherpa", "", []string{}, true); path != "" || explicit {
		t.Errorf("skipDefault: got (%q,%v), want (\"\",false)", path, explicit)
	}
	// default path is "<configDir>/<app>/theme.conf" (literal filename), not explicit
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		want := filepath.Join(dir, "ssherpa", "theme.conf")
		if path, explicit := ResolveThemeFile("ssherpa", "", []string{}, false); path != want || explicit {
			t.Errorf("default: got (%q,%v), want (%q,false)", path, explicit, want)
		}
	}
}

func TestEnvNoColor(t *testing.T) {
	cases := []struct {
		name string
		app  string
		env  []string
		opt  bool
		want bool
	}{
		{"none", "passage", []string{}, false, false},
		{"opt", "passage", []string{}, true, true},
		{"NO_COLOR any", "passage", []string{"NO_COLOR=1"}, false, true},
		{"NO_COLOR empty ignored", "passage", []string{"NO_COLOR="}, false, false},
		{"app truthy", "passage", []string{"PASSAGE_NO_COLOR=true"}, false, true},
		{"app falsy", "passage", []string{"PASSAGE_NO_COLOR=0"}, false, false},
		{"app scoped to name", "ssherpa", []string{"SSHERPA_NO_COLOR=1"}, false, true},
		{"other app's var ignored", "ssherpa", []string{"PASSAGE_NO_COLOR=1"}, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EnvNoColor(tc.app, tc.env, tc.opt); got != tc.want {
				t.Errorf("EnvNoColor(%q,%v,%v) = %v, want %v", tc.app, tc.env, tc.opt, got, tc.want)
			}
		})
	}
}

package termtheme

import "testing"

func TestApply(t *testing.T) {
	if got := Apply(false, "1;31", "hi"); got != "\x1b[1;31mhi\x1b[0m" {
		t.Errorf("Apply = %q", got)
	}
	if got := Apply(true, "1;31", "hi"); got != "hi" {
		t.Errorf("Apply(noColor) = %q, want plain", got)
	}
	if got := Apply(false, "", "hi"); got != "hi" {
		t.Errorf("Apply(no code) = %q, want plain", got)
	}
}

func TestVisibleWidth(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"hello", 5},
		{"\x1b[1;31mhello\x1b[0m", 5}, // SGR ignored
		{"日本", 4},                     // wide runes
		{"🛡️", 2},                     // emoji + variation selector = 2 cells, one cluster
		{"", 0},
	}
	for _, c := range cases {
		if got := VisibleWidth(c.in); got != c.want {
			t.Errorf("VisibleWidth(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestTruncateGraphemeSafe(t *testing.T) {
	// "🛡️x" is 3 cells (emoji cluster = 2 + x = 1). Truncating to 2 with an
	// empty marker must keep the whole emoji cluster, never split the VS16.
	got := TruncateWith("🛡️x", 2, "")
	if VisibleWidth(got) > 2 {
		t.Errorf("TruncateWith width = %d, want <= 2", VisibleWidth(got))
	}
	if got != "🛡️" {
		t.Errorf("TruncateWith = %q, want the whole shield cluster", got)
	}
}

func TestTruncateMarkerAndReset(t *testing.T) {
	// Marker width is reserved from the budget.
	if got := Truncate("abcdef", 4); got != "abc~" {
		t.Errorf("Truncate = %q, want abc~", got)
	}
	// A cut mid-SGR appends a reset so styling never leaks.
	got := Truncate("\x1b[31mabcdef\x1b[0m", 4)
	if got[len(got)-len(reset):] != reset {
		t.Errorf("Truncate styled = %q, want trailing reset", got)
	}
}

func TestPadRight(t *testing.T) {
	if got := PadRight("hi", 5); got != "hi   " {
		t.Errorf("PadRight = %q", got)
	}
	if got := PadRight("toolong", 3); got != "toolong" {
		t.Errorf("PadRight(over) = %q, want unchanged", got)
	}
}

func TestSanitizeDropsC1Control(t *testing.T) {
	// U+009B (the C1 CSI introducer) written as the real code point — UTF-8
	// bytes C2 9B, not a bare \x9b that would decode to U+FFFD — must be
	// dropped, since xterm-class terminals treat it as an escape introducer.
	if got := Sanitize("a\u009bb"); got != "ab" {
		t.Errorf("Sanitize C1 = %q, want %q", got, "ab")
	}
}

func TestSanitizeDropsControls(t *testing.T) {
	// CSI escape stripped; tab kept; BEL (C0) and U+009B (C1) dropped.
	if got := Sanitize("ok\x1b[31m\twork\x07bad"); got != "ok\tworkbad" {
		t.Errorf("Sanitize = %q, want %q", got, "ok\tworkbad")
	}
}

package termtheme

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseStyleSpec converts a human style spec into a normalized SGR parameter
// string (the bytes between CSI and 'm'). It accepts:
//
//   - the empty spec, "none", or "plain"  -> "" (no styling / inherit)
//   - a raw SGR string                    -> normalized verbatim ("1;31")
//   - space/comma/'+'-separated tokens    -> style + color names, e.g.
//     "bold red", "bold cyan, underline", "bright-white reverse"
//
// Color tokens may be prefixed "fg-"/"bg-" to force fore/background. An
// unrecognized token is a hard error so a typo in a config never silently
// renders as plain.
func ParseStyleSpec(value string) (string, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == "none" || value == "plain" {
		return "", nil
	}
	if rawSGR(value) {
		return normalizeSGR(value), nil
	}

	tokens := strings.FieldsFunc(value, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ',' || r == '+'
	})
	var parts []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if rawSGR(token) {
			parts = append(parts, strings.Split(normalizeSGR(token), ";")...)
			continue
		}
		if code, ok := styleTokenCodes[token]; ok {
			parts = append(parts, code)
			continue
		}
		if code, ok := colorTokenCode(token, false); ok {
			parts = append(parts, code)
			continue
		}
		if strings.HasPrefix(token, "fg-") {
			if code, ok := colorTokenCode(strings.TrimPrefix(token, "fg-"), false); ok {
				parts = append(parts, code)
				continue
			}
		}
		if strings.HasPrefix(token, "bg-") {
			if code, ok := colorTokenCode(strings.TrimPrefix(token, "bg-"), true); ok {
				parts = append(parts, code)
				continue
			}
		}
		return "", fmt.Errorf("unknown style token %q", token)
	}
	return strings.Join(parts, ";"), nil
}

// rawSGR reports whether value is already a bare SGR parameter string: one or
// more ';'-separated integers in [0,255].
func rawSGR(value string) bool {
	if value == "" {
		return false
	}
	parts := strings.Split(value, ";")
	for _, part := range parts {
		if part == "" {
			return false
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 || n > 255 {
			return false
		}
	}
	return true
}

func normalizeSGR(value string) string {
	parts := strings.Split(value, ";")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return strings.Join(parts, ";")
}

// colorTokenCode resolves a color name to its SGR code, converting to the
// background range when background is true.
func colorTokenCode(token string, background bool) (string, bool) {
	token = strings.ReplaceAll(token, "_", "-")
	code, ok := colorTokenCodes[token]
	if !ok {
		return "", false
	}
	if !background {
		return code, true
	}
	n, err := strconv.Atoi(code)
	if err != nil {
		return "", false
	}
	switch {
	case n == 39:
		return "49", true
	case n >= 30 && n <= 37:
		return strconv.Itoa(n + 10), true
	case n >= 90 && n <= 97:
		return strconv.Itoa(n + 10), true
	default:
		return "", false
	}
}

var styleTokenCodes = map[string]string{
	"reset":     "0",
	"bold":      "1",
	"faint":     "2",
	"dim":       "2",
	"italic":    "3",
	"underline": "4",
	"reverse":   "7",
	"inverse":   "7",
}

var colorTokenCodes = map[string]string{
	"default":        "39",
	"foreground":     "39",
	"fg":             "39",
	"black":          "30",
	"red":            "31",
	"green":          "32",
	"yellow":         "33",
	"blue":           "34",
	"magenta":        "35",
	"purple":         "35",
	"cyan":           "36",
	"white":          "37",
	"gray":           "90",
	"grey":           "90",
	"bright-black":   "90",
	"bright-red":     "91",
	"bright-green":   "92",
	"bright-yellow":  "93",
	"bright-blue":    "94",
	"bright-magenta": "95",
	"bright-purple":  "95",
	"bright-cyan":    "96",
	"bright-white":   "97",
}

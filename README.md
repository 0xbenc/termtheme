# termtheme

> **Part of [termsystem](https://github.com/0xbenc/termsystem)** — the shared terminal-UI ecosystem (`termtheme` · `termnav` · `termchrome` · `termintro` powering `passage` · `ssherpa` · `dangit`). The ecosystem map, dependency graph, and the agent guide ([AGENTS.md](https://github.com/0xbenc/termsystem/blob/main/AGENTS.md)) live there.

Shared Go theming engine for terminal TUIs: semantic SGR **roles** plus a
portable, import/export-friendly **`.theme`** format.

`termtheme` is the cross-compat core that sibling TUIs (e.g. [passage] and
[ssherpa]) had each been carrying a near-identical copy of. It owns the parts
that *must* agree for themes to interchange — the role registry, the `theme.conf`
grammar, the style-spec interpreter, the SGR + grapheme-cluster render helpers,
the portable `.theme` file, and the pure, app-parameterized **environment /
config-path helpers** (`EnvMap`, `ExpandPath`, `EnvTruthy`, `ResolveThemeFile`,
`EnvNoColor`) — while leaving each app its own **builtin palettes** and its
**fail-open theme resolution** (base selection + overrides + normalization),
which genuinely differ from app to app.

```
go get github.com/0xbenc/termtheme
```

Requires Go 1.26+. The only dependency is
[`github.com/charmbracelet/x/ansi`](https://github.com/charmbracelet/x) for
cell-accurate width.

## Concepts

- **Role** — a semantic styling slot (`title`, `primary`, `danger`, `border`,
  `selected_bar`, …). Roles name what text *means*, not a color, so a theme
  retargets across terminals and apps. `Roles()` is the universal superset.
- **Theme** — a resolved `Role → SGR-code` map plus a `NoColor` switch.
  `theme.Style(role, text)` wraps text in that role's SGR sequence.
- **ThemeConfig** — the parsed contents of a theme file: an optional base palette
  name plus per-role overrides (kept as both normalized `Codes` and human
  `Specs`). `cfg.Resolve(base)` overlays it onto a caller-supplied base palette.
- **Palettes stay in the app.** termtheme ships *no* `TerminalTheme`/`VividTheme`
  — those diverge per app, so the base is always passed in via `Resolve`. This is
  the seam that lets the engine unify without silently restyling anyone.

## Style specs

A role's value is a human spec or a raw SGR string, both via `ParseStyleSpec`:

```
red            green   bright-blue   bold   dim   underline   reverse
bold red       bold cyan, underline  bright-white reverse
bg-blue        fg-magenta            1;38;2;96;221;255   (raw SGR)
"" / none / plain   → inherit (no styling)
```

Unknown role *keys* in a file are tolerated and collected as warnings (forward
compat); an invalid *spec* for a known role is a hard error.

## The portable `.theme` format

The interchange unit is the existing `theme.conf` grammar plus a versioned
header, a `format` key, and a **full-role dump** (every role emitted inline, so
the receiver never needs the producer's base palette):

```ini
# termtheme v1
# source = passage 0.6.0
format = 1
theme = vivid
title = 1;38;2;96;221;255
primary = 1;38;2;96;221;255
...
selected_bar = 48;2;45;55;72
pill = 1;38;2;25;30;38;48;2;96;221;255
```

```go
data := termtheme.Marshal(cfg, app.VividPalette(), termtheme.MarshalOptions{
    App: "passage", AppVersion: "0.6.0", Roles: app.AppRoles,
})
cfg, meta, err := termtheme.Unmarshal(data)
```

Interchange rules:

- **Missing role** → the importer fills from its *own* base palette (`Resolve`),
  so it never errors. (Exporters dump every role to avoid relying on this.)
- **Foreign role** an app doesn't render (e.g. `selected_bar` arriving at an app
  without it) → parsed, parked in `Codes`/`Specs`, **re-emitted verbatim** on the
  next export. `appA → appB → appA` is lossless.
- **Unknown future role** → tolerated with a warning, skipped.
- **Newer `format`** → read best-effort with a warning, never rejected.

A `.theme` file is byte-droppable straight into an app's live `theme.conf`: the
header lines are comments and `format` is a tolerated key.

## Status

Extracted from passage/ssherpa per their
[unified theme-engine feasibility doc](https://github.com/0xbenc/passage/blob/main/docs/unified-theme-engine-feasibility.md).
This module is the data layer + portable format; the interactive theme editor
and per-app palettes/resolution remain in each app.

[passage]: https://github.com/0xbenc/passage
[ssherpa]: https://github.com/0xbenc/ssherpa

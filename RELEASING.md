# Releasing

`termtheme` is the shared theming engine consumed by sibling TUIs:

- [passage](https://github.com/0xbenc/passage)
- [ssherpa](https://github.com/0xbenc/ssherpa)

termtheme is the **leaf** of a three-module shared-UI stack:

```
termtheme (this repo; no bubbletea)
   ├─► termnav     (navigation / list windowing)
   ├─► termchrome  (box/footer/kvrow + glyphs/countdown; depends on termtheme only)
   └─► both, plus termnav + termchrome, are consumed by ─► passage, ssherpa
```

Tag **bottom-up**: `termtheme` → `{termnav, termchrome}` → `{passage, ssherpa}`.
Each consumer pins a tagged version with **no `replace` directive**, so it can only
build against a dependency tag that **already exists on the module proxy** — which
mechanically enforces the order.

**Dev loop:** develop a module change with a local `replace => ../<dir>` in each
consumer, get both consumers green, then drop the replace and pin the new tag (one
`<app>: pin <mod> vX.Y.Z (drop local replace)` commit per app). **Pin lockstep:**
passage and ssherpa must end on identical termtheme/termnav/termchrome versions
(hotfix exception: one app may bump ahead urgently; restore lockstep next release).

## Cross-repo release order (do this when a change touches termtheme)

1. **Tag termtheme first.** Land the change on `main`, then:

   ```sh
   git tag -a vX.Y.Z -m "..."
   git push origin vX.Y.Z
   ```

2. **Bump each consumer** (passage, ssherpa):

   ```sh
   go get github.com/0xbenc/termtheme@vX.Y.Z
   go mod tidy && go test ./...
   git commit -am "Bump termtheme to vX.Y.Z"
   ```

3. **Tag each app release.** That triggers its goreleaser workflow, which
   resolves the termtheme tag from the proxy. (See each app's `RELEASING.md`.)

If a release does **not** change termtheme, skip steps 1–2 — the apps keep their
existing pin and release on their own.

## Versioning

Semantic versioning. A change to the role set, the parsed `.theme` format, or any
exported signature is at least a **minor** bump; breaking either is a **major**
bump. The portable file carries its own `format = N` integer, bumped only when
the on-disk format changes (readers already tolerate a newer `format` with a
warning, so this is rarely needed).

# Releasing

`termtheme` is the shared theming engine consumed by sibling TUIs:

- [passage](https://github.com/0xbenc/passage)
- [ssherpa](https://github.com/0xbenc/ssherpa)

Each app pins a tagged version (`require github.com/0xbenc/termtheme vX.Y.Z`)
with **no `replace` directive**. So a downstream app can only build against a
termtheme version whose tag **already exists on the module proxy** — which fixes
the cross-repo release order.

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

# Changelog

## Unreleased

## v0.6.0 - 2026-09-02

- **The module path is now `github.com/go-gui-org/go-charts`.** This is the
  first tag under the new path. Tag v0.5.8 and earlier declare
  `github.com/mike-ward/go-charts`, so a consumer that wants a tagged build must
  move to v0.6.0. Change the import paths, then run `go mod tidy`.
- **Bump go-gui v0.53.0 -> v0.66.1 and go-glyph -> v1.24.0.** This covers the
  per-scope effective IDs, the `ColorSet` per-state colors, the `Padding` and
  `Sizing` self-flag change, the single-method `View` interface, and the
  window-owned theme. The notes below describe the earlier steps of the same
  migration.
- `make prepush` runs the full local gate: race tests, lint, cross-compile, and
  the export audit.
- The showcase example records its charts at build time. It no longer walks
  `View` trees at render time.
- Structural lines use `DrawContext.Scale`, so a hairline is one physical pixel
  on a HiDPI display.
- **Bump go-gui v0.52.0 → v0.53.0.** go-gui's nine input factories now panic on
  an empty `Cfg.ID`: focus traversal and per-widget state are keyed by it, so a
  control without one renders and clicks but is unreachable by keyboard. No
  library code is affected — the one forced edit is the Export PNG button in
  `examples/basic_line`, which now carries `ID: "export_png"`.
- **BREAKING: event callbacks take a single `gui.EventCtx`.** Bump go-gui
  v0.51.1 → v0.52.0. Every callback that took
  `(*gui.Layout, *gui.Event, *gui.Window)` now takes `func(gui.EventCtx)`, which
  bundles the three as `ctx.Layout`, `ctx.Event` and `ctx.Window`. This reaches
  the public chart config surface: `OnClick`, `OnHover`, `OnMouseLeave` and
  friends on every `*Cfg`.
- **Consume-class callbacks are handled by default.** `OnClick`, `OnMouseUp` and
  `OnGesture` are marked handled by dispatch before the callback runs, so the
  trailing `e.IsHandled = true` is gone. Call `ctx.Bubble()` on a path that
  means "not mine". Hover, move, leave and scroll are unchanged and still call
  `ctx.Consume()`.
- Migration guide upstream: `docs/migration-eventctx.md` in go-gui.

## v0.5.8 - 2026-05-17

- Bump go-gui v0.17.0 → v0.19.1 (scroll phase bridge, context menu focus fix,
  animation heartbeat, Metal autorelease fix)
- Bump go-glyph v1.7.0 → v1.7.1

## v0.5.7 - 2026-04-30

- Bump go-gui v0.12.7 → v0.17.0

## v0.5.6 - 2026-04-26

- Bump go-gui v0.12.0 → v0.12.7

## v0.5.5 - 2026-04-15

- Bump go-gui v0.9.7 → v0.12.0

## v0.5.4 - 2026-04-13

- Bump go-glyph v1.6.4 → v1.6.5
- Bump go-gui v0.9.1 → v0.9.7

## v0.5.3 - 2026-04-12

- Simplify codebase with modern Go 1.26 idioms: `cmp.Or` for defaults,
  `slices.SortFunc`/`cmp.Compare`, builtin `min`/`max`, `slices.Clone`, `wg.Go`;
  extract helpers to deduplicate legend, validation, and ring-buffer logic;
  flatten guards and remove dead code (-285 net lines)

## v0.5.2 - 2026-04-10

- Extract `InteractionCfg` from `BaseCfg`;
  zoom/pan/range-select/animate-transitions fields now live only on XY chart
  configs
- Expand axis/scale test coverage: table-driven tests for `axis.Linear`,
  `axis.Category`, `scale.Linear`, `scale.Log`

## v0.5.1 - 2026-04-08

- Bump go-gui v0.9.0 → v0.9.1
- Bump go-glyph v1.6.3 → v1.6.4
- Bump golang.org/x/sys v0.42.0 → v0.43.0

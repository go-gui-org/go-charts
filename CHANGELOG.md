# Changelog

## v0.9.0 - 2026-09-05

- Bump go-gui v0.68.0 → v0.69.0 (go-glyph stays v1.25.0). Follows the
  `TableCfg.Scrollable` removal: tables always scroll now, so the data-table
  helper drops the flag. Workflow `ref:` pins move to v0.69.0 so CI exercises
  the new version.

## v0.8.0 - 2026-09-05

- Bump go-gui v0.66.1 → v0.68.0 and go-glyph v1.24.0 → v1.25.0. Fixes showcase `ScrollbarCfg.GapEdge` for `Opt[float32]` breaking change.

## v0.7.0 - 2026-09-03

- Hover tooltips are readable. Values round to one decimal, whole numbers to
  none, and `%g` is kept for magnitudes a fixed-point form would ruin, so a
  hovered point reads `24.5` rather than `24.533333333333335`.
  `BaseCfg.TooltipXLabel` and `TooltipYLabel` name the two coordinates, so a
  chart can say "sec" and "Mbps" instead of "X" and "Y". Tooltip text draws a
  step below the tick size, overridable with `Theme.TooltipStyle`, and a tooltip
  taller than the plot area reflows into columns instead of running off the
  panel.
- `GaugeCfg.PointerColor` sets the needle color, so the needle can carry a
  meaning the zone colors do not.
- **Auto-scroll follows a cleared series back to the start.** Scroll state is
  keyed by chart ID in the window state map, so it outlives the data it tracked.
  After a run reaching eighteen seconds, clearing the series and refilling it
  from zero left the window parked on the old right edge and the new curve drew
  off the left of the plot. A decreasing data max cannot come from scrolling, so
  it is now read as a reset: the window snaps to the new max rather than
  tweening a rewind across the whole previous run, and a scroll animation still
  in flight towards the old edge is dropped. This is the fix for a live
  dashboard whose chart did not reset when a second run started.
- The auto-scroll left edge is clamped to the first data point. It was derived
  from the right edge, which lags the data while the tween runs, so whenever the
  window was as wide as the data the whole plot slid sideways on every update.
  The clamp is inert once the data is wider than the window, which is the case
  the scroll is for.

- The gauge arc can sweep its color instead of stepping at each zone edge.
  `GaugeCfg.GradientZones` blends the existing `Zones` colors into one
  continuous ramp, anchoring each zone color at the middle of its own span.
  `GaugeCfg.ArcGradient` takes an explicit `[]gui.GradientStop` ramp for a dial
  with no zones, positioned over `Min..Max`. Both default off, so no existing
  dial changes. The thresholds still drive hit-testing, the tooltip, and the
  legend; only the fill changes.
- `theme.LerpOklab` interpolates two colors in the Oklab perceptual space. Use
  it instead of `theme.Lerp` for a ramp between two hues: sRGB channel averaging
  drops a red-to-green midpoint towards grey, and Oklab does not.
- The gauge value text is placeable. `GaugeCfg.ValueAnchor` selects the base
  point (`GaugeValueDefault`, `GaugeValueCentre`, `GaugeValueAboveCentre`,
  `GaugeValueBelowArc`), `ValueOffsetRatio` shifts it by a signed fraction of
  the hole radius, and `ValueLabel` draws a unit line under the value in
  `TickStyle`. The zero value of each field keeps the previous output, so no
  existing dial moves.

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

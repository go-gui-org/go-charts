# CLAUDE.md

Guidance for Claude Code when working in this repository.

## Commands

```
go test ./...                        # run all tests
go test ./chart/... -run TestFoo     # run single test
go vet ./...                         # static analysis
make lint                            # full lint (pinned version)
go build ./...                       # build all packages
```

## Architecture

Professional charting library built on go-gui. Charts render via
`gui.DrawCanvas` using immediate-mode `OnDraw(*DrawContext)` callbacks. Retained
tessellation cache skips re-render when `Version` unchanged.

```
chart.Line(LineCfg{...}) → gui.View
  → GenerateLayout() wraps gui.DrawCanvas
  → OnDraw(*DrawContext) renders axes, series, legend
  → DrawContext primitives: Line, Polyline, FilledRect, FilledArc, ...
```

### Packages

- `chart/` — chart widget views (Line, Bar, Area, Scatter, Pie); each implements
  `gui.View` via `DrawCanvas`
- `axis/` — Axis interface + Linear, Log, Time, Category axes; tick generation
  (nice-number algorithm)
- `series/` — data series types: XY, Category, OHLC
- `scale/` — Scale interface: data-to-pixel mapping (Linear, Log)
- `render/` — DrawContext adapter with chart-specific helpers
- `theme/` — chart Theme type inheriting `gui.CurrentTheme()`; palettes (Tableau
  10, Pastel, Vivid)

### Key Types

- `chart.*Cfg` — config structs for each chart type (zero-initializable)
- `axis.Axis` — interface: `Label()`, `Ticks()`, `Transform()`, `Inverse()`
- `series.Series` — interface: `Name()`, `Len()`, `Color()`
- `series.XY` — `[]Point` with `Bounds()` method
- `scale.Scale` — interface: `Map()`, `Invert()`, `SetDomain()`, `Domain()`
- `theme.Theme` — colors, text styles, palette, padding
- `render.Context` — wraps `*gui.DrawContext`

### Dependencies

- `github.com/go-gui-org/go-gui` — GUI framework (local replace `../go-gui`)
- No other external dependencies

### Pattern Notes

- All chart types follow go-gui `*Cfg` struct convention
- Charts implement `gui.View` (`Content() []View`,
  `GenerateLayout(*Window) Layout`)
- Event callbacks: `func(gui.EventCtx)` — `ctx.Layout`, `ctx.Event`,
  `ctx.Window`. Consume-class (`OnClick`, `OnMouseUp`, `OnGesture`) is handled
  by dispatch before the callback runs; `ctx.Bubble()` opts out. Everything else
  calls `ctx.Consume()` to stop propagation.
- Default sizing: `gui.FillFill`
- `gui.Hex(0xRRGGBB)` — 3-byte RGB, alpha defaults to 255
- `gui.RGBA(r, g, b, a)` — explicit alpha
- For text rendering, glyph is the underlying library (via go-gui). Consult
  go-glyph (`../go-glyph`) before writing new text-handling routines.
- Notify-class callbacks must call `ctx.Consume()` when the event is consumed to
  prevent further propagation. Consume-class callbacks are already handled on
  entry — do the opposite there and call `ctx.Bubble()` to let the event
  through.

## Coding Conventions

- **No variable shadowing.** Use `=` for existing variables, not `:=`.
- **Clean lint and format.** `golangci-lint run ./...` and `gofmt` must pass
  with zero issues.
- Comments wrap at 90 columns when practical.
- Performance improvements should favor reducing heap allocations.

## Insights

- Add under a ## Code Changes section at the top level of CLAUDE.md\n\nWhen
  fixing a bug or applying a code change, always check ALL files in the codebase
  for the same pattern before committing. Use Grep to find all instances.

- Add under a ## Pre-Commit Checks section in CLAUDE.md\n\nAlways run
  `gofmt -l .` and `golangci-lint run ./...` before committing any Go code
  changes.

- Add under a ## Debugging Guidelines section in CLAUDE.md\n\nWhen diagnosing
  rendering or visual bugs, focus on the data flow and layout/sizing first
  before modifying draw methods. Zero-dimension layouts and stale state across
  frame rebuilds are common root causes.

- Add under a ## Language & Conventions section near the top of
  CLAUDE.md\n\nThis is a Go codebase. Use Go idioms: sentinel errors,
  `min()`/`max()` builtins, proper struct literal nesting. Always use full
  semver strings for tool versions (e.g., `v1.62.0` not `v2`).

- Add under a ## Code Review section in CLAUDE.md\n\nWhen asked to review code
  or suggest improvements, verify findings before reporting them. Check if types
  already have documentation, if patterns are actually used, etc. Avoid false
  positives.

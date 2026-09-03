# Go-Charts Roadmap

Professional charting library for Go, rendering via go-gui's DrawCanvas.

## Phase 1 — Foundation (current)

Scaffold and core infrastructure.

- [x] Package structure: chart, axis, series, scale, render, theme
- [x] Axis interface + Linear, Log, Time, Category implementations
- [x] Scale interface + Linear, Log implementations
- [x] Series types: XY, Category, OHLC
- [x] Render context wrapping gui.DrawContext
- [x] Theme system inheriting gui.CurrentTheme()
- [x] Color palettes: Tableau 10, Pastel, Vivid
- [x] Nice-number tick generation algorithm
- [x] Plot area calculation (padding, axis label space)
- [x] Axis rendering (tick marks, labels, grid lines)
- [x] Axis label text rendering
- [x] Legend rendering

## Phase 2 — Core Charts

First renderable chart types.

- [x] Line chart with polyline rendering
- [x] Line chart markers and filled area under line
- [x] Bar chart (vertical, grouped)
- [x] Bar chart horizontal orientation
- [x] Bar chart stacked mode
- [x] Area chart (filled, stacked)
- [x] Scatter plot with marker shapes (circle, square, triangle, diamond, cross)
- [x] Auto-scaling axes from series data bounds
- [x] Multi-series support with palette cycling

## Phase 3 — Circular Charts

- [x] Pie chart with label placement
- [x] Donut chart (InnerRadius > 0)
- [x] Gauge chart (arc-based, value indicator)
- [x] Gauge arc gradient sweep (`GradientZones`, `ArcGradient`)
- [x] Segment hover highlight
- [x] Percentage labels

## Phase 4 — Interactivity

- [x] Tooltip on hover (value display near cursor)
- [x] Crosshair lines (vertical/horizontal tracking)
- [x] Hover highlight (series/point emphasis)
- [x] Segment hover highlight (pie/donut explode)
- [x] Click selection (OnClick callback wired on all chart types)
- [x] Legend toggle (show/hide series)
- [x] Cursor style changes on interactive elements

## Phase 5 — Advanced Charts

- [x] Candlestick chart (OHLC data, up/down colors)
- [x] Histogram (bin calculation, frequency distribution)
- [x] Box plot (quartiles, whiskers, outliers)
- [x] Waterfall chart (running total, positive/negative)
- [x] Combo chart (mixed line + bar on shared axes)

## Phase 6 — Zoom, Pan, Annotations

- [x] Scroll wheel zoom (X-axis, Y-axis, or both)
- [x] Drag pan
- [x] Zoom reset (double-click or button)
- [x] Range selection (brush/drag to select region)
- [x] Text annotations (positioned labels)
- [x] Line annotations (horizontal/vertical reference lines)
- [x] Region annotations (shaded areas)

## Phase 7 — Statistical & Specialized Charts

- [x] Radar/spider chart
- [x] Bubble chart (scatter with sized markers)
- [x] Heatmap (color-coded grid)
- [x] Treemap (nested rectangles, squarified layout)
- [x] Funnel chart
- [x] Sankey diagram
- [x] Sparklines (inline mini-charts)

## Phase 8 — Data Transforms

- [x] Moving average (simple, exponential, weighted)
- [x] Linear regression trend line
- [x] Polynomial regression
- [x] Bollinger bands
- [x] Min/max envelope
- [x] Downsampling for large datasets (LTTB algorithm)
- [x] Data grouping/binning

## Phase 9 — Animation & Real-time

- [x] Entry animation (series draw-in, bar grow)
- [x] Transition animation on data update (via go-gui animation)
- [x] Real-time data append (streaming, rolling window)
- [x] Smooth scrolling for time-series
- [x] FPS-aware rendering (skip frames under load)

## Phase 10 — Export & Accessibility

- [x] SVG export (static chart to SVG string)
- [x] PNG export (rasterize via backend)
- [ ] Keyboard navigation (tab between points, series)
- [ ] Screen reader labels (ARIA-style metadata)
- [x] High contrast mode
- [x] Data table fallback view

## Design Principles

1. **Immediate-mode** — no retained chart objects; rebuild each frame
2. **Zero-alloc hot paths** — favor stack allocation, pre-sized slices
3. **Cfg structs** — zero-initializable with sensible defaults
4. **DrawCanvas rendering** — leverage retained tessellation cache
5. **Theme inheritance** — charts match app look automatically
6. **Composable** — charts are gui.View; embed in any layout

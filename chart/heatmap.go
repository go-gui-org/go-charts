package chart

import (
	"cmp"
	"fmt"
	"log/slog"
	"math"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// HeatmapCfg configures a heatmap chart (color-coded grid).
type HeatmapCfg struct {
	BaseCfg

	// Data is a dense grid of values with row and column labels.
	Data series.Grid

	// ColorLow is the color for the minimum value.
	// Zero value defaults to blue (0x4575B4).
	ColorLow gui.Color

	// ColorHigh is the color for the maximum value.
	// Zero value defaults to red (0xD73027).
	ColorHigh gui.Color

	// CellGap is the gap in pixels between cells. Zero defaults
	// to DefaultHeatmapCellGap.
	CellGap float32

	// ShowValues renders the numeric value inside each cell.
	ShowValues bool

	// ValueFormat is the fmt format string for cell labels.
	// Zero value defaults to "%.1f".
	ValueFormat string
}

type heatmapView struct {
	cfg                                      HeatmapCfg
	xAxis                                    *axis.Category
	yAxis                                    *axis.Category
	hoverPx                                  float32
	hoverPy                                  float32
	hovering                                 bool
	lastLeft, lastRight, lastTop, lastBottom float32
	win                                      *gui.Window
}

// Heatmap creates a heatmap chart view.
func Heatmap(cfg HeatmapCfg) gui.View {
	cfg.applyDefaults()
	if cfg.ColorLow == (gui.Color{}) {
		cfg.ColorLow = gui.Hex(0x4575B4)
	}
	if cfg.ColorHigh == (gui.Color{}) {
		cfg.ColorHigh = gui.Hex(0xD73027)
	}
	cfg.CellGap = cmp.Or(cfg.CellGap, DefaultHeatmapCellGap)
	cfg.ValueFormat = cmp.Or(cfg.ValueFormat, "%.1f")
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableGrid(&cfg.BaseCfg, cfg.Data)
	}
	return &heatmapView{cfg: cfg}
}

// Draw renders the chart onto dc for headless export.
func (hv *heatmapView) Draw(dc *gui.DrawContext) { hv.draw(dc) }

func (hv *heatmapView) chartTheme() *theme.Theme { return hv.cfg.Theme }

func (hv *heatmapView) Content() []gui.View { return nil }

func (hv *heatmapView) GenerateLayout(w *gui.Window) gui.Layout {
	c := &hv.cfg
	hovV := loadHover(w, c.ID,
		&hv.hovering, &hv.hoverPx, &hv.hoverPy)
	hv.win = w
	animV := loadAnimVersion(w, c.ID)
	transV := loadTransitionVersion(w, c.ID)
	if c.Animate {
		startEntryAnimation(w, c.ID, c.AnimDuration)
	}
	width, height := resolveSize(c.Width, c.Height, w)
	return gui.DrawCanvas(gui.DrawCanvasCfg{
		ID:           c.ID,
		Sizing:       c.Sizing,
		Width:        width,
		Height:       height,
		Version:      c.Version + hovV + animV + transV,
		Clip:         true,
		OnDraw:       hv.draw,
		OnClick:      hv.internalClick,
		OnHover:      hv.internalHover,
		OnMouseLeave: hv.internalMouseLeave,
	}).GenerateLayout(w)
}

func (hv *heatmapView) internalClick(ctx gui.EventCtx) {
	if hv.cfg.OnClick != nil {
		hv.cfg.OnClick(ctx)
	}
}

func (hv *heatmapView) internalHover(ctx gui.EventCtx) {
	ctx.Consume()
	hv.hoverPx = ctx.Event.MouseX - ctx.Layout.Shape.X
	hv.hoverPy = ctx.Event.MouseY - ctx.Layout.Shape.Y
	hv.hovering = true
	saveHover(ctx.Window, ctx.Layout, hv.cfg.ID, true, hv.hoverPx, hv.hoverPy)
	row, col, ok := hv.hitTest(hv.hoverPx, hv.hoverPy)
	if ok && !hv.cfg.Data.IsNaN(row, col) {
		ctx.Window.SetMouseCursorPointingHand()
	} else {
		ctx.Window.SetMouseCursorArrow()
	}
	if hv.cfg.OnHover != nil {
		hv.cfg.OnHover(ctx)
	}
}

func (hv *heatmapView) internalMouseLeave(ctx gui.EventCtx) {
	ctx.Consume()
	hv.hovering = false
	saveHover(ctx.Window, ctx.Layout, hv.cfg.ID, false, 0, 0)
	ctx.Window.SetMouseCursorArrow()
	if hv.cfg.OnMouseLeave != nil {
		hv.cfg.OnMouseLeave(ctx)
	}
}

// hitTest returns the grid cell under (mx, my) in local coords.
func (hv *heatmapView) hitTest(mx, my float32) (row, col int, ok bool) {
	left, right := hv.lastLeft, hv.lastRight
	top, bottom := hv.lastTop, hv.lastBottom
	if mx < left || mx > right || my < top || my > bottom {
		return 0, 0, false
	}
	nr := hv.cfg.Data.NumRows()
	nc := hv.cfg.Data.NumCols()
	if nr == 0 || nc == 0 {
		return 0, 0, false
	}
	cellW := (right - left) / float32(nc)
	cellH := (bottom - top) / float32(nr)
	col = int((mx - left) / cellW)
	row = int((my - top) / cellH)
	col = min(col, nc-1)
	row = min(row, nr-1)
	return row, col, true
}

// heatmapColor maps a value to a color between ColorLow and
// ColorHigh using the data's min/max bounds.
func heatmapColor(v, vMin, vMax float64, low, high gui.Color) gui.Color {
	if vMax <= vMin {
		return theme.Lerp(low, high, 0.5)
	}
	t := (v - vMin) / (vMax - vMin)
	return theme.Lerp(low, high, t)
}

func (hv *heatmapView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &hv.cfg
	th := cfg.Theme
	data := cfg.Data

	nr := data.NumRows()
	nc := data.NumCols()
	if nr == 0 || nc == 0 {
		slog.Warn("no grid data", "chart", cfg.ID)
		return
	}

	left := th.PaddingLeft
	right := ctx.Width() - th.PaddingRight
	top := th.PaddingTop
	bottom := ctx.Height() - th.PaddingBottom

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	// Build category axes.
	hv.xAxis = axis.NewCategory(axis.CategoryCfg{Categories: data.Cols()})
	hv.yAxis = axis.NewCategory(axis.CategoryCfg{Categories: data.Rows()})

	// Reserve space for Y-axis row labels on the left.
	tickLen, tickWidth, tickColor := resolvedTickMark(th)
	tickStyle := th.TickStyle
	fh := ctx.FontHeight(tickStyle)
	maxYTickW := float32(0)
	for _, r := range data.Rows() {
		maxYTickW = max(maxYTickW, ctx.TextWidth(r, tickStyle))
	}
	left += maxYTickW + tickLen + 4

	// Reserve space for X-axis column labels on the bottom.
	maxXTickW := float32(0)
	for _, c := range data.Cols() {
		maxXTickW = max(maxXTickW, ctx.TextWidth(c, tickStyle))
	}
	if cfg.XTickRotation != 0 {
		sinA := float32(math.Abs(math.Sin(float64(cfg.XTickRotation))))
		bottom -= maxXTickW*sinA + tickLen + 4
	} else {
		bottom -= fh + tickLen + 4
	}

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	// Cache layout bounds for hit-testing.
	hv.lastLeft = left
	hv.lastRight = right
	hv.lastTop = top
	hv.lastBottom = bottom

	cellW := (right - left) / float32(nc)
	cellH := (bottom - top) / float32(nr)
	gap := cfg.CellGap

	vMin, vMax := data.Bounds()

	// Hover highlight.
	hovRow, hovCol := -1, -1
	if hv.hovering {
		r, c, ok := hv.hitTest(hv.hoverPx, hv.hoverPy)
		if ok && !data.IsNaN(r, c) {
			hovRow, hovCol = r, c
		}
	}

	progress := animProgress(hv.win, hv.cfg.ID)

	// Draw cells.
	for r := range nr {
		for c := range nc {
			v := data.At(r, c)
			if math.IsNaN(v) {
				continue
			}
			cx := left + float32(c)*cellW + gap/2
			cy := top + float32(r)*cellH + gap/2
			cw := cellW - gap
			ch := cellH - gap
			color := heatmapColor(v, vMin, vMax, cfg.ColorLow, cfg.ColorHigh)
			if hovRow >= 0 && (r != hovRow || c != hovCol) {
				color = dimColor(color, HoverDimAlpha)
			}
			color = gui.RGBA(color.R, color.G, color.B,
				uint8(float32(color.A)*progress))
			ctx.FilledRect(cx, cy, cw, ch, color)

			// Value label.
			if cfg.ShowValues {
				label := fmt.Sprintf(cfg.ValueFormat, v)
				lw := ctx.TextWidth(label, tickStyle)
				// Auto-contrast: white on dark, black on light.
				valStyle := tickStyle
				if theme.Luminance(color) < 0.5 {
					valStyle.Color = gui.Hex(0xFFFFFF)
				} else {
					valStyle.Color = gui.Hex(0x000000)
				}
				ctx.Text(cx+cw/2-lw/2, cy+ch/2-fh/2,
					label, valStyle)
			}
		}
	}

	// Highlight hovered cell border.
	if hovRow >= 0 {
		cx := left + float32(hovCol)*cellW + gap/2
		cy := top + float32(hovRow)*cellH + gap/2
		cw := cellW - gap
		ch := cellH - gap
		ctx.Rect(cx, cy, cw, ch, th.AxisColor, 2)
	}

	// Axis lines.
	ctx.Line(left, bottom, right, bottom, th.AxisColor, th.AxisWidth)
	ctx.Line(left, top, left, bottom, th.AxisColor, th.AxisWidth)

	// X-axis ticks.
	xTicks := hv.xAxis.Ticks(left, right)
	for _, t := range xTicks {
		ctx.Line(t.Position, bottom, t.Position, bottom+tickLen,
			tickColor, tickWidth)
		lw := ctx.TextWidth(t.Label, tickStyle)
		if cfg.XTickRotation != 0 {
			xts := tickStyle
			xts.RotationRadians = cfg.XTickRotation
			ctx.Text(t.Position, bottom+tickLen+2, t.Label, xts)
		} else {
			ctx.Text(t.Position-lw/2, bottom+tickLen+2,
				t.Label, tickStyle)
		}
	}

	// Y-axis ticks.
	yTicks := hv.yAxis.Ticks(top, bottom)
	for _, t := range yTicks {
		ctx.Line(left-tickLen, t.Position, left, t.Position,
			tickColor, tickWidth)
		tw := ctx.TextWidth(t.Label, tickStyle)
		ctx.Text(left-tickLen-tw-2, t.Position-fh/2,
			t.Label, tickStyle)
	}

	// Tooltip.
	if hovRow >= 0 {
		v := data.At(hovRow, hovCol)
		label := fmt.Sprintf("%s\n%s\n"+cfg.ValueFormat,
			data.Rows()[hovRow], data.Cols()[hovCol], v)
		cx := left + float32(hovCol)*cellW + cellW/2
		cy := top + float32(hovRow)*cellH + cellH/2
		drawTooltip(ctx, cx, cy, label, th,
			plotRect{left, right, top, bottom})
	}
}

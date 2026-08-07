package chart

import (
	"cmp"
	"fmt"
	"log/slog"

	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// FunnelCfg configures a funnel chart (stacked trapezoids).
type FunnelCfg struct {
	BaseCfg

	// Slices defines the funnel stages from top (widest) to
	// bottom (narrowest). Reuses PieSlice for convenience.
	Slices []PieSlice

	// SegmentGap is the vertical gap in pixels between stages.
	// Zero defaults to DefaultFunnelSegmentGap.
	SegmentGap float32

	// ShowLabels renders stage label text centered on each
	// segment.
	ShowLabels bool

	// ShowPercent appends a percentage (of total) to labels
	// and tooltips.
	ShowPercent bool

	// MinWidthRatio is the minimum bottom-edge width of the
	// last segment as a fraction of its top edge. Zero
	// defaults to 0.25.
	MinWidthRatio float32

	// ValueFormat is the fmt format string for tooltip values.
	// Zero value defaults to "%.0f".
	ValueFormat string
}

// funnelSegment is a laid-out trapezoid in the funnel.
type funnelSegment struct {
	TopY, BotY                           float32
	TopLeft, TopRight, BotLeft, BotRight float32
	Index                                int
}

type funnelView struct {
	cfg      FunnelCfg
	hoverPx  float32
	hoverPy  float32
	hovering bool
	segments []funnelSegment
	win      *gui.Window
}

// Funnel creates a funnel chart view.
func Funnel(cfg FunnelCfg) gui.View {
	cfg.applyDefaults()
	cfg.SegmentGap = cmp.Or(cfg.SegmentGap, DefaultFunnelSegmentGap)
	cfg.MinWidthRatio = cmp.Or(cfg.MinWidthRatio, 0.25)
	cfg.ValueFormat = cmp.Or(cfg.ValueFormat, "%.0f")
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableSlices(&cfg.BaseCfg, cfg.Slices)
	}
	return &funnelView{cfg: cfg}
}

// Draw renders the chart onto dc for headless export.
func (fv *funnelView) Draw(dc *gui.DrawContext) { fv.draw(dc) }

func (fv *funnelView) chartTheme() *theme.Theme { return fv.cfg.Theme }

func (fv *funnelView) Content() []gui.View { return nil }

func (fv *funnelView) GenerateLayout(w *gui.Window) gui.Layout {
	c := &fv.cfg
	hovV := loadHover(w, c.ID,
		&fv.hovering, &fv.hoverPx, &fv.hoverPy)
	fv.win = w
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
		OnDraw:       fv.draw,
		OnClick:      fv.internalClick,
		OnHover:      fv.internalHover,
		OnMouseLeave: fv.internalMouseLeave,
	}).GenerateLayout(w)
}

func (fv *funnelView) internalClick(ctx gui.EventCtx) {
	ctx.Consume()
	if fv.cfg.OnClick != nil {
		fv.cfg.OnClick(ctx)
	}
}

func (fv *funnelView) internalHover(ctx gui.EventCtx) {
	ctx.Consume()
	fv.hoverPx = ctx.Event.MouseX - ctx.Layout.Shape.X
	fv.hoverPy = ctx.Event.MouseY - ctx.Layout.Shape.Y
	fv.hovering = true
	saveHover(ctx.Window, ctx.Layout, fv.cfg.ID, true, fv.hoverPx, fv.hoverPy)
	_, ok := fv.hitTest(fv.hoverPx, fv.hoverPy)
	if ok {
		ctx.Window.SetMouseCursorPointingHand()
	} else {
		ctx.Window.SetMouseCursorArrow()
	}
	if fv.cfg.OnHover != nil {
		fv.cfg.OnHover(ctx)
	}
}

func (fv *funnelView) internalMouseLeave(ctx gui.EventCtx) {
	ctx.Consume()
	fv.hovering = false
	saveHover(ctx.Window, ctx.Layout, fv.cfg.ID, false, 0, 0)
	ctx.Window.SetMouseCursorArrow()
	if fv.cfg.OnMouseLeave != nil {
		fv.cfg.OnMouseLeave(ctx)
	}
}

// hitTest returns the segment index under (mx, my) using
// linear interpolation of trapezoid edges.
func (fv *funnelView) hitTest(mx, my float32) (int, bool) {
	for i := range fv.segments {
		s := &fv.segments[i]
		if my < s.TopY || my > s.BotY {
			continue
		}
		h := s.BotY - s.TopY
		if h <= 0 {
			continue
		}
		t := (my - s.TopY) / h
		leftX := s.TopLeft + t*(s.BotLeft-s.TopLeft)
		rightX := s.TopRight + t*(s.BotRight-s.TopRight)
		if mx >= leftX && mx <= rightX {
			return i, true
		}
	}
	return 0, false
}

// --- draw ----------------------------------------------------------

func (fv *funnelView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &fv.cfg
	th := cfg.Theme

	if len(cfg.Slices) == 0 {
		slog.Warn("no funnel data", "chart", cfg.ID)
		return
	}

	left := th.PaddingLeft
	right := ctx.Width() - th.PaddingRight
	top := th.PaddingTop
	bottom := ctx.Height() - th.PaddingBottom

	drawTitle(ctx, cfg.Title, th)

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	// Compute value bounds; skip non-finite and negative values.
	maxValue := 0.0
	totalValue := 0.0
	for _, s := range cfg.Slices {
		if !finite(s.Value) || s.Value <= 0 {
			continue
		}
		maxValue = max(maxValue, s.Value)
		totalValue += s.Value
	}
	if maxValue <= 0 {
		slog.Warn("all funnel values zero", "chart", cfg.ID)
		return
	}

	availW := right - left
	centerX := (left + right) / 2
	n := len(cfg.Slices)
	totalGap := cfg.SegmentGap * float32(n-1)
	segH := (bottom - top - totalGap) / float32(n)
	if segH <= 0 {
		slog.Warn("funnel segments too small", "chart", cfg.ID)
		return
	}

	progress := animProgress(fv.win, fv.cfg.ID)

	// Build segment layout.
	fv.segments = fv.segments[:0]
	for i, s := range cfg.Slices {
		topW := float32(0)
		if finite(s.Value) && s.Value > 0 {
			topW = float32(s.Value/maxValue) * availW * progress
		}
		var botW float32
		if i < n-1 {
			nv := cfg.Slices[i+1].Value
			if finite(nv) && nv > 0 {
				botW = float32(nv/maxValue) * availW * progress
			}
		} else {
			botW = topW * cfg.MinWidthRatio
		}

		segTop := top + float32(i)*(segH+cfg.SegmentGap)
		segBot := segTop + segH

		fv.segments = append(fv.segments, funnelSegment{
			TopY:     segTop,
			BotY:     segBot,
			TopLeft:  centerX - topW/2,
			TopRight: centerX + topW/2,
			BotLeft:  centerX - botW/2,
			BotRight: centerX + botW/2,
			Index:    i,
		})
	}

	// Hover.
	hovIdx := -1
	if fv.hovering {
		idx, ok := fv.hitTest(fv.hoverPx, fv.hoverPy)
		if ok {
			hovIdx = idx
		}
	}

	tickStyle := th.TickStyle
	fh := ctx.FontHeight(tickStyle)

	// Draw segments.
	for i := range fv.segments {
		seg := &fv.segments[i]
		color := seriesColor(cfg.Slices[i].Color, i, th.Palette)
		if hovIdx >= 0 && i != hovIdx {
			color = dimColor(color, HoverDimAlpha)
		}

		pts := [8]float32{
			seg.TopLeft, seg.TopY,
			seg.TopRight, seg.TopY,
			seg.BotRight, seg.BotY,
			seg.BotLeft, seg.BotY,
		}
		ctx.FilledPolygon(pts[:], color)

		// Label text.
		if cfg.ShowLabels {
			label := cfg.Slices[i].Label
			if cfg.ShowPercent && totalValue > 0 {
				pct := cfg.Slices[i].Value / totalValue * 100
				label = fmt.Sprintf("%s (%.1f%%)", label, pct)
			}
			tw := ctx.TextWidth(label, tickStyle)
			minW := min(
				seg.TopRight-seg.TopLeft,
				seg.BotRight-seg.BotLeft,
			)
			midY := (seg.TopY + seg.BotY) / 2
			if tw < minW*0.9 && fh < segH*0.9 {
				labelStyle := tickStyle
				if theme.Luminance(color) < 0.5 {
					labelStyle.Color = gui.Hex(0xFFFFFF)
				} else {
					labelStyle.Color = gui.Hex(0x000000)
				}
				ctx.Text(centerX-tw/2, midY-fh/2,
					label, labelStyle)
			}
		}
	}

	// Hover border (4 lines forming trapezoid outline).
	if hovIdx >= 0 {
		seg := &fv.segments[hovIdx]
		bw := float32(2)
		bc := th.AxisColor
		ctx.Line(seg.TopLeft, seg.TopY, seg.TopRight, seg.TopY, bc, bw)
		ctx.Line(seg.TopRight, seg.TopY, seg.BotRight, seg.BotY, bc, bw)
		ctx.Line(seg.BotRight, seg.BotY, seg.BotLeft, seg.BotY, bc, bw)
		ctx.Line(seg.BotLeft, seg.BotY, seg.TopLeft, seg.TopY, bc, bw)
	}

	// Tooltip.
	if hovIdx >= 0 {
		s := cfg.Slices[hovIdx]
		tipLabel := s.Label + "\n" +
			fmt.Sprintf(cfg.ValueFormat, s.Value)
		if totalValue > 0 {
			pct := s.Value / totalValue * 100
			tipLabel += fmt.Sprintf(" (%.1f%%)", pct)
		}
		drawTooltip(ctx, fv.hoverPx, fv.hoverPy,
			tipLabel, th, plotRect{left, right, top, bottom})
	}
}

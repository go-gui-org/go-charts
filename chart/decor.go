package chart

import (
	"cmp"
	"math"
	"strings"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// legendEntry describes one item in the chart legend.
type legendEntry struct {
	Name  string
	Color gui.Color
	Index int // original series index (for toggle mapping)
}

// legendBounds holds the pixel bounds of legend entries for
// click hit-testing.
type legendBounds struct {
	// EntryRects maps original series index to the entry's pixel
	// rect within the canvas.
	EntryRects []legendEntryRect
}

type legendEntryRect struct {
	Index               int // original series index
	X, Y, Width, Height float32
}

// legendStyle holds resolved legend appearance values. Constructed
// by newLegendStyle from theme defaults.
type legendStyle struct {
	Style      gui.TextStyle
	BgColor    gui.Color
	SwatchSize float32
	Padding    float32
	ItemGap    float32
	RowGap     float32
	FH         float32 // font height
	RowH       float32 // max(fh, swatchSize)
}

// newLegendStyle resolves legend style defaults from the theme.
func newLegendStyle(ctx *render.Context, th *theme.Theme) legendStyle {
	ls := th.Legend
	style := ls.TextStyle
	if !style.Color.IsSet() && style.Size == 0 {
		style = th.LabelStyle
	}
	bgColor := ls.Background
	if !bgColor.IsSet() {
		bgColor = gui.RGBA(0, 0, 0, 120)
	}
	swatchSize := cmp.Or(ls.SwatchSize, float32(12))
	padding := cmp.Or(ls.Padding, float32(6))
	itemGap := cmp.Or(ls.ItemGap, float32(4))
	rowGap := cmp.Or(ls.RowGap, float32(2))
	fh := ctx.FontHeight(style)
	return legendStyle{
		Style:      style,
		BgColor:    bgColor,
		SwatchSize: swatchSize,
		Padding:    padding,
		ItemGap:    itemGap,
		RowGap:     rowGap,
		FH:         fh,
		RowH:       max(fh, swatchSize),
	}
}

// drawLegendEntry renders a single legend entry (swatch +
// optional strikethrough + label) at the given position.
func drawLegendEntry(
	ctx *render.Context, e legendEntry, lp legendStyle,
	hidden bool, sx, ey float32,
) {
	color := e.Color
	ts := lp.Style
	if hidden {
		color = dimColor(color, HoverDimAlpha)
		ts.Color = gui.RGBA(
			lp.Style.Color.R, lp.Style.Color.G,
			lp.Style.Color.B, HoverDimAlpha)
	}
	sy := ey + (lp.RowH-lp.SwatchSize)/2
	ctx.FilledRoundedRect(sx, sy,
		lp.SwatchSize, lp.SwatchSize, 2, color)
	if hidden {
		mid := sy + lp.SwatchSize/2
		ctx.Line(sx-1, mid, sx+lp.SwatchSize+1, mid,
			gui.RGBA(200, 200, 200, 180), 1.5)
	}
	tx := sx + lp.SwatchSize + lp.ItemGap
	ty := ey + (lp.RowH-lp.FH)/2
	ctx.Text(tx, ty, e.Name, ts)
}

// resolvedTickMark returns the tick mark visual properties,
// falling back to axis defaults from the theme.
func resolvedTickMark(
	th *theme.Theme,
) (length, width float32, color gui.Color) {
	tms := th.TickMark
	length = cmp.Or(tms.Length, DefaultTickLength)
	color = tms.Color
	if !color.IsSet() {
		color = th.AxisColor
	}
	width = cmp.Or(tms.Width, th.AxisWidth)
	return
}

// drawTitle renders the chart title centered above the plot area.
func drawTitle(
	ctx *render.Context, title string, th *theme.Theme,
) {
	if title == "" {
		return
	}
	style := th.TitleStyle
	tw := ctx.TextWidth(title, style)
	x := (ctx.Width() - tw) / 2
	fh := ctx.FontHeight(style)
	y := (th.PaddingTop - fh) / 2
	ctx.Text(x, y, title, style)
}

// drawXAxisLabel renders the X axis title centered below the
// tick labels. bottom is the Y coordinate of the X axis line.
func drawXAxisLabel(
	ctx *render.Context, label string, th *theme.Theme,
	left, right, bottom float32,
) {
	if label == "" {
		return
	}
	style := th.LabelStyle
	tw := ctx.TextWidth(label, style)
	tickFh := ctx.FontHeight(th.TickStyle)
	// Position below tick marks (5px) + tick labels + gap.
	x := (left + right - tw) / 2
	y := bottom + 5 + tickFh + 6
	ctx.Text(x, y, label, style)
}

// resolveLeft returns left expanded so that Y-axis tick labels and
// the rotated Y-axis title don't overlap. Call after the Y axis
// domain is set. Returns left unchanged when yAxis has no label.
func resolveLeft(
	ctx *render.Context, th *theme.Theme,
	left, bottom, top float32,
	yAxis axis.Axis,
) float32 {
	if yAxis == nil || yAxis.Label() == "" {
		return left
	}
	ticks := yAxis.Ticks(bottom, top)
	maxTickW := float32(0)
	for _, t := range ticks {
		maxTickW = max(maxTickW, ctx.TextWidth(t.Label, th.TickStyle))
	}
	tickLen, _, _ := resolvedTickMark(th)
	labelFH := ctx.FontHeight(th.LabelStyle)
	needed := labelFH*2 + tickLen + maxTickW + 4
	return max(left, needed)
}

// drawYAxisLabel renders the Y axis title rotated 90° CCW,
// centered vertically along the left edge of the plot area.
func drawYAxisLabel(
	ctx *render.Context, label string, th *theme.Theme,
	top, bottom float32,
) {
	if label == "" {
		return
	}
	style := th.LabelStyle
	style.RotationRadians = -math.Pi / 2
	fh := ctx.FontHeight(style)
	tw := ctx.TextWidth(label, style)
	// After -90° rotation, the text's visual width becomes its
	// height and vice versa. Position so the label is centered
	// vertically in the plot area, offset from the left edge.
	x := fh / 2
	y := (top + bottom + tw) / 2
	ctx.Text(x, y, label, style)
}

// maxTickLabelWidth returns the widest tick label in pixels.
func maxTickLabelWidth(
	ctx *render.Context, ticks []axis.Tick, style gui.TextStyle,
) float32 {
	maxW := float32(0)
	for _, t := range ticks {
		maxW = max(maxW, ctx.TextWidth(t.Label, style))
	}
	return maxW
}

// resolveBottom computes the bottom padding needed for X-axis
// decorations: tick marks, tick labels (accounting for rotation),
// and the optional axis title. Returns the pixel distance from
// the canvas bottom to the X-axis line.
func resolveBottom(
	ctx *render.Context, th *theme.Theme,
	maxTickLabelWidth, xTickRotation float32,
	xAxisLabel string,
) float32 {
	tickLen, _, _ := resolvedTickMark(th)
	tickFH := ctx.FontHeight(th.TickStyle)

	var tickLabelH float32
	if xTickRotation != 0 {
		sinA := float32(math.Abs(math.Sin(float64(xTickRotation))))
		cosA := float32(math.Abs(math.Cos(float64(xTickRotation))))
		tickLabelH = sinA*maxTickLabelWidth + cosA*tickFH
	} else {
		tickLabelH = tickFH
	}

	// tick marks + gap + labels + margin
	needed := tickLen + 2 + tickLabelH + 6
	if xAxisLabel != "" {
		labelFH := ctx.FontHeight(th.LabelStyle)
		needed += 6 + labelFH
	}
	return needed
}

// legendRightReserve returns the horizontal space to subtract
// from the right edge of the plot area when LegendRight is
// active. Returns 0 for all other positions.
func legendRightReserve(
	ctx *render.Context, th *theme.Theme,
	posOverride *theme.LegendPosition,
	names []string,
) float32 {
	pos := th.Legend.Position
	if posOverride != nil {
		pos = *posOverride
	}
	if pos != theme.LegendRight {
		return 0
	}

	lp := newLegendStyle(ctx, th)
	maxW := float32(0)
	any := false
	for _, n := range names {
		if n == "" {
			continue
		}
		any = true
		w := ctx.TextWidth(n, lp.Style)
		maxW = max(maxW, w)
	}
	if !any {
		return 0
	}
	// box width + gap between plot edge and legend
	return lp.Padding*2 + lp.SwatchSize + lp.ItemGap + maxW + 8
}

// countHorizontalRows counts the number of rows and whether any
// named entries exist for a horizontal legend layout.
func countHorizontalRows(
	ctx *render.Context, lp legendStyle,
	names []string, availW float32,
) (nRows int, any bool) {
	const interItemGap = float32(12)
	nRows = 1
	rowW := float32(0)
	for _, n := range names {
		if n == "" {
			continue
		}
		any = true
		tw := ctx.TextWidth(n, lp.Style)
		w := lp.SwatchSize + lp.ItemGap + tw
		addition := w
		if rowW > 0 {
			addition += interItemGap
		}
		if rowW > 0 && rowW+addition > availW {
			nRows++
			rowW = w
		} else {
			rowW += addition
		}
	}
	return
}

// legendTopReserve returns the vertical space to add to the top
// edge of the plot area when LegendTop is active. Returns 0 for
// all other positions.
func legendTopReserve(
	ctx *render.Context, th *theme.Theme,
	posOverride *theme.LegendPosition,
	names []string,
	left, right float32,
) float32 {
	pos := th.Legend.Position
	if posOverride != nil {
		pos = *posOverride
	}
	if pos != theme.LegendTop {
		return 0
	}
	lp := newLegendStyle(ctx, th)
	nRows, any := countHorizontalRows(ctx, lp, names, right-left)
	if !any {
		return 0
	}
	return lp.Padding*2 +
		float32(nRows)*lp.RowH +
		float32(max(nRows-1, 0))*lp.RowGap + 8
}

// legendBottomReserve returns the vertical space to add to the
// bottom edge of the plot area when LegendBottom is active.
// Returns 0 for all other positions.
func legendBottomReserve(
	ctx *render.Context, th *theme.Theme,
	posOverride *theme.LegendPosition,
	names []string,
	left, right float32,
) float32 {
	pos := th.Legend.Position
	if posOverride != nil {
		pos = *posOverride
	}
	if pos != theme.LegendBottom {
		return 0
	}
	lp := newLegendStyle(ctx, th)
	nRows, any := countHorizontalRows(ctx, lp, names, right-left)
	if !any {
		return 0
	}
	return lp.Padding*2 +
		float32(nRows)*lp.RowH +
		float32(max(nRows-1, 0))*lp.RowGap
}

// drawLegend renders the legend in the plot area. Position is
// determined by the theme LegendStyle with an optional per-chart
// override. Hidden entries are drawn dimmed with a strikethrough.
// Returns bounds for click hit-testing.
func drawLegend(
	ctx *render.Context,
	entries []legendEntry,
	th *theme.Theme,
	pr plotRect,
	posOverride *theme.LegendPosition,
	hidden map[int]bool,
) legendBounds {
	// Filter to entries with names.
	named := make([]legendEntry, 0, len(entries))
	for _, e := range entries {
		if e.Name != "" {
			named = append(named, e)
		}
	}
	if len(named) == 0 {
		return legendBounds{}
	}

	// Determine position.
	pos := th.Legend.Position
	if posOverride != nil {
		pos = *posOverride
	}
	if pos == theme.LegendNone {
		return legendBounds{}
	}

	lp := newLegendStyle(ctx, th)

	if pos == theme.LegendBottom {
		return drawLegendBottom(ctx, named, hidden, lp, pr.Left, pr.Right)
	}
	if pos == theme.LegendRight {
		return drawLegendRight(ctx, named, hidden, lp, pr.Right, pr.Top)
	}
	if pos == theme.LegendTop {
		return drawLegendTop(ctx, named, hidden, lp, pr.Left, pr.Right, pr.Top)
	}

	// Measure widest entry.
	maxW := float32(0)
	for _, e := range named {
		w := ctx.TextWidth(e.Name, lp.Style)
		maxW = max(maxW, w)
	}

	boxW := lp.Padding*2 + lp.SwatchSize + lp.ItemGap + maxW
	boxH := lp.Padding*2 +
		float32(len(named))*lp.RowH +
		float32(len(named)-1)*lp.RowGap

	var bx, by float32
	switch pos {
	case theme.LegendTopLeft:
		bx = pr.Left + 4
		by = pr.Top + 4
	case theme.LegendBottomRight:
		bx = pr.Right - boxW - 4
		by = pr.Bottom - boxH - 4
	case theme.LegendBottomLeft:
		bx = pr.Left + 4
		by = pr.Bottom - boxH - 4
	default: // LegendTopRight
		bx = pr.Right - boxW - 4
		by = pr.Top + 4
	}

	// Background.
	ctx.FilledRoundedRect(bx, by, boxW, boxH, 4, lp.BgColor)

	// Entries.
	lb := legendBounds{
		EntryRects: make([]legendEntryRect, len(named)),
	}
	for i, e := range named {
		ey := by + lp.Padding +
			float32(i)*(lp.RowH+lp.RowGap)

		lb.EntryRects[i] = legendEntryRect{
			Index:  e.Index,
			X:      bx,
			Y:      ey,
			Width:  boxW,
			Height: lp.RowH,
		}

		drawLegendEntry(ctx, e, lp, hidden[e.Index],
			bx+lp.Padding, ey)
	}

	return lb
}

// drawLegendBottom renders a horizontal legend centered below the
// plot area. Items wrap to additional rows when they exceed the
// available width.
func drawLegendBottom(
	ctx *render.Context,
	entries []legendEntry,
	hidden map[int]bool,
	lp legendStyle,
	left, right float32,
) legendBounds {
	const interItemGap = float32(12)

	// Measure each item's width.
	type itemInfo struct {
		width float32
	}
	items := make([]itemInfo, len(entries))
	for i, e := range entries {
		tw := ctx.TextWidth(e.Name, lp.Style)
		items[i] = itemInfo{width: lp.SwatchSize + lp.ItemGap + tw}
	}

	// Layout rows, wrapping when the row exceeds available width.
	availW := right - left
	type layoutRow struct {
		start, end int // index range [start, end)
		width      float32
	}
	var rows []layoutRow
	rowStart := 0
	rowW := float32(0)
	for i, item := range items {
		addition := item.width
		if i > rowStart {
			addition += interItemGap
		}
		if i > rowStart && rowW+addition > availW {
			rows = append(rows, layoutRow{rowStart, i, rowW})
			rowStart = i
			rowW = item.width
		} else {
			rowW += addition
		}
	}
	if rowStart < len(items) {
		rows = append(rows, layoutRow{rowStart, len(items), rowW})
	}

	// Compute box dimensions.
	maxRowW := float32(0)
	for _, r := range rows {
		maxRowW = max(maxRowW, r.width)
	}
	boxW := maxRowW + lp.Padding*2
	boxH := lp.Padding*2 +
		float32(len(rows))*lp.RowH +
		float32(max(len(rows)-1, 0))*lp.RowGap

	// Position centered horizontally, at the bottom of the canvas.
	bx := (left + right - boxW) / 2
	by := ctx.Height() - boxH

	ctx.FilledRoundedRect(bx, by, boxW, boxH, 4, lp.BgColor)

	// Draw entries row by row.
	lb := legendBounds{
		EntryRects: make([]legendEntryRect, len(entries)),
	}
	for ri, r := range rows {
		ey := by + lp.Padding + float32(ri)*(lp.RowH+lp.RowGap)
		// Center this row within the box.
		x := bx + lp.Padding + (maxRowW-r.width)/2
		for i := r.start; i < r.end; i++ {
			if i > r.start {
				x += interItemGap
			}
			e := entries[i]
			lb.EntryRects[i] = legendEntryRect{
				Index:  e.Index,
				X:      x,
				Y:      ey,
				Width:  items[i].width,
				Height: lp.RowH,
			}

			drawLegendEntry(ctx, e, lp, hidden[e.Index], x, ey)
			x += items[i].width
		}
	}

	return lb
}

// drawLegendTop renders a horizontal legend centered between the
// title and the plot area. Items wrap to additional rows when
// they exceed the available width. top is the adjusted plot-area
// top (after legendTopReserve has been applied).
func drawLegendTop(
	ctx *render.Context,
	entries []legendEntry,
	hidden map[int]bool,
	lp legendStyle,
	left, right, top float32,
) legendBounds {
	const interItemGap = float32(12)

	type itemInfo struct{ width float32 }
	items := make([]itemInfo, len(entries))
	for i, e := range entries {
		tw := ctx.TextWidth(e.Name, lp.Style)
		items[i] = itemInfo{width: lp.SwatchSize + lp.ItemGap + tw}
	}

	availW := right - left
	type layoutRow struct {
		start, end int
		width      float32
	}
	var rows []layoutRow
	rowStart := 0
	rowW := float32(0)
	for i, item := range items {
		addition := item.width
		if i > rowStart {
			addition += interItemGap
		}
		if i > rowStart && rowW+addition > availW {
			rows = append(rows, layoutRow{rowStart, i, rowW})
			rowStart = i
			rowW = item.width
		} else {
			rowW += addition
		}
	}
	if rowStart < len(items) {
		rows = append(rows, layoutRow{rowStart, len(items), rowW})
	}

	maxRowW := float32(0)
	for _, r := range rows {
		maxRowW = max(maxRowW, r.width)
	}
	boxW := maxRowW + lp.Padding*2
	boxH := lp.Padding*2 +
		float32(len(rows))*lp.RowH +
		float32(max(len(rows)-1, 0))*lp.RowGap

	// Position centered horizontally, just above the plot area.
	bx := (left + right - boxW) / 2
	by := top - boxH - 8

	ctx.FilledRoundedRect(bx, by, boxW, boxH, 4, lp.BgColor)

	lb := legendBounds{
		EntryRects: make([]legendEntryRect, len(entries)),
	}
	for ri, r := range rows {
		ey := by + lp.Padding + float32(ri)*(lp.RowH+lp.RowGap)
		x := bx + lp.Padding + (maxRowW-r.width)/2
		for i := r.start; i < r.end; i++ {
			if i > r.start {
				x += interItemGap
			}
			e := entries[i]
			lb.EntryRects[i] = legendEntryRect{
				Index:  e.Index,
				X:      x,
				Y:      ey,
				Width:  items[i].width,
				Height: lp.RowH,
			}

			drawLegendEntry(ctx, e, lp, hidden[e.Index], x, ey)
			x += items[i].width
		}
	}

	return lb
}

// drawLegendRight renders a vertical legend outside the plot area,
// to the right and top-aligned.
func drawLegendRight(
	ctx *render.Context,
	entries []legendEntry,
	hidden map[int]bool,
	lp legendStyle,
	right, top float32,
) legendBounds {
	// Measure widest entry.
	maxW := float32(0)
	for _, e := range entries {
		w := ctx.TextWidth(e.Name, lp.Style)
		maxW = max(maxW, w)
	}

	boxW := lp.Padding*2 + lp.SwatchSize + lp.ItemGap + maxW
	boxH := lp.Padding*2 +
		float32(len(entries))*lp.RowH +
		float32(len(entries)-1)*lp.RowGap

	// Position to the right of the plot area, top-aligned.
	bx := right + 8
	by := top

	// Clamp to canvas width.
	if bx+boxW > ctx.Width() {
		bx = ctx.Width() - boxW
	}

	ctx.FilledRoundedRect(bx, by, boxW, boxH, 4, lp.BgColor)

	lb := legendBounds{
		EntryRects: make([]legendEntryRect, len(entries)),
	}
	for i, e := range entries {
		ey := by + lp.Padding + float32(i)*(lp.RowH+lp.RowGap)

		lb.EntryRects[i] = legendEntryRect{
			Index:  e.Index,
			X:      bx,
			Y:      ey,
			Width:  boxW,
			Height: lp.RowH,
		}

		drawLegendEntry(ctx, e, lp, hidden[e.Index],
			bx+lp.Padding, ey)
	}

	return lb
}

// legendHitTest returns the original series index of the legend
// entry under (mx, my), or -1 if none.
func legendHitTest(lb legendBounds, mx, my float32) int {
	for _, r := range lb.EntryRects {
		if mx >= r.X && mx <= r.X+r.Width &&
			my >= r.Y && my <= r.Y+r.Height {
			return r.Index
		}
	}
	return -1
}

// drawTooltip draws a tooltip directly on the canvas near (tx, ty),
// clamped to the plot-area bounds so it does not overlap elements
// in the padding area (e.g. the title). label may contain
// '\n'-separated lines.
func drawTooltip(
	ctx *render.Context, tx, ty float32, label string,
	th *theme.Theme, pr plotRect,
) {
	lines := strings.Split(label, "\n")
	textStyle := th.TickStyle
	textStyle.Color = gui.Hex(0xEEEEEE)

	fh := ctx.FontHeight(textStyle)
	const padding = float32(6)
	const lineGap = float32(2)

	// Measure box dimensions.
	maxW := float32(0)
	for _, ln := range lines {
		w := ctx.TextWidth(ln, textStyle)
		maxW = max(maxW, w)
	}
	boxW := maxW + padding*2
	boxH := float32(len(lines))*fh +
		float32(len(lines)-1)*lineGap + padding*2

	// Position above-right of (tx, ty), clamped to plot area.
	bx := tx + 8
	by := ty - 8 - boxH
	if bx+boxW > pr.Right {
		bx = pr.Right - boxW
	}
	if bx < pr.Left {
		bx = pr.Left
	}
	if by < pr.Top {
		by = pr.Top
	}
	if by+boxH > pr.Bottom {
		by = pr.Bottom - boxH
	}

	ctx.FilledRoundedRect(bx, by, boxW, boxH, 4,
		gui.RGBA(20, 20, 20, 220))

	for i, ln := range lines {
		lx := bx + padding
		ly := by + padding + float32(i)*(fh+lineGap)
		ctx.Text(lx, ly, ln, textStyle)
	}
}

// drawCrosshair draws dashed vertical and horizontal tracking lines at
// (mx, my). Does nothing when the cursor is outside the plot area.
func drawCrosshair(
	ctx *render.Context, th *theme.Theme,
	mx, my float32, pr plotRect,
) {
	if mx < pr.Left || mx > pr.Right || my < pr.Top || my > pr.Bottom {
		return
	}
	cs := th.Crosshair
	color := cs.Color
	if !color.IsSet() {
		color = gui.RGBA(128, 128, 128, 160)
	}
	width := cmp.Or(cs.Width, float32(1))
	dashLen := cmp.Or(cs.DashLen, float32(6))
	gapLen := cmp.Or(cs.GapLen, float32(4))
	ctx.DashedLine(mx, pr.Top, mx, pr.Bottom, color, width, dashLen, gapLen)
	ctx.DashedLine(pr.Left, my, pr.Right, my, color, width, dashLen, gapLen)
}

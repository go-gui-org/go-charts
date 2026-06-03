package chart

import (
	"cmp"
	"fmt"
	"log/slog"
	"math"
	"slices"

	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// TreemapCfg configures a treemap chart (nested rectangles).
type TreemapCfg struct {
	BaseCfg

	// Data is a slice of root-level tree nodes. Each root
	// becomes a top-level group in the treemap.
	Data []series.TreeNode

	// CellGap is the gap in pixels between cells. Zero
	// defaults to DefaultTreemapCellGap.
	CellGap float32

	// MaxDepth limits nesting depth. 0 means unlimited.
	// Default is 2.
	MaxDepth int

	// ShowHeaders renders a label bar for non-leaf nodes.
	ShowHeaders bool

	// HeaderHeight is the height of header bars in pixels.
	// Zero defaults to DefaultTreemapHeaderHeight.
	HeaderHeight float32

	// ValueFormat is the fmt format string for tooltip values.
	// Zero value defaults to "%.0f".
	ValueFormat string
}

// treemapCell is a laid-out rectangle in the treemap.
type treemapCell struct {
	X, Y, W, H float32
	Node       *series.TreeNode
	Depth      int
	GroupIndex int
	IsHeader   bool
}

type treemapView struct {
	cfg      TreemapCfg
	hoverPx  float32
	hoverPy  float32
	hovering bool
	cells    []treemapCell
	win      *gui.Window
}

// Treemap creates a treemap chart view.
func Treemap(cfg TreemapCfg) gui.View {
	cfg.applyDefaults()
	cfg.CellGap = cmp.Or(cfg.CellGap, DefaultTreemapCellGap)
	cfg.MaxDepth = cmp.Or(cfg.MaxDepth, 2)
	cfg.HeaderHeight = cmp.Or(cfg.HeaderHeight, DefaultTreemapHeaderHeight)
	cfg.ValueFormat = cmp.Or(cfg.ValueFormat, "%.0f")
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableTree(&cfg.BaseCfg, cfg.Data)
	}
	return &treemapView{cfg: cfg}
}

// Draw renders the chart onto dc for headless export.
func (tv *treemapView) Draw(dc *gui.DrawContext) { tv.draw(dc) }

func (tv *treemapView) chartTheme() *theme.Theme { return tv.cfg.Theme }

func (tv *treemapView) Content() []gui.View { return nil }

func (tv *treemapView) GenerateLayout(w *gui.Window) gui.Layout {
	c := &tv.cfg
	hovV := loadHover(w, c.ID,
		&tv.hovering, &tv.hoverPx, &tv.hoverPy)
	tv.win = w
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
		OnDraw:       tv.draw,
		OnClick:      tv.internalClick,
		OnHover:      tv.internalHover,
		OnMouseLeave: tv.internalMouseLeave,
	}).GenerateLayout(w)
}

func (tv *treemapView) internalClick(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	if tv.cfg.OnClick != nil {
		tv.cfg.OnClick(l, e, w)
	}
}

func (tv *treemapView) internalHover(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	tv.hoverPx = e.MouseX - l.Shape.X
	tv.hoverPy = e.MouseY - l.Shape.Y
	tv.hovering = true
	saveHover(w, l, tv.cfg.ID, true, tv.hoverPx, tv.hoverPy)
	idx, ok := tv.hitTest(tv.hoverPx, tv.hoverPy)
	if ok && !tv.cells[idx].IsHeader {
		w.SetMouseCursorPointingHand()
	} else {
		w.SetMouseCursorArrow()
	}
	if tv.cfg.OnHover != nil {
		tv.cfg.OnHover(l, e, w)
	}
}

func (tv *treemapView) internalMouseLeave(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	tv.hovering = false
	saveHover(w, l, tv.cfg.ID, false, 0, 0)
	w.SetMouseCursorArrow()
	if tv.cfg.OnMouseLeave != nil {
		tv.cfg.OnMouseLeave(l, e, w)
	}
}

// hitTest returns the deepest cell index under (mx, my).
func (tv *treemapView) hitTest(mx, my float32) (int, bool) {
	best := -1
	for i := range tv.cells {
		c := &tv.cells[i]
		if mx >= c.X && mx < c.X+c.W && my >= c.Y && my < c.Y+c.H {
			best = i
		}
	}
	if best < 0 {
		return 0, false
	}
	return best, true
}

// treemapColor returns the color for a cell at the given depth
// and top-level group index.
func treemapColor(
	groupIdx, depth int, palette []gui.Color,
	nodeColor gui.Color,
) gui.Color {
	if nodeColor != (gui.Color{}) {
		return nodeColor
	}
	base := seriesColor(gui.Color{}, groupIdx, palette)
	if depth <= 0 {
		return base
	}
	amount := min(float64(depth)*0.2, 0.6)
	return theme.Lighten(base, amount)
}

// --- squarified layout ------------------------------------------

// nodeArea pairs a tree node pointer with its computed pixel area.
type nodeArea struct {
	node     *series.TreeNode
	area     float64
	groupIdx int
}

// worstAspectRatio returns the worst (maximum) aspect ratio among
// items laid out in a row with the given short side length.
func worstAspectRatio(areas []float64, totalArea float64, shortSide float32) float64 {
	if len(areas) == 0 || shortSide <= 0 {
		return math.MaxFloat64
	}
	s2 := float64(shortSide) * float64(shortSide)
	sum := totalArea
	maxA := areas[0]
	minA := areas[0]
	for _, a := range areas[1:] {
		maxA = max(maxA, a)
		minA = min(minA, a)
	}
	if minA <= 0 {
		return math.MaxFloat64
	}
	r1 := s2 * maxA / (sum * sum)
	r2 := sum * sum / (s2 * minA)
	return max(r1, r2)
}

// squarify recursively lays out nodes into rect, appending
// results to tv.cells.
func (tv *treemapView) squarify(
	nodes []nodeArea,
	rx, ry, rw, rh float32,
	depth int,
) {
	if len(nodes) == 0 || rw <= 0 || rh <= 0 {
		return
	}

	// Single node: assign entire rect.
	if len(nodes) == 1 {
		n := &nodes[0]
		tv.layoutNode(n, rx, ry, rw, rh, depth)
		return
	}

	// Sort by area descending.
	slices.SortFunc(nodes, func(a, b nodeArea) int {
		return cmp.Compare(b.area, a.area)
	})

	totalArea := 0.0
	for i := range nodes {
		totalArea += nodes[i].area
	}
	if totalArea <= 0 {
		return
	}

	// Scale areas to fit rect.
	rectArea := float64(rw) * float64(rh)
	scale := rectArea / totalArea
	for i := range nodes {
		nodes[i].area *= scale
	}

	remaining := nodes
	cx, cy, cw, ch := rx, ry, rw, rh

	for len(remaining) > 0 {
		shortSide := min(cw, ch)

		// Build row greedily.
		rowAreas := []float64{remaining[0].area}
		rowTotal := remaining[0].area
		rowEnd := 1

		for rowEnd < len(remaining) {
			candidate := append(rowAreas[:len(rowAreas):len(rowAreas)],
				remaining[rowEnd].area)
			candidateTotal := rowTotal + remaining[rowEnd].area
			oldWorst := worstAspectRatio(rowAreas, rowTotal, shortSide)
			newWorst := worstAspectRatio(candidate, candidateTotal, shortSide)
			if newWorst > oldWorst {
				break
			}
			rowAreas = candidate
			rowTotal = candidateTotal
			rowEnd++
		}

		// Lay out this row.
		rowNodes := remaining[:rowEnd]
		remaining = remaining[rowEnd:]

		if shortSide == cw {
			// Horizontal strip at top.
			stripH := float32(0)
			if cw > 0 {
				stripH = float32(rowTotal / float64(cw))
			}
			stripH = min(stripH, ch)
			ox := cx
			for i := range rowNodes {
				w := float32(0)
				if stripH > 0 {
					w = float32(rowNodes[i].area / float64(stripH))
				}
				w = min(w, cx+cw-ox)
				if i == len(rowNodes)-1 {
					w = cx + cw - ox
				}
				tv.layoutNode(&rowNodes[i], ox, cy, w, stripH, depth)
				ox += w
			}
			cy += stripH
			ch -= stripH
		} else {
			// Vertical strip at left.
			stripW := float32(0)
			if ch > 0 {
				stripW = float32(rowTotal / float64(ch))
			}
			stripW = min(stripW, cw)
			oy := cy
			for i := range rowNodes {
				h := float32(0)
				if stripW > 0 {
					h = float32(rowNodes[i].area / float64(stripW))
				}
				h = min(h, cy+ch-oy)
				if i == len(rowNodes)-1 {
					h = cy + ch - oy
				}
				tv.layoutNode(&rowNodes[i], cx, oy, stripW, h, depth)
				oy += h
			}
			cx += stripW
			cw -= stripW
		}
	}
}

// layoutNode adds a cell for the node and recurses into children.
func (tv *treemapView) layoutNode(
	na *nodeArea,
	x, y, w, h float32,
	depth int,
) {
	cfg := &tv.cfg
	node := na.node

	if node.IsLeaf() || (cfg.MaxDepth > 0 && depth >= cfg.MaxDepth) {
		tv.cells = append(tv.cells, treemapCell{
			X: x, Y: y, W: w, H: h,
			Node:       node,
			Depth:      depth,
			GroupIndex: na.groupIdx,
		})
		return
	}

	// Non-leaf: optional header + recurse children.
	innerX, innerY, innerW, innerH := x, y, w, h

	if cfg.ShowHeaders {
		hh := cfg.HeaderHeight
		if hh > h/2 {
			hh = h / 2
		}
		tv.cells = append(tv.cells, treemapCell{
			X: x, Y: y, W: w, H: hh,
			Node:       node,
			Depth:      depth,
			GroupIndex: na.groupIdx,
			IsHeader:   true,
		})
		innerY += hh
		innerH -= hh
	}

	if innerW <= 0 || innerH <= 0 {
		return
	}

	// Build child nodeArea slice.
	children := make([]nodeArea, 0, len(node.Children))
	for i := range node.Children {
		v := node.Children[i].TotalValue()
		if !finite(v) || v <= 0 {
			continue
		}
		children = append(children, nodeArea{
			node:     &node.Children[i],
			area:     v,
			groupIdx: na.groupIdx,
		})
	}
	tv.squarify(children, innerX, innerY, innerW, innerH, depth+1)
}

// --- draw -------------------------------------------------------

func (tv *treemapView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &tv.cfg
	th := cfg.Theme

	if len(cfg.Data) == 0 {
		slog.Warn("no tree data", "chart", cfg.ID)
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

	// Build root nodeArea slice.
	roots := make([]nodeArea, 0, len(cfg.Data))
	for i := range cfg.Data {
		v := cfg.Data[i].TotalValue()
		if !finite(v) || v <= 0 {
			continue
		}
		roots = append(roots, nodeArea{
			node:     &cfg.Data[i],
			area:     v,
			groupIdx: i,
		})
	}

	if len(roots) == 0 {
		slog.Warn("all tree values zero", "chart", cfg.ID)
		return
	}

	// Layout.
	tv.cells = tv.cells[:0]
	tv.squarify(roots, left, top, right-left, bottom-top, 0)

	// Hover.
	hovIdx := -1
	if tv.hovering {
		idx, ok := tv.hitTest(tv.hoverPx, tv.hoverPy)
		if ok {
			hovIdx = idx
		}
	}

	progress := animProgress(tv.win, tv.cfg.ID)

	gap := cfg.CellGap
	tickStyle := th.TickStyle
	fh := ctx.FontHeight(tickStyle)

	// Draw cells.
	for i := range tv.cells {
		c := &tv.cells[i]
		color := treemapColor(
			c.GroupIndex, c.Depth, th.Palette, c.Node.NodeColor)
		if c.IsHeader {
			color = theme.Darken(color, 0.15)
		}
		if hovIdx >= 0 && i != hovIdx {
			color = dimColor(color, HoverDimAlpha)
		}
		color = gui.RGBA(color.R, color.G, color.B,
			uint8(float32(color.A)*progress))

		cx := c.X + gap/2
		cy := c.Y + gap/2
		cw := c.W - gap
		ch := c.H - gap
		if cw <= 0 || ch <= 0 {
			continue
		}
		ctx.FilledRect(cx, cy, cw, ch, color)

		// Label (name only; values appear in tooltips).
		label := c.Node.Label

		labelStyle := tickStyle
		if theme.Luminance(color) < 0.5 {
			labelStyle.Color = gui.Hex(0xFFFFFF)
		} else {
			labelStyle.Color = gui.Hex(0x000000)
		}

		tw := ctx.TextWidth(label, labelStyle)
		if tw < cw-4 && fh < ch-2 {
			if c.IsHeader {
				ctx.Text(cx+3, cy+(ch-fh)/2, label, labelStyle)
			} else {
				ctx.Text(cx+(cw-tw)/2, cy+(ch-fh)/2,
					label, labelStyle)
			}
		}
	}

	// Hover border.
	if hovIdx >= 0 {
		c := &tv.cells[hovIdx]
		ctx.Rect(c.X+gap/2, c.Y+gap/2,
			c.W-gap, c.H-gap, th.AxisColor, 2)
	}

	// Tooltip clamped to plot area so it doesn't overlap the
	// title text in the top padding.
	if hovIdx >= 0 {
		c := &tv.cells[hovIdx]
		tipLabel := c.Node.Label + "\n" +
			fmt.Sprintf(cfg.ValueFormat, c.Node.TotalValue())
		drawTooltip(ctx, c.X+c.W/2, c.Y+c.H/2,
			tipLabel, th, plotRect{left, right, top, bottom})
	}

}

package chart

import (
	"cmp"
	"fmt"
	"log/slog"
	"slices"

	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// SankeyNode defines a node in a Sankey diagram.
type SankeyNode struct {
	Label string
	Color gui.Color // zero = palette
}

// SankeyLink defines a flow between two nodes.
type SankeyLink struct {
	Source int       // index into Nodes
	Target int       // index into Nodes
	Value  float64   // flow magnitude (determines ribbon width)
	Color  gui.Color // zero = derive from source node
}

// SankeyCfg configures a Sankey diagram (flow chart with
// proportional-width ribbons connecting nodes in columns).
type SankeyCfg struct {
	BaseCfg

	Nodes []SankeyNode
	Links []SankeyLink

	// NodeWidth is the width in pixels of node rectangles.
	// Zero defaults to DefaultSankeyNodeWidth.
	NodeWidth float32

	// NodeGap is the vertical gap in pixels between nodes in the
	// same column. Zero defaults to DefaultSankeyNodeGap.
	NodeGap float32

	// ShowLabels renders node label text beside each rectangle.
	ShowLabels bool

	// ValueFormat is the fmt format string for tooltip values.
	// Zero value defaults to "%.0f".
	ValueFormat string
}

// sankeyLayoutNode holds computed layout for one node.
type sankeyLayoutNode struct {
	Col     int
	X, Y    float32 // top-left of rectangle
	W, H    float32 // width and height
	Index   int
	InFlow  float64 // sum of incoming link values
	OutFlow float64 // sum of outgoing link values
}

// sankeyLayoutLink holds computed layout for one link ribbon.
type sankeyLayoutLink struct {
	Index int
	Poly  []float32    // concave polygon for hit-testing
	Quads [][8]float32 // convex quads for rendering
	SrcY  float32      // center Y at source side
	DstY  float32      // center Y at target side
}

type sankeyView struct {
	cfg      SankeyCfg
	hoverPx  float32
	hoverPy  float32
	hovering bool
	nodes    []sankeyLayoutNode
	links    []sankeyLayoutLink
	win      *gui.Window
}

// Sankey creates a Sankey diagram view.
func Sankey(cfg SankeyCfg) gui.View {
	cfg.applyDefaults()
	cfg.NodeWidth = cmp.Or(cfg.NodeWidth, DefaultSankeyNodeWidth)
	cfg.NodeGap = cmp.Or(cfg.NodeGap, DefaultSankeyNodeGap)
	cfg.ValueFormat = cmp.Or(cfg.ValueFormat, "%.0f")
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableSankey(&cfg.BaseCfg, cfg.Nodes, cfg.Links)
	}
	return &sankeyView{cfg: cfg}
}

// Draw renders the chart onto dc for headless export.
func (sv *sankeyView) Draw(dc *gui.DrawContext) { sv.draw(dc) }

func (sv *sankeyView) chartTheme() *theme.Theme { return sv.cfg.Theme }

func (sv *sankeyView) Content() []gui.View { return nil }

func (sv *sankeyView) GenerateLayout(w *gui.Window) gui.Layout {
	c := &sv.cfg
	hovV := loadHover(w, c.ID,
		&sv.hovering, &sv.hoverPx, &sv.hoverPy)
	sv.win = w
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
		OnDraw:       sv.draw,
		OnClick:      sv.internalClick,
		OnHover:      sv.internalHover,
		OnMouseLeave: sv.internalMouseLeave,
	}).GenerateLayout(w)
}

func (sv *sankeyView) internalClick(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	if sv.cfg.OnClick != nil {
		sv.cfg.OnClick(l, e, w)
	}
}

func (sv *sankeyView) internalHover(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	sv.hoverPx = e.MouseX - l.Shape.X
	sv.hoverPy = e.MouseY - l.Shape.Y
	sv.hovering = true
	saveHover(w, l, sv.cfg.ID, true, sv.hoverPx, sv.hoverPy)
	_, _, ok := sv.hitTest(sv.hoverPx, sv.hoverPy)
	if ok {
		w.SetMouseCursorPointingHand()
	} else {
		w.SetMouseCursorArrow()
	}
	if sv.cfg.OnHover != nil {
		sv.cfg.OnHover(l, e, w)
	}
}

func (sv *sankeyView) internalMouseLeave(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	sv.hovering = false
	saveHover(w, l, sv.cfg.ID, false, 0, 0)
	w.SetMouseCursorArrow()
	if sv.cfg.OnMouseLeave != nil {
		sv.cfg.OnMouseLeave(l, e, w)
	}
}

// hitTest returns ("node", index, true) or ("link", index, true)
// for the element under (mx, my), or ("", 0, false) on miss.
func (sv *sankeyView) hitTest(
	mx, my float32,
) (string, int, bool) {
	// Check nodes first (they are on top).
	for i := range sv.nodes {
		n := &sv.nodes[i]
		if mx >= n.X && mx <= n.X+n.W &&
			my >= n.Y && my <= n.Y+n.H {
			return "node", i, true
		}
	}
	// Check links in reverse order (last drawn = on top).
	for i := len(sv.links) - 1; i >= 0; i-- {
		if pointInPolygon(mx, my, sv.links[i].Poly) {
			return "link", i, true
		}
	}
	return "", 0, false
}

// pointInPolygon tests if (px, py) is inside a polygon defined
// by flat [x0,y0, x1,y1, ...] vertices using ray casting.
func pointInPolygon(px, py float32, poly []float32) bool {
	n := len(poly) / 2
	if n < 3 {
		return false
	}
	inside := false
	jx := poly[(n-1)*2]
	jy := poly[(n-1)*2+1]
	for i := range n {
		ix := poly[i*2]
		iy := poly[i*2+1]
		if (iy > py) != (jy > py) {
			slope := (px-ix)*(jy-iy) - (jx-ix)*(py-iy)
			if slope == 0 {
				return true // on edge
			}
			if (slope < 0) != (jy < iy) {
				inside = !inside
			}
		}
		jx = ix
		jy = iy
	}
	return inside
}

// --- layout -----------------------------------------------------------

// sankeyLinkValid reports whether a link has a positive finite
// value and valid, distinct source/target indices.
func sankeyLinkValid(lk *SankeyLink, n int) bool {
	return finite(lk.Value) && lk.Value > 0 &&
		lk.Source >= 0 && lk.Source < n &&
		lk.Target >= 0 && lk.Target < n &&
		lk.Source != lk.Target
}

// sankeyAssignColumns assigns column indices via topological
// ordering. Returns the number of columns.
func sankeyAssignColumns(
	nodes []sankeyLayoutNode, links []SankeyLink,
) int {
	n := len(nodes)
	col := make([]int, n)
	// Initialize all to 0; for each link, target column is at
	// least source column + 1. Limit iterations to n to guard
	// against cycles that escaped validation.
	for range n {
		changed := false
		for i := range links {
			if !sankeyLinkValid(&links[i], n) {
				continue
			}
			want := col[links[i].Source] + 1
			if col[links[i].Target] < want {
				col[links[i].Target] = want
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	maxCol := 0
	for i := range nodes {
		nodes[i].Col = col[i]
		maxCol = max(maxCol, col[i])
	}
	return maxCol + 1
}

// sankeyComputeFlows sums incoming and outgoing flows per node.
func sankeyComputeFlows(
	nodes []sankeyLayoutNode, links []SankeyLink,
) {
	n := len(nodes)
	for i := range nodes {
		nodes[i].InFlow = 0
		nodes[i].OutFlow = 0
	}
	for i := range links {
		if !sankeyLinkValid(&links[i], n) {
			continue
		}
		nodes[links[i].Source].OutFlow += links[i].Value
		nodes[links[i].Target].InFlow += links[i].Value
	}
}

// sankeyNodeThroughput returns the throughput of a node (max of
// incoming and outgoing flow).
func sankeyNodeThroughput(n *sankeyLayoutNode) float64 {
	return max(n.InFlow, n.OutFlow)
}

// sankeyLayoutNodes computes node positions within the plot area.
func sankeyLayoutNodes(
	nodes []sankeyLayoutNode, numCols int,
	nodeWidth, nodeGap float32,
	left, right, top, bottom float32,
) {
	availW := right - left
	availH := bottom - top

	// Horizontal spacing.
	colWidth := float32(0)
	if numCols > 1 {
		colWidth = (availW - nodeWidth) / float32(numCols-1)
	}

	// Group node indices by column.
	colNodes := make([][]int, numCols)
	for i := range nodes {
		c := nodes[i].Col
		colNodes[c] = append(colNodes[c], i)
	}

	// Compute total throughput per column for scaling.
	for c := range numCols {
		indices := colNodes[c]
		if len(indices) == 0 {
			continue
		}

		totalFlow := 0.0
		for _, idx := range indices {
			totalFlow += sankeyNodeThroughput(&nodes[idx])
		}
		if totalFlow <= 0 {
			continue
		}

		totalGap := nodeGap * float32(len(indices)-1)
		usable := availH - totalGap
		if usable < 1 {
			usable = 1
		}

		// Sort by throughput descending for stable layout.
		slices.SortFunc(indices, func(a, b int) int {
			if c := cmp.Compare(
				sankeyNodeThroughput(&nodes[b]),
				sankeyNodeThroughput(&nodes[a])); c != 0 {
				return c
			}
			return cmp.Compare(a, b)
		})

		y := top
		for _, idx := range indices {
			flow := sankeyNodeThroughput(&nodes[idx])
			h := float32(flow/totalFlow) * usable
			if h < 2 {
				h = 2
			}
			nodes[idx].X = left + float32(nodes[idx].Col)*colWidth
			nodes[idx].Y = y
			nodes[idx].W = nodeWidth
			nodes[idx].H = h
			nodes[idx].Index = idx
			y += h + nodeGap
		}
	}
}

// --- bezier ribbon ----------------------------------------------------

// cubicBezier evaluates a 1D cubic bezier at parameter t.
func cubicBezier(t, p0, p1, p2, p3 float32) float32 {
	u := 1 - t
	return u*u*u*p0 + 3*u*u*t*p1 + 3*u*t*t*p2 + t*t*t*p3
}

const sankeyBezierSamples = 24

// sankeyRibbonGeom builds geometry for a flow ribbon between
// source right edge and target left edge. Returns a concave
// polygon for hit-testing and convex quads for rendering
// (FilledPolygon requires convex input).
func sankeyRibbonGeom(
	sx, sy0, sy1 float32, // source right edge, top/bottom Y
	tx, ty0, ty1 float32, // target left edge, top/bottom Y
) ([]float32, [][8]float32) {
	n := sankeyBezierSamples

	// Control point X at 1/3 and 2/3 of horizontal distance.
	cx1 := sx + (tx-sx)/3
	cx2 := tx - (tx-sx)/3

	// Sample top and bottom edges in a single allocation.
	// Layout: [topX0 topY0 botY0  topX1 topY1 botY1  ...]
	// (botX == topX so not stored separately.)
	samples := make([]float32, (n+1)*3)
	for i := 0; i <= n; i++ {
		t := float32(i) / float32(n)
		off := i * 3
		samples[off] = cubicBezier(t, sx, cx1, cx2, tx)
		samples[off+1] = cubicBezier(t, sy0, sy0, ty0, ty0)
		samples[off+2] = cubicBezier(t, sy1, sy1, ty1, ty1)
	}

	// Build concave polygon for hit-testing (ray-cast is fine
	// with concave shapes).
	poly := make([]float32, 0, (2*n+2)*2)
	for i := 0; i <= n; i++ {
		off := i * 3
		poly = append(poly, samples[off], samples[off+1])
	}
	for i := n; i >= 0; i-- {
		off := i * 3
		poly = append(poly, samples[off], samples[off+2])
	}

	// Build convex quads: one per adjacent pair of samples.
	quads := make([][8]float32, n)
	for i := range n {
		a := i * 3
		b := a + 3
		quads[i] = [8]float32{
			samples[a], samples[a+1],
			samples[b], samples[b+1],
			samples[b], samples[b+2],
			samples[a], samples[a+2],
		}
	}
	return poly, quads
}

// --- link layout ------------------------------------------------------

// sankeyLayoutLinks computes ribbon polygons for all valid links.
// Reuses the backing array of dst to reduce allocations.
func sankeyLayoutLinks(
	nodes []sankeyLayoutNode, cfgLinks []SankeyLink,
	dst []sankeyLayoutLink,
) []sankeyLayoutLink {
	n := len(nodes)

	// Track consumed port space per node edge.
	srcOffset := make([]float32, n) // right edge offset
	dstOffset := make([]float32, n) // left edge offset

	links := dst[:0]
	for i := range cfgLinks {
		lk := &cfgLinks[i]
		if !sankeyLinkValid(lk, n) {
			continue
		}

		src := &nodes[lk.Source]
		tgt := &nodes[lk.Target]

		// Ribbon width proportional to node height.
		srcThru := sankeyNodeThroughput(src)
		tgtThru := sankeyNodeThroughput(tgt)
		if srcThru <= 0 || tgtThru <= 0 {
			continue
		}

		srcH := float32(lk.Value/srcThru) * src.H
		tgtH := float32(lk.Value/tgtThru) * tgt.H

		sy0 := src.Y + srcOffset[lk.Source]
		sy1 := sy0 + srcH
		srcOffset[lk.Source] += srcH

		dy0 := tgt.Y + dstOffset[lk.Target]
		dy1 := dy0 + tgtH
		dstOffset[lk.Target] += tgtH

		sx := src.X + src.W // right edge of source
		tx := tgt.X         // left edge of target

		poly, quads := sankeyRibbonGeom(sx, sy0, sy1, tx, dy0, dy1)
		links = append(links, sankeyLayoutLink{
			Index: i,
			Poly:  poly,
			Quads: quads,
			SrcY:  (sy0 + sy1) / 2,
			DstY:  (dy0 + dy1) / 2,
		})
	}
	return links
}

// --- draw -------------------------------------------------------------

func (sv *sankeyView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &sv.cfg
	th := cfg.Theme
	progress := animProgress(sv.win, sv.cfg.ID)

	if len(cfg.Nodes) == 0 || len(cfg.Links) == 0 {
		slog.Warn("no sankey data", "chart", cfg.ID)
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

	// Build layout nodes.
	sv.nodes = sv.nodes[:0]
	for i := range cfg.Nodes {
		sv.nodes = append(sv.nodes, sankeyLayoutNode{Index: i})
	}

	numCols := sankeyAssignColumns(sv.nodes, cfg.Links)
	sankeyComputeFlows(sv.nodes, cfg.Links)
	sankeyLayoutNodes(sv.nodes, numCols,
		cfg.NodeWidth, cfg.NodeGap,
		left, right, top, bottom)

	sv.links = sankeyLayoutLinks(sv.nodes, cfg.Links, sv.links)

	// Determine hover target.
	hovKind := ""
	hovIdx := -1
	if sv.hovering {
		kind, idx, ok := sv.hitTest(sv.hoverPx, sv.hoverPy)
		if ok {
			hovKind = kind
			hovIdx = idx
		}
	}

	// Build hover highlight sets.
	hlNode, hlLink := sv.buildHighlights(hovKind, hovIdx)
	anyHover := hovIdx >= 0

	// Draw links, nodes, labels, hover border, tooltip.
	sv.drawLinks(ctx, th, hlLink, anyHover, progress)
	sv.drawNodes(ctx, th, hlNode, anyHover, progress)
	sv.drawLabels(ctx, th, left, right)
	sv.drawHoverBorder(ctx, th, hovKind, hovIdx, anyHover)
	sv.drawSankeyTooltip(ctx, th, hovKind, hovIdx,
		anyHover, plotRect{left, right, top, bottom})
}

// buildHighlights returns per-node and per-link highlight flags.
func (sv *sankeyView) buildHighlights(
	hovKind string, hovIdx int,
) ([]bool, []bool) {
	hlNode := make([]bool, len(sv.nodes))
	hlLink := make([]bool, len(sv.links))
	if hovIdx < 0 {
		return hlNode, hlLink
	}
	cfg := &sv.cfg
	if hovKind == "node" {
		hlNode[hovIdx] = true
		for i := range sv.links {
			lk := &cfg.Links[sv.links[i].Index]
			if lk.Source == hovIdx || lk.Target == hovIdx {
				hlLink[i] = true
			}
		}
	} else {
		hlLink[hovIdx] = true
		lk := &cfg.Links[sv.links[hovIdx].Index]
		if lk.Source >= 0 && lk.Source < len(sv.nodes) {
			hlNode[lk.Source] = true
		}
		if lk.Target >= 0 && lk.Target < len(sv.nodes) {
			hlNode[lk.Target] = true
		}
	}
	return hlNode, hlLink
}

func (sv *sankeyView) drawLinks(
	ctx *render.Context, th *theme.Theme,
	hlLink []bool, anyHover bool, progress float32,
) {
	cfg := &sv.cfg
	for i := range sv.links {
		ll := &sv.links[i]
		lk := &cfg.Links[ll.Index]

		color := lk.Color
		if color == (gui.Color{}) {
			if lk.Source >= 0 && lk.Source < len(cfg.Nodes) {
				color = seriesColor(
					cfg.Nodes[lk.Source].Color,
					lk.Source, th.Palette)
			} else {
				color = gui.Hex(0x808080)
			}
		}
		color = gui.RGBA(color.R, color.G, color.B,
			DefaultSankeyLinkAlpha)

		if anyHover && !hlLink[i] {
			color = dimColor(color, HoverDimAlpha)
		}
		// Render as convex quad strips (FilledPolygon
		// requires convex input).
		for j := range ll.Quads {
			if progress >= 1 {
				ctx.FilledPolygon(ll.Quads[j][:], color)
			} else {
				q := ll.Quads[j]
				// Scale ribbon height toward vertical
				// midpoint of each quad.
				midY := (q[1] + q[3] + q[5] + q[7]) / 4
				q[1] = midY + (q[1]-midY)*progress
				q[3] = midY + (q[3]-midY)*progress
				q[5] = midY + (q[5]-midY)*progress
				q[7] = midY + (q[7]-midY)*progress
				ctx.FilledPolygon(q[:], color)
			}
		}
	}
}

func (sv *sankeyView) drawNodes(
	ctx *render.Context, th *theme.Theme,
	hlNode []bool, anyHover bool, progress float32,
) {
	cfg := &sv.cfg
	for i := range sv.nodes {
		n := &sv.nodes[i]
		color := seriesColor(
			cfg.Nodes[i].Color, i, th.Palette)
		if anyHover && !hlNode[i] {
			color = dimColor(color, HoverDimAlpha)
		}
		h := n.H * progress
		y := n.Y + (n.H-h)/2
		ctx.FilledRect(n.X, y, n.W, h, color)
	}
}

func (sv *sankeyView) drawLabels(
	ctx *render.Context, th *theme.Theme,
	left, right float32,
) {
	cfg := &sv.cfg
	if !cfg.ShowLabels {
		return
	}
	tickStyle := th.TickStyle
	fh := ctx.FontHeight(tickStyle)
	for i := range sv.nodes {
		n := &sv.nodes[i]
		if fh > n.H {
			continue
		}
		label := cfg.Nodes[i].Label
		tw := ctx.TextWidth(label, tickStyle)
		midY := n.Y + n.H/2 - fh/2

		// Place label to the right of left-column nodes,
		// to the left of right-column nodes, otherwise right.
		if n.X+n.W+4+tw <= right {
			ctx.Text(n.X+n.W+4, midY, label, tickStyle)
		} else if n.X-4-tw >= left {
			ctx.Text(n.X-4-tw, midY, label, tickStyle)
		}
	}
}

func (sv *sankeyView) drawHoverBorder(
	ctx *render.Context, th *theme.Theme,
	hovKind string, hovIdx int, anyHover bool,
) {
	if !anyHover || hovKind != "node" {
		return
	}
	n := &sv.nodes[hovIdx]
	ctx.Rect(n.X, n.Y, n.W, n.H, th.AxisColor, 2)
}

func (sv *sankeyView) drawSankeyTooltip(
	ctx *render.Context, th *theme.Theme,
	hovKind string, hovIdx int, anyHover bool,
	pr plotRect,
) {
	if !anyHover {
		return
	}
	cfg := &sv.cfg
	if hovKind == "node" {
		n := &sv.nodes[hovIdx]
		flow := sankeyNodeThroughput(n)
		tip := cfg.Nodes[hovIdx].Label + "\n" +
			fmt.Sprintf(cfg.ValueFormat, flow)
		drawTooltip(ctx, sv.hoverPx, sv.hoverPy,
			tip, th, pr)
		return
	}
	lk := &cfg.Links[sv.links[hovIdx].Index]
	src := ""
	dst := ""
	if lk.Source >= 0 && lk.Source < len(cfg.Nodes) {
		src = cfg.Nodes[lk.Source].Label
	}
	if lk.Target >= 0 && lk.Target < len(cfg.Nodes) {
		dst = cfg.Nodes[lk.Target].Label
	}
	tip := src + " → " + dst + "\n" +
		fmt.Sprintf(cfg.ValueFormat, lk.Value)
	drawTooltip(ctx, sv.hoverPx, sv.hoverPy,
		tip, th, pr)
}

// hasCycle detects a cycle in the directed graph defined by
// nodes and links using iterative DFS.
func hasCycle(numNodes int, links []SankeyLink) bool {
	// Build adjacency list.
	adj := make([][]int, numNodes)
	for _, lk := range links {
		if lk.Source < 0 || lk.Source >= numNodes ||
			lk.Target < 0 || lk.Target >= numNodes {
			continue
		}
		adj[lk.Source] = append(adj[lk.Source], lk.Target)
	}

	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make([]int, numNodes)

	for start := range numNodes {
		if color[start] != white {
			continue
		}
		stack := []int{start}
		color[start] = gray
		for len(stack) > 0 {
			u := stack[len(stack)-1]
			found := false
			for len(adj[u]) > 0 {
				v := adj[u][0]
				adj[u] = adj[u][1:]
				if color[v] == gray {
					return true
				}
				if color[v] == white {
					color[v] = gray
					stack = append(stack, v)
					found = true
					break
				}
			}
			if !found {
				color[u] = black
				stack = stack[:len(stack)-1]
			}
		}
	}
	return false
}

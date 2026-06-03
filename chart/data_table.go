package chart

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

// idScrollHash returns a non-zero uint32 from a string ID for
// use as a gui.Table IDScroll key.
func idScrollHash(id string) uint32 {
	h := uint32(2166136261)
	for i := range id {
		h ^= uint32(id[i])
		h *= 16777619
	}
	if h == 0 {
		h = 1
	}
	return h
}

// fmtVal formats a float64 for display in a data table.
func fmtVal(v float64) string {
	return strconv.FormatFloat(v, 'g', 6, 64)
}

// dataTableView builds a gui.Table from headers and rows using
// chart config for sizing and identity.
func dataTableView(
	cfg *BaseCfg, headers []string, rows [][]string,
) gui.View {
	hAlign := gui.HAlignRight
	data := make([]gui.TableRowCfg, 0, 1+len(rows))

	// Header row.
	hCells := make([]gui.TableCellCfg, len(headers))
	for i, h := range headers {
		hCells[i] = gui.TableCellCfg{Value: h, HeadCell: true}
	}
	data = append(data, gui.TableRowCfg{Cells: hCells})

	// Data rows.
	for _, row := range rows {
		cells := make([]gui.TableCellCfg, len(row))
		for j, v := range row {
			cells[j] = gui.TableCellCfg{Value: v, HAlign: &hAlign}
		}
		// Left-align first column (labels).
		if len(cells) > 0 {
			cells[0].HAlign = nil
		}
		data = append(data, gui.TableRowCfg{Cells: cells})
	}

	id := cfg.ID
	if id != "" {
		id += "-table"
	}

	return gui.Table(gui.TableCfg{
		ID:           id,
		IDScroll:     idScrollHash(id),
		Data:         data,
		ColorBorder:  gui.Gray,
		SizeBorder:   1,
		BorderStyle:  gui.TableBorderAll,
		FreezeHeader: true,
		Sizing:       gui.FillFit,
		MaxHeight:    cfg.Height,
	})
}

// --- XY series (Line, Area, Scatter) ---

func xyTableData(ss []series.XY) ([]string, [][]string) {
	if len(ss) == 0 {
		return []string{"X", "Y"}, nil
	}

	// Single series: simple X, Y columns.
	if len(ss) == 1 {
		s := ss[0]
		headers := []string{"X", "Y"}
		if s.Name() != "" {
			headers[1] = s.Name()
		}
		rows := make([][]string, len(s.Points))
		for i, p := range s.Points {
			rows[i] = []string{fmtVal(p.X), fmtVal(p.Y)}
		}
		return headers, rows
	}

	// Multi-series: X, Y(s1), Y(s2), ...
	// Rows keyed by point index (per-series).
	headers := make([]string, 0, 1+len(ss))
	headers = append(headers, "X")
	for _, s := range ss {
		headers = append(headers, cmp.Or(s.Name(), "Y"))
	}

	maxLen := 0
	for _, s := range ss {
		maxLen = max(maxLen, len(s.Points))
	}

	rows := make([][]string, maxLen)
	for i := range maxLen {
		row := make([]string, 1+len(ss))
		// Use X from the first series that has this index.
		xSet := false
		for si, s := range ss {
			if i < len(s.Points) {
				if !xSet {
					row[0] = fmtVal(s.Points[i].X)
					xSet = true
				}
				row[1+si] = fmtVal(s.Points[i].Y)
			}
		}
		rows[i] = row
	}
	return headers, rows
}

func dataTableXY(cfg *BaseCfg, ss []series.XY) gui.View {
	h, r := xyTableData(ss)
	return dataTableView(cfg, h, r)
}

// --- XYZ series (Bubble) ---

func xyzTableData(ss []series.XYZ) ([]string, [][]string) {
	if len(ss) == 0 {
		return []string{"Series", "X", "Y", "Z"}, nil
	}

	headers := []string{"Series", "X", "Y", "Z"}
	var rows [][]string
	for _, s := range ss {
		name := s.Name()
		for _, p := range s.Points {
			rows = append(rows, []string{
				name, fmtVal(p.X), fmtVal(p.Y), fmtVal(p.Z),
			})
		}
	}
	return headers, rows
}

func dataTableXYZ(cfg *BaseCfg, ss []series.XYZ) gui.View {
	h, r := xyzTableData(ss)
	return dataTableView(cfg, h, r)
}

// --- Category series (Bar) ---

func categoryTableData(
	ss []series.Category,
) ([]string, [][]string) {
	if len(ss) == 0 {
		return []string{"Label", "Value"}, nil
	}

	headers := make([]string, 0, 1+len(ss))
	headers = append(headers, "Label")
	for _, s := range ss {
		headers = append(headers, cmp.Or(s.Name(), "Value"))
	}

	// Use labels from first series as row keys.
	n := 0
	for _, s := range ss {
		n = max(n, len(s.Values))
	}

	rows := make([][]string, n)
	for i := range n {
		row := make([]string, 1+len(ss))
		for si, s := range ss {
			if i < len(s.Values) {
				if si == 0 || row[0] == "" {
					row[0] = s.Values[i].Label
				}
				row[1+si] = fmtVal(s.Values[i].Value)
			}
		}
		rows[i] = row
	}
	return headers, rows
}

func dataTableCategory(
	cfg *BaseCfg, ss []series.Category,
) gui.View {
	h, r := categoryTableData(ss)
	return dataTableView(cfg, h, r)
}

// --- OHLC series (Candlestick) ---

func ohlcTableData(
	ss []series.OHLCSeries, timeFmt string,
) ([]string, [][]string) {
	timeFmt = cmp.Or(timeFmt, "2006-01-02")
	headers := []string{"Time", "Open", "High", "Low", "Close", "Volume"}
	var rows [][]string
	for _, s := range ss {
		for _, p := range s.Points {
			rows = append(rows, []string{
				p.Time.Format(timeFmt),
				fmtVal(p.Open), fmtVal(p.High),
				fmtVal(p.Low), fmtVal(p.Close),
				fmtVal(p.Volume),
			})
		}
	}
	return headers, rows
}

func dataTableOHLC(
	cfg *BaseCfg, ss []series.OHLCSeries, timeFmt string,
) gui.View {
	h, r := ohlcTableData(ss, timeFmt)
	return dataTableView(cfg, h, r)
}

// --- PieSlice (Pie, Funnel) ---

func sliceTableData(slices []PieSlice) ([]string, [][]string) {
	headers := []string{"Label", "Value", "%"}
	total := 0.0
	for _, s := range slices {
		if finite(s.Value) && s.Value > 0 {
			total += s.Value
		}
	}
	rows := make([][]string, len(slices))
	for i, s := range slices {
		pct := ""
		if total > 0 && finite(s.Value) {
			pct = fmt.Sprintf("%.1f", s.Value/total*100)
		}
		rows[i] = []string{s.Label, fmtVal(s.Value), pct}
	}
	return headers, rows
}

func dataTableSlices(cfg *BaseCfg, slices []PieSlice) gui.View {
	h, r := sliceTableData(slices)
	return dataTableView(cfg, h, r)
}

// --- Gauge ---

func gaugeTableData(
	value, mn, mx float64,
) ([]string, [][]string) {
	return []string{"Metric", "Value"}, [][]string{
		{"Value", fmtVal(value)},
		{"Min", fmtVal(mn)},
		{"Max", fmtVal(mx)},
	}
}

func dataTableGauge(
	cfg *BaseCfg, value, mn, mx float64,
) gui.View {
	h, r := gaugeTableData(value, mn, mx)
	return dataTableView(cfg, h, r)
}

// --- Histogram ---

func histogramTableData(
	data []float64, numBins int, edges []float64, normalized bool,
) ([]string, [][]string) {
	binEdges, counts := calcBins(data, numBins, edges)
	label := "Count"
	if normalized {
		label = "Density"
	}
	headers := []string{"Bin Range", label}
	rows := make([][]string, len(counts))
	for i, c := range counts {
		lo := fmtVal(binEdges[i])
		hi := fmtVal(binEdges[i+1])
		val := fmtVal(float64(c))
		if normalized && len(binEdges) > i+1 && len(data) > 0 {
			w := binEdges[i+1] - binEdges[i]
			if w > 0 {
				val = fmtVal(float64(c) / float64(len(data)) / w)
			}
		}
		rows[i] = []string{
			fmt.Sprintf("[%s, %s]", lo, hi), val,
		}
	}
	return headers, rows
}

func dataTableHistogram(
	cfg *BaseCfg, data []float64, bins int,
	edges []float64, normalized bool,
) gui.View {
	h, r := histogramTableData(data, bins, edges, normalized)
	return dataTableView(cfg, h, r)
}

// --- Boxplot ---

func boxplotTableData(data []BoxData) ([]string, [][]string) {
	headers := []string{
		"Name", "Min", "Q1", "Median", "Q3", "Max", "Outliers",
	}
	rows := make([][]string, 0, len(data))
	for _, d := range data {
		stats, ok := computeBoxStats(d.Values)
		if !ok {
			continue
		}
		outliers := make([]string, len(stats.Outliers))
		for i, o := range stats.Outliers {
			outliers[i] = fmtVal(o)
		}
		rows = append(rows, []string{
			d.Label,
			fmtVal(stats.Min), fmtVal(stats.Q1),
			fmtVal(stats.Median), fmtVal(stats.Q3),
			fmtVal(stats.Max),
			strings.Join(outliers, ", "),
		})
	}
	return headers, rows
}

func dataTableBoxplot(cfg *BaseCfg, data []BoxData) gui.View {
	h, r := boxplotTableData(data)
	return dataTableView(cfg, h, r)
}

// --- Waterfall ---

func waterfallTableData(
	values []WaterfallValue,
) ([]string, [][]string) {
	headers := []string{"Label", "Value", "Type"}
	rows := make([][]string, len(values))
	for i, v := range values {
		typ := "Delta"
		if v.IsTotal {
			typ = "Total"
		}
		rows[i] = []string{v.Label, fmtVal(v.Value), typ}
	}
	return headers, rows
}

func dataTableWaterfall(
	cfg *BaseCfg, values []WaterfallValue,
) gui.View {
	h, r := waterfallTableData(values)
	return dataTableView(cfg, h, r)
}

// --- Combo ---

func comboTableData(
	ss []ComboSeries,
) ([]string, [][]string) {
	if len(ss) == 0 {
		return []string{"Label", "Value"}, nil
	}

	headers := make([]string, 0, 1+len(ss))
	headers = append(headers, "Label")
	for _, s := range ss {
		headers = append(headers, cmp.Or(s.Name(), "Value"))
	}

	n := 0
	for _, s := range ss {
		n = max(n, len(s.Values))
	}

	rows := make([][]string, n)
	for i := range n {
		row := make([]string, 1+len(ss))
		for si, s := range ss {
			if i < len(s.Values) {
				if row[0] == "" {
					row[0] = s.Values[i].Label
				}
				row[1+si] = fmtVal(s.Values[i].Value)
			}
		}
		rows[i] = row
	}
	return headers, rows
}

func dataTableCombo(
	cfg *BaseCfg, ss []ComboSeries,
) gui.View {
	h, r := comboTableData(ss)
	return dataTableView(cfg, h, r)
}

// --- Radar ---

func radarTableData(
	axes []RadarAxis, ss []RadarSeries,
) ([]string, [][]string) {
	headers := make([]string, 0, 1+len(ss))
	headers = append(headers, "Axis")
	for _, s := range ss {
		headers = append(headers, cmp.Or(s.Name, "Value"))
	}

	rows := make([][]string, len(axes))
	for i, a := range axes {
		row := make([]string, 1+len(ss))
		row[0] = a.Label
		for si, s := range ss {
			if i < len(s.Values) {
				row[1+si] = fmtVal(s.Values[i])
			}
		}
		rows[i] = row
	}
	return headers, rows
}

func dataTableRadar(
	cfg *BaseCfg, axes []RadarAxis, ss []RadarSeries,
) gui.View {
	h, r := radarTableData(axes, ss)
	return dataTableView(cfg, h, r)
}

// --- Grid (Heatmap) ---

func gridTableData(g series.Grid) ([]string, [][]string) {
	cols := g.Cols()
	headers := make([]string, 0, 1+len(cols))
	headers = append(headers, "")
	headers = append(headers, cols...)

	rowLabels := g.Rows()
	rows := make([][]string, len(rowLabels))
	for r := range rowLabels {
		row := make([]string, 1+len(cols))
		row[0] = rowLabels[r]
		for c := range cols {
			row[1+c] = fmtVal(g.At(r, c))
		}
		rows[r] = row
	}
	return headers, rows
}

func dataTableGrid(cfg *BaseCfg, g series.Grid) gui.View {
	h, r := gridTableData(g)
	return dataTableView(cfg, h, r)
}

// --- Treemap ---

func treeTableData(
	roots []series.TreeNode,
) ([]string, [][]string) {
	const maxDepth = 64
	headers := []string{"Label", "Value"}
	var rows [][]string
	var walk func(nodes []series.TreeNode, depth int)
	walk = func(nodes []series.TreeNode, depth int) {
		if depth >= maxDepth {
			return
		}
		for _, n := range nodes {
			indent := strings.Repeat("  ", depth)
			rows = append(rows, []string{
				indent + n.Label, fmtVal(n.Value),
			})
			if len(n.Children) > 0 {
				walk(n.Children, depth+1)
			}
		}
	}
	walk(roots, 0)
	return headers, rows
}

func dataTableTree(
	cfg *BaseCfg, roots []series.TreeNode,
) gui.View {
	h, r := treeTableData(roots)
	return dataTableView(cfg, h, r)
}

// --- Sankey ---

func sankeyTableData(
	nodes []SankeyNode, links []SankeyLink,
) ([]string, [][]string, []string, [][]string) {
	nh := []string{"Index", "Label"}
	nRows := make([][]string, len(nodes))
	for i, n := range nodes {
		nRows[i] = []string{strconv.Itoa(i), n.Label}
	}

	lh := []string{"Source", "Target", "Value"}
	lRows := make([][]string, len(links))
	for i, l := range links {
		src := strconv.Itoa(l.Source)
		if l.Source >= 0 && l.Source < len(nodes) {
			src = nodes[l.Source].Label
		}
		tgt := strconv.Itoa(l.Target)
		if l.Target >= 0 && l.Target < len(nodes) {
			tgt = nodes[l.Target].Label
		}
		lRows[i] = []string{src, tgt, fmtVal(l.Value)}
	}
	return nh, nRows, lh, lRows
}

func dataTableSankey(
	cfg *BaseCfg, nodes []SankeyNode, links []SankeyLink,
) gui.View {
	nh, nRows, lh, lRows := sankeyTableData(nodes, links)

	// Build two tables stacked vertically.
	nodesCfg := *cfg
	nodesCfg.ID = cfg.ID + "-nodes"
	nodesCfg.Title = ""

	linksCfg := *cfg
	linksCfg.ID = cfg.ID + "-links"
	linksCfg.Title = ""

	return gui.Column(gui.ContainerCfg{
		Sizing:  cfg.Sizing,
		Width:   cfg.Width,
		Height:  cfg.Height,
		Spacing: gui.SomeF(8),
		Content: []gui.View{
			dataTableView(&nodesCfg, nh, nRows),
			dataTableView(&linksCfg, lh, lRows),
		},
	})
}

// --- Sparkline ---

func sparklineTableData(
	values []float64, xy series.XY,
) ([]string, [][]string) {
	if xy.Len() > 0 {
		return xyTableData([]series.XY{xy})
	}
	headers := []string{"Index", "Value"}
	rows := make([][]string, len(values))
	for i, v := range values {
		rows[i] = []string{strconv.Itoa(i), fmtVal(v)}
	}
	return headers, rows
}

func dataTableSparkline(
	cfg *BaseCfg, values []float64, xy series.XY,
) gui.View {
	h, r := sparklineTableData(values, xy)
	return dataTableView(cfg, h, r)
}

package chart

import (
	"strconv"

	"github.com/go-gui-org/go-charts/series"
)

// checkSeriesLengths returns an error string if any series has
// a different length than the first. Returns "" if all lengths
// match or there are fewer than 2 entries.
func checkSeriesLengths(lengths []int) string {
	if len(lengths) < 2 {
		return ""
	}
	n := lengths[0]
	for i, l := range lengths[1:] {
		if l != n {
			return "series length mismatch: series 0 has " +
				strconv.Itoa(n) + " values, series " +
				strconv.Itoa(i+1) + " has " +
				strconv.Itoa(l)
		}
	}
	return ""
}

// categoryLengths returns the value counts for each category
// series.
func categoryLengths(ss []series.Category) []int {
	out := make([]int, len(ss))
	for i, s := range ss {
		out[i] = len(s.Values)
	}
	return out
}

// Validate checks LineCfg for invalid or contradictory settings.
// Returns nil when valid.
func (c *LineCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Series) == 0 && len(c.ErrorSeries) == 0 {
		errs = append(errs, "no series data")
	}
	if c.LineWidth < 0 {
		errs = append(errs, "negative LineWidth")
	}
	return buildError("chart.Line", errs)
}

// Validate checks BarCfg for invalid or contradictory settings.
// Returns nil when valid.
func (c *BarCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Series) == 0 {
		errs = append(errs, "no series data")
	}
	if c.BarWidth < 0 {
		errs = append(errs, "negative BarWidth")
	}
	if c.BarGap < 0 {
		errs = append(errs, "negative BarGap")
	}
	if c.Radius < 0 {
		errs = append(errs, "negative Radius")
	}
	if msg := checkSeriesLengths(categoryLengths(c.Series)); msg != "" {
		errs = append(errs, msg)
	}
	return buildError("chart.Bar", errs)
}

// Validate checks AreaCfg for invalid or contradictory settings.
// Returns nil when valid.
func (c *AreaCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Series) == 0 {
		errs = append(errs, "no series data")
	}
	if c.LineWidth < 0 {
		errs = append(errs, "negative LineWidth")
	}
	if c.Opacity < 0 || c.Opacity > 1 {
		errs = append(errs, "Opacity out of range [0,1]")
	}
	return buildError("chart.Area", errs)
}

// Validate checks ScatterCfg for invalid or contradictory
// settings. Returns nil when valid.
func (c *ScatterCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Series) == 0 && len(c.ErrorSeries) == 0 {
		errs = append(errs, "no series data")
	}
	if c.MarkerSize < 0 {
		errs = append(errs, "negative MarkerSize")
	}
	return buildError("chart.Scatter", errs)
}

// Validate checks BubbleCfg for invalid or contradictory
// settings. Returns nil when valid.
func (c *BubbleCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Series) == 0 {
		errs = append(errs, "no series data")
	}
	if c.MinRadius < 0 {
		errs = append(errs, "negative MinRadius")
	}
	if c.MaxRadius < 0 {
		errs = append(errs, "negative MaxRadius")
	}
	if c.MinRadius > 0 && c.MaxRadius > 0 && c.MinRadius > c.MaxRadius {
		errs = append(errs, "MinRadius exceeds MaxRadius")
	}
	return buildError("chart.Bubble", errs)
}

// Validate checks PieCfg for invalid or contradictory settings.
// Returns nil when valid.
func (c *PieCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Slices) == 0 {
		errs = append(errs, "no slice data")
	}
	if c.InnerRadius < 0 {
		errs = append(errs, "negative InnerRadius")
	}
	return buildError("chart.Pie", errs)
}

// Validate checks HistogramCfg for invalid settings.
// Returns nil when valid.
func (c *HistogramCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if c.Bins < 0 {
		errs = append(errs, "negative Bins")
	}
	if c.Radius < 0 {
		errs = append(errs, "negative Radius")
	}
	if len(c.BinEdges) == 1 {
		errs = append(errs, "BinEdges must have 0 or 2+ entries")
	}
	return buildError("chart.Histogram", errs)
}

// Validate checks WaterfallCfg for invalid settings.
// Returns nil when valid.
func (c *WaterfallCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Values) == 0 {
		errs = append(errs, "no values")
	}
	if c.BarWidth < 0 {
		errs = append(errs, "negative BarWidth")
	}
	if c.Radius < 0 {
		errs = append(errs, "negative Radius")
	}
	return buildError("chart.Waterfall", errs)
}

// Validate checks BoxPlotCfg for invalid settings.
// Returns nil when valid.
func (c *BoxPlotCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Data) == 0 {
		errs = append(errs, "no data")
	}
	if c.BoxWidth < 0 {
		errs = append(errs, "negative BoxWidth")
	}
	if c.OutlierRadius < 0 {
		errs = append(errs, "negative OutlierRadius")
	}
	return buildError("chart.BoxPlot", errs)
}

// Validate checks HeatmapCfg for invalid settings.
// Returns nil when valid.
func (c *HeatmapCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if c.Data.NumRows() == 0 || c.Data.NumCols() == 0 {
		errs = append(errs, "empty grid data")
	}
	if c.CellGap < 0 {
		errs = append(errs, "negative CellGap")
	}
	return buildError("chart.Heatmap", errs)
}

// Validate checks ComboCfg for invalid settings.
// Returns nil when valid.
func (c *ComboCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Series) == 0 {
		errs = append(errs, "no series data")
	}
	if c.BarWidth < 0 {
		errs = append(errs, "negative BarWidth")
	}
	if c.BarGap < 0 {
		errs = append(errs, "negative BarGap")
	}
	if c.Radius < 0 {
		errs = append(errs, "negative Radius")
	}
	if c.LineWidth < 0 {
		errs = append(errs, "negative LineWidth")
	}
	if len(c.Series) > 1 {
		lengths := make([]int, len(c.Series))
		for i, s := range c.Series {
			lengths[i] = len(s.Values)
		}
		if msg := checkSeriesLengths(lengths); msg != "" {
			errs = append(errs, msg)
		}
	}
	return buildError("chart.Combo", errs)
}

// Validate checks FunnelCfg for invalid settings.
// Returns nil when valid.
func (c *FunnelCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Slices) == 0 {
		errs = append(errs, "no slice data")
	}
	if c.SegmentGap < 0 {
		errs = append(errs, "negative SegmentGap")
	}
	if c.MinWidthRatio < 0 || c.MinWidthRatio > 1 {
		errs = append(errs, "MinWidthRatio out of range [0,1]")
	}
	for _, s := range c.Slices {
		if s.Value < 0 {
			errs = append(errs, "negative slice value")
			break
		}
	}
	return buildError("chart.Funnel", errs)
}

// Validate checks TreemapCfg for invalid settings.
// Returns nil when valid.
func (c *TreemapCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Data) == 0 {
		errs = append(errs, "no tree data")
	}
	if c.CellGap < 0 {
		errs = append(errs, "negative CellGap")
	}
	if c.MaxDepth < 0 {
		errs = append(errs, "negative MaxDepth")
	}
	if c.HeaderHeight < 0 {
		errs = append(errs, "negative HeaderHeight")
	}
	for i := range c.Data {
		if hasNegativeLeaf(&c.Data[i]) {
			errs = append(errs, "negative leaf value")
			break
		}
	}
	return buildError("chart.Treemap", errs)
}

// Validate checks SankeyCfg for invalid settings.
// Returns nil when valid.
func (c *SankeyCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Nodes) == 0 {
		errs = append(errs, "no node data")
	}
	if len(c.Links) == 0 {
		errs = append(errs, "no link data")
	}
	if c.NodeWidth < 0 {
		errs = append(errs, "negative NodeWidth")
	}
	if c.NodeGap < 0 {
		errs = append(errs, "negative NodeGap")
	}
	n := len(c.Nodes)
	for _, lk := range c.Links {
		if lk.Value < 0 {
			errs = append(errs, "negative link value")
			break
		}
	}
	for _, lk := range c.Links {
		if lk.Source < 0 || lk.Source >= n ||
			lk.Target < 0 || lk.Target >= n {
			errs = append(errs, "link index out of range")
			break
		}
		if lk.Source == lk.Target {
			errs = append(errs, "self-loop link")
			break
		}
	}
	if len(errs) == 0 && n > 0 && len(c.Links) > 0 {
		if hasCycle(n, c.Links) {
			errs = append(errs, "cycle detected in links")
		}
	}
	return buildError("chart.Sankey", errs)
}

// Validate checks SparklineCfg for invalid settings.
// Returns nil when valid.
func (c *SparklineCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Values) == 0 && c.Series.Len() == 0 {
		errs = append(errs, "no data (provide Values or Series)")
	}
	if c.LineWidth < 0 {
		errs = append(errs, "negative LineWidth")
	}
	if c.MarkerRadius < 0 {
		errs = append(errs, "negative MarkerRadius")
	}
	if c.Type < SparklineLine || c.Type > SparklineArea {
		errs = append(errs, "invalid Type")
	}
	return buildError("chart.Sparkline", errs)
}

// hasNegativeLeaf reports whether any leaf in the subtree has
// a negative value.
func hasNegativeLeaf(n *series.TreeNode) bool {
	if n.IsLeaf() {
		return n.Value < 0
	}
	for i := range n.Children {
		if hasNegativeLeaf(&n.Children[i]) {
			return true
		}
	}
	return false
}

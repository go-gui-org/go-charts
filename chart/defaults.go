package chart

// Default values for chart configuration. Applied when the
// corresponding Cfg field is zero-valued.
const (
	DefaultLineWidth   float32 = 2
	DefaultBarGap      float32 = 4
	DefaultMarkerSize  float32 = 6
	DefaultAreaOpacity float32 = 0.3
	DefaultTickLength  float32 = 5

	HoverDimAlpha    uint8   = 60 // alpha for non-hovered elements (~24%)
	HoverExplodeDist float32 = 8  // px offset for exploded pie segment

	DefaultGaugeArcAngle   float32 = 3 * 3.14159265 / 2 // 270°
	DefaultGaugeInnerRatio float32 = 0.7

	// DefaultGaugeRadiusRatio is how much of half the smaller plot
	// side the dial takes. The remainder is the built-in allowance
	// for graduation labels drawn outside the arc.
	DefaultGaugeRadiusRatio float32 = 0.85

	// DefaultGaugePointerHubRatio is the hub circle radius as
	// a fraction of the inner radius.
	DefaultGaugePointerHubRatio float32 = 0.15

	// DefaultGaugePointerWidthPx is the half-width of the
	// needle base in pixels.
	DefaultGaugePointerWidthPx float32 = 4

	// DefaultGaugeValueDropRatio is how far below the centre the
	// value text sits, as a fraction of the inner radius. The
	// needle pivots on the centre, so text drawn there is crossed
	// by the hub and by the needle itself.
	DefaultGaugeValueDropRatio float32 = 0.45

	// DefaultGaugeMajorTickPx is the length of a labelled
	// graduation, measured outward from the arc.
	DefaultGaugeMajorTickPx float32 = 6

	// DefaultGaugeMinorTickPx is the length of an unlabelled
	// graduation between two major ones.
	DefaultGaugeMinorTickPx float32 = 3

	// DefaultGaugeTickGapPx is the clearance between the arc and
	// the graduations, and between a graduation and its label.
	DefaultGaugeTickGapPx float32 = 3

	// DefaultCandleWidthRatio is the fraction of the slot width used
	// for the candle body when CandleWidth is 0.
	DefaultCandleWidthRatio float32 = 0.6

	// DefaultBoxWidthRatio is the fraction of the slot width used
	// for the box body when BoxWidth is 0.
	DefaultBoxWidthRatio float32 = 0.5

	// DefaultOutlierRadius is the radius of outlier dots when
	// OutlierRadius is 0.
	DefaultOutlierRadius float32 = 3

	// DefaultWaterfallWidthRatio is the fraction of the slot
	// width used for bars when BarWidth is 0.
	DefaultWaterfallWidthRatio float32 = 0.6

	// DefaultConnectorWidth is the line width for waterfall
	// connector lines between bars.
	DefaultConnectorWidth float32 = 1

	// DefaultAnnotationLineWidth is the line width for
	// annotation reference lines when Width is 0.
	DefaultAnnotationLineWidth float32 = 1.5

	// DefaultBubbleMinRadius is the minimum bubble marker
	// radius in pixels.
	DefaultBubbleMinRadius float32 = 4

	// DefaultBubbleMaxRadius is the maximum bubble marker
	// radius in pixels.
	DefaultBubbleMaxRadius float32 = 30

	// DefaultHeatmapCellGap is the gap in pixels between
	// heatmap cells.
	DefaultHeatmapCellGap float32 = 1

	// DefaultFunnelSegmentGap is the vertical gap in pixels
	// between funnel segments.
	DefaultFunnelSegmentGap float32 = 4

	// DefaultSankeyNodeWidth is the width in pixels of node
	// rectangles in Sankey diagrams.
	DefaultSankeyNodeWidth float32 = 20

	// DefaultSankeyNodeGap is the vertical gap in pixels
	// between nodes in the same column.
	DefaultSankeyNodeGap float32 = 10

	// DefaultSankeyLinkAlpha is the alpha channel value for
	// link ribbons (semi-transparent).
	DefaultSankeyLinkAlpha uint8 = 128

	// DefaultTreemapCellGap is the gap in pixels between
	// treemap cells.
	DefaultTreemapCellGap float32 = 2

	// DefaultTreemapHeaderHeight is the height in pixels of
	// group header bars in treemap charts.
	DefaultTreemapHeaderHeight float32 = 18

	// DefaultZoomFactor is the scale factor per scroll tick.
	// >1 zooms in; 1/factor zooms out.
	DefaultZoomFactor float64 = 1.15

	// DefaultMinDragPx is the minimum pixel distance to
	// distinguish a drag from a click.
	DefaultMinDragPx float32 = 4

	// DefaultMinZoomRange is the smallest allowed domain span
	// to prevent degenerate transforms.
	DefaultMinZoomRange float64 = 1e-12

	// zoomDoubleClickMs is the wall-clock threshold in
	// milliseconds for detecting a mouse double-click.
	zoomDoubleClickMs int64 = 400

	// DefaultSparklineHeight is the default height in pixels
	// for sparkline charts.
	DefaultSparklineHeight float32 = 20

	// DefaultSparklineLineWidth is the default stroke width
	// for sparkline line/area variants.
	DefaultSparklineLineWidth float32 = 1.5

	// DefaultSparklineMarkerRadius is the default radius for
	// min/max/last marker dots on sparklines.
	DefaultSparklineMarkerRadius float32 = 2.5

	// DefaultSparklineBarGap is the gap in pixels between
	// sparkline bars.
	DefaultSparklineBarGap float32 = 1

	// DefaultFPSBudgetMs is the frame time threshold in
	// milliseconds above which detail reduction kicks in.
	DefaultFPSBudgetMs float32 = 20

	// fpsEMAAlpha is the exponential moving average smoothing
	// factor for frame time tracking.
	fpsEMAAlpha float32 = 0.3
)

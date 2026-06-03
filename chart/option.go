package chart

import (
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// Option applies a configuration override to a BaseCfg. Works
// with any chart Cfg type that embeds BaseCfg.
type Option func(*BaseCfg)

// WithID sets the chart ID.
func WithID(id string) Option {
	return func(b *BaseCfg) { b.ID = id }
}

// WithTitle sets the chart title.
func WithTitle(title string) Option {
	return func(b *BaseCfg) { b.Title = title }
}

// WithSize sets explicit width and height.
func WithSize(w, h float32) Option {
	return func(b *BaseCfg) { b.Width = w; b.Height = h }
}

// WithTheme sets the chart theme.
func WithTheme(t *theme.Theme) Option {
	return func(b *BaseCfg) { b.Theme = t }
}

// WithSizing sets the sizing policy.
func WithSizing(s gui.Sizing) Option {
	return func(b *BaseCfg) { b.Sizing = s }
}

// WithXTickRotation sets X-axis tick label rotation (radians).
func WithXTickRotation(radians float32) Option {
	return func(b *BaseCfg) { b.XTickRotation = radians }
}

// WithLegendPosition overrides the theme legend position.
func WithLegendPosition(pos theme.LegendPosition) Option {
	return func(b *BaseCfg) { b.LegendPosition = &pos }
}

// Legend position constants re-exported for convenience.
const (
	LegendTopRight    = theme.LegendTopRight
	LegendTopLeft     = theme.LegendTopLeft
	LegendBottomRight = theme.LegendBottomRight
	LegendBottomLeft  = theme.LegendBottomLeft
	LegendNone        = theme.LegendNone
	LegendBottom      = theme.LegendBottom
	LegendRight       = theme.LegendRight
	LegendTop         = theme.LegendTop
)

// Apply applies options to a BaseCfg. Chart constructors can
// call cfg.Apply(opts...) to process options.
func (b *BaseCfg) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(b)
	}
}

// LineOption applies a configuration override to a LineCfg.
type LineOption func(*LineCfg)

// WithLineWidth returns a LineOption that sets line width.
func WithLineWidth(w float32) LineOption {
	return func(c *LineCfg) { c.LineWidth = w }
}

// WithMarkers returns a LineOption that enables data markers.
func WithMarkers() LineOption {
	return func(c *LineCfg) { c.ShowMarkers = true }
}

// WithArea returns a LineOption that enables filled area.
func WithArea() LineOption {
	return func(c *LineCfg) { c.ShowArea = true }
}

// LineWith creates a line chart, applying options to the Cfg.
func LineWith(cfg LineCfg, opts ...LineOption) gui.View {
	for _, opt := range opts {
		opt(&cfg)
	}
	return Line(cfg)
}

// BarOption applies a configuration override to a BarCfg.
type BarOption func(*BarCfg)

// WithBarWidth returns a BarOption that sets bar width.
func WithBarWidth(w float32) BarOption {
	return func(c *BarCfg) { c.BarWidth = w }
}

// WithBarGap returns a BarOption that sets gap between bars.
func WithBarGap(g float32) BarOption {
	return func(c *BarCfg) { c.BarGap = g }
}

// BarWith creates a bar chart, applying options to the Cfg.
func BarWith(cfg BarCfg, opts ...BarOption) gui.View {
	for _, opt := range opts {
		opt(&cfg)
	}
	return Bar(cfg)
}

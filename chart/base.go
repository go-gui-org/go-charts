package chart

import (
	"errors"
	"strings"
	"time"

	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// BaseCfg contains fields common to all chart configuration
// structs. Embed it to inherit ID, Title, sizing, theme,
// interaction callbacks, and version tracking.
type BaseCfg struct {
	ID     string
	Title  string
	Sizing gui.Sizing
	Width  float32
	Height float32

	Theme        *theme.Theme
	OnClick      func(*gui.Layout, *gui.Event, *gui.Window)
	OnHover      func(*gui.Layout, *gui.Event, *gui.Window)
	OnMouseLeave func(*gui.Layout, *gui.Event, *gui.Window)

	// XTickRotation rotates X-axis tick labels (radians).
	// 0 = horizontal.
	XTickRotation float32

	// LegendPosition overrides the theme legend position for
	// this chart. nil = use theme default.
	LegendPosition *theme.LegendPosition

	// Annotations adds reference lines, text labels, and shaded
	// regions to the chart. Ignored by pie and gauge charts.
	Annotations Annotations

	// Animate enables entry animation on first render. Series
	// draw in progressively over DefaultAnimDuration.
	Animate bool

	// ShowDataTable replaces the chart with an accessible data
	// table showing all series values in tabular form.
	ShowDataTable bool

	// AnimDuration overrides the default entry animation
	// duration. Zero uses DefaultAnimDuration (500ms).
	AnimDuration time.Duration

	Version uint64
}

// InteractionCfg holds XY-chart-specific interaction flags.
// Embed in XY chart configs (Line, Bar, Area, Scatter, etc.)
// alongside BaseCfg. Non-XY charts (Pie, Gauge, Funnel, …)
// do not embed this.
type InteractionCfg struct {
	// EnableZoom enables scroll-wheel zoom on the chart axes.
	EnableZoom bool
	// EnablePan enables LMB-drag panning of the chart axes.
	EnablePan bool
	// EnableRangeSelect enables shift+LMB brush-to-zoom.
	EnableRangeSelect bool
	// AnimateTransitions enables smooth transitions when data
	// changes (Version bump). Old values interpolate to new
	// over DefaultTransitionDuration.
	AnimateTransitions bool
}

// applyDefaults sets sensible zero-value defaults.
func (b *BaseCfg) applyDefaults() {
	if b.Sizing == (gui.Sizing{}) {
		b.Sizing = gui.FillFill
	}
	if b.Theme == nil {
		b.Theme = theme.Default()
	}
}

// Validate checks BaseCfg for invalid settings. Returns nil
// when valid.
func (b *BaseCfg) Validate() error {
	var errs []string
	if b.Width < 0 {
		errs = append(errs, "negative Width")
	}
	if b.Height < 0 {
		errs = append(errs, "negative Height")
	}
	return buildError("chart", errs)
}

// buildError joins error strings with a prefix, returning nil
// when empty.
func buildError(prefix string, errs []string) error {
	if len(errs) == 0 {
		return nil
	}
	return errors.New(prefix + ": " + strings.Join(errs, "; "))
}

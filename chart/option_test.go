package chart

import (
	"testing"

	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

func TestWithID(t *testing.T) {
	var b BaseCfg
	WithID("my-id")(&b)
	if b.ID != "my-id" {
		t.Errorf("got %q, want %q", b.ID, "my-id")
	}
}

func TestWithTitle(t *testing.T) {
	var b BaseCfg
	WithTitle("chart title")(&b)
	if b.Title != "chart title" {
		t.Errorf("got %q, want %q", b.Title, "chart title")
	}
}

func TestWithSize(t *testing.T) {
	var b BaseCfg
	WithSize(800, 600)(&b)
	if b.Width != 800 || b.Height != 600 {
		t.Errorf("got %vx%v, want 800x600", b.Width, b.Height)
	}
}

func TestWithTheme(t *testing.T) {
	var b BaseCfg
	th := theme.Default()
	WithTheme(th)(&b)
	if b.Theme != th {
		t.Error("theme not set")
	}
}

func TestWithSizing(t *testing.T) {
	var b BaseCfg
	WithSizing(gui.FillFill)(&b)
	if b.Sizing != gui.FillFill {
		t.Errorf("got %v, want FillFill", b.Sizing)
	}
}

func TestWithXTickRotation(t *testing.T) {
	var b BaseCfg
	WithXTickRotation(0.5)(&b)
	if b.XTickRotation != 0.5 {
		t.Errorf("got %v, want 0.5", b.XTickRotation)
	}
}

func TestWithLegendPosition(t *testing.T) {
	var b BaseCfg
	WithLegendPosition(LegendTopLeft)(&b)
	if b.LegendPosition == nil || *b.LegendPosition != LegendTopLeft {
		t.Error("legend position not set correctly")
	}
}

func TestApply(t *testing.T) {
	var b BaseCfg
	b.Apply(WithID("x"), WithTitle("y"))
	if b.ID != "x" || b.Title != "y" {
		t.Errorf("Apply failed: ID=%q Title=%q", b.ID, b.Title)
	}
}

func TestWithLineWidth(t *testing.T) {
	var c LineCfg
	WithLineWidth(3)(&c)
	if c.LineWidth != 3 {
		t.Errorf("got %v, want 3", c.LineWidth)
	}
}

func TestWithMarkers(t *testing.T) {
	var c LineCfg
	WithMarkers()(&c)
	if !c.ShowMarkers {
		t.Error("ShowMarkers not set")
	}
}

func TestWithArea(t *testing.T) {
	var c LineCfg
	WithArea()(&c)
	if !c.ShowArea {
		t.Error("ShowArea not set")
	}
}

func TestLineWith(t *testing.T) {
	s := series.XYFromYValues("s", []float64{1, 2, 3})
	cfg := LineCfg{Series: []series.XY{s}}
	v := LineWith(cfg, WithLineWidth(4), WithMarkers())
	if v == nil {
		t.Fatal("LineWith returned nil")
	}
}

func TestWithBarWidth(t *testing.T) {
	var c BarCfg
	WithBarWidth(20)(&c)
	if c.BarWidth != 20 {
		t.Errorf("got %v, want 20", c.BarWidth)
	}
}

func TestWithBarGap(t *testing.T) {
	var c BarCfg
	WithBarGap(8)(&c)
	if c.BarGap != 8 {
		t.Errorf("got %v, want 8", c.BarGap)
	}
}

func TestBarWith(t *testing.T) {
	s := series.CategoryFromMap("s", map[string]float64{"a": 1, "b": 2})
	cfg := BarCfg{Series: []series.Category{s}}
	v := BarWith(cfg, WithBarWidth(15), WithBarGap(5))
	if v == nil {
		t.Fatal("BarWith returned nil")
	}
}

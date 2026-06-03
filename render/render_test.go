package render

import (
	"testing"

	"github.com/go-gui-org/go-glyph"
	"github.com/go-gui-org/go-gui/gui"
)

// stubMeasurer satisfies gui.TextMeasurer with zero returns.
type stubMeasurer struct{}

func (stubMeasurer) TextWidth(_ string, _ gui.TextStyle) float32  { return 0 }
func (stubMeasurer) TextHeight(_ string, _ gui.TextStyle) float32 { return 0 }
func (stubMeasurer) FontHeight(_ gui.TextStyle) float32           { return 0 }
func (stubMeasurer) FontAscent(_ gui.TextStyle) float32           { return 0 }
func (stubMeasurer) LayoutText(_ string, _ gui.TextStyle, _ float32) (glyph.Layout, error) {
	return glyph.Layout{}, nil
}

func newTestContext(w, h float32) *Context {
	dc := gui.NewDrawContext(w, h, stubMeasurer{})
	return NewContext(dc)
}

func TestNewContext(t *testing.T) {
	c := newTestContext(200, 100)
	if c == nil {
		t.Fatal("NewContext returned nil")
	}
}

func TestWidthHeight(t *testing.T) {
	c := newTestContext(320, 240)
	if c.Width() != 320 {
		t.Errorf("Width: got %v, want 320", c.Width())
	}
	if c.Height() != 240 {
		t.Errorf("Height: got %v, want 240", c.Height())
	}
}

func TestDrawingMethodsDoNotPanic(t *testing.T) {
	c := newTestContext(400, 300)
	col := gui.Hex(0xFF0000)

	c.Line(0, 0, 10, 10, col, 1)
	c.Polyline([]float32{0, 0, 10, 10}, col, 1)
	c.FilledRect(0, 0, 50, 50, col)
	c.Rect(0, 0, 50, 50, col, 1)
	c.FilledCircle(50, 50, 10, col)
	c.Circle(50, 50, 10, col, 1)
	c.FilledArc(50, 50, 20, 20, 0, 1, col)
	c.Arc(50, 50, 20, 20, 0, 1, col, 1)
	c.FilledPolygon([]float32{0, 0, 10, 0, 5, 10}, col)
	c.FilledRoundedRect(0, 0, 50, 50, 5, col)
	c.RoundedRect(0, 0, 50, 50, 5, col, 1)
	c.DashedLine(0, 0, 100, 0, col, 1, 5, 3)
	c.DashedPolyline([]float32{0, 0, 100, 0, 100, 100}, col, 1, 5, 3)
	c.PolylineJoined([]float32{0, 0, 10, 10, 20, 0}, col, 1)
}

func TestTextMethodsDoNotPanic(t *testing.T) {
	c := newTestContext(400, 300)
	style := gui.TextStyle{}

	c.Text(10, 10, "hello", style)
	_ = c.TextWidth("hello", style)
	_ = c.FontHeight(style)
}

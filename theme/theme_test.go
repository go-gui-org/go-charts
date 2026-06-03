package theme

import (
	"testing"

	"github.com/go-gui-org/go-gui/gui"
)

func TestHighContrastLength(t *testing.T) {
	got := HighContrast()
	if len(got) != 10 {
		t.Errorf("len = %d, want 10", len(got))
	}
}

func TestHighContrastClone(t *testing.T) {
	a := HighContrast()
	b := HighContrast()
	a[0] = gui.Hex(0xFFFFFF)
	if a[0] == b[0] {
		t.Error("HighContrast did not clone")
	}
}

func TestHighContrastDistinct(t *testing.T) {
	colors := HighContrast()
	for i := range colors {
		for j := i + 1; j < len(colors); j++ {
			if colors[i] == colors[j] {
				t.Errorf("duplicate at index %d and %d", i, j)
			}
		}
	}
}

func TestHighContrastTheme(t *testing.T) {
	th := HighContrastTheme()
	if th.AxisWidth < 2 {
		t.Errorf("AxisWidth = %v, want >= 2", th.AxisWidth)
	}
	if th.GridWidth < 1 {
		t.Errorf("GridWidth = %v, want >= 1", th.GridWidth)
	}
	if len(th.Palette) != 10 {
		t.Errorf("Palette len = %d, want 10", len(th.Palette))
	}
	if th.PaddingBottom != DefaultPaddingBottom+10 {
		t.Errorf("PaddingBottom = %v, want %v",
			th.PaddingBottom, DefaultPaddingBottom+10)
	}
	if th.PaddingLeft != DefaultPaddingLeft+10 {
		t.Errorf("PaddingLeft = %v, want %v",
			th.PaddingLeft, DefaultPaddingLeft+10)
	}
	if th.TickMark.Length != 8 {
		t.Errorf("TickMark.Length = %v, want 8", th.TickMark.Length)
	}
}

package series

import (
	"testing"

	"github.com/go-gui-org/go-gui/gui"
)

func TestXYBounds(t *testing.T) {
	s := NewXY(XYCfg{
		Name:  "test",
		Color: gui.Blue,
		Points: []Point{
			{X: 1, Y: 10},
			{X: 5, Y: 2},
			{X: 3, Y: 8},
		},
	})
	minX, maxX, minY, maxY := s.Bounds()
	if minX != 1 || maxX != 5 || minY != 2 || maxY != 10 {
		t.Errorf("Bounds = (%v,%v,%v,%v), want (1,5,2,10)",
			minX, maxX, minY, maxY)
	}
}

func TestXYEmpty(t *testing.T) {
	s := NewXY(XYCfg{Name: "empty"})
	if s.Len() != 0 {
		t.Errorf("Len = %d, want 0", s.Len())
	}
}

func TestXYFromSlices(t *testing.T) {
	s, err := XYFromSlices("test", []float64{1, 2, 3}, []float64{10, 20, 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name() != "test" {
		t.Errorf("Name = %q, want %q", s.Name(), "test")
	}
	if s.Len() != 3 {
		t.Fatalf("Len = %d, want 3", s.Len())
	}
	if s.Points[1].X != 2 || s.Points[1].Y != 20 {
		t.Errorf("Points[1] = %v, want {2, 20}", s.Points[1])
	}
}

func TestXYFromSlicesEmpty(t *testing.T) {
	s, err := XYFromSlices("empty", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Len() != 0 {
		t.Errorf("Len = %d, want 0", s.Len())
	}
}

func TestXYFromSlicesError(t *testing.T) {
	_, err := XYFromSlices("bad", []float64{1}, []float64{1, 2})
	if err == nil {
		t.Error("expected error on mismatched lengths")
	}
}

func TestXYFromYValues(t *testing.T) {
	s := XYFromYValues("auto", []float64{10, 20, 30})
	if s.Len() != 3 {
		t.Fatalf("Len = %d, want 3", s.Len())
	}
	for i, p := range s.Points {
		if p.X != float64(i) {
			t.Errorf("Points[%d].X = %v, want %v", i, p.X, float64(i))
		}
	}
	if s.Points[2].Y != 30 {
		t.Errorf("Points[2].Y = %v, want 30", s.Points[2].Y)
	}
}

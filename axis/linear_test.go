package axis

import (
	"math"
	"testing"
)

func TestLinearTicks(t *testing.T) {
	tests := []struct {
		name     string
		min, max float64
		wantMin  int // minimum tick count
		wantMax  int // maximum tick count
	}{
		{"normal range", 0, 100, 3, 12},
		{"negative range", -50, 50, 3, 12},
		{"small range", 1.0, 1.1, 2, 12},
		{"single value", 5, 5, 0, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewLinear(LinearCfg{Min: tt.min, Max: tt.max})
			ticks := a.Ticks(0, 500)
			if len(ticks) < tt.wantMin || len(ticks) > tt.wantMax {
				t.Errorf("Ticks count = %d, want [%d, %d]",
					len(ticks), tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestLinearTicksAutoRange(t *testing.T) {
	a := NewLinear(LinearCfg{Min: 1, Max: 9, AutoRange: true})
	ticks := a.Ticks(0, 500)
	if len(ticks) == 0 {
		t.Fatal("expected ticks, got none")
	}
	// Auto-range should expand domain to nice boundaries.
	dMin, dMax := a.Domain()
	if dMin > 1 || dMax < 9 {
		t.Errorf("auto-range domain [%v, %v] did not expand to cover [1, 9]",
			dMin, dMax)
	}
}

func TestLinearTicksOverrideDomain(t *testing.T) {
	a := NewLinear(LinearCfg{Min: 0, Max: 100, AutoRange: true})
	a.SetOverrideDomain(true)
	a.SetRange(20, 60)
	ticks := a.Ticks(0, 500)
	for _, tk := range ticks {
		if tk.Value < 20-1e-9 || tk.Value > 60+1e-9 {
			t.Errorf("tick %v outside override domain [20, 60]", tk.Value)
		}
	}
}

func TestLinearTransformInvertRoundTrip(t *testing.T) {
	tests := []struct {
		name           string
		min, max       float64
		value          float64
		pixMin, pixMax float32
	}{
		{"mid-range", 0, 100, 50, 0, 400},
		{"negative domain", -100, 100, 0, 0, 400},
		{"at min", 10, 90, 10, 50, 250},
		{"at max", 10, 90, 90, 50, 250},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewLinear(LinearCfg{Min: tt.min, Max: tt.max})
			px := a.Transform(tt.value, tt.pixMin, tt.pixMax)
			got := a.Invert(px, tt.pixMin, tt.pixMax)
			if math.Abs(got-tt.value) > 1e-9 {
				t.Errorf("round-trip: Transform(%v) = %v, Invert = %v, want %v",
					tt.value, px, got, tt.value)
			}
		})
	}
}

func TestLinearTickPositions(t *testing.T) {
	a := NewLinear(LinearCfg{Min: 0, Max: 100})
	ticks := a.Ticks(0, 100)
	for _, tk := range ticks {
		if tk.Position < -1e-3 || tk.Position > 100+1e-3 {
			t.Errorf("tick position %v outside pixel range [0, 100]",
				tk.Position)
		}
	}
}

func TestLinearTickTarget(t *testing.T) {
	// A short plot asks for fewer ticks. The count is approximate —
	// ticks land on round numbers — so this asserts the direction, not
	// an exact number.
	a := NewLinear(LinearCfg{Min: 0, Max: 100})
	dense := len(a.Ticks(0, 200))

	a = NewLinear(LinearCfg{Min: 0, Max: 100, TickTarget: 3})
	sparse := len(a.Ticks(0, 200))

	if sparse >= dense {
		t.Errorf("TickTarget ignored: %d ticks vs %d at the default", sparse, dense)
	}
	if sparse < 2 {
		t.Errorf("an axis must keep its endpoints: %d ticks", sparse)
	}
}

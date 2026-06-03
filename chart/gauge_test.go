package chart

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-gui/gui"
)

func TestGaugeCfgValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     GaugeCfg
		wantErr bool
	}{
		{
			name: "valid defaults",
			cfg: GaugeCfg{
				Value: 50,
			},
			wantErr: false,
		},
		{
			name: "min >= max",
			cfg: GaugeCfg{
				Value: 50,
				Min:   100,
				Max:   0,
			},
			wantErr: true,
		},
		{
			name: "arc angle zero",
			cfg: GaugeCfg{
				Value:    50,
				ArcAngle: -1,
			},
			wantErr: true,
		},
		{
			name: "arc angle > 2pi",
			cfg: GaugeCfg{
				Value:    50,
				ArcAngle: 2*math.Pi + 0.1,
			},
			wantErr: true,
		},
		{
			name: "inner ratio out of range",
			cfg: GaugeCfg{
				Value:      50,
				InnerRatio: 1.5,
			},
			wantErr: true,
		},
		{
			name: "valid zones",
			cfg: GaugeCfg{
				Value: 75,
				Zones: []GaugeZone{
					{Threshold: 30, Color: gui.Hex(0x00FF00)},
					{Threshold: 70, Color: gui.Hex(0xFFFF00)},
					{Threshold: 100, Color: gui.Hex(0xFF0000)},
				},
			},
			wantErr: false,
		},
		{
			name: "zone threshold not ascending",
			cfg: GaugeCfg{
				Value: 50,
				Zones: []GaugeZone{
					{Threshold: 70, Color: gui.Hex(0x00FF00)},
					{Threshold: 30, Color: gui.Hex(0xFF0000)},
				},
			},
			wantErr: true,
		},
		{
			name: "zone threshold exceeds max",
			cfg: GaugeCfg{
				Value: 50,
				Zones: []GaugeZone{
					{Threshold: 150, Color: gui.Hex(0xFF0000)},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.applyGaugeDefaults()
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}

func TestGaugeValueFraction(t *testing.T) {
	tests := []struct {
		value, min, max float64
		want            float64
	}{
		{50, 0, 100, 0.5},
		{0, 0, 100, 0},
		{100, 0, 100, 1},
		{-10, 0, 100, 0},                   // clamped
		{200, 0, 100, 1},                   // clamped
		{50, 50, 50, 0},                    // degenerate
		{math.NaN(), 0, 100, 0},            // NaN → 0
		{math.Inf(1), 0, 100, 1},           // +Inf → clamped
		{math.Inf(-1), 0, 100, 0},          // -Inf → clamped
		{50, math.NaN(), 100, 0},           // NaN min → 0
		{50, 0, math.NaN(), 0},             // NaN max → 0
		{50, math.NaN(), math.NaN(), 0},    // NaN both → 0
		{50, math.Inf(-1), math.Inf(1), 0}, // Inf bounds → 0
	}
	for _, tt := range tests {
		got := gaugeValueFraction(tt.value, tt.min, tt.max)
		if math.IsNaN(got) || math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("gaugeValueFraction(%g, %g, %g) = %g, want %g",
				tt.value, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestGaugeStartAngle(t *testing.T) {
	// 270° arc: gap of 90° centered at bottom (π/2).
	// Start should be at π/2 + 45° = 3π/4 ≈ 2.356
	got := gaugeStartAngle(3 * math.Pi / 2)
	want := float32(math.Pi/2 + math.Pi/4)
	if math.Abs(float64(got-want)) > 1e-5 {
		t.Errorf("gaugeStartAngle(270°) = %g, want %g", got, want)
	}
}

func TestGaugeConstructor(t *testing.T) {
	v := Gauge(GaugeCfg{
		BaseCfg: BaseCfg{ID: "g1"},
		Value:   42,
	})
	if v == nil {
		t.Fatal("Gauge returned nil")
	}
	gv, ok := v.(*gaugeView)
	if !ok {
		t.Fatal("Gauge did not return *gaugeView")
	}
	if gv.cfg.Max != 100 {
		t.Errorf("default Max = %g, want 100", gv.cfg.Max)
	}
	if gv.cfg.ArcAngle == 0 {
		t.Error("default ArcAngle not set")
	}
}

func TestGaugeConstructor_ShowPointer(t *testing.T) {
	v := Gauge(GaugeCfg{
		BaseCfg:     BaseCfg{ID: "gp1"},
		Value:       42,
		ShowPointer: true,
	})
	gv, ok := v.(*gaugeView)
	if !ok {
		t.Fatal("Gauge did not return *gaugeView")
	}
	if !gv.cfg.ShowPointer {
		t.Error("ShowPointer not preserved")
	}
}

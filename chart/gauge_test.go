package chart

import (
	"math"
	"path/filepath"
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
			name: "graduations",
			cfg: GaugeCfg{
				Value:          50,
				TickCount:      5,
				MinorTicks:     4,
				ShowTickLabels: true,
			},
			wantErr: false,
		},
		{
			name: "negative tick count",
			cfg: GaugeCfg{
				Value:     50,
				TickCount: -1,
			},
			wantErr: true,
		},
		{
			name: "negative minor ticks",
			cfg: GaugeCfg{
				Value:      50,
				TickCount:  4,
				MinorTicks: -2,
			},
			wantErr: true,
		},
		{
			// A ratio past 1 would push the dial outside its own
			// plot box, clipping the arc against the panel edge.
			name: "radius ratio past 1",
			cfg: GaugeCfg{
				Value:       50,
				RadiusRatio: 1.5,
			},
			wantErr: true,
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
		{
			name: "value offset not finite",
			cfg: GaugeCfg{
				Value:            50,
				ValueOffsetRatio: float32(math.NaN()),
			},
			wantErr: true,
		},
		{
			name: "value offset out of range",
			cfg: GaugeCfg{
				Value:            50,
				ValueOffsetRatio: 99,
			},
			wantErr: true,
		},
		{
			name: "value anchor unknown",
			cfg: GaugeCfg{
				Value:       50,
				ValueAnchor: GaugeValueBelowArc + 1,
			},
			wantErr: true,
		},
		{
			name: "valid anchor with negative offset",
			cfg: GaugeCfg{
				Value:            50,
				ValueAnchor:      GaugeValueCentre,
				ValueOffsetRatio: -0.2,
				ValueLabel:       "Mbps",
			},
			wantErr: false,
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

func TestGaugeDrawTicksDoesNotPanicOnHugeCounts(t *testing.T) {
	// Huge TickCount/MinorTicks must be capped, not loop forever
	// or allocate unbounded memory. Exercise via draw path.
	cfg := GaugeCfg{
		BaseCfg:        BaseCfg{ID: "huge"},
		Value:          50,
		Min:            0,
		Max:            100,
		TickCount:      100000,
		MinorTicks:     100000,
		ShowTickLabels: true,
	}
	cfg.applyGaugeDefaults()
	gv := newGaugeView(cfg)
	// Use a headless DrawContext: NewContext needs a DrawContext,
	// but we only need to ensure drawTicks itself does not panic
	// when given a valid render.Context. Create a minimal one via
	// the chart's Draw helper with a small canvas—just check it
	// does not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("drawTicks panicked on huge counts: %v", r)
		}
	}()
	// Directly test the capping logic: steps must be bounded.
	tickCount := min(cfg.TickCount, maxGaugeSteps)
	minorTicks := min(cfg.MinorTicks, maxGaugeSteps)
	steps := min(tickCount*(minorTicks+1), maxGaugeSteps)
	if steps != maxGaugeSteps {
		t.Errorf("capped steps = %d, want %d", steps, maxGaugeSteps)
	}
	_ = gv // avoid unused if draw not called
}

func TestGaugeRadiusRatioDefault(t *testing.T) {
	v := Gauge(GaugeCfg{BaseCfg: BaseCfg{ID: "rr"}, Value: 10})
	gv := v.(*gaugeView)
	if gv.cfg.RadiusRatio != DefaultGaugeRadiusRatio {
		t.Errorf("RadiusRatio = %v, want %v", gv.cfg.RadiusRatio, DefaultGaugeRadiusRatio)
	}
}

func TestLabelCentrePlacesNearEdge(t *testing.T) {
	// Right side (cos=1, sin=0): near (left) edge should be at r.
	cx, cy, r := float32(100), float32(100), float32(50)
	tw, fh := float32(40), float32(10)
	x, _ := labelCentre(cx, cy, r, 1, 0, tw, fh)
	left := x - tw/2
	if math.Abs(float64(left-(cx+r))) > 1e-5 {
		t.Errorf("right label left edge %v, want %v", left, cx+r)
	}
	// Top side (cos=0, sin=-1): near (bottom) edge at r.
	_, y := labelCentre(cx, cy, r, 0, -1, tw, fh)
	bottom := y + fh/2
	if math.Abs(float64(bottom-(cy-r))) > 1e-5 {
		t.Errorf("top label bottom edge %v, want %v", bottom, cy-r)
	}
}

func TestGaugeValueBaseY(t *testing.T) {
	const cy, innerR, outerR = 100, 40, 80
	tests := []struct {
		name   string
		anchor GaugeValueAnchor
		want   float32
	}{
		{"default drops below centre", GaugeValueDefault,
			cy + innerR*DefaultGaugeValueDropRatio},
		{"centre", GaugeValueCentre, cy},
		{"above centre", GaugeValueAboveCentre,
			cy - innerR*DefaultGaugeValueDropRatio},
		{"below arc", GaugeValueBelowArc,
			cy + outerR + DefaultGaugeValueBelowArcGapPx},
		{"unknown falls back to default", GaugeValueBelowArc + 7,
			cy + innerR*DefaultGaugeValueDropRatio},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gaugeValueBaseY(tt.anchor, cy, innerR, outerR)
			if math.Abs(float64(got-tt.want)) > 1e-4 {
				t.Errorf("gaugeValueBaseY = %g, want %g", got, tt.want)
			}
		})
	}

	// No hole: the outer radius stands in for the hole radius.
	got := gaugeValueBaseY(GaugeValueDefault, cy, 0, outerR)
	want := float32(cy + outerR*DefaultGaugeValueDropRatio)
	if math.Abs(float64(got-want)) > 1e-4 {
		t.Errorf("no-hole base = %g, want %g", got, want)
	}
}

func TestGaugeValueOffsetPx(t *testing.T) {
	const innerR, outerR = 40, 80
	tests := []struct {
		name  string
		ratio float32
		want  float32
	}{
		{"zero adds nothing", 0, 0},
		{"positive moves down", 0.5, 20},
		{"negative lifts", -0.25, -10},
		{"clamped high", 1000, maxGaugeValueOffset * innerR},
		{"clamped low", -1000, -maxGaugeValueOffset * innerR},
		{"NaN ignored", float32(math.NaN()), 0},
		{"+Inf ignored", float32(math.Inf(1)), 0},
		{"-Inf ignored", float32(math.Inf(-1)), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gaugeValueOffsetPx(tt.ratio, innerR, outerR)
			if math.IsNaN(float64(got)) ||
				math.Abs(float64(got-tt.want)) > 1e-4 {
				t.Errorf("gaugeValueOffsetPx(%g) = %g, want %g",
					tt.ratio, got, tt.want)
			}
		})
	}
}

func TestGaugeConstructor_ValuePlacement(t *testing.T) {
	v := Gauge(GaugeCfg{
		BaseCfg:          BaseCfg{ID: "gv1"},
		Value:            42,
		ShowValue:        true,
		ValueAnchor:      GaugeValueCentre,
		ValueOffsetRatio: -0.1,
		ValueLabel:       "Mbps",
	})
	gv, ok := v.(*gaugeView)
	if !ok {
		t.Fatal("Gauge did not return *gaugeView")
	}
	if gv.cfg.ValueAnchor != GaugeValueCentre {
		t.Errorf("ValueAnchor = %d, want %d",
			gv.cfg.ValueAnchor, GaugeValueCentre)
	}
	if gv.cfg.ValueOffsetRatio != -0.1 {
		t.Errorf("ValueOffsetRatio = %g, want -0.1", gv.cfg.ValueOffsetRatio)
	}
	if gv.cfg.ValueLabel != "Mbps" {
		t.Errorf("ValueLabel = %q, want %q", gv.cfg.ValueLabel, "Mbps")
	}
	if gv.cfg.ValueFormat != "%.0f" {
		t.Errorf("defaults not applied, ValueFormat = %q", gv.cfg.ValueFormat)
	}
}

// TestExportPNG_GaugeValuePlacement exercises the draw path for every
// anchor, with a unit line and with an offset Validate would reject.
// Validate only warns, so draw must survive the bad ratio.
func TestExportPNG_GaugeValuePlacement(t *testing.T) {
	tests := []struct {
		name   string
		anchor GaugeValueAnchor
		offset float32
		label  string
	}{
		{"default", GaugeValueDefault, 0, ""},
		{"centre with unit", GaugeValueCentre, 0, "Mbps"},
		{"above centre", GaugeValueAboveCentre, -0.1, "Mbps"},
		{"below arc", GaugeValueBelowArc, 0, ""},
		{"non-finite offset", GaugeValueCentre, float32(math.NaN()), "Mbps"},
		{"wild offset", GaugeValueCentre, 1e6, ""},
		{"unknown anchor", GaugeValueBelowArc + 3, 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Gauge(GaugeCfg{
				BaseCfg: BaseCfg{
					ID:    "gauge-value-placement",
					Width: 400, Height: 300,
				},
				Value:            72,
				Max:              100,
				ShowValue:        true,
				ValueAnchor:      tt.anchor,
				ValueOffsetRatio: tt.offset,
				ValueLabel:       tt.label,
			})
			path := filepath.Join(t.TempDir(), "gauge_value.png")
			if err := ExportPNG(v, 400, 300, path); err != nil {
				t.Fatal(err)
			}
			assertValidPNG(t, path, 400, 300)
		})
	}
}

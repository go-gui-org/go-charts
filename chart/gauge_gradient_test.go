package chart

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// zoneCfg builds a three-zone gauge with explicit colors, so a test
// can assert on the exact stop colors rather than palette lookups.
func zoneCfg(gradient bool) GaugeCfg {
	cfg := GaugeCfg{
		BaseCfg: BaseCfg{ID: "grad"},
		Value:   50,
		Min:     0,
		Max:     100,
		Zones: []GaugeZone{
			{Label: "slow", Threshold: 30, Color: gui.Hex(0xE03131)},
			{Label: "fair", Threshold: 70, Color: gui.Hex(0xF59F00)},
			{Label: "fast", Threshold: 100, Color: gui.Hex(0x2F9E44)},
		},
		GradientZones: gradient,
	}
	cfg.applyGaugeDefaults()
	return cfg
}

func TestGaugeGradientStopsOffByDefault(t *testing.T) {
	cfg := zoneCfg(false)
	if got := gaugeGradientStops(&cfg, cfg.Theme); got != nil {
		t.Errorf("stops = %v, want nil when GradientZones is off", got)
	}
}

func TestGaugeGradientStopsZoneMidpoints(t *testing.T) {
	cfg := zoneCfg(true)
	stops := gaugeGradientStops(&cfg, cfg.Theme)
	if len(stops) != 3 {
		t.Fatalf("len(stops) = %d, want 3", len(stops))
	}
	// Ends pinned; middle zone (30..70) anchored at its midpoint 0.5.
	wantPos := []float32{0, 0.5, 1}
	for i, want := range wantPos {
		if math.Abs(float64(stops[i].Pos-want)) > 1e-6 {
			t.Errorf("stop %d Pos = %v, want %v", i, stops[i].Pos, want)
		}
	}
	wantColor := []gui.Color{
		gui.Hex(0xE03131), gui.Hex(0xF59F00), gui.Hex(0x2F9E44),
	}
	for i, want := range wantColor {
		if stops[i].Color != want {
			t.Errorf("stop %d Color = %v, want %v", i, stops[i].Color, want)
		}
	}
}

func TestGaugeGradientStopsUnsetZoneColorUsesPalette(t *testing.T) {
	cfg := zoneCfg(true)
	cfg.Zones[1].Color = gui.Color{}
	stops := gaugeGradientStops(&cfg, cfg.Theme)
	want := seriesColor(gui.Color{}, 1, cfg.Theme.Palette)
	if stops[1].Color != want {
		t.Errorf("stop 1 Color = %v, want palette %v", stops[1].Color, want)
	}
}

func TestGaugeGradientStopsNeedsTwoZones(t *testing.T) {
	cfg := zoneCfg(true)
	cfg.Zones = cfg.Zones[:1]
	if got := gaugeGradientStops(&cfg, cfg.Theme); got != nil {
		t.Errorf("stops = %v, want nil for a single zone", got)
	}
	cfg.Zones = nil
	if got := gaugeGradientStops(&cfg, cfg.Theme); got != nil {
		t.Errorf("stops = %v, want nil for no zones", got)
	}
}

func TestGaugeGradientArcGradientWins(t *testing.T) {
	cfg := zoneCfg(true)
	cfg.ArcGradient = []gui.GradientStop{
		{Color: gui.Hex(0x000000), Pos: 0},
		{Color: gui.Hex(0xFFFFFF), Pos: 1},
	}
	stops := gaugeGradientStops(&cfg, cfg.Theme)
	if len(stops) != 2 || stops[0].Color != gui.Hex(0x000000) {
		t.Errorf("stops = %v, want the explicit ArcGradient", stops)
	}
}

// ArcGradient must work with GradientZones off and no zones at all.
func TestGaugeGradientArcGradientWithoutZones(t *testing.T) {
	cfg := GaugeCfg{BaseCfg: BaseCfg{ID: "ag"}, Value: 50, Max: 100}
	cfg.applyGaugeDefaults()
	cfg.ArcGradient = []gui.GradientStop{
		{Color: gui.Hex(0x111111), Pos: 0},
		{Color: gui.Hex(0xEEEEEE), Pos: 1},
	}
	if got := gaugeGradientStops(&cfg, cfg.Theme); len(got) != 2 {
		t.Errorf("stops = %v, want 2", got)
	}
}

func TestSanitizeStopsHostileInput(t *testing.T) {
	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	in := []gui.GradientStop{
		{Color: gui.Hex(0x00FF00), Pos: 0.9},
		{Color: gui.Hex(0xFF0000), Pos: nan},
		{Color: gui.Hex(0x0000FF), Pos: inf},
		{Color: gui.Hex(0xFFFFFF), Pos: -5},
		{Color: gui.Hex(0x000000), Pos: 5},
	}
	got := sanitizeStops(in)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3 (two non-finite dropped)", len(got))
	}
	// Sorted ascending, out-of-range clamped into [0, 1].
	if got[0].Pos != 0 || got[len(got)-1].Pos != 1 {
		t.Errorf("positions = %v, want clamped ends 0 and 1", got)
	}
	for i := 1; i < len(got); i++ {
		if got[i].Pos < got[i-1].Pos {
			t.Errorf("stops not sorted: %v", got)
		}
	}
}

func TestSanitizeStopsTruncatesAndDropsAllBad(t *testing.T) {
	huge := make([]gui.GradientStop, 10000)
	for i := range huge {
		huge[i] = gui.GradientStop{
			Color: gui.Hex(0x123456),
			Pos:   float32(i) / float32(len(huge)),
		}
	}
	if got := len(sanitizeStops(huge)); got != maxArcGradientStops {
		t.Errorf("len = %d, want %d", got, maxArcGradientStops)
	}

	allBad := []gui.GradientStop{{Pos: float32(math.NaN())}}
	if got := sanitizeStops(allBad); got != nil {
		t.Errorf("stops = %v, want nil when nothing survives", got)
	}
}

func TestGaugeSampleStops(t *testing.T) {
	black := gui.Hex(0x000000)
	white := gui.Hex(0xFFFFFF)
	stops := []gui.GradientStop{
		{Color: black, Pos: 0},
		{Color: white, Pos: 1},
	}
	if got := gaugeSampleStops(stops, 0); got != black {
		t.Errorf("sample(0) = %v, want black", got)
	}
	if got := gaugeSampleStops(stops, 1); got != white {
		t.Errorf("sample(1) = %v, want white", got)
	}
	// Out-of-range positions clamp to the ends.
	if got := gaugeSampleStops(stops, -3); got != black {
		t.Errorf("sample(-3) = %v, want black", got)
	}
	if got := gaugeSampleStops(stops, 9); got != white {
		t.Errorf("sample(9) = %v, want white", got)
	}
	mid := gaugeSampleStops(stops, 0.5)
	if mid.R == 0 || mid.R == 255 {
		t.Errorf("sample(0.5) = %v, want a blend", mid)
	}
	if got := gaugeSampleStops(nil, 0.5); got.IsSet() {
		t.Errorf("sample(nil) = %v, want the zero color", got)
	}
}

// Coincident stops are a hard step, which is how a caller asks for a
// band edge inside an otherwise smooth ramp.
func TestGaugeSampleStopsCoincident(t *testing.T) {
	stops := []gui.GradientStop{
		{Color: gui.Hex(0xFF0000), Pos: 0.5},
		{Color: gui.Hex(0x00FF00), Pos: 0.5},
	}
	if got := gaugeSampleStops(stops, 0.5); got != gui.Hex(0xFF0000) {
		t.Errorf("sample at the step = %v, want the left color", got)
	}
	if got := gaugeSampleStops(stops, 0.6); got != gui.Hex(0x00FF00) {
		t.Errorf("sample past the step = %v, want the right color", got)
	}
}

func TestGaugeSegments(t *testing.T) {
	tests := []struct {
		name        string
		outerR, arc float32
		want        int
	}{
		{name: "tiny dial clamps to min", outerR: 4, arc: 1,
			want: gaugeMinGradientSegments},
		{name: "zero radius clamps to min", outerR: 0, arc: 3,
			want: gaugeMinGradientSegments},
		{name: "huge dial clamps to max", outerR: 10000, arc: 4.71,
			want: gaugeMaxGradientSegments},
		{name: "NaN sweep clamps to min", outerR: 100,
			arc: float32(math.NaN()), want: gaugeMinGradientSegments},
		{name: "Inf radius clamps to min", outerR: float32(math.Inf(1)),
			arc: 3, want: gaugeMinGradientSegments},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gaugeSegments(tt.outerR, tt.arc); got != tt.want {
				t.Errorf("gaugeSegments(%v, %v) = %d, want %d",
					tt.outerR, tt.arc, got, tt.want)
			}
		})
	}

	// A mid-size dial lands between the clamps, and the count scales
	// with the arc length rather than sitting at a fixed number.
	small := gaugeSegments(60, DefaultGaugeArcAngle)
	large := gaugeSegments(120, DefaultGaugeArcAngle)
	if small <= gaugeMinGradientSegments || small >= gaugeMaxGradientSegments {
		t.Errorf("small dial = %d, want strictly between the clamps", small)
	}
	if large <= small {
		t.Errorf("segments did not grow with radius: %d then %d", small, large)
	}
}

func TestGaugeGradientDrawDoesNotPanic(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*GaugeCfg)
	}{
		{"gradient zones", func(c *GaugeCfg) { c.GradientZones = true }},
		{"gradient off", func(c *GaugeCfg) {}},
		{"explicit ramp", func(c *GaugeCfg) {
			c.ArcGradient = []gui.GradientStop{
				{Color: gui.Hex(0xE03131), Pos: 0},
				{Color: gui.Hex(0x2F9E44), Pos: 1},
			}
		}},
		{"hostile ramp", func(c *GaugeCfg) {
			c.ArcGradient = []gui.GradientStop{
				{Color: gui.Hex(0xE03131), Pos: float32(math.NaN())},
				{Color: gui.Hex(0x2F9E44), Pos: 1e9},
			}
		}},
		{"gradient with pointer and value off the top", func(c *GaugeCfg) {
			c.GradientZones = true
			c.ShowPointer = true
			c.Value = 1e12
		}},
		{"gradient with NaN value", func(c *GaugeCfg) {
			c.GradientZones = true
			c.ShowPointer = true
			c.Value = math.NaN()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := zoneCfg(false)
			cfg.Width, cfg.Height = 400, 300
			tt.mut(&cfg)
			path := filepath.Join(t.TempDir(), "gauge_gradient.png")
			if err := ExportPNG(Gauge(cfg), 400, 300, path); err != nil {
				t.Fatal(err)
			}
			assertValidPNG(t, path, 400, 300)
		})
	}
}

// A gradient dial must still validate, hit-test and label by the
// discrete thresholds: only the fill changes.
func TestGaugeGradientKeepsDiscreteThresholds(t *testing.T) {
	cfg := zoneCfg(true)
	if err := cfg.Validate(); err != nil {
		t.Fatalf("gradient config did not validate: %v", err)
	}
	gv := newGaugeView(cfg)
	gv.cx, gv.cy, gv.outerR, gv.innerR = 100, 100, 50, 35
	if !gv.gaugeHitTest(100, 60) {
		t.Error("hit-test failed on the ring above the centre")
	}
	if gv.gaugeHitTest(100, 100) {
		t.Error("hit-test hit the hole")
	}
}

func TestGaugeCfgValidateArcGradient(t *testing.T) {
	tests := []struct {
		name    string
		stops   []gui.GradientStop
		wantErr bool
	}{
		{"valid", []gui.GradientStop{{Pos: 0}, {Pos: 1}}, false},
		{"nil", nil, false},
		{"NaN pos", []gui.GradientStop{{Pos: float32(math.NaN())}}, true},
		{"Inf pos", []gui.GradientStop{{Pos: float32(math.Inf(-1))}}, true},
		{"pos below range", []gui.GradientStop{{Pos: -0.1}}, true},
		{"pos above range", []gui.GradientStop{{Pos: 1.1}}, true},
		{"too many stops",
			make([]gui.GradientStop, maxArcGradientStops+1), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GaugeCfg{Value: 50, Max: 100, ArcGradient: tt.stops}
			cfg.applyGaugeDefaults()
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// drawGradientArc must refuse degenerate geometry rather than loop or
// divide by zero. It draws through a real context via ExportPNG
// elsewhere; here the guards are checked directly.
func TestDrawGradientArcGuards(t *testing.T) {
	stops := []gui.GradientStop{{Color: gui.Hex(0xFFFFFF), Pos: 0}}
	// A nil context would panic if any guard failed to return early.
	bg := gui.Hex(0x000000)
	drawGradientArc(nil, 0, 0, 0, 0, 1, stops, 0, 1, 1, bg)
	drawGradientArc(nil, 0, 0, 10, 0, 0, stops, 0, 1, 1, bg)
	drawGradientArc(nil, 0, 0, 10, 0, float32(math.NaN()), stops, 0, 1, 1, bg)
	drawGradientArc(nil, 0, 0, 10, 0, float32(math.Inf(1)), stops, 0, 1, 1, bg)
	drawGradientArc(nil, 0, 0, 10, 0, 1, nil, 0, 1, 1, bg)
	// A non-finite radius passes r <= 0, so it needs its own guard:
	// without it the overlap division poisons every segment angle.
	drawGradientArc(nil, 0, 0, float32(math.NaN()), 0, 1, stops, 0, 1, 1, bg)
	drawGradientArc(nil, 0, 0, float32(math.Inf(1)), 0, 1, stops, 0, 1, 1, bg)
}

// The ramp is built at construction, not per frame. A nil cache on a
// gradient dial would silently fall back to the stepped fill.
func TestNewGaugeViewCachesRamp(t *testing.T) {
	gv := newGaugeView(zoneCfg(true))
	if len(gv.stops) != 3 {
		t.Fatalf("cached stops = %d, want 3", len(gv.stops))
	}
	if gv.stops[0].Pos != 0 || gv.stops[2].Pos != 1 {
		t.Errorf("cached ramp not pinned at the ends: %v", gv.stops)
	}
	if flat := newGaugeView(zoneCfg(false)); flat.stops != nil {
		t.Errorf("stops = %v, want nil with GradientZones off", flat.stops)
	}
}

// gaugeZoneColor picks the needle color on the stepped path.
func TestGaugeZoneColor(t *testing.T) {
	cfg := zoneCfg(false)
	tests := []struct {
		name  string
		value float64
		want  gui.Color
	}{
		{"first zone", 10, gui.Hex(0xE03131)},
		{"on a threshold takes that zone", 30, gui.Hex(0xE03131)},
		{"middle zone", 50, gui.Hex(0xF59F00)},
		{"last zone", 90, gui.Hex(0x2F9E44)},
		// Past every threshold there is no zone to name, so the
		// first palette entry stands in.
		{"above all zones", 500, seriesColor(gui.Color{}, 0, cfg.Theme.Palette)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.Value = tt.value
			if got := gaugeZoneColor(&cfg, cfg.Theme); got != tt.want {
				t.Errorf("gaugeZoneColor(%v) = %v, want %v",
					tt.value, got, tt.want)
			}
		})
	}
}

// An unset zone color falls back to that zone's palette entry, and a
// gauge with no zones at all falls back to the first.
func TestGaugeZoneColorPaletteFallback(t *testing.T) {
	cfg := zoneCfg(false)
	cfg.Zones[1].Color = gui.Color{}
	cfg.Value = 50
	want := seriesColor(gui.Color{}, 1, cfg.Theme.Palette)
	if got := gaugeZoneColor(&cfg, cfg.Theme); got != want {
		t.Errorf("unset zone color = %v, want palette[1] %v", got, want)
	}

	cfg.Zones = nil
	want = seriesColor(gui.Color{}, 0, cfg.Theme.Palette)
	if got := gaugeZoneColor(&cfg, cfg.Theme); got != want {
		t.Errorf("no zones = %v, want palette[0] %v", got, want)
	}
}

// The flattened track must land on the same pixel the stepped zones
// produce, or a dial changes weight when GradientZones is flipped.
// The stepped path draws the zone color at alpha 60 over the
// background, which resolves to exactly this blend.
func TestFlattenOntoMatchesSteppedZoneTint(t *testing.T) {
	bg := gui.Hex(0x1A1B1E)
	zone := gui.Hex(0x59A14F)
	want := theme.Lerp(bg, gui.RGBA(zone.R, zone.G, zone.B, 60), 60.0/255)
	got := flattenOnto(bg, zone, gaugeTrackAlpha)
	for i, d := range [][2]uint8{{got.R, want.R}, {got.G, want.G}, {got.B, want.B}} {
		if diff := int(d[0]) - int(d[1]); diff > 1 || diff < -1 {
			t.Errorf("channel %d = %d, want %d (stepped-zone tint)", i, d[0], d[1])
		}
	}
	if got.A != 255 {
		t.Errorf("alpha = %d, want 255: segments must be opaque to overlap", got.A)
	}
}

// A translucent stop must still read as translucent against the
// background, even though the drawn segment is opaque.
func TestFlattenOntoHonoursStopAlpha(t *testing.T) {
	bg := gui.Hex(0x000000)
	half := gui.RGBA(255, 255, 255, 128)
	got := flattenOnto(bg, half, 1)
	if got.R < 120 || got.R > 136 {
		t.Errorf("R = %d, want about half way to white", got.R)
	}
	if got.A != 255 {
		t.Errorf("alpha = %d, want 255", got.A)
	}
}

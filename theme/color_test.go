package theme

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-gui/gui"
)

func TestWithAlpha(t *testing.T) {
	c := gui.Hex(0xFF0000)
	got := WithAlpha(c, 0.5)
	if got.A != 127 {
		t.Errorf("A = %d, want 127", got.A)
	}
	if got.R != 255 || got.G != 0 || got.B != 0 {
		t.Errorf("RGB changed: %v", got)
	}
}

func TestWithAlphaClamped(t *testing.T) {
	c := gui.Hex(0x000000)
	if got := WithAlpha(c, -1); got.A != 0 {
		t.Errorf("A = %d, want 0 for negative", got.A)
	}
	if got := WithAlpha(c, 2); got.A != 255 {
		t.Errorf("A = %d, want 255 for >1", got.A)
	}
}

func TestLighten(t *testing.T) {
	c := gui.Hex(0x000000)
	got := Lighten(c, 1.0)
	if got.R != 255 || got.G != 255 || got.B != 255 {
		t.Errorf("Lighten(black, 1.0) = %v, want white", got)
	}
}

func TestLightenZero(t *testing.T) {
	c := gui.Hex(0x804020)
	got := Lighten(c, 0)
	if got.R != c.R || got.G != c.G || got.B != c.B {
		t.Errorf("Lighten(c, 0) changed color: %v", got)
	}
}

func TestDarken(t *testing.T) {
	c := gui.Hex(0xFFFFFF)
	got := Darken(c, 1.0)
	if got.R != 0 || got.G != 0 || got.B != 0 {
		t.Errorf("Darken(white, 1.0) = %v, want black", got)
	}
}

func TestDarkenZero(t *testing.T) {
	c := gui.Hex(0x804020)
	got := Darken(c, 0)
	if got.R != c.R || got.G != c.G || got.B != c.B {
		t.Errorf("Darken(c, 0) changed color: %v", got)
	}
}

func TestLerp(t *testing.T) {
	c1 := gui.RGBA(0, 0, 0, 255)
	c2 := gui.RGBA(100, 200, 50, 255)

	got := Lerp(c1, c2, 0)
	if got.R != 0 || got.G != 0 || got.B != 0 {
		t.Errorf("Lerp(t=0) = %v, want black", got)
	}
	got = Lerp(c1, c2, 1)
	if got.R != 100 || got.G != 200 || got.B != 50 {
		t.Errorf("Lerp(t=1) = %v, want c2", got)
	}
	got = Lerp(c1, c2, 0.5)
	if got.R != 50 || got.G != 100 || got.B != 25 {
		t.Errorf("Lerp(t=0.5) = %v, want midpoint", got)
	}
}

func TestLerpClamped(t *testing.T) {
	c1 := gui.Hex(0x000000)
	c2 := gui.Hex(0xFFFFFF)
	if got := Lerp(c1, c2, -1); got.R != 0 {
		t.Errorf("Lerp(t=-1) R=%d, want 0", got.R)
	}
	if got := Lerp(c1, c2, 2); got.R != 255 {
		t.Errorf("Lerp(t=2) R=%d, want 255", got.R)
	}
}

func TestLuminance(t *testing.T) {
	if got := Luminance(gui.Hex(0x000000)); got != 0 {
		t.Errorf("Luminance(black) = %v, want 0", got)
	}
	if got := Luminance(gui.Hex(0xFFFFFF)); got != 1 {
		t.Errorf("Luminance(white) = %v, want 1", got)
	}
}

func TestLerpOklabEndpoints(t *testing.T) {
	c1 := gui.Hex(0xE03131)
	c2 := gui.Hex(0x2F9E44)
	// Endpoints are returned unchanged, not round-tripped through
	// Oklab, so a stop always renders as the color the caller gave.
	if got := LerpOklab(c1, c2, 0); got != c1 {
		t.Errorf("LerpOklab(t=0) = %v, want %v", got, c1)
	}
	if got := LerpOklab(c1, c2, 1); got != c2 {
		t.Errorf("LerpOklab(t=1) = %v, want %v", got, c2)
	}
}

func TestLerpOklabClamped(t *testing.T) {
	c1 := gui.Hex(0x000000)
	c2 := gui.Hex(0xFFFFFF)
	if got := LerpOklab(c1, c2, -1); got != c1 {
		t.Errorf("LerpOklab(t=-1) = %v, want c1", got)
	}
	if got := LerpOklab(c1, c2, 2); got != c2 {
		t.Errorf("LerpOklab(t=2) = %v, want c2", got)
	}
	// A NaN t fails both bound checks, so it must be caught on its
	// own or it would produce a NaN color.
	if got := LerpOklab(c1, c2, math.NaN()); got != c1 {
		t.Errorf("LerpOklab(t=NaN) = %v, want c1", got)
	}
}

func TestLerpOklabStaysSaturated(t *testing.T) {
	// The point of Oklab: an sRGB red-to-green midpoint drops both
	// endpoints towards grey. Oklab holds more of the light.
	red := gui.Hex(0xE03131)
	green := gui.Hex(0x2F9E44)
	srgb := Lerp(red, green, 0.5)
	oklab := LerpOklab(red, green, 0.5)
	if oklab.R <= srgb.R {
		t.Errorf("Oklab midpoint R = %d, want brighter than sRGB %d",
			oklab.R, srgb.R)
	}
	if Luminance(oklab) <= Luminance(srgb) {
		t.Errorf("Oklab midpoint luminance %v, want above sRGB %v",
			Luminance(oklab), Luminance(srgb))
	}
}

func TestLerpOklabAlpha(t *testing.T) {
	c1 := gui.RGBA(255, 0, 0, 0)
	c2 := gui.RGBA(255, 0, 0, 200)
	got := LerpOklab(c1, c2, 0.5)
	if got.A != 100 {
		t.Errorf("alpha = %d, want 100 (linear midpoint)", got.A)
	}
}

func TestOklabRoundTrip(t *testing.T) {
	// Interpolating a color with itself must return that color, or a
	// single-color ramp would drift. Checked across the gamut.
	colors := []gui.Color{
		gui.Hex(0x000000), gui.Hex(0xFFFFFF), gui.Hex(0xFF0000),
		gui.Hex(0x00FF00), gui.Hex(0x0000FF), gui.Hex(0x808080),
		gui.Hex(0x1C7ED6), gui.Hex(0xF59F00),
	}
	for _, c := range colors {
		got := LerpOklab(c, c, 0.5)
		for _, d := range [][2]uint8{{got.R, c.R}, {got.G, c.G}, {got.B, c.B}} {
			if diff := int(d[0]) - int(d[1]); diff > 1 || diff < -1 {
				t.Errorf("round trip %v -> %v drifted", c, got)
				break
			}
		}
	}
}

package theme

import "github.com/go-gui-org/go-gui/gui"

// WithAlpha returns a copy of the color with the given alpha
// (0.0 = fully transparent, 1.0 = fully opaque).
func WithAlpha(c gui.Color, alpha float64) gui.Color {
	a := uint8(max(0, min(255, alpha*255)))
	return gui.RGBA(c.R, c.G, c.B, a)
}

// Lighten returns a lighter version of the color. Amount is
// 0.0 (unchanged) to 1.0 (white).
func Lighten(c gui.Color, amount float64) gui.Color {
	f := max(0, min(1, amount))
	r := c.R + uint8(float64(255-c.R)*f)
	g := c.G + uint8(float64(255-c.G)*f)
	b := c.B + uint8(float64(255-c.B)*f)
	return gui.RGBA(r, g, b, c.A)
}

// Darken returns a darker version of the color. Amount is
// 0.0 (unchanged) to 1.0 (black).
func Darken(c gui.Color, amount float64) gui.Color {
	f := 1 - max(0, min(1, amount))
	r := uint8(float64(c.R) * f)
	g := uint8(float64(c.G) * f)
	b := uint8(float64(c.B) * f)
	return gui.RGBA(r, g, b, c.A)
}

// Lerp linearly interpolates between two colors. t is clamped
// to [0, 1] where 0 returns c1 and 1 returns c2.
func Lerp(c1, c2 gui.Color, t float64) gui.Color {
	t = max(0, min(1, t))
	r := uint8(float64(c1.R) + t*float64(int(c2.R)-int(c1.R)))
	g := uint8(float64(c1.G) + t*float64(int(c2.G)-int(c1.G)))
	b := uint8(float64(c1.B) + t*float64(int(c2.B)-int(c1.B)))
	a := uint8(float64(c1.A) + t*float64(int(c2.A)-int(c1.A)))
	return gui.RGBA(r, g, b, a)
}

// Luminance returns the relative luminance of a color using
// the sRGB approximation (0.299R + 0.587G + 0.114B). Result
// is in [0, 1].
func Luminance(c gui.Color) float64 {
	return (0.299*float64(c.R) + 0.587*float64(c.G) +
		0.114*float64(c.B)) / 255
}

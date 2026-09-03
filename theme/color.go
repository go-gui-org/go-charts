package theme

import (
	"math"

	"github.com/go-gui-org/go-gui/gui"
)

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

// LerpOklab interpolates between two colors in the Oklab perceptual
// space. A red-to-green ramp keeps its saturation instead of passing
// through mud, which plain Lerp cannot avoid: sRGB channel averaging
// drops both endpoints towards grey in the middle.
//
// t is clamped to [0, 1]. The endpoints are returned unchanged rather
// than round-tripped, so a stop always renders as the color the
// caller gave. Alpha is interpolated linearly, outside Oklab, because
// Oklab has no alpha axis.
func LerpOklab(c1, c2 gui.Color, t float64) gui.Color {
	// Exact endpoints, returned as given rather than round-tripped.
	// A NaN t fails both comparisons and reaches the check below.
	if t <= 0 {
		return c1
	}
	if t >= 1 {
		return c2
	}
	if math.IsNaN(t) {
		return c1
	}

	l1, a1, b1 := srgbToOklab(c1)
	l2, a2, b2 := srgbToOklab(c2)
	l := l1 + t*(l2-l1)
	a := a1 + t*(a2-a1)
	b := b1 + t*(b2-b1)
	alpha := uint8(float64(c1.A) + t*float64(int(c2.A)-int(c1.A)))
	return oklabToSRGB(l, a, b, alpha)
}

// srgbDecode converts one gamma-encoded sRGB channel (0–255) to
// linear light in [0, 1].
func srgbDecode(v uint8) float64 {
	c := float64(v) / 255
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// srgbEncode converts linear light back to a gamma-encoded sRGB
// channel, saturating out-of-gamut values instead of wrapping.
func srgbEncode(c float64) uint8 {
	if math.IsNaN(c) || c <= 0 {
		return 0
	}
	if c >= 1 {
		return 255
	}
	if c <= 0.0031308 {
		c *= 12.92
	} else {
		c = 1.055*math.Pow(c, 1/2.4) - 0.055
	}
	// c is strictly inside (0, 1) here: both ends returned above.
	return uint8(math.Round(c * 255))
}

// srgbToOklab converts a color to Oklab (Björn Ottosson's matrices).
// Alpha is not carried: Oklab has no alpha axis.
func srgbToOklab(c gui.Color) (l, a, b float64) {
	r := srgbDecode(c.R)
	g := srgbDecode(c.G)
	bl := srgbDecode(c.B)

	lc := 0.4122214708*r + 0.5363325363*g + 0.0514459929*bl
	mc := 0.2119034982*r + 0.6806995451*g + 0.1073969566*bl
	sc := 0.0883024619*r + 0.2817188376*g + 0.6299787005*bl

	// Cbrt, not Pow(1/3): Cbrt is defined for negative inputs, which
	// rounding can produce from a channel at zero.
	lr := math.Cbrt(lc)
	mr := math.Cbrt(mc)
	sr := math.Cbrt(sc)

	l = 0.2104542553*lr + 0.7936177850*mr - 0.0040720468*sr
	a = 1.9779984951*lr - 2.4285922050*mr + 0.4505937099*sr
	b = 0.0259040371*lr + 0.7827717662*mr - 0.8086757660*sr
	return l, a, b
}

// oklabToSRGB is the inverse of srgbToOklab, with alpha attached.
// Out-of-gamut results are saturated per channel by srgbEncode.
func oklabToSRGB(l, a, b float64, alpha uint8) gui.Color {
	lr := l + 0.3963377774*a + 0.2158037573*b
	mr := l - 0.1055613458*a - 0.0638541728*b
	sr := l - 0.0894841775*a - 1.2914855480*b

	lc := lr * lr * lr
	mc := mr * mr * mr
	sc := sr * sr * sr

	r := 4.0767416621*lc - 3.3077115913*mc + 0.2309699292*sc
	g := -1.2684380046*lc + 2.6097574011*mc - 0.3413193965*sc
	bl := -0.0041960863*lc - 0.7034186147*mc + 1.7076147010*sc

	return gui.RGBA(srgbEncode(r), srgbEncode(g), srgbEncode(bl), alpha)
}

// Luminance returns the relative luminance of a color using
// the sRGB approximation (0.299R + 0.587G + 0.114B). Result
// is in [0, 1].
func Luminance(c gui.Color) float64 {
	return (0.299*float64(c.R) + 0.587*float64(c.G) +
		0.114*float64(c.B)) / 255
}

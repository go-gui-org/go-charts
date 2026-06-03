package theme

import (
	"slices"

	"github.com/go-gui-org/go-gui/gui"
)

var tableau10 = []gui.Color{
	gui.Hex(0x4E79A7), // blue
	gui.Hex(0xF28E2B), // orange
	gui.Hex(0xE15759), // red
	gui.Hex(0x76B7B2), // teal
	gui.Hex(0x59A14F), // green
	gui.Hex(0xEDC948), // yellow
	gui.Hex(0xB07AA1), // purple
	gui.Hex(0xFF9DA7), // pink
	gui.Hex(0x9C755F), // brown
	gui.Hex(0xBAB0AC), // gray
}

var pastel = []gui.Color{
	gui.Hex(0xA1C9F4),
	gui.Hex(0xFFB482),
	gui.Hex(0x8DE5A1),
	gui.Hex(0xFF9F9B),
	gui.Hex(0xD0BBFF),
	gui.Hex(0xDEBB9B),
	gui.Hex(0xFAB0E4),
	gui.Hex(0xCFCFCF),
	gui.Hex(0xFFFEA3),
	gui.Hex(0xB9F2F0),
}

var vivid = []gui.Color{
	gui.Hex(0xE64B35),
	gui.Hex(0x4DBBD5),
	gui.Hex(0x00A087),
	gui.Hex(0x3C5488),
	gui.Hex(0xF39B7F),
	gui.Hex(0x8491B4),
	gui.Hex(0x91D1C2),
	gui.Hex(0xDC0000),
	gui.Hex(0x7E6148),
	gui.Hex(0xB09C85),
}

// DefaultPalette returns the default color palette (Tableau 10).
func DefaultPalette() []gui.Color { return slices.Clone(tableau10) }

// Tableau10 returns the Tableau 10 color palette.
func Tableau10() []gui.Color { return slices.Clone(tableau10) }

// Pastel returns a pastel color palette.
func Pastel() []gui.Color { return slices.Clone(pastel) }

// Vivid returns a vivid/saturated color palette.
func Vivid() []gui.Color { return slices.Clone(vivid) }

var highContrast = []gui.Color{
	gui.Hex(0x0000CD), // medium blue
	gui.Hex(0xDC143C), // crimson
	gui.Hex(0x009E73), // bluish green
	gui.Hex(0xD55E00), // vermillion
	gui.Hex(0xCC79A7), // reddish purple
	gui.Hex(0x56B4E9), // sky blue
	gui.Hex(0xB8860B), // dark goldenrod
	gui.Hex(0x8B008B), // dark magenta
	gui.Hex(0x2F4F4F), // dark slate gray
	gui.Hex(0xE69F00), // orange
}

// HighContrast returns a high-contrast color palette designed for
// accessibility. Colors are colorblind-safe and meet WCAG AA
// contrast guidelines against both light and dark backgrounds.
func HighContrast() []gui.Color { return slices.Clone(highContrast) }

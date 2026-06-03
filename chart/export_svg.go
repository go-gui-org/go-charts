package chart

import (
	"errors"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/go-gui-org/go-glyph"
	"github.com/go-gui-org/go-gui/gui"
)

// ExportSVG renders a chart view to an SVG file at the given
// pixel dimensions.
func ExportSVG(v gui.View, width, height int, path string) error {
	svg, err := ExportSVGString(v, width, height)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(svg), 0o644)
}

// ExportSVGString renders a chart view to an SVG string.
func ExportSVGString(v gui.View, width, height int) (string, error) {
	d, ok := v.(Drawer)
	if !ok {
		return "", errors.New("chart: view does not implement Drawer")
	}
	if width <= 0 || height <= 0 {
		return "", errors.New("chart: dimensions must be positive")
	}
	const maxDim = 16384
	if width > maxDim || height > maxDim {
		return "", errors.New("chart: dimensions exceed 16384")
	}

	// Set up text measurement (needed for layout).
	backend := &imageBackend{
		target:   nil, // not rasterizing
		textures: make(map[glyph.TextureID]atlasPage),
	}
	textSys, err := glyph.NewTextSystem(backend)
	if err != nil {
		return "", err
	}
	defer textSys.Free()

	measurer := &pngTextMeasurer{textSys: textSys}
	dc := gui.NewDrawContext(float32(width), float32(height), measurer)

	rec := &svgRecorder{measurer: measurer}
	dc.SetRecorder(rec)
	d.Draw(dc)

	// Build SVG.
	var b strings.Builder
	b.Grow(4096)

	// Header with viewBox for scalability.
	fmt.Fprintf(&b,
		"<svg xmlns=\"http://www.w3.org/2000/svg\""+
			" width=\"%d\" height=\"%d\""+
			" viewBox=\"0 0 %d %d\">\n",
		width, height, width, height)

	// Background.
	bg := gui.Color{R: 255, G: 255, B: 255, A: 255}
	if t, ok := v.(themer); ok && t.chartTheme() != nil {
		c := t.chartTheme().Background
		if c.IsSet() {
			bg = c
		}
	}
	fmt.Fprintf(&b,
		"<rect width=\"%d\" height=\"%d\" fill=\"%s\"/>\n",
		width, height, colorToCSS(bg))

	// Emit recorded commands.
	for _, cmd := range rec.cmds {
		cmd.writeSVG(&b)
	}

	b.WriteString("</svg>\n")
	return b.String(), nil
}

// -----------------------------------------------------------
// SVG command types
// -----------------------------------------------------------

type svgCmd interface {
	writeSVG(b *strings.Builder)
}

type svgRecorder struct {
	cmds     []svgCmd
	measurer gui.TextMeasurer
}

func (r *svgRecorder) Line(x0, y0, x1, y1 float32, color gui.Color, width float32) {
	r.cmds = append(r.cmds, &svgLine{x0, y0, x1, y1, color, width})
}

func (r *svgRecorder) Polyline(points []float32, color gui.Color, width float32) {
	r.cmds = append(r.cmds, &svgPolyline{
		points: copyPoints(points), color: color,
		width: width,
	})
}

func (r *svgRecorder) FilledRect(x, y, w, h float32, color gui.Color) {
	r.cmds = append(r.cmds, &svgFilledRect{x, y, w, h, color})
}

func (r *svgRecorder) Rect(x, y, w, h float32, color gui.Color, width float32) {
	r.cmds = append(r.cmds, &svgRect{x, y, w, h, color, width})
}

func (r *svgRecorder) FilledCircle(cx, cy, radius float32, color gui.Color) {
	r.cmds = append(r.cmds, &svgFilledCircle{cx, cy, radius, color})
}

func (r *svgRecorder) Circle(cx, cy, radius float32, color gui.Color, width float32) {
	r.cmds = append(r.cmds, &svgCircle{cx, cy, radius, color, width})
}

func (r *svgRecorder) FilledArc(cx, cy, rx, ry, start, sweep float32, color gui.Color) {
	r.cmds = append(r.cmds, &svgFilledArc{cx, cy, rx, ry, start, sweep, color})
}

func (r *svgRecorder) Arc(cx, cy, rx, ry, start, sweep float32, color gui.Color, width float32) {
	r.cmds = append(r.cmds, &svgArc{cx, cy, rx, ry, start, sweep, color, width})
}

func (r *svgRecorder) FilledPolygon(points []float32, color gui.Color) {
	r.cmds = append(r.cmds, &svgFilledPolygon{
		points: copyPoints(points), color: color,
	})
}

func (r *svgRecorder) FilledRoundedRect(x, y, w, h, radius float32, color gui.Color) {
	r.cmds = append(r.cmds, &svgFilledRoundedRect{x, y, w, h, radius, color})
}

func (r *svgRecorder) RoundedRect(x, y, w, h, radius float32, color gui.Color, width float32) {
	r.cmds = append(r.cmds, &svgRoundedRect{x, y, w, h, radius, color, width})
}

func (r *svgRecorder) DashedLine(
	x0, y0, x1, y1 float32,
	color gui.Color, width, dashLen, gapLen float32,
) {
	r.cmds = append(r.cmds, &svgDashedLine{
		x0, y0, x1, y1, color, width, dashLen, gapLen,
	})
}

func (r *svgRecorder) DashedPolyline(
	points []float32,
	color gui.Color, width, dashLen, gapLen float32,
) {
	r.cmds = append(r.cmds, &svgDashedPolyline{
		points: copyPoints(points), color: color,
		width: width, dashLen: dashLen, gapLen: gapLen,
	})
}

func (r *svgRecorder) PolylineJoined(points []float32, color gui.Color, width float32) {
	r.cmds = append(r.cmds, &svgPolylineJoined{
		points: copyPoints(points), color: color,
		width: width,
	})
}

func (r *svgRecorder) QuadBezier(
	x0, y0, cx, cy, x1, y1 float32, color gui.Color, width float32,
) {
	r.cmds = append(r.cmds, &svgQuadBezier{
		x0, y0, cx, cy, x1, y1, color, width,
	})
}

func (r *svgRecorder) CubicBezier(
	x0, y0, c1x, c1y, c2x, c2y, x1, y1 float32, color gui.Color, width float32,
) {
	r.cmds = append(r.cmds, &svgCubicBezier{
		x0, y0, c1x, c1y, c2x, c2y, x1, y1, color, width,
	})
}

func (r *svgRecorder) Text(x, y float32, text string, style gui.TextStyle) {
	ascent := style.Size * 0.8
	if r.measurer != nil {
		ascent = r.measurer.FontAscent(style)
	}
	r.cmds = append(r.cmds, &svgText{x, y, text, style, ascent})
}

// -----------------------------------------------------------
// SVG command implementations
// -----------------------------------------------------------

type svgLine struct {
	x0, y0, x1, y1 float32
	color          gui.Color
	width          float32
}

func (c *svgLine) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\""+
			" stroke=\"%s\" stroke-width=\"%.1f\"",
		c.x0, c.y0, c.x1, c.y1, colorToCSS(c.color), c.width)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgPolyline struct {
	points []float32
	color  gui.Color
	width  float32
}

func (c *svgPolyline) writeSVG(b *strings.Builder) {
	b.WriteString("<polyline points=\"")
	writePoints(b, c.points)
	fmt.Fprintf(b,
		"\" fill=\"none\" stroke=\"%s\" stroke-width=\"%.1f\"",
		colorToCSS(c.color), c.width)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgFilledRect struct {
	x, y, w, h float32
	color      gui.Color
}

func (c *svgFilledRect) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\""+
			" fill=\"%s\"",
		c.x, c.y, c.w, c.h, colorToCSS(c.color))
	writeFillOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgRect struct {
	x, y, w, h float32
	color      gui.Color
	width      float32
}

func (c *svgRect) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\""+
			" fill=\"none\" stroke=\"%s\" stroke-width=\"%.1f\"",
		c.x, c.y, c.w, c.h, colorToCSS(c.color), c.width)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgFilledCircle struct {
	cx, cy, radius float32
	color          gui.Color
}

func (c *svgFilledCircle) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<circle cx=\"%.1f\" cy=\"%.1f\" r=\"%.1f\" fill=\"%s\"",
		c.cx, c.cy, c.radius, colorToCSS(c.color))
	writeFillOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgCircle struct {
	cx, cy, radius float32
	color          gui.Color
	width          float32
}

func (c *svgCircle) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<circle cx=\"%.1f\" cy=\"%.1f\" r=\"%.1f\""+
			" fill=\"none\" stroke=\"%s\" stroke-width=\"%.1f\"",
		c.cx, c.cy, c.radius, colorToCSS(c.color), c.width)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgFilledArc struct {
	cx, cy, rx, ry float32
	start, sweep   float32
	color          gui.Color
}

func (c *svgFilledArc) writeSVG(b *strings.Builder) {
	b.WriteString("<path d=\"")
	writeArcPath(b, c.cx, c.cy, c.rx, c.ry, c.start, c.sweep, true)
	fmt.Fprintf(b, "\" fill=\"%s\"", colorToCSS(c.color))
	writeFillOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgArc struct {
	cx, cy, rx, ry float32
	start, sweep   float32
	color          gui.Color
	width          float32
}

func (c *svgArc) writeSVG(b *strings.Builder) {
	b.WriteString("<path d=\"")
	writeArcPath(b, c.cx, c.cy, c.rx, c.ry, c.start, c.sweep, false)
	fmt.Fprintf(b,
		"\" fill=\"none\" stroke=\"%s\" stroke-width=\"%.1f\"",
		colorToCSS(c.color), c.width)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgFilledPolygon struct {
	points []float32
	color  gui.Color
}

func (c *svgFilledPolygon) writeSVG(b *strings.Builder) {
	b.WriteString("<polygon points=\"")
	writePoints(b, c.points)
	fmt.Fprintf(b, "\" fill=\"%s\"", colorToCSS(c.color))
	writeFillOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgFilledRoundedRect struct {
	x, y, w, h, radius float32
	color              gui.Color
}

func (c *svgFilledRoundedRect) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\""+
			" rx=\"%.1f\" fill=\"%s\"",
		c.x, c.y, c.w, c.h, c.radius, colorToCSS(c.color))
	writeFillOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgRoundedRect struct {
	x, y, w, h, radius float32
	color              gui.Color
	width              float32
}

func (c *svgRoundedRect) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\""+
			" rx=\"%.1f\" fill=\"none\" stroke=\"%s\""+
			" stroke-width=\"%.1f\"",
		c.x, c.y, c.w, c.h, c.radius, colorToCSS(c.color), c.width)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgDashedLine struct {
	x0, y0, x1, y1         float32
	color                  gui.Color
	width, dashLen, gapLen float32
}

func (c *svgDashedLine) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\""+
			" stroke=\"%s\" stroke-width=\"%.1f\""+
			" stroke-dasharray=\"%.1f %.1f\"",
		c.x0, c.y0, c.x1, c.y1,
		colorToCSS(c.color), c.width, c.dashLen, c.gapLen)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgDashedPolyline struct {
	points                 []float32
	color                  gui.Color
	width, dashLen, gapLen float32
}

func (c *svgDashedPolyline) writeSVG(b *strings.Builder) {
	b.WriteString("<polyline points=\"")
	writePoints(b, c.points)
	fmt.Fprintf(b,
		"\" fill=\"none\" stroke=\"%s\" stroke-width=\"%.1f\""+
			" stroke-dasharray=\"%.1f %.1f\"",
		colorToCSS(c.color), c.width, c.dashLen, c.gapLen)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgPolylineJoined struct {
	points []float32
	color  gui.Color
	width  float32
}

func (c *svgPolylineJoined) writeSVG(b *strings.Builder) {
	b.WriteString("<polyline points=\"")
	writePoints(b, c.points)
	fmt.Fprintf(b,
		"\" fill=\"none\" stroke=\"%s\" stroke-width=\"%.1f\""+
			" stroke-linejoin=\"miter\"",
		colorToCSS(c.color), c.width)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgQuadBezier struct {
	x0, y0, cx, cy, x1, y1 float32
	color                  gui.Color
	width                  float32
}

func (c *svgQuadBezier) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<path d=\"M%.1f %.1f Q%.1f %.1f %.1f %.1f\""+
			" fill=\"none\" stroke=\"%s\" stroke-width=\"%.1f\"",
		c.x0, c.y0, c.cx, c.cy, c.x1, c.y1,
		colorToCSS(c.color), c.width)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgCubicBezier struct {
	x0, y0, c1x, c1y, c2x, c2y, x1, y1 float32
	color                              gui.Color
	width                              float32
}

func (c *svgCubicBezier) writeSVG(b *strings.Builder) {
	fmt.Fprintf(b,
		"<path d=\"M%.1f %.1f C%.1f %.1f %.1f %.1f %.1f %.1f\""+
			" fill=\"none\" stroke=\"%s\" stroke-width=\"%.1f\"",
		c.x0, c.y0, c.c1x, c.c1y, c.c2x, c.c2y, c.x1, c.y1,
		colorToCSS(c.color), c.width)
	writeOpacity(b, c.color)
	b.WriteString("/>\n")
}

type svgText struct {
	x, y   float32
	text   string
	style  gui.TextStyle
	ascent float32
}

func (c *svgText) writeSVG(b *strings.Builder) {
	// SVG default baseline is alphabetic. Chart text uses
	// top-left origin, so offset y by font ascent.
	svgY := c.y + c.ascent

	fmt.Fprintf(b, "<text x=\"%.1f\" y=\"%.1f\"", c.x, svgY)

	// Font attributes.
	family := c.style.Family
	if family == "" {
		family = "sans-serif"
	}
	fmt.Fprintf(b, " font-family=\"%s\"", family)
	if c.style.Size > 0 {
		fmt.Fprintf(b, " font-size=\"%.1f\"", c.style.Size)
	}
	fmt.Fprintf(b, " fill=\"%s\"", colorToCSS(c.style.Color))
	writeFillOpacity(b, c.style.Color)

	// Rotation around the original top-left origin.
	if c.style.RotationRadians != 0 {
		deg := c.style.RotationRadians * 180 / math.Pi
		fmt.Fprintf(b, " transform=\"rotate(%.1f %.1f %.1f)\"",
			deg, c.x, c.y)
	}

	b.WriteByte('>')
	writeEscaped(b, c.text)
	b.WriteString("</text>\n")
}

// -----------------------------------------------------------
// Helpers
// -----------------------------------------------------------

// colorToCSS returns an SVG color string (#RRGGBB).
func colorToCSS(c gui.Color) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// writeOpacity writes stroke-opacity if alpha < 255.
func writeOpacity(b *strings.Builder, c gui.Color) {
	if c.A < 255 {
		fmt.Fprintf(b, " stroke-opacity=\"%.2f\"", float32(c.A)/255)
	}
}

// writeFillOpacity writes fill-opacity if alpha < 255.
func writeFillOpacity(b *strings.Builder, c gui.Color) {
	if c.A < 255 {
		fmt.Fprintf(b, " fill-opacity=\"%.2f\"", float32(c.A)/255)
	}
}

// writePoints writes space-separated x,y pairs.
func writePoints(b *strings.Builder, pts []float32) {
	for i := 0; i+1 < len(pts); i += 2 {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(b, "%.1f,%.1f", pts[i], pts[i+1])
	}
}

// writeArcPath writes an SVG arc path command. If filled, the
// path closes through the center (pie slice).
func writeArcPath(
	b *strings.Builder,
	cx, cy, rx, ry, start, sweep float32,
	filled bool,
) {
	// Guard non-finite or degenerate values.
	if rx <= 0 || ry <= 0 || sweep == 0 {
		return
	}
	if isNonFinite32(cx) || isNonFinite32(cy) ||
		isNonFinite32(rx) || isNonFinite32(ry) ||
		isNonFinite32(start) || isNonFinite32(sweep) {
		return
	}

	// Start point.
	sx := cx + rx*float32(math.Cos(float64(start)))
	sy := cy + ry*float32(math.Sin(float64(start)))
	// End point.
	end := start + sweep
	ex := cx + rx*float32(math.Cos(float64(end)))
	ey := cy + ry*float32(math.Sin(float64(end)))

	largeArc := 0
	if math.Abs(float64(sweep)) > math.Pi {
		largeArc = 1
	}
	sweepFlag := 1
	if sweep < 0 {
		sweepFlag = 0
	}

	if filled {
		fmt.Fprintf(b, "M%.1f %.1f ", cx, cy)
		fmt.Fprintf(b, "L%.1f %.1f ", sx, sy)
	} else {
		fmt.Fprintf(b, "M%.1f %.1f ", sx, sy)
	}
	fmt.Fprintf(b, "A%.1f %.1f 0 %d %d %.1f %.1f",
		rx, ry, largeArc, sweepFlag, ex, ey)
	if filled {
		b.WriteByte('Z')
	}
}

// writeEscaped writes XML-escaped text.
func writeEscaped(b *strings.Builder, s string) {
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		default:
			b.WriteRune(r)
		}
	}
}

// isNonFinite32 reports whether f is NaN or +/-Inf.
func isNonFinite32(f float32) bool {
	return f != f || f-f != 0 // NaN or Inf
}

// copyPoints returns a copy of the float32 slice (points may
// be reused by the caller's buffer). Capped to prevent
// excessive allocation from degenerate input.
func copyPoints(pts []float32) []float32 {
	const maxPoints = 1 << 20 // ~4 MB
	return slices.Clone(pts[:min(len(pts), maxPoints)])
}

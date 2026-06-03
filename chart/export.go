package chart

import (
	"errors"
	"image"
	"image/png"
	"math"
	"os"

	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-glyph"
	"github.com/go-gui-org/go-gui/gui"
)

// themer is satisfied by chart views that expose their theme.
type themer interface {
	chartTheme() *theme.Theme
}

// Drawer is implemented by chart views that support
// headless export via ExportPNG.
type Drawer interface {
	Draw(dc *gui.DrawContext)
}

// ExportPNG renders a chart view to a PNG file at the given
// pixel dimensions. v must be a chart view created by Line,
// Bar, Area, Scatter, Pie, Combo, or other chart type.
func ExportPNG(v gui.View, width, height int, path string) error {
	d, ok := v.(Drawer)
	if !ok {
		return errors.New("chart: view does not implement Drawer")
	}
	if width <= 0 || height <= 0 {
		return errors.New("chart: dimensions must be positive")
	}
	const maxDim = 16384
	if width > maxDim || height > maxDim {
		return errors.New("chart: dimensions exceed 16384")
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill background from theme, fall back to white.
	bg := gui.Color{R: 255, G: 255, B: 255, A: 255}
	if t, ok := v.(themer); ok && t.chartTheme() != nil {
		c := t.chartTheme().Background
		if c.IsSet() {
			bg = c
		}
	}
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}

	backend := &imageBackend{
		target:   img,
		textures: make(map[glyph.TextureID]atlasPage),
	}

	textSys, err := glyph.NewTextSystem(backend)
	if err != nil {
		return err
	}
	defer textSys.Free()

	measurer := &pngTextMeasurer{textSys: textSys}
	dc := gui.NewDrawContext(float32(width), float32(height), measurer)
	d.Draw(dc)

	rasterizeTriBatches(img, dc.Batches())

	for _, t := range dc.Texts() {
		cfg := toGlyphConfig(t.Style)
		if t.Style.RotationRadians != 0 {
			layout, lerr := textSys.LayoutText(t.Text, cfg)
			if lerr == nil {
				textSys.DrawLayoutRotated(
					layout, t.X, t.Y, t.Style.RotationRadians)
			}
		} else {
			_ = textSys.DrawText(t.X, t.Y, t.Text, cfg)
		}
	}
	textSys.Commit()
	backend.flush()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	encErr := png.Encode(f, img)
	closeErr := f.Close()
	if encErr != nil {
		return encErr
	}
	return closeErr
}

// -----------------------------------------------------------
// Triangle rasterizer
// -----------------------------------------------------------

// 16x MSAA sub-pixel offsets (4×4 grid).
var msaaOffsets = [16][2]float32{
	{0.125, 0.125}, {0.375, 0.125}, {0.625, 0.125}, {0.875, 0.125},
	{0.125, 0.375}, {0.375, 0.375}, {0.625, 0.375}, {0.875, 0.375},
	{0.125, 0.625}, {0.375, 0.625}, {0.625, 0.625}, {0.875, 0.625},
	{0.125, 0.875}, {0.375, 0.875}, {0.625, 0.875}, {0.875, 0.875},
}

const msaaSamples = uint32(len(msaaOffsets))

// edgeInside returns true when all three edge values have the
// same sign (>= 0 or <= 0), indicating the sample is inside.
func edgeInside(e0, e1, e2 float32) bool {
	return (e0 >= 0 && e1 >= 0 && e2 >= 0) ||
		(e0 <= 0 && e1 <= 0 && e2 <= 0)
}

// rasterizeTriBatches fills triangles onto img using
// edge-function rasterization with 16x MSAA and src-over
// compositing. Hits are accumulated across all triangles in
// a batch before blending so that shared interior edges
// between tessellated triangles are invisible.
func rasterizeTriBatches(img *image.RGBA, batches []gui.DrawCanvasTriBatch) {
	bounds := img.Bounds()
	stride := img.Stride
	pix := img.Pix
	imgW := bounds.Max.X

	// Reusable per-row coverage buffer. Each entry holds
	// accumulated MSAA hits (clamped to msaaSamples).
	coverage := make([]uint8, imgW)

	for _, batch := range batches {
		cr := uint32(batch.Color.R)
		cg := uint32(batch.Color.G)
		cb := uint32(batch.Color.B)
		ca := uint32(batch.Color.A)
		tris := batch.Triangles

		if len(tris) < 6 {
			continue
		}

		// Compute bounding box of the entire batch.
		batchMinX, batchMinY := tris[0], tris[1]
		batchMaxX, batchMaxY := tris[0], tris[1]
		for i := 0; i+1 < len(tris); i += 2 {
			batchMinX = min(batchMinX, tris[i])
			batchMinY = min(batchMinY, tris[i+1])
			batchMaxX = max(batchMaxX, tris[i])
			batchMaxY = max(batchMaxY, tris[i+1])
		}

		// Skip batch if bounds contain NaN or Inf.
		if math.IsNaN(float64(batchMinX)) || math.IsNaN(float64(batchMinY)) ||
			math.IsNaN(float64(batchMaxX)) || math.IsNaN(float64(batchMaxY)) ||
			math.IsInf(float64(batchMinX), 0) || math.IsInf(float64(batchMinY), 0) ||
			math.IsInf(float64(batchMaxX), 0) || math.IsInf(float64(batchMaxY), 0) {
			continue
		}

		bMinX := max(int(math.Floor(float64(batchMinX))), bounds.Min.X)
		bMinY := max(int(math.Floor(float64(batchMinY))), bounds.Min.Y)
		bMaxX := min(int(math.Ceil(float64(batchMaxX))), bounds.Max.X-1)
		bMaxY := min(int(math.Ceil(float64(batchMaxY))), bounds.Max.Y-1)

		// Precompute per-triangle bounding boxes and edge deltas.
		type triInfo struct {
			mnX, mnY, mxX, mxY     int
			dx01, dy01             float32
			dx12, dy12             float32
			dx20, dy20             float32
			x0, y0, x1, y1, x2, y2 float32
		}
		nTris := len(tris) / 6
		infos := make([]triInfo, 0, nTris)
		for i := 0; i+5 < len(tris); i += 6 {
			x0, y0 := tris[i], tris[i+1]
			x1, y1 := tris[i+2], tris[i+3]
			x2, y2 := tris[i+4], tris[i+5]

			mnX := max(int(math.Floor(float64(min(x0, x1, x2)))), bounds.Min.X)
			mnY := max(int(math.Floor(float64(min(y0, y1, y2)))), bounds.Min.Y)
			mxX := min(int(math.Ceil(float64(max(x0, x1, x2)))), bounds.Max.X-1)
			mxY := min(int(math.Ceil(float64(max(y0, y1, y2)))), bounds.Max.Y-1)

			infos = append(infos, triInfo{
				mnX: mnX, mnY: mnY, mxX: mxX, mxY: mxY,
				dx01: x1 - x0, dy01: y1 - y0,
				dx12: x2 - x1, dy12: y2 - y1,
				dx20: x0 - x2, dy20: y0 - y2,
				x0: x0, y0: y0, x1: x1, y1: y1, x2: x2, y2: y2,
			})
		}

		// Process row by row: accumulate MSAA hits from all
		// triangles, then blend once per pixel.
		for py := bMinY; py <= bMaxY; py++ {
			// Clear coverage for this row's active range.
			clear(coverage[bMinX : bMaxX+1])

			for ti := range infos {
				t := &infos[ti]
				if py < t.mnY || py > t.mxY {
					continue
				}
				for px := t.mnX; px <= t.mxX; px++ {
					hits := uint8(0)
					for _, off := range msaaOffsets {
						sx := float32(px) + off[0]
						sy := float32(py) + off[1]
						e0 := t.dx01*(sy-t.y0) - t.dy01*(sx-t.x0)
						e1 := t.dx12*(sy-t.y1) - t.dy12*(sx-t.x1)
						e2 := t.dx20*(sy-t.y2) - t.dy20*(sx-t.x2)
						if edgeInside(e0, e1, e2) {
							hits++
						}
					}
					coverage[px] = min(coverage[px]+hits, uint8(msaaSamples))
				}
			}

			// Blend pixels with non-zero coverage.
			rowOff := py*stride + bMinX*4
			for px := bMinX; px <= bMaxX; px++ {
				if coverage[px] > 0 {
					sa := ca * uint32(coverage[px]) / msaaSamples
					blendPixel(pix, rowOff, cr, cg, cb, sa)
				}
				rowOff += 4
			}
		}
	}
}

// blendPixel applies src-over compositing at pix offset off.
func blendPixel(pix []byte, off int, sr, sg, sb, sa uint32) {
	if sa == 0 {
		return
	}
	if sa == 255 {
		pix[off] = uint8(sr)
		pix[off+1] = uint8(sg)
		pix[off+2] = uint8(sb)
		pix[off+3] = 255
		return
	}
	inv := 255 - sa
	pix[off] = uint8((sr*sa + uint32(pix[off])*inv) / 255)
	pix[off+1] = uint8((sg*sa + uint32(pix[off+1])*inv) / 255)
	pix[off+2] = uint8((sb*sa + uint32(pix[off+2])*inv) / 255)
	pix[off+3] = uint8(sa + uint32(pix[off+3])*inv/255)
}

// -----------------------------------------------------------
// Image-backed glyph DrawBackend
// -----------------------------------------------------------

type atlasPage struct {
	data          []byte
	width, height int
}

type pendingDraw struct {
	id       glyph.TextureID
	src, dst glyph.Rect
	color    glyph.Color
	xform    glyph.AffineTransform
	hasXform bool
}

// imageBackend implements glyph.DrawBackend, compositing glyph
// quads from atlas pages onto a target *image.RGBA.
type imageBackend struct {
	target   *image.RGBA
	textures map[glyph.TextureID]atlasPage
	nextID   glyph.TextureID
	pending  []pendingDraw
}

func (b *imageBackend) DPIScale() float32 { return 1.0 }

func (b *imageBackend) NewTexture(width, height int) glyph.TextureID {
	id := b.nextID
	b.nextID++
	b.textures[id] = atlasPage{
		data:   make([]byte, width*height*4),
		width:  width,
		height: height,
	}
	return id
}

func (b *imageBackend) UpdateTexture(id glyph.TextureID, data []byte) {
	if page, ok := b.textures[id]; ok {
		copy(page.data, data)
	}
}

func (b *imageBackend) DeleteTexture(id glyph.TextureID) {
	delete(b.textures, id)
}

func (b *imageBackend) DrawTexturedQuad(
	id glyph.TextureID, src, dst glyph.Rect, c glyph.Color,
) {
	b.pending = append(b.pending, pendingDraw{
		id: id, src: src, dst: dst, color: c,
	})
}

func (b *imageBackend) DrawFilledRect(dst glyph.Rect, c glyph.Color) {
	if c.A == 0 {
		return
	}
	img := b.target
	x0 := max(int(dst.X), 0)
	y0 := max(int(dst.Y), 0)
	x1 := min(int(dst.X+dst.Width), img.Bounds().Max.X)
	y1 := min(int(dst.Y+dst.Height), img.Bounds().Max.Y)
	cr, cg, cb, ca := uint32(c.R), uint32(c.G), uint32(c.B), uint32(c.A)
	for py := y0; py < y1; py++ {
		off := py*img.Stride + x0*4
		for px := x0; px < x1; px++ {
			blendPixel(img.Pix, off, cr, cg, cb, ca)
			off += 4
		}
	}
}

func (b *imageBackend) DrawTexturedQuadTransformed(
	id glyph.TextureID, src, dst glyph.Rect,
	c glyph.Color, t glyph.AffineTransform,
) {
	b.pending = append(b.pending, pendingDraw{
		id: id, src: src, dst: dst, color: c, xform: t,
		hasXform: true,
	})
}

// flush replays recorded draw commands after atlas textures
// have been committed via UpdateTexture.
func (b *imageBackend) flush() {
	for _, d := range b.pending {
		if d.hasXform {
			b.blitQuadTransformed(d.id, d.src, d.dst, d.color, d.xform)
		} else {
			b.blitQuad(d.id, d.src, d.dst, d.color)
		}
	}
	b.pending = b.pending[:0]
}

func (b *imageBackend) blitQuad(
	id glyph.TextureID, src, dst glyph.Rect, c glyph.Color,
) {
	page, ok := b.textures[id]
	if !ok || dst.Width <= 0 || dst.Height <= 0 {
		return
	}

	img := b.target
	imgW := img.Bounds().Max.X
	imgH := img.Bounds().Max.Y

	dstX0 := max(int(dst.X), 0)
	dstY0 := max(int(dst.Y), 0)
	dstX1 := min(int(math.Ceil(float64(dst.X+dst.Width))), imgW)
	dstY1 := min(int(math.Ceil(float64(dst.Y+dst.Height))), imgH)

	// Scale factors: src pixels per dst pixel.
	scaleX := src.Width / dst.Width
	scaleY := src.Height / dst.Height

	cr, cg, cb, ca := uint32(c.R), uint32(c.G), uint32(c.B), uint32(c.A)

	for py := dstY0; py < dstY1; py++ {
		// Nearest-neighbor sample in atlas.
		sy := int(src.Y + float32(py-int(dst.Y))*scaleY)
		if sy < 0 || sy >= page.height {
			continue
		}
		srcRowOff := sy * page.width * 4
		imgOff := py*img.Stride + dstX0*4

		for px := dstX0; px < dstX1; px++ {
			sx := int(src.X + float32(px-int(dst.X))*scaleX)
			if sx < 0 || sx >= page.width {
				imgOff += 4
				continue
			}
			sOff := srcRowOff + sx*4
			texA := uint32(page.data[sOff+3])
			if texA == 0 {
				imgOff += 4
				continue
			}

			// Glyph alpha modulated by tint alpha.
			ga := texA * ca / 255
			if ga == 0 {
				imgOff += 4
				continue
			}

			// Src-over with tint color.
			inv := 255 - ga
			img.Pix[imgOff] = uint8((cr*ga + uint32(img.Pix[imgOff])*inv) / 255)
			img.Pix[imgOff+1] = uint8((cg*ga + uint32(img.Pix[imgOff+1])*inv) / 255)
			img.Pix[imgOff+2] = uint8((cb*ga + uint32(img.Pix[imgOff+2])*inv) / 255)
			img.Pix[imgOff+3] = uint8(ga + uint32(img.Pix[imgOff+3])*inv/255)
			imgOff += 4
		}
	}
}

// blitQuadTransformed draws a textured quad with an affine
// transform applied. The transform maps local glyph coords to
// image coords. We iterate over the bounding box of the
// transformed quad and inverse-map each pixel to source coords.
func (b *imageBackend) blitQuadTransformed(
	id glyph.TextureID, src, dst glyph.Rect,
	c glyph.Color, xf glyph.AffineTransform,
) {
	page, ok := b.textures[id]
	if !ok || dst.Width <= 0 || dst.Height <= 0 {
		return
	}

	img := b.target
	imgW := img.Bounds().Max.X
	imgH := img.Bounds().Max.Y

	// Transform the four corners of the dst rect to find
	// the bounding box in image space.
	corners := [4][2]float32{
		{dst.X, dst.Y},
		{dst.X + dst.Width, dst.Y},
		{dst.X, dst.Y + dst.Height},
		{dst.X + dst.Width, dst.Y + dst.Height},
	}
	bMinX, bMinY := float32(math.MaxFloat32), float32(math.MaxFloat32)
	bMaxX, bMaxY := float32(-math.MaxFloat32), float32(-math.MaxFloat32)
	for _, corner := range corners {
		tx, ty := xf.Apply(corner[0], corner[1])
		bMinX = min(bMinX, tx)
		bMinY = min(bMinY, ty)
		bMaxX = max(bMaxX, tx)
		bMaxY = max(bMaxY, ty)
	}

	mnX := max(int(math.Floor(float64(bMinX))), 0)
	mnY := max(int(math.Floor(float64(bMinY))), 0)
	mxX := min(int(math.Ceil(float64(bMaxX))), imgW-1)
	mxY := min(int(math.Ceil(float64(bMaxY))), imgH-1)

	// Inverse transform: map image pixel back to local dst coords.
	det := xf.XX*xf.YY - xf.XY*xf.YX
	if det == 0 {
		return
	}
	invDet := 1.0 / det
	inv := glyph.AffineTransform{
		XX: xf.YY * invDet,
		XY: -xf.XY * invDet,
		YX: -xf.YX * invDet,
		YY: xf.XX * invDet,
		X0: (xf.XY*xf.Y0 - xf.YY*xf.X0) * invDet,
		Y0: (xf.YX*xf.X0 - xf.XX*xf.Y0) * invDet,
	}

	scaleX := src.Width / dst.Width
	scaleY := src.Height / dst.Height
	cr, cg, cb, ca := uint32(c.R), uint32(c.G), uint32(c.B), uint32(c.A)

	for py := mnY; py <= mxY; py++ {
		for px := mnX; px <= mxX; px++ {
			// Map image pixel to local dst space.
			lx, ly := inv.Apply(float32(px)+0.5, float32(py)+0.5)
			// Check if inside the dst rect.
			if lx < dst.X || lx >= dst.X+dst.Width ||
				ly < dst.Y || ly >= dst.Y+dst.Height {
				continue
			}
			// Map to source texture coords.
			sx := int(src.X + (lx-dst.X)*scaleX)
			sy := int(src.Y + (ly-dst.Y)*scaleY)
			if sx < 0 || sx >= page.width || sy < 0 || sy >= page.height {
				continue
			}
			sOff := (sy*page.width + sx) * 4
			texA := uint32(page.data[sOff+3])
			if texA == 0 {
				continue
			}
			ga := texA * ca / 255
			if ga == 0 {
				continue
			}
			imgOff := py*img.Stride + px*4
			iinv := 255 - ga
			img.Pix[imgOff] = uint8((cr*ga + uint32(img.Pix[imgOff])*iinv) / 255)
			img.Pix[imgOff+1] = uint8((cg*ga + uint32(img.Pix[imgOff+1])*iinv) / 255)
			img.Pix[imgOff+2] = uint8((cb*ga + uint32(img.Pix[imgOff+2])*iinv) / 255)
			img.Pix[imgOff+3] = uint8(ga + uint32(img.Pix[imgOff+3])*iinv/255)
		}
	}
}

// -----------------------------------------------------------
// TextMeasurer wrapping glyph.TextSystem
// -----------------------------------------------------------

type pngTextMeasurer struct {
	textSys *glyph.TextSystem
}

func (m *pngTextMeasurer) TextWidth(text string, style gui.TextStyle) float32 {
	w, err := m.textSys.TextWidth(text, toGlyphConfig(style))
	if err != nil {
		return 0
	}
	return w
}

func (m *pngTextMeasurer) TextHeight(text string, style gui.TextStyle) float32 {
	h, err := m.textSys.TextHeight(text, toGlyphConfig(style))
	if err != nil {
		return 0
	}
	return h
}

func (m *pngTextMeasurer) FontHeight(style gui.TextStyle) float32 {
	h, err := m.textSys.FontHeight(toGlyphConfig(style))
	if err != nil {
		return style.Size * 1.4
	}
	return h
}

func (m *pngTextMeasurer) FontAscent(style gui.TextStyle) float32 {
	met, err := m.textSys.FontMetrics(toGlyphConfig(style))
	if err != nil {
		return style.Size * 0.8
	}
	return met.Ascender
}

func (m *pngTextMeasurer) LayoutText(
	text string, style gui.TextStyle, wrapWidth float32,
) (glyph.Layout, error) {
	cfg := toGlyphConfig(style)
	if wrapWidth > 0 {
		cfg.Block.Width = wrapWidth
		cfg.Block.Wrap = glyph.WrapWord
	} else if wrapWidth < 0 {
		cfg.Block.Width = -wrapWidth
		cfg.Block.Wrap = glyph.WrapNone
	}
	return m.textSys.LayoutText(text, cfg)
}

// -----------------------------------------------------------
// Style conversion (inlined from glyphconv)
// -----------------------------------------------------------

func toGlyphConfig(s gui.TextStyle) glyph.TextConfig {
	align := glyph.AlignLeft
	switch s.Align {
	case gui.TextAlignCenter:
		align = glyph.AlignCenter
	case gui.TextAlignRight:
		align = glyph.AlignRight
	}
	return glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName:      s.Family,
			Size:          s.Size,
			Color:         glyph.Color{R: s.Color.R, G: s.Color.G, B: s.Color.B, A: s.Color.A},
			BgColor:       glyph.Color{R: s.BgColor.R, G: s.BgColor.G, B: s.BgColor.B, A: s.BgColor.A},
			Typeface:      s.Typeface,
			Underline:     s.Underline,
			Strikethrough: s.Strikethrough,
			LetterSpacing: s.LetterSpacing,
			StrokeWidth:   s.StrokeWidth,
			StrokeColor:   glyph.Color{R: s.StrokeColor.R, G: s.StrokeColor.G, B: s.StrokeColor.B, A: s.StrokeColor.A},
			Features:      s.Features,
		},
		Block: glyph.BlockStyle{
			Align:       align,
			Wrap:        glyph.WrapWord,
			Width:       -1,
			LineSpacing: s.LineSpacing,
		},
		Gradient: s.Gradient,
	}
}

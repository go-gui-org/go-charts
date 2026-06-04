module github.com/go-gui-org/go-charts

go 1.26.4

require (
	github.com/go-gui-org/go-glyph v1.9.0
	github.com/go-gui-org/go-gui v0.21.0
)

require (
	github.com/alecthomas/chroma/v2 v2.23.1 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/go-gl/gl v0.0.0-20260331235117-4566fea9a276 // indirect
	github.com/go-pdf/fpdf v0.9.0 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/tdewolff/parse/v2 v2.8.12 // indirect
	github.com/veandco/go-sdl2 v0.4.40 // indirect
	github.com/yuin/goldmark v1.8.2 // indirect
	github.com/yuin/goldmark-emoji v1.0.6 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace (
	github.com/go-gui-org/go-glyph => ../go-glyph
	github.com/go-gui-org/go-gui => ../go-gui
)

// Keep CI workflows in sync with the require versions above:
//   .github/workflows/ci.yml
//   .github/workflows/gallery.yml
// Update the `ref:` on the go-gui / go-glyph checkout steps.

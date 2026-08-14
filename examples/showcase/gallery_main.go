//go:build gallery

// Gallery generator. Builds with: go run -tags gallery ./examples/showcase
// Iterates every registered demo, renders each chart to PNG via
// chart.ExportPNG, and emits an index.html grouped by category.
package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

func main() {
	out := flag.String("out", "docs/gallery", "output directory")
	width := flag.Int("w", 960, "image width")
	height := flag.Int("h", 600, "image height")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}

	w := gui.NewWindow(gui.WindowCfg{Width: *width, Height: *height})

	type item struct {
		entry DemoEntry
		files []string
	}
	byGroup := map[string][]item{}
	var groups []string
	total := 0

	for _, e := range demoEntries {
		if e.Group == groupAnimation || e.Group == groupTypes {
			continue
		}
		files := renderEntry(w, e, *width, *height, *out)
		if len(files) == 0 {
			continue
		}
		if _, ok := byGroup[e.Group]; !ok {
			groups = append(groups, e.Group)
		}
		byGroup[e.Group] = append(byGroup[e.Group], item{entry: e, files: files})
		total += len(files)
	}
	sort.Strings(groups)

	idxPath := filepath.Join(*out, "index.html")
	f, err := os.Create(idxPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Fprintln(f, `<!doctype html><html><head><meta charset="utf-8">`)
	fmt.Fprintln(f, `<title>go-charts gallery</title>`)
	fmt.Fprintln(f, `<style>
body{font-family:system-ui,sans-serif;margin:24px;background:#fafafa;color:#222}
h1{margin:0 0 8px}h2{margin-top:32px;text-transform:capitalize;border-bottom:1px solid #ccc;padding-bottom:4px}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(420px,1fr));gap:16px}
figure{margin:0;background:#fff;border:1px solid #ddd;border-radius:6px;padding:8px}
figure img{width:100%;height:auto;display:block;border-radius:4px}
figcaption{font-size:13px;margin-top:6px}
figcaption b{display:block;font-size:14px}
figcaption span{color:#666}
nav a{display:inline-block;margin:0 8px 4px 0;font-size:13px}
</style></head><body>`)
	fmt.Fprintf(f, "<h1>go-charts gallery</h1><p>%d images across %d demos.</p>\n", total, len(demoEntries))

	fmt.Fprintln(f, `<nav>`)
	for _, g := range groups {
		fmt.Fprintf(f, `<a href="#%s">%s</a>`, g, html.EscapeString(g))
	}
	fmt.Fprintln(f, `</nav>`)

	for _, g := range groups {
		fmt.Fprintf(f, `<h2 id=%q>%s</h2><div class="grid">`+"\n", g, html.EscapeString(g))
		for _, it := range byGroup[g] {
			for _, name := range it.files {
				fmt.Fprintf(f,
					`<figure><img src=%q alt=%q><figcaption><b>%s</b><span>%s</span></figcaption></figure>`+"\n",
					name, html.EscapeString(it.entry.Label),
					html.EscapeString(it.entry.Label),
					html.EscapeString(it.entry.Summary),
				)
			}
		}
		fmt.Fprintln(f, `</div>`)
	}
	fmt.Fprintln(f, `</body></html>`)

	log.Printf("wrote %d images to %s", total, *out)
	log.Printf("index: %s", idxPath)
}

// galleryID converts a DemoEntry.ID (underscores) to the id form the
// demos pass to demoWithCode and demoWithCodeCharts (hyphens).
func galleryID(id string) string {
	return strings.ReplaceAll(id, "_", "-")
}

// renderEntry resolves a demo by ID, exports every chart the demo
// registered via demoWithCode/demoWithCodeCharts, and writes one PNG
// per chart. Panics in any single demo are recovered so one bad demo
// cannot abort the run.
func renderEntry(w *gui.Window, e DemoEntry, width, height int, outDir string) (files []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("skip %s: panic: %v", e.ID, r)
			files = nil
		}
	}()

	// The demo must run: demoWithCode registers its charts in
	// demoCharts as a side effect of building the view. Demos key
	// the registry by their hyphenated id (as passed to
	// demoWithCode), so translate the underscore form of DemoEntry.ID.
	_ = componentDemo(w, e.ID)
	charts := demoCharts[galleryID(e.ID)]
	for i, c := range charts {
		name := e.ID + ".png"
		if len(charts) > 1 {
			name = fmt.Sprintf("%s_%d.png", e.ID, i)
		}
		path := filepath.Join(outDir, name)
		if err := chart.ExportPNG(c, width, height, path); err != nil {
			log.Printf("skip %s: %v", e.ID, err)
			continue
		}
		files = append(files, name)
	}
	if len(charts) == 0 {
		log.Printf("skip %s: no chart.Drawer found", e.ID)
	}
	return files
}

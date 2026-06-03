package main

import "github.com/go-gui-org/go-gui/gui"

func demoBaseCfg(w *gui.Window) gui.View {
	source := `**BaseCfg** is embedded by every chart config struct
(` + "`LineCfg`" + `, ` + "`BarCfg`" + `, ` + "`PieCfg`" + `, etc.).
It provides the fields common to all chart types.

` + "```go" + `
type BaseCfg struct {
    ID      string         // unique identifier for the chart widget
    Title   string         // centered title above the plot area
    Sizing  gui.Sizing     // layout sizing policy (default: FillFill)
    Width   float32        // explicit width; 0 = fill from parent
    Height  float32        // explicit height; 0 = fill from parent
    Theme   *theme.Theme   // chart theme; nil = theme.Default()
    OnClick func(*gui.Layout, *gui.Event, *gui.Window)
    OnHover func(*gui.Layout, *gui.Event, *gui.Window)
    Version uint64         // bump to invalidate cached axes/ticks
}
` + "```" + `

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| ID | string | "" | Widget identifier for hit testing and scrolling |
| Title | string | "" | Centered above the plot area; omitted when empty |
| Sizing | gui.Sizing | FillFill | Controls how the chart fills its parent |
| Width | float32 | 0 | Explicit pixel width; 0 uses parent/window width |
| Height | float32 | 0 | Explicit pixel height; 0 uses parent/window height |
| Theme | *theme.Theme | Default() | Colors, fonts, padding, palette for rendering |
| OnClick | func | nil | Called on mouse click with layout and event |
| OnHover | func | nil | Called on mouse hover with layout and event |
| Version | uint64 | 0 | Bump to force axis/tick recomputation |

### Functional Options

BaseCfg fields can also be set via functional options:

` + "```go" + `
// Option functions for BaseCfg
chart.WithID("my-chart")
chart.WithTitle("Revenue")
chart.WithSize(800, 400)
chart.WithSizing(gui.FillFixed)
chart.WithTheme(myTheme)
` + "```" + `

### Usage

Embed BaseCfg in any chart constructor:

` + "```go" + `
chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        ID:     "revenue",
        Title:  "Monthly Revenue",
        Sizing: gui.FillFixed,
        Height: 350,
    },
    Series: []series.XY{ ... },
})
` + "```" + `

### Notes

- When both Width and Height are 0, the chart fills the
  window. Set explicit Height with FillFixed sizing when
  placing charts inside scrollable containers.
- Theme defaults to ` + "`theme.Default()`" + ` which inherits
  the current gui theme's colors.
- Bump Version when series data changes to force the chart
  to recompute axes and tick marks from the new data bounds.
`

	return typesDoc(w, "basecfg", source)
}

func demoSeriesXY(w *gui.Window) gui.View {
	source := `**series.XY** holds a named sequence of (X, Y) data points.
Used by Line, Area, and Scatter charts.

` + "```go" + `
type Point struct {
    X, Y float64
}

type XY struct {
    Points []Point  // exported; name and color via accessor
}
` + "```" + `

### Constructors

` + "```go" + `
// From explicit config
s := series.NewXY(series.XYCfg{
    Name:   "Revenue",
    Color:  gui.Hex(0x4E79A7),
    Points: []series.Point{
        {X: 1, Y: 12}, {X: 2, Y: 19}, {X: 3, Y: 15},
    },
})

// From parallel slices
s, err := series.XYFromSlices("Revenue",
    []float64{1, 2, 3},     // X values
    []float64{12, 19, 15},  // Y values
)
if err != nil {
    log.Fatal(err)
}

// From Y values only (X auto-indexed 0, 1, 2, ...)
s := series.XYFromYValues("Revenue",
    []float64{12, 19, 15},
)
` + "```" + `

### Methods

| Method | Returns | Description |
|--------|---------|-------------|
| Name() | string | Series name for legends |
| Len() | int | Number of data points |
| Color() | gui.Color | Series color; zero = use palette |
| Bounds() | minX, maxX, minY, maxY | Min/max across finite points |

### Notes

- Non-finite points (NaN, +/-Inf) are skipped by Bounds()
  and silently ignored during rendering.
- Color can be zero-valued; the chart falls back to the
  theme palette based on series index.
- Points is exported and can be modified directly.
`

	return typesDoc(w, "series-xy", source)
}

func demoSeriesCategory(w *gui.Window) gui.View {
	source := `**series.Category** holds labeled categorical data.
Used by Bar charts.

` + "```go" + `
type CategoryValue struct {
    Label string
    Value float64
}

type Category struct {
    Values []CategoryValue  // exported; name and color via accessor
}
` + "```" + `

### Constructors

` + "```go" + `
// From explicit config
s := series.NewCategory(series.CategoryCfg{
    Name:  "Q1",
    Color: gui.Hex(0x4E79A7),
    Values: []series.CategoryValue{
        {Label: "North", Value: 45},
        {Label: "South", Value: 32},
    },
})

// From a map (labels sorted alphabetically)
s := series.CategoryFromMap("Q1", map[string]float64{
    "North": 45,
    "South": 32,
})
` + "```" + `

### Methods

| Method | Returns | Description |
|--------|---------|-------------|
| Name() | string | Series name for legends |
| Len() | int | Number of category values |
| Color() | gui.Color | Series color; zero = use palette |

### Notes

- All series in a bar chart must have the same category
  labels in the same order. The first series defines the
  X-axis labels.
- CategoryFromMap sorts labels alphabetically for
  deterministic rendering order.
`

	return typesDoc(w, "series-cat", source)
}

func demoTheme(w *gui.Window) gui.View {
	source := `**theme.Theme** controls the visual style of charts:
colors, text styles, axis appearance, padding, and the
color palette used for series.

` + "```go" + `
type Theme struct {
    Background gui.Color      // chart background
    TitleStyle gui.TextStyle  // chart title text
    LabelStyle gui.TextStyle  // axis label text
    TickStyle  gui.TextStyle  // tick label text
    AxisColor  gui.Color      // axis line color
    AxisWidth  float32        // axis line width
    GridColor  gui.Color      // grid line color
    GridWidth  float32        // grid line width
    Palette    []gui.Color    // series color cycle
    PaddingTop    float32     // space above plot area
    PaddingRight  float32     // space right of plot area
    PaddingBottom float32     // space below plot area
    PaddingLeft   float32     // space left of plot area
}
` + "```" + `

### Default Values

` + "`theme.Default()`" + ` derives from ` + "`gui.CurrentTheme()`" + `:

| Field | Default |
|-------|---------|
| Background | theme's ColorBackground |
| TitleStyle | theme's B1 (bold heading) |
| LabelStyle | theme's TextStyleDef |
| TickStyle | theme's TextStyleDef |
| AxisColor | theme's ColorBorder |
| AxisWidth | 1 |
| GridColor | RGBA(128, 128, 128, 40) |
| GridWidth | 0.5 |
| Palette | Tableau 10 |
| PaddingTop | 40 |
| PaddingRight | 20 |
| PaddingBottom | 40 |
| PaddingLeft | 60 |

### Palettes

Three built-in palettes cycle colors for series:

` + "```go" + `
theme.Tableau10()  // bold, high-contrast (default)
theme.Pastel()     // soft, muted tones
theme.Vivid()      // saturated, high-energy
` + "```" + `

### Custom Theme

` + "```go" + `
t := theme.Default()
t.Palette = theme.Vivid()
t.GridColor = gui.RGBA(200, 200, 200, 60)
t.PaddingLeft = 80

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Theme: t,
    },
    Series: []series.XY{ ... },
})
` + "```" + `

### Global Default

` + "```go" + `
// Set once at startup; all charts inherit it
theme.SetDefault(myTheme)

// Revert to auto-generated from gui theme
theme.SetDefault(nil)
` + "```" + `
`

	return typesDoc(w, "theme", source)
}

func demoAxisLinear(w *gui.Window) gui.View {
	source := `**axis.Linear** is a linear numeric axis with nice-number
tick generation. Charts auto-create axes from series bounds
when not provided; supply a custom axis for explicit control.

` + "```go" + `
type LinearCfg struct {
    Title      string          // axis title (currently unused)
    Min        float64         // domain minimum
    Max        float64         // domain maximum
    AutoRange  bool            // expand domain to nice tick bounds
    TickFormat TickFormat       // custom tick label formatter
}
` + "```" + `

### Construction

` + "```go" + `
// Auto-ranged axis (expands domain to nice tick bounds)
yAxis := axis.NewLinear(axis.LinearCfg{AutoRange: true})
yAxis.SetRange(0, 97)  // ticks: 0, 20, 40, 60, 80, 100

// Fixed-range axis
yAxis := axis.NewLinear(axis.LinearCfg{Min: 0, Max: 100})

// Custom tick formatting
yAxis := axis.NewLinear(axis.LinearCfg{
    AutoRange: true,
    TickFormat: func(v float64) string {
        return fmt.Sprintf("$%.0f", v)
    },
})
` + "```" + `

### Methods

| Method | Description |
|--------|-------------|
| SetRange(min, max) | Update the data domain |
| Label() | Return axis title |
| Ticks(pxMin, pxMax) | Generate tick positions for pixel range |
| Transform(v, pxMin, pxMax) | Map data value to pixel |
| Inverse(px, pxMin, pxMax) | Map pixel back to data value |

### TickFormat

` + "```go" + `
// TickFormat converts a numeric value to a display string.
// When nil, the axis formats integers as "42" and
// decimals as compact %g notation.
type TickFormat func(float64) string
` + "```" + `

### Usage with Charts

` + "```go" + `
chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{Title: "Revenue"},
    YAxis: axis.NewLinear(axis.LinearCfg{
        AutoRange:  true,
        TickFormat: func(v float64) string {
            return fmt.Sprintf("$%.0fk", v)
        },
    }),
    Series: []series.XY{ ... },
})
` + "```" + `

### Notes

- AutoRange expands the domain to encompass the outermost
  nice tick values. Without it, data points may sit on the
  axis boundary.
- Tick generation targets ~8 ticks using the nice-number
  algorithm (1, 2, 5 multiples).
- Charts auto-create axes when XAxis/YAxis are nil; explicit
  axes override auto-creation entirely.
`

	return typesDoc(w, "axis-linear", source)
}

func demoSeriesXYZ(w *gui.Window) gui.View {
	source := `**series.XYZ** holds a named sequence of (X, Y, Z) data points.
Used by Bubble charts where Z controls marker size.

` + "```go" + `
type XYZPoint struct {
    X, Y, Z float64
}

type XYZ struct {
    Points []XYZPoint  // exported; name and color via accessor
}
` + "```" + `

### Constructors

` + "```go" + `
// From explicit config
s := series.NewXYZ(series.XYZCfg{
    Name:  "Cities",
    Color: gui.Hex(0x4E79A7),
    Points: []series.XYZPoint{
        {X: 12, Y: 75, Z: 331},
        {X: 42, Y: 83, Z: 83},
    },
})

// From parallel slices
s, err := series.XYZFromSlices("Cities",
    []float64{12, 42},   // X values
    []float64{75, 83},   // Y values
    []float64{331, 83},  // Z values (size)
)
` + "```" + `

### Methods

| Method | Returns | Description |
|--------|---------|-------------|
| Name() | string | Series name for legends |
| Len() | int | Number of data points |
| Color() | gui.Color | Series color; zero = use palette |
| Bounds() | minX, maxX, minY, maxY | Min/max X,Y across finite points |
| ZBounds() | minZ, maxZ | Min/max Z across finite points |

### Notes

- Z controls bubble marker size via sqrt scaling (area proportional to Z).
- Bounds() excludes Z since it maps to size, not position.
- Non-finite points (NaN, +/-Inf) are skipped in Bounds(),
  ZBounds(), and rendering.
`

	return typesDoc(w, "series-xyz", source)
}

func demoDataCSV(w *gui.Window) gui.View {
	source := `**CSV Parsers** convert CSV data into typed series via
` + "`io.Reader`" + `. All functions live in the ` + "`series`" + ` package.

### CSVCfg

` + "```go" + `
type CSVCfg struct {
    Delimiter  rune // 0 defaults to ','
    Comment    rune // 0 means no comment character
    HasHeader  bool // first row is column headers
    TrimSpace  bool // trim whitespace from fields
    SkipErrors bool // skip unparseable rows
}
` + "```" + `

### Column Selection

Columns are identified by index or header name:

` + "```go" + `
series.ColIdx(0)       // first column (0-based)
series.ColName("price") // column named "price"
` + "```" + `

### XY from CSV

` + "```go" + `
// Single series
r := strings.NewReader("x,y\n1,10\n2,20\n3,30")
s, err := series.XYFromCSV(r, "Revenue",
    series.ColName("x"), series.ColName("y"),
    series.CSVCfg{HasHeader: true})

// Multiple series sharing an X column
r = strings.NewReader("month,2024,2025\n1,10,12\n2,15,18")
ss, err := series.XYMultiFromCSV(r,
    series.ColName("month"),
    []series.Col{
        series.ColName("2024"),
        series.ColName("2025"),
    },
    nil, // names from headers
    series.CSVCfg{HasHeader: true})
` + "```" + `

### Category from CSV

` + "```go" + `
r := strings.NewReader("region,sales\nNorth,45\nSouth,32")
s, err := series.CategoryFromCSV(r, "Q1",
    series.ColName("region"), series.ColName("sales"),
    series.CSVCfg{HasHeader: true})
` + "```" + `

### OHLC from CSV

` + "```go" + `
s, err := series.OHLCFromCSV(r, "AAPL",
    series.OHLCCSVCfg{
        CSVCfg:     series.CSVCfg{HasHeader: true},
        TimeCol:    series.ColName("date"),
        OpenCol:    series.ColName("open"),
        HighCol:    series.ColName("high"),
        LowCol:     series.ColName("low"),
        CloseCol:   series.ColName("close"),
        VolumeCol:  series.ColName("volume"),
        TimeLayout: "2006-01-02",
    })
` + "```" + `

### Grid from CSV

First column = row labels, header row = column labels:

` + "```go" + `
r := strings.NewReader(",Mon,Tue,Wed\nAlice,1,2,3\nBob,4,5,6")
g, err := series.GridFromCSV(r, "schedule",
    series.CSVCfg{HasHeader: true})
` + "```" + `

### Notes

- All parsers accept ` + "`io.Reader`" + ` (files, strings, HTTP bodies).
- SkipErrors silently drops rows that fail to parse.
- Tab-separated: set ` + "`Delimiter: '\\t'`" + `.
`

	return typesDoc(w, "data-csv", source)
}

func demoDataJSON(w *gui.Window) gui.View {
	source := `**JSON Parsers** convert JSON arrays of objects into typed
series via ` + "`io.Reader`" + `. All functions live in the
` + "`series`" + ` package.

### XY from JSON

` + "```go" + `
data := ` + "`" + `[{"x": 1, "y": 10}, {"x": 2, "y": 20}]` + "`" + `
s, err := series.XYFromJSON(
    strings.NewReader(data), "Revenue", "x", "y")

// Multiple series sharing an X field
data = ` + "`" + `[{"month":1,"a":10,"b":100},{"month":2,"a":20,"b":200}]` + "`" + `
ss, err := series.XYMultiFromJSON(
    strings.NewReader(data), "month", []string{"a", "b"})
` + "```" + `

### Category from JSON

` + "```go" + `
data := ` + "`" + `[
    {"region": "North", "sales": 45},
    {"region": "South", "sales": 32}
]` + "`" + `
s, err := series.CategoryFromJSON(
    strings.NewReader(data), "Q1", "region", "sales")
` + "```" + `

### XYZ from JSON

` + "```go" + `
data := ` + "`" + `[{"x":12,"y":75,"z":331},{"x":42,"y":83,"z":83}]` + "`" + `
s, err := series.XYZFromJSON(
    strings.NewReader(data), "Cities", "x", "y", "z")
` + "```" + `

### OHLC from JSON

` + "```go" + `
s, err := series.OHLCFromJSON(r, "AAPL",
    series.OHLCJSONCfg{
        TimeLayout:  "2006-01-02",
        VolumeField: "vol",
    })
// Default field names: time, open, high, low, close
` + "```" + `

### Grid from JSON

` + "```go" + `
data := ` + "`" + `{
    "rows": ["Alice", "Bob"],
    "cols": ["Mon", "Tue"],
    "values": [[1, 2], [3, 4]]
}` + "`" + `
g, err := series.GridFromJSON(
    strings.NewReader(data), "schedule")
` + "```" + `

### TreeNode from JSON

` + "```go" + `
data := ` + "`" + `{
    "label": "root",
    "value": 0,
    "children": [
        {"label": "A", "value": 10},
        {"label": "B", "value": 0, "children": [
            {"label": "B1", "value": 5},
            {"label": "B2", "value": 3}
        ]}
    ]
}` + "`" + `
n, err := series.TreeNodeFromJSON(strings.NewReader(data))
` + "```" + `

### Notes

- All parsers decode arrays of objects (not arrays of arrays).
- JSON number fields accept both integers and floats.
- OHLCJSONCfg defaults: field names are "time", "open",
  "high", "low", "close"; layout is RFC3339.
- GridFromJSON validates dimensions match via NewGrid.
`

	return typesDoc(w, "data-json", source)
}

func typesDoc(w *gui.Window, id, source string) gui.View {
	return w.Markdown(gui.MarkdownCfg{
		ID:      "doc-" + id,
		Source:  source,
		Padding: gui.NoPadding,
		Style:   gui.DefaultMarkdownStyle(),
	})
}

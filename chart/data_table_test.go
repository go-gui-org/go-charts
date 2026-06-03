package chart

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/series"
)

func TestFmtVal(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{1.0, "1"},
		{3.14159, "3.14159"},
		{1234567.89, "1.23457e+06"},
		{0, "0"},
		{-42.5, "-42.5"},
	}
	for _, tt := range tests {
		got := fmtVal(tt.v)
		if got != tt.want {
			t.Errorf("fmtVal(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestXYTableDataSingle(t *testing.T) {
	ss := []series.XY{
		series.NewXY(series.XYCfg{
			Name:   "Sales",
			Points: []series.Point{{X: 1, Y: 10}, {X: 2, Y: 20}},
		}),
	}
	h, r := xyTableData(ss)
	if len(h) != 2 || h[0] != "X" || h[1] != "Sales" {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 2 {
		t.Fatalf("rows = %d, want 2", len(r))
	}
	if r[0][0] != "1" || r[0][1] != "10" {
		t.Errorf("row[0] = %v", r[0])
	}
}

func TestXYTableDataMulti(t *testing.T) {
	ss := []series.XY{
		series.NewXY(series.XYCfg{
			Name:   "A",
			Points: []series.Point{{X: 1, Y: 10}, {X: 2, Y: 20}},
		}),
		series.NewXY(series.XYCfg{
			Name:   "B",
			Points: []series.Point{{X: 1, Y: 30}},
		}),
	}
	h, r := xyTableData(ss)
	if len(h) != 3 {
		t.Fatalf("headers = %v, want 3 columns", h)
	}
	if h[1] != "A" || h[2] != "B" {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 2 {
		t.Fatalf("rows = %d, want 2", len(r))
	}
	// Second series has no second point.
	if r[1][2] != "" {
		t.Errorf("row[1][2] = %q, want empty", r[1][2])
	}
}

func TestXYTableDataEmpty(t *testing.T) {
	h, r := xyTableData(nil)
	if len(h) != 2 {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 0 {
		t.Errorf("rows = %d, want 0", len(r))
	}
}

func TestXYTableDataNaN(t *testing.T) {
	ss := []series.XY{
		series.NewXY(series.XYCfg{
			Points: []series.Point{{X: math.NaN(), Y: 5}},
		}),
	}
	h, r := xyTableData(ss)
	if len(h) != 2 {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 1 {
		t.Fatalf("rows = %d, want 1", len(r))
	}
	if r[0][0] != "NaN" {
		t.Errorf("row[0][0] = %q, want NaN", r[0][0])
	}
}

func TestCategoryTableData(t *testing.T) {
	ss := []series.Category{
		series.NewCategory(series.CategoryCfg{
			Name: "Q1",
			Values: []series.CategoryValue{
				{Label: "A", Value: 10},
				{Label: "B", Value: 20},
			},
		}),
		series.NewCategory(series.CategoryCfg{
			Name: "Q2",
			Values: []series.CategoryValue{
				{Label: "A", Value: 30},
				{Label: "B", Value: 40},
			},
		}),
	}
	h, r := categoryTableData(ss)
	if len(h) != 3 || h[0] != "Label" || h[1] != "Q1" || h[2] != "Q2" {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 2 {
		t.Fatalf("rows = %d, want 2", len(r))
	}
	if r[0][0] != "A" || r[0][1] != "10" || r[0][2] != "30" {
		t.Errorf("row[0] = %v", r[0])
	}
}

func TestSliceTableData(t *testing.T) {
	slices := []PieSlice{
		{Label: "Cat", Value: 30},
		{Label: "Dog", Value: 70},
	}
	h, r := sliceTableData(slices)
	if len(h) != 3 || h[2] != "%" {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 2 {
		t.Fatalf("rows = %d, want 2", len(r))
	}
	if r[0][2] != "30.0" {
		t.Errorf("pct = %q, want 30.0", r[0][2])
	}
}

func TestSliceTableDataZeroTotal(t *testing.T) {
	slices := []PieSlice{{Label: "X", Value: 0}}
	_, r := sliceTableData(slices)
	if len(r) != 1 {
		t.Fatalf("rows = %d, want 1", len(r))
	}
	if r[0][2] != "" {
		t.Errorf("pct = %q, want empty", r[0][2])
	}
}

func TestGaugeTableData(t *testing.T) {
	h, r := gaugeTableData(75, 0, 100)
	if len(h) != 2 {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 3 {
		t.Fatalf("rows = %d, want 3", len(r))
	}
	if r[0][1] != "75" {
		t.Errorf("value = %q, want 75", r[0][1])
	}
}

func TestGridTableData(t *testing.T) {
	g, err := series.NewGrid(series.GridCfg{
		Rows:   []string{"R1", "R2"},
		Cols:   []string{"C1", "C2", "C3"},
		Values: [][]float64{{1, 2, 3}, {4, 5, 6}},
	})
	if err != nil {
		t.Fatal(err)
	}
	h, r := gridTableData(g)
	if len(h) != 4 {
		t.Errorf("headers = %v, want 4 cols", h)
	}
	if len(r) != 2 {
		t.Fatalf("rows = %d, want 2", len(r))
	}
	if r[0][0] != "R1" || r[1][3] != "6" {
		t.Errorf("row[0]=%v, row[1]=%v", r[0], r[1])
	}
}

func TestHistogramTableData(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	h, r := histogramTableData(data, 2, nil, false)
	if len(h) != 2 || h[0] != "Bin Range" || h[1] != "Count" {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 2 {
		t.Fatalf("rows = %d, want 2", len(r))
	}
}

func TestBoxplotTableData(t *testing.T) {
	data := []BoxData{
		{Label: "A", Values: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}
	h, r := boxplotTableData(data)
	if len(h) != 7 {
		t.Errorf("headers = %v, want 7 cols", h)
	}
	if len(r) != 1 {
		t.Fatalf("rows = %d, want 1", len(r))
	}
	if r[0][0] != "A" {
		t.Errorf("name = %q, want A", r[0][0])
	}
}

func TestTreeTableData(t *testing.T) {
	roots := []series.TreeNode{
		{Label: "Root", Value: 100, Children: []series.TreeNode{
			{Label: "Child", Value: 60},
			{Label: "Other", Value: 40},
		}},
	}
	h, r := treeTableData(roots)
	if len(h) != 2 {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 3 {
		t.Fatalf("rows = %d, want 3", len(r))
	}
	if r[0][0] != "Root" {
		t.Errorf("row[0] = %q, want Root", r[0][0])
	}
	if r[1][0] != "  Child" {
		t.Errorf("row[1] = %q, want '  Child'", r[1][0])
	}
}

func TestSankeyTableData(t *testing.T) {
	nodes := []SankeyNode{{Label: "A"}, {Label: "B"}, {Label: "C"}}
	links := []SankeyLink{
		{Source: 0, Target: 1, Value: 10},
		{Source: 1, Target: 2, Value: 5},
	}
	nh, nRows, lh, lRows := sankeyTableData(nodes, links)
	if len(nh) != 2 || len(nRows) != 3 {
		t.Errorf("nodes: headers=%v, rows=%d", nh, len(nRows))
	}
	if len(lh) != 3 || len(lRows) != 2 {
		t.Errorf("links: headers=%v, rows=%d", lh, len(lRows))
	}
	if lRows[0][0] != "A" || lRows[0][1] != "B" {
		t.Errorf("link[0] = %v", lRows[0])
	}
}

func TestWaterfallTableData(t *testing.T) {
	vals := []WaterfallValue{
		{Label: "Start", Value: 100, IsTotal: true},
		{Label: "Add", Value: 20},
	}
	h, r := waterfallTableData(vals)
	if len(h) != 3 {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 2 {
		t.Fatalf("rows = %d, want 2", len(r))
	}
	if r[0][2] != "Total" || r[1][2] != "Delta" {
		t.Errorf("types = %q, %q", r[0][2], r[1][2])
	}
}

func TestRadarTableData(t *testing.T) {
	axes := []RadarAxis{{Label: "Speed"}, {Label: "Power"}}
	ss := []RadarSeries{
		{Name: "A", Values: []float64{80, 60}},
	}
	h, r := radarTableData(axes, ss)
	if len(h) != 2 || h[0] != "Axis" || h[1] != "A" {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 2 {
		t.Fatalf("rows = %d, want 2", len(r))
	}
	if r[0][0] != "Speed" || r[0][1] != "80" {
		t.Errorf("row[0] = %v", r[0])
	}
}

func TestSparklineTableDataValues(t *testing.T) {
	h, r := sparklineTableData([]float64{1, 2, 3}, series.XY{})
	if len(h) != 2 || h[0] != "Index" {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 3 {
		t.Fatalf("rows = %d, want 3", len(r))
	}
	if r[0][0] != "0" || r[0][1] != "1" {
		t.Errorf("row[0] = %v", r[0])
	}
}

func TestSparklineTableDataXY(t *testing.T) {
	xy := series.NewXY(series.XYCfg{
		Points: []series.Point{{X: 10, Y: 20}},
	})
	h, r := sparklineTableData(nil, xy)
	if len(h) != 2 || h[0] != "X" {
		t.Errorf("headers = %v", h)
	}
	if len(r) != 1 {
		t.Fatalf("rows = %d, want 1", len(r))
	}
}

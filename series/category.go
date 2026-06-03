package series

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/go-gui-org/go-gui/gui"
)

// CategoryValue represents a labeled value.
type CategoryValue struct {
	Label string
	Value float64
}

// Category is a series of labeled categorical data.
type Category struct {
	name   string
	color  gui.Color
	Values []CategoryValue
}

// CategoryCfg configures a category series.
type CategoryCfg struct {
	Name   string
	Color  gui.Color
	Values []CategoryValue
}

// NewCategory creates a new category data series.
func NewCategory(cfg CategoryCfg) Category {
	return Category{
		name:   cfg.Name,
		color:  cfg.Color,
		Values: cfg.Values,
	}
}

// CategoryFromMap creates a Category series from a map. Labels
// are sorted alphabetically for deterministic rendering.
func CategoryFromMap(name string, m map[string]float64) Category {
	vals := make([]CategoryValue, 0, len(m))
	for label, value := range m {
		vals = append(vals, CategoryValue{Label: label, Value: value})
	}
	slices.SortFunc(vals, func(a, b CategoryValue) int {
		return cmp.Compare(a.Label, b.Label)
	})
	return Category{name: name, Values: vals}
}

// String implements fmt.Stringer.
func (v CategoryValue) String() string {
	return fmt.Sprintf("%s:%.4g", v.Label, v.Value)
}

// Name implements Series.
func (s Category) Name() string { return s.name }

// Len implements Series.
func (s Category) Len() int { return len(s.Values) }

// Color implements Series.
func (s Category) Color() gui.Color { return s.color }

// String implements fmt.Stringer.
func (s Category) String() string {
	return fmt.Sprintf("Category{%q, %d values}",
		s.name, len(s.Values))
}

package chart

import (
	"testing"

	"github.com/go-gui-org/go-gui/gui"
)

func TestChartCreatesView(t *testing.T) {
	v := Chart(Cfg{BaseCfg: BaseCfg{ID: "test-chart"}})
	if v == nil {
		t.Fatal("Chart() returned nil")
	}
}

func TestChartDefaultSizing(t *testing.T) {
	cv := Chart(Cfg{BaseCfg: BaseCfg{ID: "test"}}).(*chartView)
	if cv.cfg.Sizing != gui.FillFill {
		t.Errorf("expected FillFill, got %v", cv.cfg.Sizing)
	}
}

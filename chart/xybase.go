package chart

import (
	"log/slog"

	"github.com/go-gui-org/go-gui/gui"
)

// xyBase holds state and event handlers shared by all XY-axis chart types.
// Embed in a chart view struct to inherit generateLayout and the seven internal
// event handlers. After constructing the view, set base and interaction to
// point into the view's own cfg (e.g. lv.base = &lv.cfg.BaseCfg;
// lv.interaction = &lv.cfg.InteractionCfg); this cannot be done inside xyBase
// because it does not know which concrete Cfg type embeds them.
type xyBase struct {
	base        *BaseCfg        // points into the chart's embedded BaseCfg
	interaction *InteractionCfg // points into the chart's embedded InteractionCfg

	// Per-frame state loaded from/saved to StateMap.
	hovering bool
	hoverPx  float32
	hoverPy  float32
	hidden   map[int]bool // legend toggle state
	lastLB   legendBounds // legend bounds for click/hover hit-testing
	lastPA   plotArea     // set in draw(); consumed by event handlers
	win      *gui.Window

	// zoomX/zoomY control which axes respond to pan/zoom/range-select.
	zoomX, zoomY bool

	// nearestFn returns true when the cursor is over a data element.
	// nil means no cursor upgrade (cursor stays arrow). Set in the
	// chart constructor.
	nearestFn func(px, py float32) bool

	// extraVersionFn contributes additional version bits beyond the
	// common set (e.g. scroll version for line/area charts with
	// AutoScroll). nil contributes 0.
	extraVersionFn func(w *gui.Window) uint64
}

// generateLayout builds a DrawCanvas layout with all common state
// loaded and all event handlers wired to the xyBase methods.
// drawFn is the chart's own draw callback.
func (xb *xyBase) generateLayout(
	w *gui.Window, drawFn func(*gui.DrawContext),
) gui.Layout {
	if xb.base == nil || xb.interaction == nil {
		slog.Error("xyBase.base and interaction must be set by chart constructor")
		return gui.Layout{}
	}
	c := xb.base
	hv := loadHover(w, c.ID, &xb.hovering, &xb.hoverPx, &xb.hoverPy)
	var hidV uint64
	xb.hidden, hidV = loadHiddenState(w, c.ID)
	xb.lastLB = loadLegendBounds(w, c.ID)
	xb.win = w
	zv := loadZoomVersion(w, c.ID)
	av := loadAnimVersion(w, c.ID)
	tv := loadTransitionVersion(w, c.ID)
	var ev uint64
	if xb.extraVersionFn != nil {
		ev = xb.extraVersionFn(w)
	}
	if c.Animate {
		startEntryAnimation(w, c.ID, c.AnimDuration)
	}
	width, height := resolveSize(c.Width, c.Height, w)
	return gui.DrawCanvas(gui.DrawCanvasCfg{
		ID:            c.ID,
		Sizing:        c.Sizing,
		Width:         width,
		Height:        height,
		Version:       c.Version + hv + hidV + zv + av + tv + ev,
		Clip:          true,
		OnDraw:        drawFn,
		OnClick:       xb.internalClick,
		OnHover:       xb.internalHover,
		OnMouseMove:   xb.internalMouseMove,
		OnMouseUp:     xb.internalMouseUp,
		OnMouseLeave:  xb.internalMouseLeave,
		OnMouseScroll: xb.internalScroll,
		OnGesture:     xb.internalGesture,
	}).GenerateLayout(w)
}

func (xb *xyBase) internalScroll(l *gui.Layout, e *gui.Event, w *gui.Window) {
	if !xb.interaction.EnableZoom {
		return
	}
	handleZoomScroll(w, l, e, xb.base.ID, xb.lastPA, xb.zoomX, xb.zoomY)
}

func (xb *xyBase) internalGesture(l *gui.Layout, e *gui.Event, w *gui.Window) {
	if !xb.interaction.EnableZoom {
		return
	}
	handleZoomGesture(w, l, e, xb.base.ID, xb.lastPA, xb.zoomX, xb.zoomY)
}

func (xb *xyBase) internalClick(l *gui.Layout, e *gui.Event, w *gui.Window) {
	if xb.interaction.EnableZoom && handleDoubleClickCheck(w, l, e, xb.base.ID) {
		e.IsHandled = true
		return
	}
	if idx := legendHitTest(xb.lastLB, e.MouseX, e.MouseY); idx >= 0 {
		e.IsHandled = true
		l.Shape.Version = toggleHidden(w, xb.base.ID, idx)
		return
	}
	if xb.base.OnClick != nil {
		xb.base.OnClick(l, e, w)
	}
}

func (xb *xyBase) internalMouseMove(l *gui.Layout, e *gui.Event, w *gui.Window) {
	if (xb.interaction.EnablePan || xb.interaction.EnableRangeSelect) &&
		handleDragHover(w, l, e, xb.base.ID, xb.lastPA,
			xb.interaction.EnablePan, xb.interaction.EnableRangeSelect,
			xb.zoomX, xb.zoomY) {
		return
	}
}

func (xb *xyBase) internalMouseUp(l *gui.Layout, e *gui.Event, w *gui.Window) {
	if xb.interaction.EnablePan || xb.interaction.EnableRangeSelect {
		handleDragEnd(w, l, e, xb.base.ID, xb.lastPA, xb.zoomX, xb.zoomY)
	}
}

func (xb *xyBase) internalHover(l *gui.Layout, e *gui.Event, w *gui.Window) {
	if isDragging(w, xb.base.ID) {
		xb.hovering = false
		saveHover(w, l, xb.base.ID, false, 0, 0)
		return
	}
	e.IsHandled = true
	xb.hoverPx = e.MouseX - l.Shape.X
	xb.hoverPy = e.MouseY - l.Shape.Y
	xb.hovering = true
	saveHover(w, l, xb.base.ID, true, xb.hoverPx, xb.hoverPy)
	if legendHitTest(xb.lastLB, xb.hoverPx, xb.hoverPy) >= 0 {
		w.SetMouseCursorPointingHand()
	} else if xb.nearestFn != nil && xb.nearestFn(xb.hoverPx, xb.hoverPy) {
		w.SetMouseCursorPointingHand()
	} else {
		w.SetMouseCursorArrow()
	}
	if xb.base.OnHover != nil {
		xb.base.OnHover(l, e, w)
	}
}

func (xb *xyBase) internalMouseLeave(l *gui.Layout, e *gui.Event, w *gui.Window) {
	e.IsHandled = true
	xb.hovering = false
	saveHover(w, l, xb.base.ID, false, 0, 0)
	w.SetMouseCursorArrow()
	if xb.base.OnMouseLeave != nil {
		xb.base.OnMouseLeave(l, e, w)
	}
}

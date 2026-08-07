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

func (xb *xyBase) internalScroll(ctx gui.EventCtx) {
	if !xb.interaction.EnableZoom {
		return
	}
	handleZoomScroll(ctx.Window, ctx.Layout, ctx.Event, xb.base.ID, xb.lastPA, xb.zoomX, xb.zoomY)
}

func (xb *xyBase) internalGesture(ctx gui.EventCtx) {
	if !xb.interaction.EnableZoom {
		return
	}
	handleZoomGesture(ctx.Window, ctx.Layout, ctx.Event, xb.base.ID, xb.lastPA, xb.zoomX, xb.zoomY)
}

func (xb *xyBase) internalClick(ctx gui.EventCtx) {
	if xb.interaction.EnableZoom && handleDoubleClickCheck(ctx.Window, ctx.Layout, ctx.Event, xb.base.ID) {
		ctx.Consume()
		return
	}
	if idx := legendHitTest(xb.lastLB, ctx.Event.MouseX, ctx.Event.MouseY); idx >= 0 {
		ctx.Consume()
		ctx.Layout.Shape.Version = toggleHidden(ctx.Window, xb.base.ID, idx)
		return
	}
	if xb.base.OnClick != nil {
		xb.base.OnClick(ctx)
	}
}

func (xb *xyBase) internalMouseMove(ctx gui.EventCtx) {
	if (xb.interaction.EnablePan || xb.interaction.EnableRangeSelect) &&
		handleDragHover(ctx.Window, ctx.Layout, ctx.Event, xb.base.ID, xb.lastPA,
			xb.interaction.EnablePan, xb.interaction.EnableRangeSelect,
			xb.zoomX, xb.zoomY) {
		return
	}
}

func (xb *xyBase) internalMouseUp(ctx gui.EventCtx) {
	if xb.interaction.EnablePan || xb.interaction.EnableRangeSelect {
		handleDragEnd(ctx.Window, ctx.Layout, ctx.Event, xb.base.ID, xb.lastPA, xb.zoomX, xb.zoomY)
	}
}

func (xb *xyBase) internalHover(ctx gui.EventCtx) {
	if isDragging(ctx.Window, xb.base.ID) {
		xb.hovering = false
		saveHover(ctx.Window, ctx.Layout, xb.base.ID, false, 0, 0)
		return
	}
	ctx.Consume()
	xb.hoverPx = ctx.Event.MouseX - ctx.Layout.Shape.X
	xb.hoverPy = ctx.Event.MouseY - ctx.Layout.Shape.Y
	xb.hovering = true
	saveHover(ctx.Window, ctx.Layout, xb.base.ID, true, xb.hoverPx, xb.hoverPy)
	if legendHitTest(xb.lastLB, xb.hoverPx, xb.hoverPy) >= 0 {
		ctx.Window.SetMouseCursorPointingHand()
	} else if xb.nearestFn != nil && xb.nearestFn(xb.hoverPx, xb.hoverPy) {
		ctx.Window.SetMouseCursorPointingHand()
	} else {
		ctx.Window.SetMouseCursorArrow()
	}
	if xb.base.OnHover != nil {
		xb.base.OnHover(ctx)
	}
}

func (xb *xyBase) internalMouseLeave(ctx gui.EventCtx) {
	ctx.Consume()
	xb.hovering = false
	saveHover(ctx.Window, ctx.Layout, xb.base.ID, false, 0, 0)
	ctx.Window.SetMouseCursorArrow()
	if xb.base.OnMouseLeave != nil {
		xb.base.OnMouseLeave(ctx)
	}
}

package render

import "github.com/go-gui-org/go-gui/gui"

// Text draws text at (x, y) using the given style.
func (c *Context) Text(x, y float32, text string, style gui.TextStyle) {
	c.DC.Text(x, y, text, style)
}

// TextWidth returns the measured width of text in the given style.
func (c *Context) TextWidth(text string, style gui.TextStyle) float32 {
	return c.DC.TextWidth(text, style)
}

// FontHeight returns the line height for the given text style.
func (c *Context) FontHeight(style gui.TextStyle) float32 {
	return c.DC.FontHeight(style)
}

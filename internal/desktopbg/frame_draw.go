package desktopbg

import (
	"image"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

// Draw implements the drawable contract for [lipgloss.Canvas.Compose] (lipgloss v2
// compositing is cell-based; types live in charmbracelet/ultraviolet).
func (f *Frame) Draw(scr uv.Screen, area image.Rectangle) {
	if f == nil {
		return
	}
	pix, tw, th := f.pix, f.tw, f.th
	if pix == nil || tw <= 0 || th <= 0 {
		return
	}
	b := image.Rect(0, 0, tw, th)
	DrawRasterCells(scr, pix, tw, th, area.Intersect(b))

	tmp := lipgloss.NewCanvas(tw, th)
	lipgloss.NewLayer(f.Overlay).Draw(tmp, tmp.Bounds())

	for y := 0; y < th; y++ {
		for x := 0; x < tw; {
			c := tmp.CellAt(x, y)
			// Wide continuation slots are Cell{} (IsZero). Never SetCell them: scr
			// already has the correct placeholders from the primary wide SetCell, and
			// writing Cell{} here triggers Line.Set's wide-cell cleanup and corrupts
			// the row (looks like a second bad draw on top of the correct one).
			if c == nil || c.IsZero() {
				x++
				continue
			}
			if keepWallpaperCell(c) {
				x++
				continue
			}
			cc := c.Clone()
			scr.SetCell(x, y, cc)
			if cc.Width > 1 {
				x += cc.Width
			} else {
				x++
			}
		}
	}
}

func keepWallpaperCell(c *uv.Cell) bool {
	if c == nil {
		return true
	}
	if c.Link.URL != "" {
		return false
	}
	// Default cleared / blank lipgloss cells: leave half-block wallpaper visible.
	if c.Width == 1 && c.Style.IsZero() && (c.Content == " " || c.Content == "") {
		return true
	}
	return false
}

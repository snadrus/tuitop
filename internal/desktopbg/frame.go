package desktopbg

import (
	"image"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// OverlayTopLines returns the first n logical lines of overlay (newline-separated),
// padding with empty lines at the end if overlay has fewer than n lines.
// CR in overlay is normalized to LF for splitting only.
func OverlayTopLines(overlay string, n int) string {
	if n <= 0 {
		return ""
	}
	norm := strings.ReplaceAll(overlay, "\r\n", "\n")
	parts := strings.Split(norm, "\n")
	if len(parts) > n {
		parts = parts[:n]
	}
	out := strings.Join(parts, "\n")
	if len(parts) < n {
		out += strings.Repeat("\n", n-len(parts))
	}
	return out
}

// Frame draws a Wallpaper then overlays a lipgloss/ANSI string without clearing the wallpaper.
// (uv.StyledString.Draw clears its bounds first, which erases a lower z-index layer in the same buffer.)
type Frame struct {
	Bg      *Wallpaper
	Overlay string
}

// Bounds implements lipgloss layer sizing.
func (f *Frame) Bounds() image.Rectangle {
	if f == nil || f.Bg == nil {
		return image.Rectangle{}
	}
	return f.Bg.Bounds()
}

// Draw implements uv.Drawable.
func (f *Frame) Draw(scr uv.Screen, area image.Rectangle) {
	if f == nil || f.Bg == nil {
		return
	}
	b := f.Bg.Bounds()
	if b.Empty() {
		return
	}
	f.Bg.Draw(scr, area.Intersect(b))

	tw, th := b.Dx(), b.Dy()
	tmp := uv.NewScreenBuffer(tw, th)
	tmp.Method = ansi.GraphemeWidth
	uv.NewStyledString(f.Overlay).Draw(tmp, b)

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

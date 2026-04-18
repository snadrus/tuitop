package desktopbg

import (
	"image"
	"strings"
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

// NewFrame snapshots the wallpaper raster and bundles it with the overlay string.
// All state needed for rendering is captured here (on the main goroutine) so that
// Draw—potentially called on a render goroutine—never races with SetTerminalSize.
func NewFrame(wp *Wallpaper, overlay string) *Frame {
	if wp == nil {
		return &Frame{Overlay: overlay}
	}
	pix, tw, th := wp.Raster()
	return &Frame{pix: pix, tw: tw, th: th, Overlay: overlay}
}

// Frame draws a pre-snapshotted wallpaper raster then overlays a lipgloss/ANSI
// string without clearing the wallpaper.
type Frame struct {
	pix     image.Image
	tw, th  int
	Overlay string
}

// Bounds implements lipgloss layer sizing.
func (f *Frame) Bounds() image.Rectangle {
	if f == nil {
		return image.Rectangle{}
	}
	return image.Rect(0, 0, f.tw, f.th)
}

package desktopbg

import (
	"image"
	"image/color"
	"sync"

	"github.com/gdamore/tcell/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

// Lower half block (U+2584): upper cell area uses background color, lower uses foreground.
const lowerHalfBlock = "▄"

// Wallpaper draws a source image using tcellblit's raster sizing and tcell.FromImageColor
// (same pairing as tcellblit.render).
type Wallpaper struct {
	Src  image.Image
	Fill bool // true = cover terminal (tcellblit fill mode), false = letterbox

	mu     sync.Mutex
	tw, th int
	pix    image.Image // termW × (2*termH)
}

// SetTerminalSize rebuilds the scaled raster when the terminal size changes.
func (w *Wallpaper) SetTerminalSize(tw, th int) {
	if w == nil || w.Src == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if tw <= 0 || th <= 0 {
		w.tw, w.th = tw, th
		w.pix = nil
		return
	}
	if w.tw == tw && w.th == th && w.pix != nil {
		return
	}
	w.tw, w.th = tw, th
	w.pix = HalfBlockRaster(w.Src, tw, th, w.Fill)
}

// Bounds implements sizing for lipgloss.NewLayer.
func (w *Wallpaper) Bounds() image.Rectangle {
	if w == nil {
		return image.Rectangle{}
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return image.Rect(0, 0, w.tw, w.th)
}

// Raster returns a snapshot of the current scaled pixel buffer and terminal
// dimensions under a single lock acquisition. Use this to capture a consistent
// (pix, tw, th) triple that won't be mutated by a concurrent SetTerminalSize.
func (w *Wallpaper) Raster() (pix image.Image, tw, th int) {
	if w == nil {
		return nil, 0, 0
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.pix, w.tw, w.th
}

// Draw implements uv.Drawable (wallpaper cells only; use Frame to composite lipgloss on top).
func (w *Wallpaper) Draw(scr uv.Screen, area image.Rectangle) {
	if w == nil {
		return
	}
	pix, tw, th := w.Raster()
	DrawRasterCells(scr, pix, tw, th, area)
}

// DrawRasterCells renders half-block wallpaper cells from a pre-scaled raster.
// pix must be tw × (2*th) pixels. This is the shared rendering core used by
// both Wallpaper.Draw and Frame.Draw.
func DrawRasterCells(scr uv.Screen, pix image.Image, tw, th int, area image.Rectangle) {
	if pix == nil || tw <= 0 || th <= 0 {
		return
	}
	pb := pix.Bounds()
	pw, ph := pb.Dx(), pb.Dy()
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			if x < 0 || y < 0 || x >= tw || y >= th {
				continue
			}
			y2 := y * 2
			if x >= pw || y2+1 >= ph {
				continue
			}
			px := pb.Min.X + x
			py := pb.Min.Y + y2
			tcUp := tcell.FromImageColor(pix.At(px, py))
			tcDown := tcell.FromImageColor(pix.At(px, py+1))
			c := &uv.Cell{
				Content: lowerHalfBlock,
				Width:   1,
				Style: uv.Style{
					Fg: tcellToColor(tcDown),
					Bg: tcellToColor(tcUp),
				},
			}
			scr.SetCell(x, y, c)
		}
	}
}

func tcellToColor(tc tcell.Color) color.Color {
	if !tc.Valid() {
		return color.NRGBA{A: 255}
	}
	h := tc.TrueColor().Hex()
	if h < 0 {
		return color.NRGBA{A: 255}
	}
	return color.NRGBA{
		R: uint8((h >> 16) & 0xff),
		G: uint8((h >> 8) & 0xff),
		B: uint8(h & 0xff),
		A: 255,
	}
}

package desktopbg

import (
	"image"
	stdraw "image/draw"
	"math"

	xdraw "golang.org/x/image/draw"

	_ "github.com/snadrus/tcellblit" // optional dep kept for tooling / version alignment
)

const (
	// defaultMaxSourceEdge caps the longest edge before the final high-quality scale.
	defaultMaxSourceEdge = 4096
	// progressiveTargetFactor: halve the source until its longest edge is at most
	// factor×max(dw,dh). Reduces moiré and kernel artifacts when shrinking huge images a lot.
	progressiveTargetFactor = 4
	// catmullPixelBudget: above this many pixels in the cover intermediate, use ApproxBiLinear
	// instead of CatmullRom (faster, less risk of subtle ringing at odd dimensions).
	catmullPixelBudget = 14_000_000
)

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// progressivePrefilter repeatedly halves the image (bilinear) while it is still
// much larger than the destination. Stabilizes resampling when the shrink ratio is huge.
func progressivePrefilter(src image.Image, dw, dh int) image.Image {
	if src == nil || dw <= 0 || dh <= 0 {
		return src
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return src
	}
	limit := maxInt(maxInt(dw, dh)*progressiveTargetFactor, 256)

	cur := src
	curB := b
	for range 24 {
		if maxInt(curB.Dx(), curB.Dy()) <= limit {
			break
		}
		nw := maxInt(1, curB.Dx()/2)
		nh := maxInt(1, curB.Dy()/2)
		tmp := image.NewNRGBA(image.Rect(0, 0, nw, nh))
		xdraw.ApproxBiLinear.Scale(tmp, tmp.Bounds(), cur, curB, xdraw.Src, nil)
		cur = tmp
		curB = tmp.Bounds()
	}
	return cur
}

// limitSourceMaxEdge shrinks the image (preserving aspect) so the longest edge is maxEdge.
func limitSourceMaxEdge(src image.Image, maxEdge int) image.Image {
	if maxEdge <= 0 || src == nil {
		return src
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return src
	}
	m := maxInt(w, h)
	if m <= maxEdge {
		return src
	}
	var nw, nh int
	if w >= h {
		nw = maxEdge
		nh = maxInt(1, int(math.Round(float64(h)*float64(maxEdge)/float64(w))))
	} else {
		nh = maxEdge
		nw = maxInt(1, int(math.Round(float64(w)*float64(maxEdge)/float64(h))))
	}
	out := image.NewNRGBA(image.Rect(0, 0, nw, nh))
	xdraw.ApproxBiLinear.Scale(out, out.Bounds(), src, b, xdraw.Src, nil)
	return out
}

func pickFinalScaler(rw, rh int) xdraw.Scaler {
	if rw <= 0 || rh <= 0 {
		return xdraw.ApproxBiLinear
	}
	if rw*rh > catmullPixelBudget {
		return xdraw.ApproxBiLinear
	}
	return xdraw.CatmullRom
}

// halfBlockResizeCover scales src to cover termW×(2×termH), center-crops.
func halfBlockResizeCover(src image.Image, termW, termH int) image.Image {
	dw, dh := termW, termH*2
	if dw <= 0 || dh <= 0 {
		return nil
	}

	work := progressivePrefilter(src, dw, dh)
	work = limitSourceMaxEdge(work, defaultMaxSourceEdge)

	wb := work.Bounds()
	sw, sh := float64(wb.Dx()), float64(wb.Dy())
	if sw <= 0 || sh <= 0 {
		return nil
	}
	twf, thf := float64(dw), float64(dh)
	k := twf / sw
	if k2 := thf / sh; k2 > k {
		k = k2
	}
	// Ceil so the scaled image never undershoots the crop window (avoids size-dependent edge glitches).
	rw := maxInt(1, int(math.Ceil(sw*k)))
	rh := maxInt(1, int(math.Ceil(sh*k)))

	scaled := image.NewNRGBA(image.Rect(0, 0, rw, rh))
	pickFinalScaler(rw, rh).Scale(scaled, scaled.Bounds(), work, wb, xdraw.Src, nil)

	out := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	sb := scaled.Bounds()
	x0 := sb.Min.X + maxInt(0, sb.Dx()-dw)/2
	y0 := sb.Min.Y + maxInt(0, sb.Dy()-dh)/2
	stdraw.Draw(out, out.Bounds(), scaled, image.Pt(x0, y0), stdraw.Src)
	return out
}

// halfBlockResizeContain fits src inside termW×(2×termH) with letterboxing on black.
func halfBlockResizeContain(src image.Image, termW, termH int) image.Image {
	dw, dh := termW, termH*2
	if dw <= 0 || dh <= 0 {
		return nil
	}

	work := progressivePrefilter(src, dw, dh)
	work = limitSourceMaxEdge(work, defaultMaxSourceEdge)

	wb := work.Bounds()
	sw, sh := float64(wb.Dx()), float64(wb.Dy())
	if sw <= 0 || sh <= 0 {
		return nil
	}
	twf, thf := float64(dw), float64(dh)
	k := twf / sw
	if k2 := thf / sh; k2 < k {
		k = k2
	}
	rw := maxInt(1, int(math.Floor(sw*k+0.5)))
	rh := maxInt(1, int(math.Floor(sh*k+0.5)))
	if rw > dw {
		rw = dw
	}
	if rh > dh {
		rh = dh
	}

	scaled := image.NewNRGBA(image.Rect(0, 0, rw, rh))
	pickFinalScaler(rw, rh).Scale(scaled, scaled.Bounds(), work, wb, xdraw.Src, nil)

	out := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	sx := (dw - rw) / 2
	sy := (dh - rh) / 2
	stdraw.Draw(out, image.Rect(sx, sy, sx+rw, sy+rh), scaled, scaled.Bounds().Min, stdraw.Src)
	return out
}

// HalfBlockRaster builds a termW×(2×termH) NRGBA image for half-block sampling.
func HalfBlockRaster(src image.Image, termW, termH int, fill bool) image.Image {
	if src == nil || termW <= 0 || termH <= 0 {
		return nil
	}
	if fill {
		return halfBlockResizeCover(src, termW, termH)
	}
	return halfBlockResizeContain(src, termW, termH)
}

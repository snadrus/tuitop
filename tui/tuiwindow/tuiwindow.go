package tuiwindow

import (
	"os/exec"
	"path"
	"sync/atomic"

	"github.com/snadrus/cview"
	"github.com/snadrus/tuitop/tui/cterm"
)

type TuiWindowCfg struct {
	closeHandler func(exitStatus int)
}

func WithCloseHandler(f func(exitStatus int)) func(*TuiWindowCfg) {
	return func(w *TuiWindowCfg) {
		w.closeHandler = f
	}
}

type CreateWindow func(cmd string, opts ...func(*TuiWindowCfg))

func MkCreateWindow(wm *cview.WindowManager) CreateWindow {
	return func(cmd string, opts ...func(*TuiWindowCfg)) {
		cfg := &TuiWindowCfg{}
		for _, opt := range opts {
			opt(cfg)
		}
		w := cview.NewWindow(cterm.NewTerminal(exec.Command(cmd)))
		_, file := path.Split(cmd)
		w.SetTitle(file)

		windowWidth, windowHt := 54, 12
		bestX, bestY := bestXY(wm, windowWidth, windowHt)
		w.SetRect(bestX, bestY, windowWidth, windowHt)
		wm.Add(w)
	}
}

var location int64 = 0

func bestXY(wm *cview.WindowManager, windowWidth, windowHt int) (x, y int) {
	// TODO fix me: GetRect() only usable after app starts. Need to know screen size.
	//_, _, screenW, screenH := wm.GetRect()
	loc := atomic.AddInt64(&location, 1)
	if loc > 5 {
		atomic.StoreInt64(&location, 0)
	}
	x = 1 + int(loc)*4
	y = int(loc) * 3
	/*if cannot do this, not ready: x+windowWidth > screenW || y+windowHt > screenH {
		atomic.StoreInt64(&location, 0)
		x = 1
		y = 1
	}*/
	return x, y // TODO do better than random: avoid overlap.
}

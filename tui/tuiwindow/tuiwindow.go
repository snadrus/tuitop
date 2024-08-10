package tuiwindow

import (
	"log"
	"os/exec"
	"path"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/snadrus/tuitop/deps/cterm"
	"github.com/snadrus/tuitop/deps/cview"
	"github.com/snadrus/tuitop/deps/tcellterm"
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
		cmdExec := exec.Command(cmd)
		t := cterm.NewTerminal(cmdExec)

		w := cview.NewWindow(t)
		_, file := path.Split(cmd)
		w.SetTitle(file)

		windowWidth, windowHt := 54, 12
		bestX, bestY := bestXY(wm, windowWidth, windowHt)
		w.SetRect(bestX, bestY, windowWidth, windowHt)
		wm.Add(w)
		t.Attach(func(ev tcell.Event) {
			switch ev.(type) {
			case *tcellterm.EventClosed:
				log.Printf("closed")
				if cfg.closeHandler != nil {
					cfg.closeHandler(cmdExec.ProcessState.ExitCode())
				}
				wm.Remove(w)
			}
		})
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

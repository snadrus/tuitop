package tuiwindow

import (
	"os/exec"
	"path"

	"golang.org/x/exp/rand"

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
		_, _, screenW, screenH := wm.GetRect()
		w := cview.NewWindow(cterm.NewTerminal(exec.Command(cmd)))
		_, file := path.Split(cmd)
		w.SetTitle(file)

		windowWidth, windowHt := 54, 12
		w.SetRect(rand.Int()%(screenW-windowWidth), rand.Int()%(screenH-windowHt), windowWidth, windowHt)
		wm.Add(w)
	}
}

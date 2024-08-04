package cterm

import (
	"os/exec"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/snadrus/cview"
	"github.com/snadrus/tuitop/tcellterm"
)

type Terminal struct {
	*cview.Box

	term    *tcellterm.VT
	running bool
	cmd     *exec.Cmd
	screen  tcell.Screen

	sync.Once
	sync.RWMutex
}

func NewTerminal(cmd *exec.Cmd) *Terminal {
	n := tcellterm.New()
	t := &Terminal{
		Box:  cview.NewBox(),
		term: n,
		cmd:  cmd,
	}
	return t
}

func (t *Terminal) Draw(s tcell.Screen) {
	if !t.GetVisible() {
		return
	}
	t.Box.Draw(s)
	t.Lock()
	defer t.Unlock()

	x, y, w, h := t.GetInnerRect()
	view := views.NewViewPort(s, x, y, w, h)
	t.term.SetSurface(view)
	t.screen = s

	t.Once.Do(func() {
		//t.term.Watch(t)		// TODO !
		go func() {
			//attr := &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 1}err := t.term.RunWithAttrs(t.cmd, attr);
			err := t.term.Start(t.cmd)
			if err != nil {
				panic(err)
			}
			t.running = false
		}()
	})
	t.term.Draw()
}

func (t *Terminal) HandleEvent(ev tcell.Event) bool {
	switch ev.(type) {
	case *views.EventWidgetContent:
		t.Draw(t.screen)
		if t.HasFocus() {

			x, y, style, vis := t.term.Cursor()
			gx, gy, _, _ := t.Box.GetInnerRect()
			if vis {
				t.screen.ShowCursor(x+gx, y+gy)
				t.screen.SetCursorStyle(style)
			} else {
				t.screen.HideCursor()
			}
		}
		t.screen.Show()
		return true
	}
	return false
}

func (t *Terminal) GetFocusable() cview.Focusable {
	return t
}

func (t *Terminal) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		t.term.HandleEvent(event)
	})
}

func (t *Terminal) MouseHandler() func(action cview.MouseAction, event *tcell.EventMouse, setFocus func(p cview.Primitive)) (consumed bool, capture cview.Primitive) {
	return t.WrapMouseHandler(func(action cview.MouseAction, event *tcell.EventMouse, setFocus func(p cview.Primitive)) (consumed bool, capture cview.Primitive) {
		return t.term.HandleEvent(event), nil
	})
}

package sub

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/snadrus/cview"
	"github.com/snadrus/tuitop/tcellterm"
)

type SubVt100 struct {
	*cview.Box
	term     *tcellterm.VT
	s        views.View
	termView *cview.Box

	lock     sync.Mutex //for: hasFocus
	hasFocus bool
}

// Update is the main event handler. It should only be called by the main thread
func (m *SubVt100) Update(ev tcell.Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r, string(debug.Stack()))
		}
	}()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlC:
			m.term.Close()
			return
		}
		if m.term != nil {
			m.term.HandleEvent(ev)
		}

		var padtblr [4]int
		padtblr[0], padtblr[1], padtblr[2], padtblr[3] = m.Box.GetPadding()
		m.term.Draw(m.s, padtblr)

		//m.s.Show()
	case *tcell.EventResize:
		if m.term != nil {
			//m.termView.SetRect(0, 2, -1, -1)
			w, h := ev.Size()
			m.term.Resize(w, h-2)
		}

		m.Box.Draw(m.s)
		var padtblr [4]int
		padtblr[0], padtblr[1], padtblr[2], padtblr[3] = m.Box.GetPadding()

		m.term.Draw(m.s, padtblr)
		return
	case *tcellterm.EventRedraw:
		// row, col, style, vis := m.term.Cursor()
		// if vis {
		// 	//m.s.SetCursorStyle(style)
		// 	m.s.ShowCursor(col, row+2)

		// } else {
		// 	m.s.HideCursor()
		// }
		// m.s.Show()
		return
	case *tcellterm.EventClosed:
		m.s.Clear()
		return
	case *tcell.EventPaste:
		m.term.HandleEvent(ev)
		return
	case *tcell.EventMouse:
		// Translate the coordinates to our global coordinates (y-2)
		x, y := ev.Position()
		if y-2 < 0 {
			// Event is outside our view
			return
		}
		e := tcell.NewEventMouse(x, y-2, ev.Buttons(), ev.Modifiers())
		m.term.HandleEvent(e)
		return

	case *tcellterm.EventPanic:
		m.s.Clear()
		fmt.Println(ev.Error)
	}
	return
}

// HandleEvent is used to handle events from underlying widgets. Any events
// which redraw must be executed in the main goroutine by posting the event back
// to tcell
func (m *SubVt100) HandleEvent(ev tcell.Event) {
	//m.s.PostEvent(ev)
}

func NewSubVT100(s views.View, logger io.Writer) *SubVt100 {
	m := &SubVt100{
		s: s,
	}
	m.term = tcellterm.New()
	m.term.Attach(m.HandleEvent)
	m.term.Logger = log.New(logger, "", log.Flags())

	cmd := exec.Command(os.Getenv("SHELL"))
	err := m.term.Start(cmd)
	if err != nil {
		panic(err)
	}
	return m
}

// / Interface satisfying'
// Draw draws this primitive onto the screen. Implementers can call the
// screen's ShowCursor() function but should only do so when they have focus.
// (They will need to keep track of this themselves.)
func (m *SubVt100) Draw(screen tcell.Screen) {
	m.Box.Draw(screen)
	padtblr := [4]int{}
	padtblr[0], padtblr[1], padtblr[2], padtblr[3] = m.Box.GetPadding()
	m.term.Draw(screen, padtblr)
	// TODO
	// handle cursor show/hide based on focus
	// TEST: do we need to handle repaint with sync?
}

// InputHandler returns a handler which receives key events when it has focus.
// It is called by the Application class.
//
// A value of nil may also be returned, in which case this primitive cannot
// receive focus and will not process any key events.
//
// The handler will receive the key event and a function that allows it to
// set the focus to a different primitive, so that future key events are sent
// to that primitive.
//
// The Application's Draw() function will be called automatically after the
// handler returns.
//
// The Box class provides functionality to intercept keyboard input. If you
// subclass from Box, it is recommended that you wrap your handler using
// Box.WrapInputHandler() so you inherit that functionality.
func (m *SubVt100) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		m.Box.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
			_ = m.term.HandleEvent(event)
		})
	}
}

// Focus is called by the application when the primitive receives focus.
// Implementers may call delegate() to pass the focus on to another primitive.
func (m *SubVt100) Focus() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.hasFocus = true
}

// Blur is called by the application when the primitive loses focus.
func (m *SubVt100) Blur() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.hasFocus = true
}

func (m *SubVt100) HasFocus() bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.hasFocus
}

// MouseHandler returns a handler which receives mouse events.
// It is called by the Application class.
//
// A value of nil may also be returned to stop the downward propagation of
// mouse events.
//
// The Box class provides functionality to intercept mouse events. If you
// subclass from Box, it is recommended that you wrap your handler using
// Box.WrapMouseHandler() so you inherit that functionality.
func (m *SubVt100) MouseHandler() func(action cview.MouseAction, event *tcell.EventMouse, setFocus func(p cview.Primitive)) (consumed bool, capture cview.Primitive) {
	return m.WrapMouseHandler(func(action cview.MouseAction, event *tcell.EventMouse, setFocus func(p cview.Primitive)) (consumed bool, capture cview.Primitive) {
		x, y := event.Position()

		// Should the window painter handle this instead?
		if !m.InRect(x, y) {
			return false, nil
		}
		top, bottom, left, right := m.Box.GetPadding()
		_ = bottom
		_ = right
		newEvt := tcell.NewEventMouse(x-left, y-top, event.Buttons(), event.Modifiers())
		m.term.HandleEvent(newEvt)
		return true, nil
	})
}

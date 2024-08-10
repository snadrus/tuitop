package termshim

import (
	"fmt"
	golog "log"
	"os"
	"os/exec"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/snadrus/tuitop/deps/cview"
	"github.com/snadrus/tuitop/deps/tcellterm"
)

var log = logToFile //func(s string) {}

var logToFile = func(s string) {
	appendfile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer appendfile.Close()
	appendfile.WriteString(time.Now().Local().Format("04:05") + " ")
	appendfile.WriteString(s)
}

type TermShim struct {
	*cview.Box
	vt        *tcellterm.VT
	initiated bool
	innerX    int
	innerY    int
	innerW    int
	innerH    int
	*tcSurf
}

func NewTermShim() *TermShim {
	// REMOVE ME
	f, _ := os.Create("log.txt")
	golog.SetOutput(f)

	// INIT
	vt := tcellterm.New()
	b := cview.NewBox()
	// TODO Setup Input & Mouse handler.
	// vt.Attach(m.HandleEvent)  <--->  b.Wrap....()
	l := golog.New(f, "", golog.Flags())
	vt.Logger = l
	return &TermShim{
		Box: b,
		vt:  vt,
	}
}

func (t *TermShim) SetRect(x, y, width, height int) {
	log(fmt.Sprintf("SetRect(%d, %d, %d, %d)\n", x, y, width, height))
	t.Box.SetRect(x, y, width, height)
	t.innerX, t.innerY, t.innerW, t.innerH = t.Box.GetInnerRect()

	if t.initiated {
		t.tcSurf.width, t.tcSurf.height = t.innerW, t.innerH
		t.tcSurf.x, t.tcSurf.y = t.innerX, t.innerY
		t.vt.Resize(t.innerW, t.innerH)
	} else {
		tcSurf := &tcSurf{
			width:  t.innerW,
			height: t.innerH,
			x:      t.innerX,
			y:      t.innerY}
		t.vt.SetSurface(tcSurf)
		t.tcSurf = tcSurf
	}
	// in-case we scribbled atop it while resizing.
	t.Box.SetRect(x, y, width, height)
}
func (t *TermShim) Draw(screen tcell.Screen) {
	log("Draw\n")
	if !t.initiated {
		log("INIT within Draw()\n")
		t.tcSurf.SetSurface(screen)
		cmd := exec.Command(os.Getenv("SHELL"))
		err := t.vt.Start(cmd)
		if err != nil {
			panic(err)
		}
		t.initiated = true
	}
	t.vt.Draw()
	t.Box.Draw(screen)
}

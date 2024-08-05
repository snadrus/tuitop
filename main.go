package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/exec"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/snadrus/cview"
	"github.com/snadrus/tuitop/verifications/6-cterm/cterm"
	"golang.org/x/exp/rand"
)

// Window returns the window page.
func Window() *cview.WindowManager {
	wm := cview.NewWindowManager()

	t := cview.NewTextView()
	t.SetText("Ctrl-C exits.")
	w3 := cview.NewWindow(t)
	w3.SetRect(12, 12, 17, 7)
	w3.SetBorder(false)
	w3.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		return 0, nil
	})

	AddBash(wm)
	AddBash(wm)
	return wm
}

var count = 1

func AddBash(wm *cview.WindowManager) {
	w := cview.NewWindow(cterm.NewTerminal(exec.Command("bash")))
	w.SetTitle(fmt.Sprintf("Window %d", count))
	count++
	w.SetRect(rand.Int()%12, rand.Int()%8, 54, 12)
	wm.Add(w)

}

var ColorWindowsBlue = tcell.NewRGBColor(49, 119, 217)

func Btm(wm *cview.WindowManager) cview.Primitive {
	btm := cview.NewFlex()
	btm.SetDirection(cview.FlexColumn)
	btn1 := cview.NewTextView()
	btn1.SetText(" TuiTop")
	btn1.SetBackgroundColor(tcell.ColorWhite)
	btn1.SetTextColor(tcell.ColorGreen)
	btm.AddItem(btn1, 8, 1, false)

	spc1 := cview.NewTextView()
	spc1.SetText(" ")
	spc1.SetBackgroundColor(ColorWindowsBlue)
	btm.AddItem(spc1, 1, 0, false)

	btn2 := cview.NewTextView()
	btn2.SetBackgroundColor(tcell.ColorBlack)
	btn2.SetTextColor(tcell.ColorGreen)
	btn2.SetHighlightForegroundColor(tcell.ColorGreen)
	btn2.SetText(">_")
	btn2.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		if action == cview.MouseLeftClick {
			AddBash(wm)
		}
		return action, event
	})
	btm.AddItem(btn2, 2, 0, false)

	drawer := cview.NewTextView()
	drawer.SetBackgroundColor(ColorWindowsBlue) //#3177d9
	btm.AddItem(drawer, 0, 100, false)
	tray := cview.NewTextView()
	tray.SetBackgroundColor(tcell.NewRGBColor(256-((256-49)/4), 256-((256-119)/4), 256-((256-217)/4))) //#3177d9
	tray.SetTextColor(tcell.ColorBlack)
	tray.SetText(getTime())

	go func() {
		for {
			tray.SetText(getTime())
			app.Draw()
			<-time.After(time.Second)
		}
	}()
	btm.AddItem(tray, 7, 0, false)
	return btm
}
func getTime() string {
	return time.Now().Format(" 15:04 ")
}

var TuiTopWindowColor = tcell.NewRGBColor(0, 106, 255)

// The application.
var app = cview.NewApplication()

// Starting point for the presentation.
func main() {
	defer app.HandlePanic()

	var debugPort int
	flag.IntVar(&debugPort, "debug", 0, "port to serve debug info")
	flag.Parse()

	if debugPort > 0 {
		go func() {
			log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", debugPort), nil))
		}()
	}

	app.EnableMouse(true)

	// Shortcuts to navigate the slides.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})

	wm := Window()
	btm := Btm(wm)
	flex := cview.NewFlex()
	flex.SetDirection(cview.FlexRow)
	flex.AddItem(wm, 0, 1, true)
	flex.AddItem(btm, 1, 0, false)

	// Start the application.
	app.SetRoot(flex, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

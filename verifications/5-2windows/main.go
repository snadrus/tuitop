package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"
	"github.com/snadrus/tuitop/sub"
)

func main() {
	// TODO use 4's code as a library to wrangle vt100s
	// use a simple CView with 2 windows that overlap each with a shell

	// The application.
	var app = cview.NewApplication()

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

	// Start the application.
	app.SetRoot(wm, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

const loremIpsumText = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

// Window returns the window page.
func Window() cview.Primitive {
	wm := cview.NewWindowManager()

	sub1 := sub.NewSubVT100()
	w1 := cview.NewWindow(sub1)
	w1.SetTitle("List")

	loremIpsum := cview.NewTextView()
	loremIpsum.SetText(loremIpsumText)

	w2 := cview.NewWindow(loremIpsum)
	w2.SetRect(7, 4, 12, 12)

	w2.SetTitle("Lorem Ipsum")

	t := cview.NewTextView()
	t.SetText("Ctrl-C exits.")
	w3 := cview.NewWindow(t)
	w3.SetRect(12, 12, 17, 7)
	w3.SetBorder(false)
	w3.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		return 0, nil
	})

	wm.Add(w1, w2, w3)

	return wm
}

var TuiTopWindowColor = tcell.NewRGBColor(0, 106, 255)

func NewTuiTopWindow(name string, in cview.Primitive, x, y, width, height int) cview.Primitive {
	flex := cview.NewFlex()
	flex.SetDirection(cview.FlexRow)

	topper := cview.NewFlex()
	{
		topper.SetBackgroundColor(TuiTopWindowColor)
		topper.SetDirection(cview.FlexColumn)
		topper.SetBorder(false)
		nameItem := cview.NewTextView()
		nameItem.SetText(" " + name + " ")
		topper.AddItem(nameItem, len(name)+2, 0, false)

		// growable section with foreground our "random color"
		growableSection := cview.NewFlex()
		growableSection.SetBorder(false)
		growableSection.SetDirection(cview.FlexRow)
		// TODO foreground our randm color for just the right length
		topper.AddItem(growableSection, 0, 1, true)

		// a fixed-size section with x and resize buttons
		fixedSection := cview.NewTextView()
		fixedSection.SetBorder(false)
		fixedSection.SetTextColor(tcell.ColorWhite)
		fixedSection.SetBackgroundColor(tcell.ColorDarkRed)
		fixedSection.SetText("X")
		// TODO handle X click

		// TODO handle resize click
		resizer := cview.NewTextView()
		resizer.SetBorder(false)
		resizer.SetTextColor(tcell.ColorWhite)
		resizer.SetBackgroundColor(TuiTopWindowColor)
		resizer.SetText(" ↗")

		topper.AddItem(resizer, 2, 0, false)

	}

	flex.AddItem(topper, 1, 0, false)
	flex.AddItem(in, 0, 1, true)
	w := cview.NewWindow(flex)
	// todo: smart sizing & locating
	w.SetRect(x, y, width, height+1)
	w.SetBorder(false)
	return w
}

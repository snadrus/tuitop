/*
A presentation of the cview package, implemented with cview.

# Navigation

The presentation will advance to the next slide when the primitive demonstrated
in the current slide is left (usually by hitting Enter or Escape). Additionally,
the following shortcuts can be used:

  - Ctrl-N: Jump to next slide
  - Ctrl-P: Jump to previous slide
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gdamore/tcell/v2"
	"github.com/snadrus/cview"
	"github.com/snadrus/tuitop/verifications/5-termshim/termshim"
)

const loremIpsumText = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

// Window returns the window page.
func Window() cview.Primitive {
	wm := cview.NewWindowManager()

	// TODO LATER make transparent atop a verifications/3-mydecor and use SetBeforeDrawFunc()
	//wm.SetBackgroundTransparent(true)

	list := cview.NewList()
	list.ShowSecondaryText(false)
	list.AddItem(cview.NewListItem("Item #1"))
	list.AddItem(cview.NewListItem("Item #2"))
	list.AddItem(cview.NewListItem("Item #3"))
	list.AddItem(cview.NewListItem("Item #4"))
	list.AddItem(cview.NewListItem("Item #5"))
	list.AddItem(cview.NewListItem("Item #6"))
	list.AddItem(cview.NewListItem("Item #7"))

	loremIpsum := cview.NewTextView()
	loremIpsum.SetText(loremIpsumText)

	w1 := cview.NewWindow(list)
	w1.SetRect(2, 2, 10, 7)
	w1.SetBackgroundColor(TuiTopWindowColor)

	w2 := cview.NewWindow(loremIpsum)
	w2.SetRect(7, 4, 12, 12)

	w1.SetTitle("List")
	w2.SetTitle("Lorem Ipsum")
	w2.SetBackgroundColor(TuiTopWindowColor)

	t := cview.NewTextView()
	t.SetText("Ctrl-C exits.")
	w3 := cview.NewWindow(t)
	w3.SetRect(12, 12, 17, 7)
	w3.SetBorder(false)
	w3.SetBackgroundColor(TuiTopWindowColor)
	w3.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		return 0, nil
	})

	w4 := cview.NewWindow(termshim.NewTermShim())
	w4.SetRect(4, 3, 54, 12)
	w4.SetTitle("TermShim")
	w4.SetBackgroundColor(TuiTopWindowColor)

	wm.Add(w1, w2, w3, w4)

	return wm
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

	// Start the application.
	app.SetRoot(wm, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

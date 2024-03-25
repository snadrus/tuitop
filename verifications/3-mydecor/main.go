/*
Looks like it was time to hack on cview directly.
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
)

const loremIpsumText = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

// Window returns the window page.
func Window() cview.Primitive {
	wm := cview.NewWindowManager()

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

	w1 := NewTuiTopWindow("List", list, 2, 2, 10, 7)

	w2 := NewTuiTopWindow("Lorem Ipsum", loremIpsum, 7, 4, 12, 12)

	t := cview.NewTextView()
	t.SetText("Ctrl-C exits.")
	w3 := NewTuiTopWindow("Text", t, 12, 12, 17, 7)

	wm.Add(w1, w2, w3)

	return wm
}

var TuiTopWindowColor = tcell.NewRGBColor(0, 106, 255)

func NewTuiTopWindow(name string, in cview.Primitive, x, y, width, height int) *cview.Window {
	flex := cview.NewFlex()
	flex.SetDirection(cview.FlexRow)
	flex.SetBorder(false)

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

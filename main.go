package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gdamore/tcell/v2"
	"github.com/snadrus/cview"
	"github.com/snadrus/tuitop/tui/tuiwm"
)

func main() {
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

	xp := tuiwm.MakeXP(app)

	// Start the application.
	app.SetRoot(xp, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gdamore/tcell/v2"
	"github.com/snadrus/tuitop/deps/cview"
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
		logBuf := bytes.NewBuffer(nil)
		log.SetOutput(logBuf)
		log.Println("Starting debug server on port", debugPort)
		go func() {
			log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", debugPort), http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(200)
					w.Write(logBuf.Bytes())
				})))
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

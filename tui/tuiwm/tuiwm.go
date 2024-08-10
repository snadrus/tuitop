package tuiwm

import (
	_ "net/http/pprof"
	"os"
	"os/exec"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/snadrus/cview"
	"github.com/snadrus/tuitop/tui/installer"
	"github.com/snadrus/tuitop/tui/tuiwindow"
)

// CreateWindowManager returns the window page.
func CreateWindowManager() *cview.WindowManager {
	wm := cview.NewWindowManager()

	t := cview.NewTextView()
	t.SetText("Ctrl-C exits.")
	w3 := cview.NewWindow(t)
	w3.SetRect(12, 12, 17, 7)
	w3.SetBorder(false)
	w3.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		return 0, nil
	})

	wm.Add(w3)
	createWindow := tuiwindow.MkCreateWindow(wm)
	AddShell(createWindow)
	AddShell(createWindow)

	// TODO replace wm to have Remove() and GetWindows() methods
	return wm
}

func AddShell(createWindow tuiwindow.CreateWindow) {
	sh := os.Getenv("SHELL")
	if sh == "" {
		var err error
		sh, err = exec.LookPath("bash")
		if err != nil {
			panic(err)
		}
	}

	createWindow(sh)
}

var ColorWindowsBlue = tcell.NewRGBColor(49, 119, 217)

func CreateBottomLayout(app *cview.Application, wm *cview.WindowManager, createWindow tuiwindow.CreateWindow) cview.Primitive {
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
			AddShell(createWindow)
		}
		return action, event
	})
	btm.AddItem(btn2, 2, 0, false)

	drawer := cview.NewTextView()
	drawer.SetBackgroundColor(ColorWindowsBlue) //#3177d9
	btm.AddItem(drawer, 0, 100, false)
	tray := cview.NewTextView()
	tray.SetBackgroundColor(Light(ColorWindowsBlue, 4)) //#3177d9
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

func Light(baseColor tcell.Color, howLight int32) tcell.Color {
	r, g, b := baseColor.RGB()
	r = 256 - r/howLight
	g = 256 - g/howLight
	b = 256 - b/howLight
	return tcell.NewRGBColor(r, g, b)
}

var TuiTopWindowColor = tcell.NewRGBColor(0, 106, 255)

type XP struct {
	*cview.Flex
	inst *installer.Installer
}

func MakeXP(app *cview.Application) cview.Primitive {
	wm := CreateWindowManager()
	createWindow := tuiwindow.MkCreateWindow(wm)
	btm := CreateBottomLayout(app, wm, createWindow)
	flex := cview.NewFlex()
	flex.SetDirection(cview.FlexRow)
	flex.AddItem(wm, 0, 1, true)
	flex.AddItem(btm, 1, 0, false)

	i := installer.New(createWindow)
	return &XP{flex, i}
}

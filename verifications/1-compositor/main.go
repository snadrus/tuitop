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
	"strconv"
	"strings"
	"time"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"
)

const (
	appInfo      = "Next slide: Ctrl-N  Previous: Ctrl-P  Exit: Ctrl-C  (Navigate with your keyboard and mouse)"
	listInfo     = "Next item: J, Down  Previous item: K, Up  Open context menu: Alt+Enter"
	textViewInfo = "Scroll down: J, Down, PageDown  Scroll up: K, Up, PageUp"
	sliderInfo   = "Decrease: H, J, Left, Down  Increase: K, L, Right, Up"
	formInfo     = "Next field: Tab  Previous field: Shift+Tab  Select: Enter"
	windowInfo   = "Windows may be dragged and resized using the mouse."
)

// Center returns a new primitive which shows the provided primitive in its
// center, given the provided primitive's size.
func Center(width, height int, p cview.Primitive) cview.Primitive {
	subFlex := cview.NewFlex()
	subFlex.SetDirection(cview.FlexRow)
	subFlex.AddItem(cview.NewBox(), 0, 1, false)
	subFlex.AddItem(p, height, 1, true)
	subFlex.AddItem(cview.NewBox(), 0, 1, false)

	flex := cview.NewFlex()
	flex.AddItem(cview.NewBox(), 0, 1, false)
	flex.AddItem(subFlex, width, 1, true)
	flex.AddItem(cview.NewBox(), 0, 1, false)

	return flex
}

// The width of the code window.
const codeWidth = 56

// Code returns a primitive which displays the given primitive (with the given
// size) on the left side and its source code on the right side.
func Code(p cview.Primitive, width, height int, code string) cview.Primitive {
	// Set up code view.
	codeView := cview.NewTextView()
	codeView.SetWrap(false)
	codeView.SetDynamicColors(true)
	codeView.SetPadding(1, 1, 2, 0)
	fmt.Fprint(codeView, code)

	f := cview.NewFlex()
	f.AddItem(Center(width, height, p), 0, 1, true)
	f.AddItem(codeView, codeWidth, 1, false)
	return f
}

const colorsText = `You can use color tags almost everywhere to partially change the color of a string.

Simply put a color name or hex string in square brackets to change the following characters' color.

H[green]er[white]e i[yellow]s a[darkcyan]n ex[red]amp[white]le.

The [black:red]tags [black:green]look [black:yellow]like [::u]this:

[cyan[]
[blue:yellow:u[]
[#00ff00[]`

// Colors demonstrates how to use colors.
func Colors(nextSlide func()) (title string, info string, content cview.Primitive) {
	tv := cview.NewTextView()
	tv.SetBorder(true)
	tv.SetTitle("A [red]c[yellow]o[green]l[darkcyan]o[blue]r[darkmagenta]f[red]u[yellow]l[white] [black:red]c[:yellow]o[:green]l[:darkcyan]o[:blue]r[:darkmagenta]f[:red]u[:yellow]l[white:] [::bu]title")
	tv.SetDynamicColors(true)
	tv.SetWordWrap(true)
	tv.SetText(colorsText)
	tv.SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})

	return "Colors", "", Center(44, 16, tv)
}

const logo = `
 ======= ===  === === ======== ===  ===  ===
===      ===  === === ===      ===  ===  ===
===      ===  === === ======   ===  ===  ===
===       ======  === ===       ===========
 =======    ==    === ========   ==== ====
`

const subtitle = "Terminal-based user interface toolkit"

// Cover returns the cover page.
func Cover(nextSlide func()) (title string, info string, content cview.Primitive) {
	// What's the size of the logo?
	lines := strings.Split(logo, "\n")
	logoWidth := 0
	logoHeight := len(lines)
	for _, line := range lines {
		if len(line) > logoWidth {
			logoWidth = len(line)
		}
	}
	logoBox := cview.NewTextView()
	logoBox.SetTextColor(tcell.ColorGreen.TrueColor())
	logoBox.SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	fmt.Fprint(logoBox, logo)

	// Create a frame for the subtitle and navigation infos.
	frame := cview.NewFrame(cview.NewBox())
	frame.SetBorders(0, 0, 0, 0, 0, 0)
	frame.AddText(subtitle, true, cview.AlignCenter, tcell.ColorDarkMagenta.TrueColor())

	// Create a Flex layout that centers the logo and subtitle.
	subFlex := cview.NewFlex()
	subFlex.AddItem(cview.NewBox(), 0, 1, false)
	subFlex.AddItem(logoBox, logoWidth, 1, true)
	subFlex.AddItem(cview.NewBox(), 0, 1, false)

	flex := cview.NewFlex()
	flex.SetDirection(cview.FlexRow)
	flex.AddItem(cview.NewBox(), 0, 7, false)
	flex.AddItem(subFlex, logoHeight, 1, true)
	flex.AddItem(frame, 0, 10, false)

	return "Start", appInfo, flex
}

// End shows the final slide.
func End(nextSlide func()) (title string, info string, content cview.Primitive) {
	textView := cview.NewTextView()
	textView.SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	url := "https://code.rocketnine.space/tslocum/cview"
	fmt.Fprint(textView, url)
	return "End", "", Center(len(url), 1, textView)
}

func demoBox(title string) *cview.Box {
	b := cview.NewBox()
	b.SetBorder(true)
	b.SetTitle(title)
	return b
}

// Flex demonstrates flexbox layout.
func Flex(nextSlide func()) (title string, info string, content cview.Primitive) {
	modalShown := false
	panels := cview.NewPanels()

	textView := cview.NewTextView()
	textView.SetBorder(true)
	textView.SetTitle("Flexible width, twice of middle column")
	textView.SetDoneFunc(func(key tcell.Key) {
		if modalShown {
			nextSlide()
			modalShown = false
		} else {
			panels.ShowPanel("modal")
			modalShown = true
		}
	})

	subFlex := cview.NewFlex()
	subFlex.SetDirection(cview.FlexRow)
	subFlex.AddItem(demoBox("Flexible width"), 0, 1, false)
	subFlex.AddItem(demoBox("Fixed height"), 15, 1, false)
	subFlex.AddItem(demoBox("Flexible height"), 0, 1, false)

	flex := cview.NewFlex()
	flex.AddItem(textView, 0, 2, true)
	flex.AddItem(subFlex, 0, 1, false)
	flex.AddItem(demoBox("Fixed width"), 30, 1, false)

	modal := cview.NewModal()
	modal.SetText("Resize the window to see the effect of the flexbox parameters")
	modal.AddButtons([]string{"Ok"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		panels.HidePanel("modal")
	})

	panels.AddPanel("flex", flex, true, true)
	panels.AddPanel("modal", modal, false, false)
	return "Flex", "", panels
}

const form = `[green]package[white] main

[green]import[white] (
    [red]"code.rocketnine.space/tslocum/cview"[white]
)

[green]func[white] [yellow]main[white]() {
    form := cview.[yellow]NewForm[white]().
        [yellow]AddInputField[white]([red]"First name:"[white], [red]""[white], [red]20[white], nil, nil).
        [yellow]AddInputField[white]([red]"Last name:"[white], [red]""[white], [red]20[white], nil, nil).
        [yellow]AddDropDown[white]([red]"Role:"[white], [][green]string[white]{
            [red]"Engineer"[white],
            [red]"Manager"[white],
            [red]"Administration"[white],
        }, [red]0[white], nil).
        [yellow]AddCheckBox[white]([red]"On vacation:"[white], false, nil).
        [yellow]AddPasswordField[white]([red]"Password:"[white], [red]""[white], [red]10[white], [red]'*'[white], nil).
        [yellow]AddButton[white]([red]"Save"[white], [yellow]func[white]() { [blue]/* Save data */[white] }).
        [yellow]AddButton[white]([red]"Cancel"[white], [yellow]func[white]() { [blue]/* Cancel */[white] })
    cview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](form, true).
        [yellow]Run[white]()
}`

// Form demonstrates forms.
func Form(nextSlide func()) (title string, info string, content cview.Primitive) {
	f := cview.NewForm()
	f.AddInputField("First name:", "", 20, nil, nil)
	f.AddInputField("Last name:", "", 20, nil, nil)
	f.AddDropDownSimple("Role:", 0, nil, "Engineer", "Manager", "Administration")
	f.AddPasswordField("Password:", "", 10, '*', nil)
	f.AddCheckBox("", "On vacation", false, nil)
	f.AddButton("Save", nextSlide)
	f.AddButton("Cancel", nextSlide)
	f.SetBorder(true)
	f.SetTitle("Employee Information")
	return "Form", formInfo, Code(f, 36, 15, form)
}

// Grid demonstrates the grid layout.
func Grid(nextSlide func()) (title string, info string, content cview.Primitive) {
	modalShown := false
	panels := cview.NewPanels()

	newPrimitive := func(text string) cview.Primitive {
		tv := cview.NewTextView()
		tv.SetTextAlign(cview.AlignCenter)
		tv.SetText(text)
		tv.SetDoneFunc(func(key tcell.Key) {
			if modalShown {
				nextSlide()
				modalShown = false
			} else {
				panels.ShowPanel("modal")
				modalShown = true
			}
		})
		return tv
	}

	menu := newPrimitive("Menu")
	main := newPrimitive("Main content")
	sideBar := newPrimitive("Side Bar")

	grid := cview.NewGrid()
	grid.SetRows(3, 0, 3)
	grid.SetColumns(0, -4, 0)
	grid.SetBorders(true)
	grid.AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, true)
	grid.AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false)
	grid.AddItem(main, 1, 0, 1, 3, 0, 0, false)
	grid.AddItem(sideBar, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false)
	grid.AddItem(main, 1, 1, 1, 1, 0, 100, false)
	grid.AddItem(sideBar, 1, 2, 1, 1, 0, 100, false)

	modal := cview.NewModal()
	modal.SetText("Resize the window to see how the grid layout adapts")
	modal.AddButtons([]string{"Ok"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		panels.HidePanel("modal")
	})

	panels.AddPanel("grid", grid, true, true)
	panels.AddPanel("modal", modal, false, false)

	return "Grid", "", panels
}

const inputField = `[green]package[white] main

[green]import[white] (
    [red]"strconv"[white]

    [red]"github.com/gdamore/tcell/v2"[white]
    [red]"code.rocketnine.space/tslocum/cview"[white]
)

[green]func[white] [yellow]main[white]() {
    input := cview.[yellow]NewInputField[white]().
        [yellow]SetLabel[white]([red]"Enter a number: "[white]).
        [yellow]SetAcceptanceFunc[white](
            cview.InputFieldInteger,
        ).[yellow]SetDoneFunc[white]([yellow]func[white](key tcell.Key) {
            text := input.[yellow]GetText[white]()
            n, _ := strconv.[yellow]Atoi[white](text)
            [blue]// We have a number.[white]
        })
    cview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](input, true).
        [yellow]Run[white]()
}`

// InputField demonstrates the InputField.
func InputField(nextSlide func()) (title string, info string, content cview.Primitive) {
	input := cview.NewInputField()
	input.SetLabel("Enter a number: ")
	input.SetAcceptanceFunc(cview.InputFieldInteger)
	input.SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	return "InputField", "", Code(input, 30, 1, inputField)
}

// Introduction returns a cview.List with the highlights of the cview package.
func Introduction(nextSlide func()) (title string, info string, content cview.Primitive) {
	list := cview.NewList()

	listText := [][]string{
		{"A Go package for terminal based UIs", "with a special focus on rich interactive widgets"},
		{"Based on github.com/gdamore/tcell", "Supports Linux, FreeBSD, MacOS and Windows"},
		{"Designed to be simple", `"Hello world" is less than 20 lines of code`},
		{"Good for data entry", `For charts, use "termui" - for low-level views, use "gocui" - ...`},
		{"Supports context menus", "Right click on one of these items or press Alt+Enter"},
		{"Extensive documentation", "Demo code is available for each widget"},
	}

	reset := func() {
		list.Clear()

		for i, itemText := range listText {
			item := cview.NewListItem(itemText[0])
			item.SetSecondaryText(itemText[1])
			item.SetShortcut(rune('1' + i))
			item.SetSelectedFunc(nextSlide)
			list.AddItem(item)
		}

		list.ContextMenuList().SetItemEnabled(3, false)
	}

	list.AddContextItem("Delete item", 'i', func(index int) {
		list.RemoveItem(index)

		if list.GetItemCount() == 0 {
			list.ContextMenuList().SetItemEnabled(0, false)
			list.ContextMenuList().SetItemEnabled(1, false)
		}
		list.ContextMenuList().SetItemEnabled(3, true)
	})

	list.AddContextItem("Delete all", 'a', func(index int) {
		list.Clear()

		list.ContextMenuList().SetItemEnabled(0, false)
		list.ContextMenuList().SetItemEnabled(1, false)
		list.ContextMenuList().SetItemEnabled(3, true)
	})

	list.AddContextItem("", 0, nil)

	list.AddContextItem("Reset", 'r', func(index int) {
		reset()

		list.ContextMenuList().SetItemEnabled(0, true)
		list.ContextMenuList().SetItemEnabled(1, true)
		list.ContextMenuList().SetItemEnabled(3, false)
	})

	reset()
	return "Introduction", listInfo, Center(80, 12, list)
}

const sliderCode = `[green]package[white] main

[green]import[white] (
    [red]"fmt"[white]

    [red]"github.com/gdamore/tcell/v2"[white]
    [red]"code.rocketnine.space/tslocum/cview"[white]
)

[green]func[white] [yellow]main[white]() {
    slider := cview.[yellow]NewSlider[white]()
    slider.[yellow]SetLabel[white]([red]"Volume:   0%"[white])
    slider.[yellow][yellow]SetChangedFunc[white]([yellow]func[white](key tcell.Key) {
        label := fmt.[yellow]Sprintf[white]("Volume: %3d%%", value)
        slider.[yellow]SetLabel[white](label)
    })
    slider.[yellow][yellow]SetDoneFunc[white]([yellow]func[white](key tcell.Key) {
        [yellow]nextSlide[white]()
    })
    app := cview.[yellow]NewApplication[white]()
    app.[yellow]SetRoot[white](slider, true)
    app.[yellow]Run[white]()
}`

// Slider demonstrates the Slider.
func Slider(nextSlide func()) (title string, info string, content cview.Primitive) {
	slider := cview.NewSlider()
	slider.SetLabel("Volume:   0%")
	slider.SetChangedFunc(func(value int) {
		slider.SetLabel(fmt.Sprintf("Volume: %3d%%", value))
	})
	slider.SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	return "Slider", sliderInfo, Code(slider, 30, 1, sliderCode)
}

const tableData = `OrderDate|Region|Rep|Item|Units|UnitCost|Total
1/6/2017|East|Jones|Pencil|95|1.99|189.05
1/23/2017|Central|Kivell|Binder|50|19.99|999.50
2/9/2017|Central|Jardine|Pencil|36|4.99|179.64
2/26/2017|Central|Gill|Pen|27|19.99|539.73
3/15/2017|West|Sorvino|Pencil|56|2.99|167.44
4/1/2017|East|Jones|Binder|60|4.99|299.40
4/18/2017|Central|Andrews|Pencil|75|1.99|149.25
5/5/2017|Central|Jardine|Pencil|90|4.99|449.10
5/22/2017|West|Thompson|Pencil|32|1.99|63.68
6/8/2017|East|Jones|Binder|60|8.99|539.40
6/25/2017|Central|Morgan|Pencil|90|4.99|449.10
7/12/2017|East|Howard|Binder|29|1.99|57.71
7/29/2017|East|Parent|Binder|81|19.99|1,619.19
8/15/2017|East|Jones|Pencil|35|4.99|174.65
9/1/2017|Central|Smith|Desk|2|125.00|250.00
9/18/2017|East|Jones|Pen Set|16|15.99|255.84
10/5/2017|Central|Morgan|Binder|28|8.99|251.72
10/22/2017|East|Jones|Pen|64|8.99|575.36
11/8/2017|East|Parent|Pen|15|19.99|299.85
11/25/2017|Central|Kivell|Pen Set|96|4.99|479.04
12/12/2017|Central|Smith|Pencil|67|1.29|86.43
12/29/2017|East|Parent|Pen Set|74|15.99|1,183.26
1/15/2018|Central|Gill|Binder|46|8.99|413.54
2/1/2018|Central|Smith|Binder|87|15.00|1,305.00
2/18/2018|East|Jones|Binder|4|4.99|19.96
3/7/2018|West|Sorvino|Binder|7|19.99|139.93
3/24/2018|Central|Jardine|Pen Set|50|4.99|249.50
4/10/2018|Central|Andrews|Pencil|66|1.99|131.34
4/27/2018|East|Howard|Pen|96|4.99|479.04
5/14/2018|Central|Gill|Pencil|53|1.29|68.37
5/31/2018|Central|Gill|Binder|80|8.99|719.20
6/17/2018|Central|Kivell|Desk|5|125.00|625.00
7/4/2018|East|Jones|Pen Set|62|4.99|309.38
7/21/2018|Central|Morgan|Pen Set|55|12.49|686.95
8/7/2018|Central|Kivell|Pen Set|42|23.95|1,005.90
8/24/2018|West|Sorvino|Desk|3|275.00|825.00
9/10/2018|Central|Gill|Pencil|7|1.29|9.03
9/27/2018|West|Sorvino|Pen|76|1.99|151.24
10/14/2018|West|Thompson|Binder|57|19.99|1,139.43
10/31/2018|Central|Andrews|Pencil|14|1.29|18.06
11/17/2018|Central|Jardine|Binder|11|4.99|54.89
12/4/2018|Central|Jardine|Binder|94|19.99|1,879.06
12/21/2018|Central|Andrews|Binder|28|4.99|139.72`

const tableBasic = `[green]func[white] [yellow]main[white]() {
    table := cview.[yellow]NewTable[white]().
        [yellow]SetFixed[white]([red]1[white], [red]1[white])
    [yellow]for[white] row := [red]0[white]; row < [red]40[white]; row++ {
        [yellow]for[white] column := [red]0[white]; column < [red]7[white]; column++ {
            color := tcell.ColorWhite.TrueColor()
            [yellow]if[white] row == [red]0[white] {
                color = tcell.ColorYellow.TrueColor()
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] {
                color = tcell.ColorDarkCyan.TrueColor()
            }
            align := cview.AlignLeft
            [yellow]if[white] row == [red]0[white] {
                align = cview.AlignCenter
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] || column >= [red]4[white] {
                align = cview.AlignRight
            }
            table.[yellow]SetCell[white](row,
                column,
                &cview.TableCell{
                    Text:  [red]"..."[white],
                    Color: color,
                    Align: align,
                })
        }
    }
    cview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](table, true).
        [yellow]Run[white]()
}`

const tableSeparator = `[green]func[white] [yellow]main[white]() {
    table := cview.[yellow]NewTable[white]().
        [yellow]SetFixed[white]([red]1[white], [red]1[white]).
        [yellow]SetSeparator[white](Borders.Vertical)
    [yellow]for[white] row := [red]0[white]; row < [red]40[white]; row++ {
        [yellow]for[white] column := [red]0[white]; column < [red]7[white]; column++ {
            color := tcell.ColorWhite.TrueColor()
            [yellow]if[white] row == [red]0[white] {
                color = tcell.ColorYellow.TrueColor()
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] {
                color = tcell.ColorDarkCyan.TrueColor()
            }
            align := cview.AlignLeft
            [yellow]if[white] row == [red]0[white] {
                align = cview.AlignCenter
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] || column >= [red]4[white] {
                align = cview.AlignRight
            }
            table.[yellow]SetCell[white](row,
                column,
                &cview.TableCell{
                    Text:  [red]"..."[white],
                    Color: color,
                    Align: align,
                })
        }
    }
    cview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](table, true).
        [yellow]Run[white]()
}`

const tableBorders = `[green]func[white] [yellow]main[white]() {
    table := cview.[yellow]NewTable[white]().
        [yellow]SetFixed[white]([red]1[white], [red]1[white]).
        [yellow]SetBorders[white](true)
    [yellow]for[white] row := [red]0[white]; row < [red]40[white]; row++ {
        [yellow]for[white] column := [red]0[white]; column < [red]7[white]; column++ {
            color := tcell.ColorWhite.TrueColor()
            [yellow]if[white] row == [red]0[white] {
                color = tcell.ColorYellow.TrueColor()
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] {
                color = tcell.ColorDarkCyan.TrueColor()
            }
            align := cview.AlignLeft
            [yellow]if[white] row == [red]0[white] {
                align = cview.AlignCenter
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] || column >= [red]4[white] {
                align = cview.AlignRight
            }
            table.[yellow]SetCell[white](row,
                column,
                &cview.TableCell{
                    Text:  [red]"..."[white],
                    Color: color,
                    Align: align,
                })
        }
    }
    cview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](table, true).
        [yellow]Run[white]()
}`

const tableSelectRow = `[green]func[white] [yellow]main[white]() {
    table := cview.[yellow]NewTable[white]().
        [yellow]SetFixed[white]([red]1[white], [red]1[white]).
        [yellow]SetSelectable[white](true, false)
    [yellow]for[white] row := [red]0[white]; row < [red]40[white]; row++ {
        [yellow]for[white] column := [red]0[white]; column < [red]7[white]; column++ {
            color := tcell.ColorWhite.TrueColor()
            [yellow]if[white] row == [red]0[white] {
                color = tcell.ColorYellow.TrueColor()
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] {
                color = tcell.ColorDarkCyan.TrueColor()
            }
            align := cview.AlignLeft
            [yellow]if[white] row == [red]0[white] {
                align = cview.AlignCenter
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] || column >= [red]4[white] {
                align = cview.AlignRight
            }
            table.[yellow]SetCell[white](row,
                column,
                &cview.TableCell{
                    Text:          [red]"..."[white],
                    Color:         color,
                    Align:         align,
                    NotSelectable: row == [red]0[white] || column == [red]0[white],
                })
        }
    }
    cview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](table, true).
        [yellow]Run[white]()
}`

const tableSelectColumn = `[green]func[white] [yellow]main[white]() {
    table := cview.[yellow]NewTable[white]().
        [yellow]SetFixed[white]([red]1[white], [red]1[white]).
        [yellow]SetSelectable[white](false, true)
    [yellow]for[white] row := [red]0[white]; row < [red]40[white]; row++ {
        [yellow]for[white] column := [red]0[white]; column < [red]7[white]; column++ {
            color := tcell.ColorWhite.TrueColor()
            [yellow]if[white] row == [red]0[white] {
                color = tcell.ColorYellow.TrueColor()
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] {
                color = tcell.ColorDarkCyan.TrueColor()
            }
            align := cview.AlignLeft
            [yellow]if[white] row == [red]0[white] {
                align = cview.AlignCenter
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] || column >= [red]4[white] {
                align = cview.AlignRight
            }
            table.[yellow]SetCell[white](row,
                column,
                &cview.TableCell{
                    Text:          [red]"..."[white],
                    Color:         color,
                    Align:         align,
                    NotSelectable: row == [red]0[white] || column == [red]0[white],
                })
        }
    }
    cview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](table, true).
        [yellow]Run[white]()
}`

const tableSelectCell = `[green]func[white] [yellow]main[white]() {
    table := cview.[yellow]NewTable[white]().
        [yellow]SetFixed[white]([red]1[white], [red]1[white]).
        [yellow]SetSelectable[white](true, true)
    [yellow]for[white] row := [red]0[white]; row < [red]40[white]; row++ {
        [yellow]for[white] column := [red]0[white]; column < [red]7[white]; column++ {
            color := tcell.ColorWhite.TrueColor()
            [yellow]if[white] row == [red]0[white] {
                color = tcell.ColorYellow.TrueColor()
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] {
                color = tcell.ColorDarkCyan.TrueColor()
            }
            align := cview.AlignLeft
            [yellow]if[white] row == [red]0[white] {
                align = cview.AlignCenter
            } [yellow]else[white] [yellow]if[white] column == [red]0[white] || column >= [red]4[white] {
                align = cview.AlignRight
            }
            table.[yellow]SetCell[white](row,
                column,
                &cview.TableCell{
                    Text:          [red]"..."[white],
                    Color:         color,
                    Align:         align,
                    NotSelectable: row == [red]0[white] || column == [red]0[white],
                })
        }
    }
    cview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](table, true).
        [yellow]Run[white]()
}`

// Table demonstrates the Table.
func Table(nextSlide func()) (title string, info string, content cview.Primitive) {
	table := cview.NewTable()
	table.SetFixed(1, 1)
	for row, line := range strings.Split(tableData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			color := cview.Styles.PrimaryTextColor
			if row == 0 {
				color = cview.Styles.SecondaryTextColor
			} else if column == 0 {
				color = cview.Styles.TertiaryTextColor
			}
			align := cview.AlignLeft
			if row == 0 {
				align = cview.AlignCenter
			} else if column == 0 || column >= 4 {
				align = cview.AlignRight
			}
			tableCell := cview.NewTableCell(cell)
			tableCell.SetTextColor(color)
			tableCell.SetAlign(align)
			tableCell.SetSelectable(row != 0 && column != 0)
			if column >= 1 && column <= 3 {
				tableCell.SetExpansion(1)
			}
			table.SetCell(row, column, tableCell)
		}
	}
	table.SetBorder(true)
	table.SetTitle("Table")

	code := cview.NewTextView()
	code.SetWrap(false)
	code.SetDynamicColors(true)
	code.SetPadding(1, 1, 2, 0)

	list := cview.NewList()

	basic := func() {
		table.SetBorders(false)
		table.SetSelectable(false, false)
		table.SetSeparator(' ')
		code.Clear()
		fmt.Fprint(code, tableBasic)
	}

	separator := func() {
		table.SetBorders(false)
		table.SetSelectable(false, false)
		table.SetSeparator(cview.Borders.Vertical)
		code.Clear()
		fmt.Fprint(code, tableSeparator)
	}

	borders := func() {
		table.SetBorders(true)
		table.SetSelectable(false, false)
		code.Clear()
		fmt.Fprint(code, tableBorders)
	}

	selectRow := func() {
		table.SetBorders(false)
		table.SetSelectable(true, false)
		table.SetSeparator(' ')
		code.Clear()
		fmt.Fprint(code, tableSelectRow)
	}

	selectColumn := func() {
		table.SetBorders(false)
		table.SetSelectable(false, true)
		table.SetSeparator(' ')
		code.Clear()
		fmt.Fprint(code, tableSelectColumn)
	}

	selectCell := func() {
		table.SetBorders(false)
		table.SetSelectable(true, true)
		table.SetSeparator(' ')
		code.Clear()
		fmt.Fprint(code, tableSelectCell)
	}

	navigate := func() {
		app.SetFocus(table)
		table.SetDoneFunc(func(key tcell.Key) {
			app.SetFocus(list)
		})
		table.SetSelectedFunc(func(row int, column int) {
			app.SetFocus(list)
		})
	}

	list.ShowSecondaryText(false)
	list.SetPadding(1, 1, 2, 2)

	var demoTableText = []struct {
		text     string
		shortcut rune
		selected func()
	}{
		{"Basic table", 'b', basic},
		{"Table with separator", 's', separator},
		{"Table with borders", 'o', borders},
		{"Selectable rows", 'r', selectRow},
		{"Selectable columns", 'c', selectColumn},
		{"Selectable cells", 'l', selectCell},
		{"Navigate", 'n', navigate},
		{"Next slide", 'x', nextSlide},
	}

	for _, tableText := range demoTableText {
		item := cview.NewListItem(tableText.text)
		item.SetShortcut(tableText.shortcut)
		item.SetSelectedFunc(tableText.selected)
		list.AddItem(item)
	}

	basic()

	subFlex := cview.NewFlex()
	subFlex.SetDirection(cview.FlexRow)
	subFlex.AddItem(list, 10, 1, true)
	subFlex.AddItem(table, 0, 1, false)

	flex := cview.NewFlex()
	flex.AddItem(subFlex, 0, 1, true)
	flex.AddItem(code, codeWidth, 1, false)

	return "Table", "", flex
}

const textView1 = `[green]func[white] [yellow]main[white]() {
	app := cview.[yellow]NewApplication[white]()
    textView := cview.[yellow]NewTextView[white]().
        [yellow]SetTextColor[white](tcell.ColorYellow.TrueColor()).
        [yellow]SetScrollable[white](false).
        [yellow]SetChangedFunc[white]([yellow]func[white]() {
            app.[yellow]Draw[white]()
        })
    [green]go[white] [yellow]func[white]() {
        [green]var[white] n [green]int
[white]        [yellow]for[white] {
            n++
            fmt.[yellow]Fprintf[white](textView, [red]"%d "[white], n)
            time.[yellow]Sleep[white]([red]200[white] * time.Millisecond)
        }
    }()
    app.[yellow]SetRoot[white](textView, true).
        [yellow]Run[white]()
}`

// TextView1 demonstrates the basic text view.
func TextView1(nextSlide func()) (title string, info string, content cview.Primitive) {
	textView := cview.NewTextView()
	textView.SetVerticalAlign(cview.AlignBottom)
	textView.SetTextColor(tcell.ColorYellow.TrueColor())
	textView.SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	textView.SetChangedFunc(func() {
		if textView.HasFocus() {
			app.Draw()
		}
	})
	go func() {
		var n int
		for {
			n++
			if n > 512 {
				n = 1
				textView.SetText("")
			}

			fmt.Fprintf(textView, "%d ", n)
			time.Sleep(75 * time.Millisecond)
		}
	}()
	textView.SetBorder(true)
	textView.SetTitle("TextView implements io.Writer")
	textView.ScrollToEnd()
	return "TextView 1", textViewInfo, Code(textView, 36, 13, textView1)
}

const textView2 = `[green]package[white] main

[green]import[white] (
    [red]"strconv"[white]

    [red]"github.com/gdamore/tcell/v2"[white]
    [red]"code.rocketnine.space/tslocum/cview"[white]
)

[green]func[white] [yellow]main[white]() {
    ["0"]textView[""] := cview.[yellow]NewTextView[white]()
    ["1"]textView[""].[yellow]SetDynamicColors[white](true).
        [yellow]SetWrap[white](false).
        [yellow]SetRegions[white](true).
        [yellow]SetDoneFunc[white]([yellow]func[white](key tcell.Key) {
            highlights := ["2"]textView[""].[yellow]GetHighlights[white]()
            hasHighlights := [yellow]len[white](highlights) > [red]0
            [yellow]switch[white] key {
            [yellow]case[white] tcell.KeyEnter:
                [yellow]if[white] hasHighlights {
                    ["3"]textView[""].[yellow]Highlight[white]()
                } [yellow]else[white] {
                    ["4"]textView[""].[yellow]Highlight[white]([red]"0"[white]).
                        [yellow]ScrollToHighlight[white]()
                }
            [yellow]case[white] tcell.KeyTab:
                [yellow]if[white] hasHighlights {
                    current, _ := strconv.[yellow]Atoi[white](highlights[[red]0[white]])
                    next := (current + [red]1[white]) % [red]9
                    ["5"]textView[""].[yellow]Highlight[white](strconv.[yellow]Itoa[white](next)).
                        [yellow]ScrollToHighlight[white]()
                }
            [yellow]case[white] tcell.KeyBacktab:
                [yellow]if[white] hasHighlights {
                    current, _ := strconv.[yellow]Atoi[white](highlights[[red]0[white]])
                    next := (current - [red]1[white] + [red]9[white]) % [red]9
                    ["6"]textView[""].[yellow]Highlight[white](strconv.[yellow]Itoa[white](next)).
                        [yellow]ScrollToHighlight[white]()
                }
            }
        })
    fmt.[yellow]Fprint[white](["7"]textView[""], content)
    cview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](["8"]textView[""], true).
        [yellow]Run[white]()
}`

// TextView2 demonstrates the extended text view.
func TextView2(nextSlide func()) (title string, info string, content cview.Primitive) {
	codeView := cview.NewTextView()
	codeView.SetWrap(false)
	fmt.Fprint(codeView, textView2)
	codeView.SetBorder(true)
	codeView.SetTitle("Buffer content")

	textView := cview.NewTextView()
	textView.SetDynamicColors(true)
	textView.SetWrap(false)
	textView.SetRegions(true)
	textView.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			nextSlide()
			return
		}
		highlights := textView.GetHighlights()
		hasHighlights := len(highlights) > 0
		switch key {
		case tcell.KeyEnter:
			if hasHighlights {
				textView.Highlight()
			} else {
				textView.Highlight("0")
				textView.ScrollToHighlight()
			}
		case tcell.KeyTab:
			if hasHighlights {
				current, _ := strconv.Atoi(highlights[0])
				next := (current + 1) % 9
				textView.Highlight(strconv.Itoa(next))
				textView.ScrollToHighlight()
			}
		case tcell.KeyBacktab:
			if hasHighlights {
				current, _ := strconv.Atoi(highlights[0])
				next := (current - 1 + 9) % 9
				textView.Highlight(strconv.Itoa(next))
				textView.ScrollToHighlight()
			}
		}
	})
	fmt.Fprint(textView, textView2)
	textView.SetBorder(true)
	textView.SetTitle("TextView output")
	textView.SetScrollBarVisibility(cview.ScrollBarAuto)

	flex := cview.NewFlex()
	flex.AddItem(textView, 0, 1, true)
	flex.AddItem(codeView, 0, 1, false)

	return "TextView 2", textViewInfo, flex
}

const treeAllCode = `[green]package[white] main

[green]import[white] [red]"code.rocketnine.space/tslocum/cview"[white]

[green]func[white] [yellow]main[white]() {
	$$$

	root := cview.[yellow]NewTreeNode[white]([red]"Root"[white]).
		[yellow]AddChild[white](cview.[yellow]NewTreeNode[white]([red]"First child"[white]).
			[yellow]AddChild[white](cview.[yellow]NewTreeNode[white]([red]"Grandchild A"[white])).
			[yellow]AddChild[white](cview.[yellow]NewTreeNode[white]([red]"Grandchild B"[white]))).
		[yellow]AddChild[white](cview.[yellow]NewTreeNode[white]([red]"Second child"[white]).
			[yellow]AddChild[white](cview.[yellow]NewTreeNode[white]([red]"Grandchild C"[white])).
			[yellow]AddChild[white](cview.[yellow]NewTreeNode[white]([red]"Grandchild D"[white]))).
		[yellow]AddChild[white](cview.[yellow]NewTreeNode[white]([red]"Third child"[white]))

	tree.[yellow]SetRoot[white](root).
		[yellow]SetCurrentNode[white](root)

	cview.[yellow]NewApplication[white]().
		[yellow]SetRoot[white](tree, true).
		[yellow]Run[white]()
}`

const treeBasicCode = `tree := cview.[yellow]NewTreeView[white]()`

const treeTopLevelCode = `tree := cview.[yellow]NewTreeView[white]().
		[yellow]SetTopLevel[white]([red]1[white])`

const treeAlignCode = `tree := cview.[yellow]NewTreeView[white]().
		[yellow]SetAlign[white](true)`

const treePrefixCode = `tree := cview.[yellow]NewTreeView[white]().
		[yellow]SetGraphics[white](false).
		[yellow]SetTopLevel[white]([red]1[white]).
		[yellow]SetPrefixes[white]([][green]string[white]{
			[red]"[red[]* "[white],
			[red]"[darkcyan[]- "[white],
			[red]"[darkmagenta[]- "[white],
		})`

type node struct {
	text     string
	expand   bool
	selected func()
	children []*node
}

var (
	tree          = cview.NewTreeView()
	treeNextSlide func()
	treeCode      = cview.NewTextView()
)

var rootNode = &node{
	text: "Root",
	children: []*node{
		{text: "Expand all", selected: func() { tree.GetRoot().ExpandAll() }},
		{text: "Collapse all", selected: func() {
			for _, child := range tree.GetRoot().GetChildren() {
				child.CollapseAll()
			}
		}},
		{text: "Hide root node", expand: true, children: []*node{
			{text: "Tree list starts one level down"},
			{text: "Works better for lists where no top node is needed"},
			{text: "Switch to this layout", selected: func() {
				tree.SetAlign(false)
				tree.SetTopLevel(1)
				tree.SetGraphics(true)
				tree.SetPrefixes(nil)
				treeCode.SetText(strings.Replace(treeAllCode, "$$$", treeTopLevelCode, -1))
			}},
		}},
		{text: "Align node text", expand: true, children: []*node{
			{text: "For trees that are similar to lists"},
			{text: "Hierarchy shown only in line drawings"},
			{text: "Switch to this layout", selected: func() {
				tree.SetAlign(true)
				tree.SetTopLevel(0)
				tree.SetGraphics(true)
				tree.SetPrefixes(nil)
				treeCode.SetText(strings.Replace(treeAllCode, "$$$", treeAlignCode, -1))
			}},
		}},
		{text: "Prefixes", expand: true, children: []*node{
			{text: "Best for hierarchical bullet point lists"},
			{text: "You can define your own prefixes per level"},
			{text: "Switch to this layout", selected: func() {
				tree.SetAlign(false)
				tree.SetTopLevel(1)
				tree.SetGraphics(false)
				tree.SetPrefixes([]string{"[red]* ", "[darkcyan]- ", "[darkmagenta]- "})
				treeCode.SetText(strings.Replace(treeAllCode, "$$$", treePrefixCode, -1))
			}},
		}},
		{text: "Basic tree with graphics", expand: true, children: []*node{
			{text: "Lines illustrate hierarchy"},
			{text: "Basic indentation"},
			{text: "Switch to this layout", selected: func() {
				tree.SetAlign(false)
				tree.SetTopLevel(0)
				tree.SetGraphics(true)
				tree.SetPrefixes(nil)
				treeCode.SetText(strings.Replace(treeAllCode, "$$$", treeBasicCode, -1))
			}},
		}},
		{text: "Next slide", selected: func() { treeNextSlide() }},
	}}

// TreeView demonstrates the tree view.
func TreeView(nextSlide func()) (title string, info string, content cview.Primitive) {
	treeNextSlide = nextSlide
	tree.SetBorder(true)
	tree.SetTitle("TreeView")

	// Add nodes.
	var add func(target *node) *cview.TreeNode
	add = func(target *node) *cview.TreeNode {
		node := cview.NewTreeNode(target.text)
		node.SetSelectable(target.expand || target.selected != nil)
		node.SetExpanded(target == rootNode)
		node.SetReference(target)
		if target.expand {
			node.SetColor(tcell.ColorLimeGreen.TrueColor())
		} else if target.selected != nil {
			node.SetColor(tcell.ColorRed.TrueColor())
		}
		for _, child := range target.children {
			node.AddChild(add(child))
		}
		return node
	}
	root := add(rootNode)
	tree.SetRoot(root)
	tree.SetCurrentNode(root)
	tree.SetSelectedFunc(func(n *cview.TreeNode) {
		original := n.GetReference().(*node)
		if original.expand {
			n.SetExpanded(!n.IsExpanded())
		} else if original.selected != nil {
			original.selected()
		}
	})

	treeCode.SetWrap(false)
	treeCode.SetDynamicColors(true)
	treeCode.SetText(strings.Replace(treeAllCode, "$$$", treeBasicCode, -1))
	treeCode.SetPadding(1, 1, 2, 0)

	flex := cview.NewFlex()
	flex.AddItem(tree, 0, 1, true)
	flex.AddItem(treeCode, codeWidth, 1, false)

	return "TreeView", "", flex
}

const loremIpsumText = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

// Window returns the window page.
func Window(nextSlide func()) (title string, info string, content cview.Primitive) {
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

	w1 := cview.NewWindow(list)
	w1.SetRect(2, 2, 10, 7)

	w2 := cview.NewWindow(loremIpsum)
	w2.SetRect(7, 4, 12, 12)

	w1.SetTitle("List")
	w2.SetTitle("Lorem Ipsum")

	w3 := cview.NewWindow(list)
	w3.SetRect(12, 12, 10, 7)
	w3.SetBorder(false)
	w3.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		return 0, nil
	})

	wm.Add(w1, w2, w3)

	return "Window", windowInfo, wm
}

// Slide is a function which returns the slide's title, any applicable
// information and its main primitive, its. It receives a "nextSlide" function
// which can be called to advance the presentation to the next slide.
type Slide func(nextSlide func()) (title string, info string, content cview.Primitive)

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

	// The presentation slides.
	slides := []Slide{
		Cover,
		Introduction,
		Colors,
		TextView1,
		TextView2,
		InputField,
		Slider,
		Form,
		Table,
		TreeView,
		Flex,
		Grid,
		Window,
		End,
	}

	panels := cview.NewTabbedPanels()

	// Create the pages for all slides.
	previousSlide := func() {
		slide, _ := strconv.Atoi(panels.GetCurrentTab())
		slide = (slide - 1 + len(slides)) % len(slides)
		panels.SetCurrentTab(strconv.Itoa(slide))
	}
	nextSlide := func() {
		slide, _ := strconv.Atoi(panels.GetCurrentTab())
		slide = (slide + 1) % len(slides)
		panels.SetCurrentTab(strconv.Itoa(slide))
	}

	cursor := 0
	var slideRegions []int
	for index, slide := range slides {
		slideRegions = append(slideRegions, cursor)

		title, info, primitive := slide(nextSlide)

		h := cview.NewTextView()
		if info != "" {
			h.SetDynamicColors(true)
			h.SetText("  [" + cview.ColorHex(cview.Styles.SecondaryTextColor) + "]Info:[-]  " + info)
		}

		// Create a Flex layout that centers the logo and subtitle.
		f := cview.NewFlex()
		f.SetDirection(cview.FlexRow)
		f.AddItem(h, 1, 1, false)
		f.AddItem(primitive, 0, 1, true)

		panels.AddTab(strconv.Itoa(index), title, f)

		cursor += len(title) + 4
	}
	panels.SetCurrentTab("0")

	// Shortcuts to navigate the slides.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			nextSlide()
		} else if event.Key() == tcell.KeyCtrlP {
			previousSlide()
		}
		return event
	})

	// Start the application.
	app.SetRoot(panels, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

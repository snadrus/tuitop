package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gaurav-Gosain/tuios/pkg/tuios"
	tint "github.com/lrstanley/bubbletint/v2"
	"github.com/snadrus/tuitop/internal/desktopbg"
	"github.com/snadrus/tuitop/internal/startmenu"
)

//go:embed images/fake-winxp-bkgd.png
var fakeWinXPBackgroundPNG []byte

var fakeWinXPBackground image.Image

func init() {
	img, _, err := image.Decode(bytes.NewReader(fakeWinXPBackgroundPNG))
	if err != nil {
		log.Fatalf("tuitop: decode embedded desktop wallpaper: %v", err)
	}
	fakeWinXPBackground = img
}

const unsnapQuarter = 8
const statusBarHeight = 1

// wallpaperRasterHeight is the wallpaper/Frame row count. It must be identical in Update
// and View whenever SetTerminalSize is called; otherwise the raster alternates heights
// (e.g. inner h vs h-1) and the background appears to stretch an extra row every other frame.
// When inner content is taller than one row, reserve the bottom row for the taskbar strip.
func wallpaperRasterHeight(innerRenderH int) int {
	if innerRenderH <= 1 {
		return innerRenderH
	}
	return innerRenderH - 1
}

// Windows XP taskbar colors (hex)
const (
	winXPTaskbarBlue       = "#0054E3"
	winXPNotificationBlue  = "#1996E9" // notification area light blue (from XP design guidelines)
	winXPStartGreen        = "#22B14C" // Start button green
	winXPStartGreenPressed = "#166B2E" // pressed / menu-open (darker green, not taskbar blue)
	// Start menu body: left pane (white), right pane (XP-style light blue) per classic shell.
	winXPStartMenuLeftBg  = "#ffffff"
	winXPStartMenuLeftFg  = "#000000"
	winXPStartMenuRightBg = "#cde6fc"
	winXPStartMenuRightFg = "#000000"
)

// 256-color approximations for terminals where hex backgrounds break foreground (set TUITOP_USE_256_COLORS=1)
// 28=green, 26=blue; use ANSI 4 for notification (256-color 110 breaks fg on some terminals)
const (
	color256StartGreen        = "28"
	color256StartGreenPressed = "64" // darker olive-green in 256 palette (distinct from 28)
	color256TaskbarBlue       = "26"
	color256Notification      = "4" // ANSI blue - 110 renders clock text black on problem terminals
	// Lighter blue for start menu top chrome (upper half of ▄ row); distinct from taskbar 26.
	color256MenuTopUpper = "81"
	// Start menu panes in 256-color mode (approximates hex panes above).
	color256StartMenuLeftBg  = "15"
	color256StartMenuLeftFg  = "0"
	color256StartMenuRightBg = "195"
	color256StartMenuRightFg = "0"
)
const (
	terminalBtnGreenHex = "#22B14C"
	terminalBtnBlackHex = "#000000"
	terminalBtnGreen256 = "28"
	terminalBtnBlack256 = "0"
)

// nearBlackPalette: dark tints for window backgrounds (RRGGBB hex).
var nearBlackPalette = []string{
	"2a2a2a", "382a2a", "2a382a", "2a2a38",
	"38382a", "2a3838", "382a38", "342a26",
	"2a3438", "26302a", "34262a", "2a2634",
	"382a34", "2a3834", "342a38", "36362a",
}

// fixEmojiWidth replaces U+FE0F (emoji presentation selector) with a space.
// Terminals often ignore FE0F for cursor advance, so the width library over-counts
// by 1 for affected grapheme clusters. Swapping FE0F → space keeps the measured
// width equal to the physical terminal width.
func fixEmojiWidth(s string) string {
	return strings.ReplaceAll(s, "\uFE0F", " ")
}

// taskbarItem tracks a minimized-window button on the XP taskbar for click handling.
type taskbarItem struct {
	xStart, xEnd int
	windowIndex  int
}

type appModel struct {
	inner        *tuios.Model
	prevWinCount int
	// lastTermBtnBounds: [xStart, xEnd) of terminal button on status bar, from last render
	lastTermBtnBounds [2]int
	// lastStartBtnBounds: [xStart, xEnd) of Start button + divider (combined click target), from last render
	lastStartBtnBounds [2]int
	// lastMenuBounds: [xStart, yStart, xEnd, yEnd) of start menu overlay when open; used for click-outside-to-close
	lastMenuBounds    [4]int
	lastTaskbarItems  []taskbarItem
	startMenuOpen     bool
	wallpaper         *desktopbg.Wallpaper
}

func (m *appModel) Init() tea.Cmd {
	m.prevWinCount = len(m.inner.Windows)
	return m.inner.Init()
}

func (m *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Shrink the inner viewport by statusBarHeight so tuios renders above our bar
	if sz, ok := msg.(tea.WindowSizeMsg); ok {
		msg = tea.WindowSizeMsg{Width: sz.Width, Height: sz.Height - statusBarHeight}
	}

	if click, ok := msg.(tea.MouseClickMsg); ok {
		if handled, cmd := m.handleStartMenuClick(click); handled {
			return m, cmd
		}
		if m.handleStatusBarClick(click) {
			return m, nil
		}
		if m.handleRestoreClick(click) {
			return m, nil
		}
	}

	updated, cmd := m.inner.Update(msg)
	if inner, ok := updated.(*tuios.Model); ok {
		m.inner = inner
	}

	if m.wallpaper != nil {
		w := m.inner.GetRenderWidth()
		h := wallpaperRasterHeight(m.inner.GetRenderHeight())
		m.wallpaper.SetTerminalSize(w, h)
	}

	m.maybeSetNewWindowBackground()
	return m, cmd
}

func (m *appModel) maybeSetNewWindowBackground() {
	n := len(m.inner.Windows)
	if n <= m.prevWinCount {
		m.prevWinCount = n
		return
	}
	m.prevWinCount = n

	newWin := m.inner.Windows[n-1]
	hex := nearBlackPalette[rand.Intn(len(nearBlackPalette))]
	_ = m.inner.SetWindowBackgroundColor(newWin.ID, hex)
}

func (m *appModel) View() tea.View {
	innerView := m.inner.View()
	tw, th := m.inner.GetRenderWidth(), m.inner.GetRenderHeight()

	minimized := m.getMinimizedWindows()
	bar, termBtnStart, termBtnEnd, startBtnStart, startBtnEnd, tbItems := renderStatusBarWithBounds(tw, m.startMenuOpen, minimized)
	m.lastTermBtnBounds = [2]int{termBtnStart, termBtnEnd}
	m.lastStartBtnBounds = [2]int{startBtnStart, startBtnEnd}
	m.lastTaskbarItems = tbItems

	var menuLayer *lipgloss.Layer
	if m.startMenuOpen {
		menuCfg := startMenuStyleConfig()
		menuContent, menuW, menuH := startmenu.Render(menuCfg)
		menuY := th - statusBarHeight - menuH
		if menuY < 0 {
			menuY = 0
		}
		m.lastMenuBounds = [4]int{0, menuY, menuW, menuY + menuH}
		menuLayer = lipgloss.NewLayer(menuContent).
			X(0).
			Y(menuY).
			Z(100000).
			ID("tuitop-startmenu")
	} else {
		m.lastMenuBounds = [4]int{0, 0, 0, 0}
	}

	var content string

	if m.wallpaper != nil && tw > 0 && th > 1 {
		// Wallpaper composites per-cell and misaligns wide glyphs on the same row as
		// the clock. Keep the taskbar off the Frame: only rows [0, th-2] use wallpaper.
		bodyCanvas := m.inner.GetCanvas(true)
		if menuLayer != nil {
			bodyCanvas.Compose(lipgloss.NewCompositor(menuLayer))
		}
		bodyStr := lipgloss.Sprint(bodyCanvas.Render())
		bodyStr = strings.ReplaceAll(bodyStr, "\r\n", "\n")
		needLines := wallpaperRasterHeight(th)
		bodyStr = desktopbg.OverlayTopLines(bodyStr, needLines)

		m.wallpaper.SetTerminalSize(tw, needLines)
		frame := desktopbg.NewFrame(m.wallpaper, bodyStr)
		barLayer := lipgloss.NewLayer(bar).X(0).Y(th - statusBarHeight).Z(99999).ID("tuitop-bar")
		canvas := lipgloss.NewCanvas(tw, th)
		canvas.Compose(frame)
		canvas.Compose(lipgloss.NewCompositor(barLayer))
		content = lipgloss.Sprint(canvas.Render())
	} else {
		canvas := m.inner.GetCanvas(true)
		barLayer := lipgloss.NewLayer(bar).
			X(0).
			Y(th - statusBarHeight).
			Z(99999).
			ID("tuitop-bar")
		var extraLayers []*lipgloss.Layer
		extraLayers = append(extraLayers, barLayer)
		if menuLayer != nil {
			extraLayers = append(extraLayers, menuLayer)
		}
		canvas.Compose(lipgloss.NewCompositor(extraLayers...))
		content = lipgloss.Sprint(canvas.Render())

		terminalHeight := m.inner.Height + statusBarHeight
		if terminalHeight > statusBarHeight {
			innerLineCount := len(strings.Split(content, "\n"))
			requiredInnerLines := terminalHeight - statusBarHeight
			if innerLineCount < requiredInnerLines {
				content += strings.Repeat("\n", requiredInnerLines-innerLineCount)
			}
		}

		if m.wallpaper != nil && tw > 0 && th == 1 {
			h := wallpaperRasterHeight(th)
			m.wallpaper.SetTerminalSize(tw, h)
			c := lipgloss.NewCanvas(tw, h)
			c.Compose(desktopbg.NewFrame(m.wallpaper, content))
			content = lipgloss.Sprint(c.Render())
		}
	}

	var view tea.View
	view.SetContent(content)
	view.AltScreen = innerView.AltScreen
	view.MouseMode = innerView.MouseMode
	view.ReportFocus = innerView.ReportFocus
	view.DisableBracketedPasteMode = innerView.DisableBracketedPasteMode
	view.Cursor = innerView.Cursor
	return view
}

// minimizedInfo describes one minimized window for taskbar rendering.
type minimizedInfo struct {
	windowIndex int
	label       string
}

// getMinimizedWindows returns minimized windows in the current workspace, ordered by minimize time.
func (m *appModel) getMinimizedWindows() []minimizedInfo {
	type entry struct {
		idx   int
		order int64
	}
	var entries []entry
	for i, w := range m.inner.Windows {
		if w.Workspace == m.inner.CurrentWorkspace && w.Minimized {
			entries = append(entries, entry{i, w.MinimizeOrder})
		}
	}
	sort.Slice(entries, func(a, b int) bool { return entries[a].order < entries[b].order })
	out := make([]minimizedInfo, len(entries))
	for i, e := range entries {
		w := m.inner.Windows[e.idx]
		name := w.CustomName
		if name == "" && w.Title != "" {
			t := w.Title
			if len(t) > 8 {
				t = t[len(t)-8:]
			}
			name = t
		}
		if name == "" {
			name = fmt.Sprintf("Term %d", i+1)
		}
		if len(name) > 16 {
			name = name[:13] + "..."
		}
		out[i] = minimizedInfo{windowIndex: e.idx, label: name}
	}
	return out
}

// renderStatusBarWithBounds returns the status bar string, terminal button X bounds [start, end),
// start button X bounds [start, end), and positions of minimized-window taskbar buttons.
func renderStatusBarWithBounds(width int, startMenuOpen bool, minimized []minimizedInfo) (bar string, termBtnXStart, termBtnXEnd, startBtnXStart, startBtnXEnd int, items []taskbarItem) {
	t := time.Now()
	currentTime := fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
	use256 := os.Getenv("TUITOP_USE_256_COLORS") == "1" || strings.EqualFold(os.Getenv("TUITOP_USE_256_COLORS"), "true")

	var startBg, taskbarBg, notifBg, dividerFg string
	if use256 {
		startBg, taskbarBg, notifBg = color256StartGreen, color256TaskbarBlue, color256Notification
		dividerFg = "0"
	} else {
		startBg, taskbarBg, notifBg = winXPStartGreen, winXPTaskbarBlue, winXPNotificationBlue
		dividerFg = "#000000"
	}

	// When menu is open, show Start button as pressed (darker green; taskbar stays blue only on the bar)
	startBtnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color(startBg)).
		Padding(0, 1)
	if startMenuOpen {
		pressedBg := winXPStartGreenPressed
		if use256 {
			pressedBg = color256StartGreenPressed
		}
		startBtnStyle = startBtnStyle.Background(lipgloss.Color(pressedBg)).Foreground(lipgloss.Color("7"))
	}
	startBtn := startBtnStyle.Render(fixEmojiWidth(" ❄️  Start "))

	divStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(startBg)).
		Background(lipgloss.Color(taskbarBg))
	divider := divStyle.Render("\uE0BC")

	var terminalBtnFg, terminalBtnBg string
	if use256 {
		terminalBtnFg, terminalBtnBg = terminalBtnGreen256, terminalBtnBlack256
	} else {
		terminalBtnFg, terminalBtnBg = terminalBtnGreenHex, terminalBtnBlackHex
	}
	// Blue bar spacer before terminal button (space to the left of button zone)
	blueSpacer := lipgloss.NewStyle().
		Background(lipgloss.Color(taskbarBg)).
		Width(4).
		Render(" ")
	terminalBtnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(terminalBtnFg)).
		Background(lipgloss.Color(terminalBtnBg)).
		Padding(0, 1)
	terminalBtn := terminalBtnStyle.Render(">_")

	termBtnXStart = lipgloss.Width(startBtn) + lipgloss.Width(divider) + lipgloss.Width(blueSpacer)
	termBtnXEnd = termBtnXStart + lipgloss.Width(terminalBtn)

	clockStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Bold(true).
		Background(lipgloss.Color(notifBg)).
		Padding(0, 1)
	clockText := clockStyle.Render(currentTime)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(dividerFg)).
		Background(lipgloss.Color(notifBg)).
		Width(1).
		AlignHorizontal(lipgloss.Center)
	separator := separatorStyle.Render("\u258F")

	leftSection := lipgloss.JoinHorizontal(lipgloss.Top, startBtn, divider)
	leftWithTermBtn := lipgloss.JoinHorizontal(lipgloss.Top, leftSection, blueSpacer, terminalBtn)
	leftWithTermBtnWidth := lipgloss.Width(leftWithTermBtn)
	clockWidth := lipgloss.Width(clockText)
	separatorWidth := lipgloss.Width(separator)

	// Build minimized-window buttons for the middle section
	var taskbarBtnBg string
	if use256 {
		taskbarBtnBg = "25" // slightly darker blue in 256 palette
	} else {
		taskbarBtnBg = "#1941A5" // XP-style darker blue for depressed taskbar buttons
	}
	btnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color(taskbarBtnBg)).
		Padding(0, 1)

	var btnParts []string
	var btnWidths []int
	for _, mi := range minimized {
		rendered := btnStyle.Render(mi.label)
		btnParts = append(btnParts, rendered)
		btnWidths = append(btnWidths, lipgloss.Width(rendered))
	}

	// Calculate the space used by buttons (1-cell gap between each)
	buttonsWidth := 0
	for i, w := range btnWidths {
		buttonsWidth += w
		if i > 0 {
			buttonsWidth++ // 1-cell gap
		}
	}

	middleWidth := width - leftWithTermBtnWidth - separatorWidth - clockWidth
	if middleWidth < 0 {
		middleWidth = 0
	}

	// Render the middle: a small left spacer, then buttons, then fill remaining space
	const middleLeftPad = 1
	remainingAfterBtns := middleWidth - middleLeftPad - buttonsWidth
	if remainingAfterBtns < 0 {
		remainingAfterBtns = 0
	}

	padLeft := lipgloss.NewStyle().
		Background(lipgloss.Color(taskbarBg)).
		Width(middleLeftPad).
		Render(" ")
	padRight := lipgloss.NewStyle().
		Background(lipgloss.Color(taskbarBg)).
		Width(remainingAfterBtns).
		Render(" ")

	// Join buttons with 1-cell blue gaps
	gap := lipgloss.NewStyle().Background(lipgloss.Color(taskbarBg)).Width(1).Render(" ")
	var middleParts []string
	middleParts = append(middleParts, padLeft)
	cursorX := leftWithTermBtnWidth + middleLeftPad
	for i, bp := range btnParts {
		if i > 0 {
			middleParts = append(middleParts, gap)
			cursorX++
		}
		middleParts = append(middleParts, bp)
		items = append(items, taskbarItem{
			xStart:      cursorX,
			xEnd:        cursorX + btnWidths[i],
			windowIndex: minimized[i].windowIndex,
		})
		cursorX += btnWidths[i]
	}
	middleParts = append(middleParts, padRight)
	middle := lipgloss.JoinHorizontal(lipgloss.Top, middleParts...)

	bar = lipgloss.JoinHorizontal(lipgloss.Top, leftWithTermBtn, middle, separator, clockText)
	startBtnXStart = 0
	startBtnXEnd = lipgloss.Width(startBtn) + lipgloss.Width(divider)
	return bar, termBtnXStart, termBtnXEnd, startBtnXStart, startBtnXEnd, items
}

// startMenuStyleConfig matches the StyleConfig used when drawing the open start menu.
func startMenuStyleConfig() startmenu.StyleConfig {
	use256 := os.Getenv("TUITOP_USE_256_COLORS") == "1" || strings.EqualFold(os.Getenv("TUITOP_USE_256_COLORS"), "true")
	if use256 {
		return startmenu.StyleConfig{
			MenuBg:           color256TaskbarBlue,
			MenuFg:           "7",
			UserFg:           "7",
			HighlightBg:      color256Notification,
			HighlightFg:      "7",
			AllProgramsFg:    "",
			TopStripeUpperBg: color256MenuTopUpper,
			LeftColumnBg:     color256StartMenuLeftBg,
			LeftColumnFg:     color256StartMenuLeftFg,
			RightColumnBg:    color256StartMenuRightBg,
			RightColumnFg:    color256StartMenuRightFg,
			MenuWidth:        40,
			LeftColumnWidth:  18,
			RightColumnWidth: 21,
		}
	}
	return startmenu.StyleConfig{
		MenuBg:           winXPTaskbarBlue,
		MenuFg:           "#ffffff",
		UserFg:           "#ffffff",
		HighlightBg:      winXPNotificationBlue,
		HighlightFg:      "#ffffff",
		AllProgramsFg:    "",
		TopStripeUpperBg: winXPNotificationBlue,
		LeftColumnBg:     winXPStartMenuLeftBg,
		LeftColumnFg:     winXPStartMenuLeftFg,
		RightColumnBg:    winXPStartMenuRightBg,
		RightColumnFg:    winXPStartMenuRightFg,
		MenuWidth:        40,
		LeftColumnWidth:  18,
		RightColumnWidth: 21,
	}
}

// handleStartMenuClick handles start menu clicks: Exit quits; inside-menu clicks are consumed; outside closes menu.
// Must run before handleStatusBarClick.
func (m *appModel) handleStartMenuClick(click tea.MouseClickMsg) (handled bool, cmd tea.Cmd) {
	if !m.startMenuOpen || m.inner == nil || click.Button != tea.MouseLeft {
		return false, nil
	}
	mb := m.lastMenuBounds
	// Click inside menu? Consume (don't forward to windows underneath).
	if click.X >= mb[0] && click.X < mb[2] && click.Y >= mb[1] && click.Y < mb[3] {
		localX := click.X - mb[0]
		localY := click.Y - mb[1]
		if startmenu.ClickIsExit(startMenuStyleConfig(), localX, localY) {
			return true, tea.Quit
		}
		return true, nil
	}
	// Click outside menu -> close
	m.startMenuOpen = false
	m.inner.MarkAllDirty()
	return true, nil
}

func (m *appModel) handleStatusBarClick(click tea.MouseClickMsg) bool {
	if m.inner == nil || click.Button != tea.MouseLeft {
		return false
	}
	// Bar is drawn at GetRenderHeight()-1 (last row of inner content)
	statusBarY := m.inner.GetRenderHeight() - 1
	if click.Y != statusBarY {
		return false
	}
	startStart, startEnd := m.lastStartBtnBounds[0], m.lastStartBtnBounds[1]
	termStart, termEnd := m.lastTermBtnBounds[0], m.lastTermBtnBounds[1]

	// Start button + divider click -> toggle menu
	if click.X >= startStart && click.X < startEnd {
		m.startMenuOpen = !m.startMenuOpen
		m.inner.MarkAllDirty()
		return true
	}
	// Terminal button click -> new terminal
	if click.X >= termStart && click.X < termEnd {
		if m.startMenuOpen {
			m.startMenuOpen = false
		}
		m.inner.AddWindow("")
		m.inner.MarkAllDirty()
		return true
	}
	// Minimized-window taskbar button click -> restore that window
	for _, item := range m.lastTaskbarItems {
		if click.X >= item.xStart && click.X < item.xEnd {
			if m.startMenuOpen {
				m.startMenuOpen = false
			}
			m.inner.RestoreWindow(item.windowIndex)
			m.inner.MarkAllDirty()
			return true
		}
	}
	// Click elsewhere on taskbar while menu open -> close menu
	if m.startMenuOpen {
		m.startMenuOpen = false
		m.inner.MarkAllDirty()
		return true
	}
	return false
}

func (m *appModel) handleRestoreClick(click tea.MouseClickMsg) bool {
	if m.inner == nil || m.inner.AutoTiling || click.Button != tea.MouseLeft {
		return false
	}

	info, err := m.inner.GetFocusedWindowData()
	if err != nil {
		return false
	}

	fullscreen, _ := info["fullscreen"].(bool)
	x, okX := info["x"].(int)
	y, okY := info["y"].(int)
	width, okWidth := info["width"].(int)
	if !fullscreen || !okX || !okY || !okWidth {
		return false
	}

	leftMost := x + width
	if click.Y != y || click.X < leftMost-7 || click.X > leftMost-5 {
		return false
	}

	m.inner.Snap(m.inner.FocusedWindow, unsnapQuarter)
	m.inner.InteractionMode = false
	m.inner.MarkAllDirty()
	return true
}

func main() {
	config := tuios.Config.DefaultConfig()
	config.Appearance.SuppressEmptyDesktopWelcome = true
	config.Appearance.BorderStyle = "none"
	config.Appearance.WindowTitlePosition = "top"
	config.Appearance.WindowTitleFgFocused = "#ffffff"   // active window title and controls
	config.Appearance.WindowTitleFgUnfocused = "#000000" // inactive windows
	config.Appearance.HideClock = true
	snapOnDragToEdge := false
	config.Appearance.SnapOnDragToEdge = &snapOnDragToEdge

	// WithTheme("WindowsXP") will fall back to "default" since it's not built-in.
	// After New(), the bubbletint default registry exists; we register our custom
	// tint and set it active. The fork's theme.Current() reads from this registry.
	//
	// Border color mapping in the fork's theme package:
	//   Red       → unfocused borders
	//   BrightCyan → focused window-mode borders
	//   BrightGreen → focused terminal-mode borders
	model := tuios.New(tuios.WithUserConfig(config), tuios.WithBorderStyle("none"), tuios.WithTheme("WindowsXP"), tuios.WithDockbarPosition("hidden"), tuios.WithModeless(true))

	winXPTheme := *tint.TintBuiltinDark
	winXPTheme.ID = "WindowsXP"
	winXPTheme.DisplayName = "Windows XP"
	winXPTheme.Red = &tint.Color{R: 192, G: 192, B: 192, A: 255}      // #C0C0C0 Luna silver → unfocused borders
	winXPTheme.BrightCyan = &tint.Color{R: 0, G: 84, B: 227, A: 255}  // #0054E3 Windows blue → focused window borders
	winXPTheme.BrightGreen = &tint.Color{R: 0, G: 84, B: 227, A: 255} // #0054E3 Windows blue → focused terminal borders
	tint.Register(&winXPTheme)
	tint.SetTintID("WindowsXP")
	program := tea.NewProgram(&appModel{
		inner:     model,
		wallpaper: &desktopbg.Wallpaper{Src: fakeWinXPBackground, Fill: true},
	}, tuios.ProgramOptions()...)
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}

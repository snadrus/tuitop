package main

import (
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gaurav-Gosain/tuios/pkg/tuios"
	tint "github.com/lrstanley/bubbletint/v2"
)

const unsnapQuarter = 8
const statusBarHeight = 1

// Windows XP taskbar colors (hex)
const (
	winXPTaskbarBlue      = "#0054E3"
	winXPNotificationBlue = "#1996E9" // notification area light blue (from XP design guidelines)
	winXPStartGreen       = "#22B14C" // Start button green
)

// 256-color approximations for terminals where hex backgrounds break foreground (set TUITOP_USE_256_COLORS=1)
// 28=green, 26=blue; use ANSI 4 for notification (256-color 110 breaks fg on some terminals)
const (
	color256StartGreen   = "28"
	color256TaskbarBlue  = "26"
	color256Notification = "4" // ANSI blue - 110 renders clock text black on problem terminals
)

// nearBlackPalette: dark tints for window backgrounds (RRGGBB hex).
var nearBlackPalette = []string{
	"2a2a2a", "382a2a", "2a382a", "2a2a38",
	"38382a", "2a3838", "382a38", "342a26",
	"2a3438", "26302a", "34262a", "2a2634",
	"382a34", "2a3834", "342a38", "36362a",
}

type appModel struct {
	inner        *tuios.Model
	prevWinCount int
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

	if click, ok := msg.(tea.MouseClickMsg); ok && m.handleRestoreClick(click) {
		return m, nil
	}

	updated, cmd := m.inner.Update(msg)
	if inner, ok := updated.(*tuios.Model); ok {
		m.inner = inner
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
	canvas := m.inner.GetCanvas(true)

	// Add our status bar as a canvas layer (same pipeline as tuios overlays) so it renders with correct colors
	bar := renderStatusBar(m.inner.GetRenderWidth())
	barLayer := lipgloss.NewLayer(bar).
		X(0).
		Y(m.inner.GetRenderHeight() - statusBarHeight).
		Z(99999).
		ID("tuitop-bar")
	canvas.AddLayers(barLayer)

	innerContent := lipgloss.Sprint(canvas.Render())

	// Pad so total height fills terminal (inner height is already shrunk by statusBarHeight)
	terminalHeight := m.inner.Height + statusBarHeight
	if terminalHeight > statusBarHeight {
		innerLineCount := len(strings.Split(innerContent, "\n"))
		requiredInnerLines := terminalHeight - statusBarHeight
		if innerLineCount < requiredInnerLines {
			innerContent += strings.Repeat("\n", requiredInnerLines-innerLineCount)
		}
	}

	var view tea.View
	view.SetContent(innerContent)
	view.AltScreen = innerView.AltScreen
	view.MouseMode = innerView.MouseMode
	view.ReportFocus = innerView.ReportFocus
	view.DisableBracketedPasteMode = innerView.DisableBracketedPasteMode
	view.Cursor = innerView.Cursor
	return view
}

func renderStatusBar(width int) string {
	currentTime := time.Now().Format("15:04:05")
	use256 := os.Getenv("TUITOP_USE_256_COLORS") == "1" || strings.EqualFold(os.Getenv("TUITOP_USE_256_COLORS"), "true")

	var startBg, taskbarBg, notifBg, dividerFg string
	if use256 {
		startBg, taskbarBg, notifBg = color256StartGreen, color256TaskbarBlue, color256Notification
		dividerFg = "0"
	} else {
		startBg, taskbarBg, notifBg = winXPStartGreen, winXPTaskbarBlue, winXPNotificationBlue
		dividerFg = "#000000"
	}

	startBtnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color(startBg)).
		Padding(0, 1)
	startBtn := startBtnStyle.Render(" 🐧 Start ")

	divStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(startBg)).
		Background(lipgloss.Color(taskbarBg))
	divider := divStyle.Render("\uE0BC")

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
	leftWidth := lipgloss.Width(leftSection)
	clockWidth := lipgloss.Width(clockText)
	separatorWidth := lipgloss.Width(separator)
	middleWidth := width - leftWidth - separatorWidth - clockWidth
	if middleWidth < 0 {
		middleWidth = 0
	}
	middle := lipgloss.NewStyle().
		Background(lipgloss.Color(taskbarBg)).
		Width(middleWidth).
		Render(" ")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftSection, middle, separator, clockText)
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
	config.Appearance.BorderStyle = "hidden"
	config.Appearance.WindowTitlePosition = "top"
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
	model := tuios.New(tuios.WithUserConfig(config), tuios.WithBorderStyle("hidden"), tuios.WithTheme("WindowsXP"), tuios.WithDockbarPosition("hidden"))

	winXPTheme := *tint.TintBuiltinDark
	winXPTheme.ID = "WindowsXP"
	winXPTheme.DisplayName = "Windows XP"
	winXPTheme.Red = &tint.Color{R: 192, G: 192, B: 192, A: 255}      // #C0C0C0 Luna silver → unfocused borders
	winXPTheme.BrightCyan = &tint.Color{R: 0, G: 84, B: 227, A: 255}  // #0054E3 Windows blue → focused window borders
	winXPTheme.BrightGreen = &tint.Color{R: 0, G: 84, B: 227, A: 255} // #0054E3 Windows blue → focused terminal borders
	tint.Register(&winXPTheme)
	tint.SetTintID("WindowsXP")
	program := tea.NewProgram(&appModel{inner: model}, tuios.ProgramOptions()...)
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}

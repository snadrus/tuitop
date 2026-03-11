package startmenu

import (
	"os"
	"os/user"
	"strings"

	"charm.land/lipgloss/v2"
)

// Colors and widths are passed from the host (main.go) per DESIGN_PHILOSOPHY.
// This package only handles layout and rendering.

// StyleConfig holds colors and widths for the start menu (all passed from host).
type StyleConfig struct {
	MenuBg           string // background of menu (e.g. taskbar blue)
	MenuFg           string // default text
	UserFg           string // user section text
	HighlightBg      string // hover/highlight (e.g. Windows blue)
	HighlightFg      string // highlight text
	AllProgramsFg    string // "All Programs" text
	MenuWidth        int    // total menu width in cols
	LeftColumnWidth  int    // left column width
	RightColumnWidth int    // right column width
}

// MenuItem is an icon + label for a menu entry (functionality comes later).
type MenuItem struct {
	Icon  string
	Label string
}

// RightItems returns the right-column items: folders, Control Panel, Search, Help, Exit.
func RightItems() []MenuItem {
	return []MenuItem{
		{Icon: "📁", Label: "~"},
		{Icon: "🖥️", Label: "Desktop"},
		{Icon: "💻", Label: "My Computer"},
		{Icon: "⚙️", Label: "Control Panel"},
		{Icon: "🔍", Label: "Search"},
		{Icon: "❓", Label: "Help"},
		{Icon: "🚪", Label: "Exit"},
	}
}

// Render returns the start menu as a string and its dimensions (width, height).
func Render(cfg StyleConfig) (content string, width, height int) {
	if cfg.MenuWidth <= 0 {
		cfg.MenuWidth = 28
	}
	if cfg.LeftColumnWidth <= 0 {
		cfg.LeftColumnWidth = (cfg.MenuWidth - 1) / 2 // -1 for divider
	}
	// Ensure left + divider + right = MenuWidth exactly (so body matches user line width)
	cfg.RightColumnWidth = cfg.MenuWidth - cfg.LeftColumnWidth - 1

	baseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.MenuFg)).
		Background(lipgloss.Color(cfg.MenuBg))

	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.UserFg)).
		Background(lipgloss.Color(cfg.MenuBg)).
		Padding(0, 1)

	allProgramsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.AllProgramsFg)).
		Background(lipgloss.Color(cfg.MenuBg)).
		Padding(0, 1)

	// User section: 👤 host\username
	hostname, _ := os.Hostname()
	username := "user"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	userLine := userStyle.Width(cfg.MenuWidth).Render("👤 " + hostname + "\\" + username)

	halfTop := baseStyle.Width(cfg.MenuWidth).Render(
		lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.MenuBg)).Background(lipgloss.Color("#000000")).Render(strings.Repeat("▄", cfg.MenuWidth)))

	// Left column: Recent apps (blank), All Programs
	leftCell := lipgloss.NewStyle().Width(cfg.LeftColumnWidth).Padding(0, 1).Background(lipgloss.Color(cfg.MenuBg)).Foreground(lipgloss.Color(cfg.MenuFg))
	leftLines := []string{
		leftCell.Render(""), // Recent apps - blank
		leftCell.Render(""),
		leftCell.Render(allProgramsStyle.Render("All Programs  ▶")),
	}
	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftLines...)

	// Right column: folders and items
	rightCell := lipgloss.NewStyle().Width(cfg.RightColumnWidth).Padding(0, 1).Background(lipgloss.Color(cfg.MenuBg)).Foreground(lipgloss.Color(cfg.MenuFg))
	rightLines := make([]string, 0, len(RightItems()))
	for _, it := range RightItems() {
		rightLines = append(rightLines, rightCell.Render(it.Icon+" "+it.Label))
	}
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightLines...)

	// Pad column heights to match
	for len(leftLines) < len(rightLines) {
		leftLines = append(leftLines, leftCell.Render(""))
	}
	for len(rightLines) < len(leftLines) {
		rightLines = append(rightLines, rightCell.Render(""))
	}
	leftCol = lipgloss.JoinVertical(lipgloss.Left, leftLines...)
	rightCol = lipgloss.JoinVertical(lipgloss.Left, rightLines...)

	leftPadded := lipgloss.NewStyle().
		Width(cfg.LeftColumnWidth).
		Background(lipgloss.Color(cfg.MenuBg)).
		Render(leftCol)
	rightPadded := lipgloss.NewStyle().
		Width(cfg.RightColumnWidth).
		Background(lipgloss.Color(cfg.MenuBg)).
		Render(rightCol)

	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.MenuFg)).
		Background(lipgloss.Color(cfg.MenuBg)).
		Width(1).
		Render("\u2502") // box drawing light vertical

	// Build rows: each row is leftCol line + divider + rightCol line
	leftSplit := splitLines(leftPadded)
	rightSplit := splitLines(rightPadded)
	maxRows := len(leftSplit)
	if len(rightSplit) > maxRows {
		maxRows = len(rightSplit)
	}

	var bodyRows []string
	for i := 0; i < maxRows; i++ {
		var leftPart, rightPart string
		if i < len(leftSplit) {
			leftPart = leftSplit[i]
		} else {
			leftPart = lipgloss.NewStyle().Width(cfg.LeftColumnWidth).Render("")
		}
		if i < len(rightSplit) {
			rightPart = rightSplit[i]
		} else {
			rightPart = lipgloss.NewStyle().Width(cfg.RightColumnWidth).Render("")
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, leftPart, divider, rightPart)
		bodyRows = append(bodyRows, baseStyle.Width(cfg.MenuWidth).Render(row))
	}
	body := lipgloss.JoinVertical(lipgloss.Left, bodyRows...)

	// Top: user line, then body
	full := lipgloss.JoinVertical(lipgloss.Left,
		baseStyle.Width(cfg.MenuWidth).Render(halfTop),
		baseStyle.Width(cfg.MenuWidth).Render(userLine),
		body,
	)

	// Ensure consistent width
	full = baseStyle.Width(cfg.MenuWidth).Render(full)
	_, height = getDimensions(full)
	width = cfg.MenuWidth
	return full, width, height
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func getDimensions(s string) (w, h int) {
	lines := splitLines(s)
	h = len(lines)
	for _, line := range lines {
		if lipgloss.Width(line) > w {
			w = lipgloss.Width(line)
		}
	}
	return w, h
}

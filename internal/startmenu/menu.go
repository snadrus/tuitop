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
	MenuBg           string
	MenuFg           string
	UserFg           string
	HighlightBg      string
	HighlightFg      string
	AllProgramsFg    string
	LeftColumnBg     string
	LeftColumnFg     string
	RightColumnBg    string
	RightColumnFg    string
	DividerFg        string
	DividerBg        string
	TopStripeUpperBg string
	ExitBarBg        string
	ExitBarFg        string
	UserIcon         string
	MenuWidth        int
	LeftColumnWidth  int
	RightColumnWidth int
}

// MenuItem is an icon + label for a menu entry (functionality comes later).
type MenuItem struct {
	Icon  string
	Label string
}

// ClickHitRightColumnLabel reports whether (localX, localY) hits a row in the right column
// that contains the given label text (e.g. "My Computer"). Coordinates are menu-local.
func ClickHitRightColumnLabel(cfg StyleConfig, localX, localY int, label string) bool {
	if localX < 0 || localY < 0 {
		return false
	}
	if cfg.MenuWidth <= 0 {
		cfg.MenuWidth = 28
	}
	if cfg.LeftColumnWidth <= 0 {
		cfg.LeftColumnWidth = (cfg.MenuWidth - 1) / 2
	}
	content, _, h := Render(cfg)
	if localY >= h {
		return false
	}
	lines := splitLines(content)
	if localY >= len(lines) {
		return false
	}
	line := lines[localY]
	if !strings.Contains(line, label) {
		return false
	}
	// Right column starts after left pane + vertical divider (see Render).
	minX := cfg.LeftColumnWidth + 1
	return localX >= minX
}

// ClickIsExit reports whether a click at (localX, localY) hits the Exit label on the bottom bar (right-justified).
func ClickIsExit(cfg StyleConfig, localX, localY int) bool {
	if localX < 0 || localY < 0 {
		return false
	}
	if cfg.MenuWidth <= 0 {
		cfg.MenuWidth = 28
	}
	content, _, h := Render(cfg)
	if localY >= h {
		return false
	}
	lines := splitLines(content)
	if localY >= len(lines) {
		return false
	}
	items := RightItems()
	exitItem := items[len(items)-1]
	exitText := fixEmojiWidth(exitItem.Icon) + " " + exitItem.Label
	if !strings.Contains(lines[localY], exitItem.Label) {
		return false
	}
	// Exit row is full width; only the right-justified text cells count.
	tw := lipgloss.Width(exitText)
	if tw <= 0 {
		return false
	}
	startX := cfg.MenuWidth - tw
	return localX >= startX && localX < cfg.MenuWidth
}

// RightItems returns the right-column items: folders, Control Panel, Search, Help, Exit.
func RightItems() []MenuItem {
	return []MenuItem{
		{Icon: "📁", Label: "Home Folder"},
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
		cfg.LeftColumnWidth = (cfg.MenuWidth - 1) / 2
	}
	cfg.RightColumnWidth = cfg.MenuWidth - cfg.LeftColumnWidth - 1

	leftBg, leftFg := cfg.LeftColumnBg, cfg.LeftColumnFg
	if leftBg == "" {
		leftBg = cfg.MenuBg
	}
	if leftFg == "" {
		leftFg = cfg.MenuFg
	}
	rightBg, rightFg := cfg.RightColumnBg, cfg.RightColumnFg
	if rightBg == "" {
		rightBg = cfg.MenuBg
	}
	if rightFg == "" {
		rightFg = cfg.MenuFg
	}
	divFg, divBg := cfg.DividerFg, cfg.DividerBg
	if divFg == "" {
		divFg = leftFg
	}
	if divBg == "" {
		divBg = leftBg
	}
	apFg := cfg.AllProgramsFg
	if apFg == "" {
		apFg = leftFg
	}

	baseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.MenuFg)).
		Background(lipgloss.Color(cfg.MenuBg))

	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.UserFg)).
		Background(lipgloss.Color(cfg.MenuBg)).
		Padding(0, 1)

	allProgramsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(apFg)).
		Background(lipgloss.Color(leftBg)).
		Padding(0, 1)

	hostname, _ := os.Hostname()
	username := "user"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	userIcon := cfg.UserIcon
	if userIcon == "" {
		userIcon = "👤 "
	} else if !strings.HasSuffix(userIcon, " ") {
		userIcon += " "
	}
	userLine := userStyle.Width(cfg.MenuWidth).Render(userIcon + hostname + "\\" + username)

	topUpper := cfg.TopStripeUpperBg
	if topUpper == "" {
		topUpper = "#000000"
	}
	halfTop := baseStyle.Width(cfg.MenuWidth).Render(
		lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.MenuBg)).Background(lipgloss.Color(topUpper)).Render(strings.Repeat("▄", cfg.MenuWidth)))

	// --- layout restored from pre-refactor: two vertical stacks, pad heights, join per row ---
	leftCell := lipgloss.NewStyle().
		Width(cfg.LeftColumnWidth).
		Padding(0, 1).
		Background(lipgloss.Color(leftBg)).
		Foreground(lipgloss.Color(leftFg))
	leftLines := []string{
		leftCell.Render(""),
		leftCell.Render(""),
		leftCell.Render(allProgramsStyle.Render("All Programs  ▶")),
	}
	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftLines...)

	rightCell := lipgloss.NewStyle().
		Width(cfg.RightColumnWidth).
		Padding(0, 1).
		Background(lipgloss.Color(rightBg)).
		Foreground(lipgloss.Color(rightFg))
	rightItems := RightItems()
	exitItem := rightItems[len(rightItems)-1]
	rightColumnItems := rightItems[:len(rightItems)-1]
	rightLines := make([]string, 0, len(rightColumnItems))
	for _, it := range rightColumnItems {
		rightLines = append(rightLines, rightCell.Render(fixEmojiWidth(it.Icon)+" "+it.Label))
	}
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightLines...)

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
		Background(lipgloss.Color(leftBg)).
		Render(leftCol)
	rightPadded := lipgloss.NewStyle().
		Width(cfg.RightColumnWidth).
		Background(lipgloss.Color(rightBg)).
		Render(rightCol)

	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color(divFg)).
		Background(lipgloss.Color(divBg)).
		Width(1).
		Render("\u2502")

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
			leftPart = lipgloss.NewStyle().Width(cfg.LeftColumnWidth).Background(lipgloss.Color(leftBg)).Render("")
		}
		if i < len(rightSplit) {
			rightPart = rightSplit[i]
		} else {
			rightPart = lipgloss.NewStyle().Width(cfg.RightColumnWidth).Background(lipgloss.Color(rightBg)).Render("")
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, leftPart, divider, rightPart)
		// Do not wrap body rows in baseStyle: it would repaint both panes with MenuBg and hide per-column colors.
		bodyRows = append(bodyRows, row)
	}
	body := lipgloss.JoinVertical(lipgloss.Left, bodyRows...)

	exitBarBg := cfg.ExitBarBg
	if exitBarBg == "" {
		exitBarBg = cfg.MenuBg
	}
	exitBarFg := cfg.ExitBarFg
	if exitBarFg == "" {
		exitBarFg = cfg.MenuFg
	}
	exitText := fixEmojiWidth(exitItem.Icon) + " " + exitItem.Label
	exitBar := lipgloss.NewStyle().
		Width(cfg.MenuWidth).
		Align(lipgloss.Right).
		Background(lipgloss.Color(exitBarBg)).
		Foreground(lipgloss.Color(exitBarFg)).
		Render(exitText)

	full := lipgloss.JoinVertical(lipgloss.Left,
		halfTop,
		baseStyle.Width(cfg.MenuWidth).Render(userLine),
		body,
		exitBar,
	)

	_, height = getDimensions(full)
	width = cfg.MenuWidth
	return full, width, height
}

// fixEmojiWidth replaces U+FE0F (emoji presentation selector) with a space.
// Terminals often ignore FE0F for cursor advance, so the width library over-counts
// by 1 for affected grapheme clusters. Swapping FE0F → space keeps the measured
// width equal to the physical terminal width.
func fixEmojiWidth(s string) string {
	return strings.ReplaceAll(s, "\uFE0F", " ")
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

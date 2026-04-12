package startmenu

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestRenderRightColumnKeepsIcons(t *testing.T) {
	cfg := StyleConfig{
		MenuBg:           "#0054E3",
		MenuFg:           "#ffffff",
		UserFg:           "#ffffff",
		HighlightBg:      "#1996E9",
		HighlightFg:      "#ffffff",
		AllProgramsFg:    "#ffffff",
		MenuWidth:        40,
		LeftColumnWidth:  18,
		RightColumnWidth: 21,
	}
	content, _, _ := Render(cfg)
	for _, it := range RightItems() {
		if it.Icon == "" {
			continue
		}
		wantIcon := fixEmojiWidth(it.Icon)
		if !strings.Contains(content, wantIcon) {
			t.Errorf("rendered menu missing icon %q for label %q", wantIcon, it.Label)
		}
	}
}

func TestRenderExitLineIndex(t *testing.T) {
	cfg := StyleConfig{
		MenuBg:           "#0054E3",
		MenuFg:           "#ffffff",
		UserFg:           "#ffffff",
		HighlightBg:      "#1996E9",
		HighlightFg:      "#ffffff",
		AllProgramsFg:    "#ffffff",
		MenuWidth:        40,
		LeftColumnWidth:  18,
		RightColumnWidth: 21,
	}
	content, _, h := Render(cfg)
	lines := splitLines(content)
	if len(lines) != h {
		t.Fatalf("splitLines len %d != Render height %d", len(lines), h)
	}
	var exitLines []int
	for i, line := range lines {
		if strings.Contains(line, "Exit") {
			exitLines = append(exitLines, i)
		}
	}
	var controlLines int
	for _, line := range lines {
		if strings.Contains(line, "Control Panel") {
			controlLines++
		}
	}
	if controlLines != 1 {
		t.Fatalf("Control Panel should be on exactly one rendered line, got %d matches", controlLines)
	}

	if len(exitLines) != 1 {
		t.Fatalf("expected exactly one line containing Exit, got %v (lines=%d)", exitLines, len(lines))
	}
	got := exitLines[0]
	items := RightItems()
	exitItem := items[len(items)-1]
	exitText := exitItem.Icon + " " + exitItem.Label
	tw := lipgloss.Width(exitText)
	exitStartX := cfg.MenuWidth - tw
	if !ClickIsExit(cfg, exitStartX, got) {
		t.Errorf("ClickIsExit(cfg, %d, %d) = false (exit text width %d)", exitStartX, got, tw)
	}
	if ClickIsExit(cfg, exitStartX-1, got) {
		t.Error("click left of exit label should not be Exit")
	}
}

package termshim

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type tcSurf struct {
	width   int
	height  int
	x       int
	y       int
	surface tcell.Screen
	cache   []scCall
}
type scCall struct {
	x, y  int
	ch    rune
	comb  []rune
	style tcell.Style
}

// tcSurf implements tcellterm.Surface
func (t *tcSurf) Size() (int, int) {
	return t.width, t.height
}
func (t *tcSurf) SetContent(x, y int, ch rune, comb []rune, style tcell.Style) {
	if ch != 0 {
		log(fmt.Sprintf("SetContent(%d, %d, %c, %v, %v)\n", x, y, ch, comb, style))
	}
	if t.surface == nil {
		t.cache = append(t.cache, scCall{x, y, ch, comb, style})
		return
	}
	phyX, phyY := x+t.x, y+t.y
	if phyX < 0 || phyY < 0 || x >= t.width || y >= t.height {
		return
	}
	t.surface.SetContent(phyX, phyY, ch, comb, style)
}
func (t *tcSurf) SetSurface(s tcell.Screen) {
	t.surface = s
	for _, c := range t.cache {
		t.SetContent(c.x, c.y, c.ch, c.comb, c.style)
	}
	t.cache = nil
}

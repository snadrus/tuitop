package desktopbg

import "testing"

func TestOverlayTopLines(t *testing.T) {
	if got := OverlayTopLines("a\nb\nc", 2); got != "a\nb" {
		t.Fatalf("clip: %q", got)
	}
	if got := OverlayTopLines("a", 3); got != "a\n\n" {
		t.Fatalf("pad: %q", got)
	}
	if got := OverlayTopLines("x\r\ny", 2); got != "x\ny" {
		t.Fatalf("crlf: %q", got)
	}
}

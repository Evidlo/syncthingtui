package app

import (
	"regexp"
	"strings"
	"testing"
)

var ansiRE = regexp.MustCompile("\x1b\\[[0-9;]*m")

// TestScreensFitWidth statically renders every screen preset at several sizes
// and fails if any line exceeds the terminal width (AGENTS.md: min 80x60).
func TestScreensFitWidth(t *testing.T) {
	sizes := [][2]int{{80, 60}, {80, 24}, {120, 40}}
	for _, name := range ScreenNames() {
		for _, sz := range sizes {
			out, err := RenderScreen(name, sz[0], sz[1])
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			for i, line := range strings.Split(out, "\n") {
				w := len([]rune(ansiRE.ReplaceAllString(line, "")))
				if w > sz[0] {
					t.Errorf("%s at %dx%d: line %d is %d cols", name, sz[0], sz[1], i+1, w)
				}
			}
		}
	}
}

package app

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Regression for tall menus: forms/lists taller than the window must scroll
// so the focused item (and the trailing button row) stays visible.

func plain(s string) string { return ansiRE.ReplaceAllString(s, "") }

func TestFormScrollsToCursor(t *testing.T) {
	fields := make([]FormField, 30)
	for i := range fields {
		fields[i] = Bool(fmt.Sprintf("Device %02d", i), "", false)
	}
	f := NewForm([]string{"Save", "Close"}, fields...)
	f, _ = f.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if v := plain(f.View(16)); !strings.Contains(v, "Device 00") {
		t.Errorf("cursor at top: first field not visible:\n%s", v)
	}
	for range fields { // move onto the button row
		f, _ = f.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	v := plain(f.View(16))
	if !strings.Contains(v, "▸ [ Save ]") {
		t.Errorf("cursor on buttons: button row not visible:\n%s", v)
	}
	if !strings.Contains(v, "⋮") {
		t.Errorf("expected a scroll marker for clipped content above:\n%s", v)
	}
}

func TestListScrollsToCursor(t *testing.T) {
	m := NewMain(nil)
	m.folders = nil
	for i := 0; i < 40; i++ {
		m.folders = append(m.folders, Folder{ID: fmt.Sprint(i), Label: fmt.Sprintf("folder%02d", i), State: "Up to Date"})
	}
	m.width, m.height = 80, 24
	m.folderIdx = len(m.folders) + 1 // Rescan All, the very last row
	if v := plain(m.viewFolders()); !strings.Contains(v, "▸ ⟳ Rescan All") {
		t.Errorf("last list item not visible when selected:\n%s", v)
	}
	m.folderIdx = 0
	if v := plain(m.viewFolders()); !strings.Contains(v, "▸ ● folder00") {
		t.Errorf("first list item not visible when selected:\n%s", v)
	}
}

func TestScrollToCursorMarkers(t *testing.T) {
	body := ""
	for i := 0; i < 30; i++ {
		body += fmt.Sprintf("line%02d\n", i)
	}
	mid := plain(scrollToCursor(body, 15, 10))
	for _, want := range []string{"⋮", "line15"} {
		if !strings.Contains(mid, want) {
			t.Errorf("mid scroll: missing %q:\n%s", want, mid)
		}
	}
	if top := plain(scrollToCursor(body, 0, 10)); !strings.HasPrefix(top, "line00") {
		t.Errorf("top scroll: cursor line hidden by marker:\n%s", top)
	}
	if bot := plain(scrollToCursor(body, 29, 10)); !strings.HasSuffix(strings.TrimRight(bot, "\n"), "line29") {
		t.Errorf("bottom scroll: cursor line hidden by marker:\n%s", bot)
	}
	if short := scrollToCursor("a\nb\n", 0, 10); short != "a\nb\n" {
		t.Errorf("short body must pass through unchanged, got %q", short)
	}
}

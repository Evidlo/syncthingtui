package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Subviews must show the live rates forwarded by the App, not a constant.
func TestSubviewsShowForwardedRates(t *testing.T) {
	const rates = "↓7.7KiB/s ↑8.8KiB/s"
	subs := map[string]tea.Model{
		"FormView": NewSettings(),
		"ShowID":   NewShowID(DeviceID),
		"About":    NewAbout("v2.0.13", AboutPaths),
	}
	for name, sub := range subs {
		// wide enough that the status bar keeps rates next to the key hints
		sub, _ = sub.Update(tea.WindowSizeMsg{Width: 100, Height: 60})
		sub, _ = sub.Update(ratesMsg(rates))
		if v := plain(sub.View()); !strings.Contains(v, rates) {
			t.Errorf("%s: forwarded rates not shown", name)
		}
	}
}

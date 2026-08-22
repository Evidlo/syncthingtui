// syncthingtui - a terminal user interface for Syncthing
// Copyright (C) 2026 Evan Widloski
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: GPL-3.0-or-later

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

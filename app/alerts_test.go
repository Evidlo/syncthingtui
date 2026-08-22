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
)

// TestAlertButtonsWired presses every button of every alert kind and fails if
// any falls through to an unwired/TODO stub. This is the regression test for
// the bug where "Share" silently did nothing (the alert vanished locally and
// reappeared on the next poll).
func TestAlertButtonsWired(t *testing.T) {
	for _, alert := range Alerts { // fake data covers all kinds incl. New Folder
		for btn, label := range alert.Buttons {
			m := NewMain(nil)
			m.active = tabAlerts
			m.alerts = []Alert{alert}
			m.alertIdx, m.alertBtn = 0, btn
			res, _ := m.updateAlerts("enter")
			mm := res.(MainModel)
			for _, bad := range []string{"not wired", "TODO", "not in mockup"} {
				if strings.Contains(mm.flash, bad) {
					t.Errorf("%s/%s: unwired handler (flash %q)", alert.Title, label, mm.flash)
				}
			}
			if len(mm.alerts) != 0 {
				t.Errorf("%s/%s: alert not consumed in fake mode", alert.Title, label)
			}
		}
	}
}

// Every alert kind used by the live layer must exist in the fake data, or the
// dispatch test above silently loses coverage.
func TestFakeAlertsCoverLiveButtons(t *testing.T) {
	want := []string{"Add Device", "Share", "Add", "Ignore", "Dismiss", "OK"}
	have := map[string]bool{}
	for _, a := range Alerts {
		for _, b := range a.Buttons {
			have[b] = true
		}
	}
	for _, b := range want {
		if !have[b] {
			t.Errorf("fake alert data missing button %q", b)
		}
	}
}

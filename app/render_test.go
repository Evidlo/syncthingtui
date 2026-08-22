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

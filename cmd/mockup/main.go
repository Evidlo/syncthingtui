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

// Command mockup renders a named mockup screen at a fixed size for static
// layout review (see AGENTS.md: static TUI rendering requirement).
//
//	go run ./cmd/mockup <screen> [WxH]   # default 80x60
//	go run ./cmd/mockup -list
package main

import (
	"fmt"
	"os"

	"github.com/evidlo/syncthingtui/mockups"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "-list" {
		for _, n := range mockups.Names() {
			fmt.Println(n)
		}
		return
	}
	width, height := 80, 60
	if len(args) > 1 {
		fmt.Sscanf(args[1], "%dx%d", &width, &height)
	}
	render, ok := mockups.Screens[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown screen %q; run with -list\n", args[0])
		os.Exit(1)
	}
	fmt.Println(render(width, height))
}

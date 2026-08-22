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

// Read-only check against the real local syncthing via Discover().
package main

import (
	"fmt"

	"github.com/evidlo/syncthingtui/client"
)

func main() {
	addr, key, err := client.Discover()
	fmt.Printf("discover: addr=%q key=%q... err=%v\n", addr, key[:min(8, len(key))], err)
	if err != nil {
		return
	}
	c := client.New(addr, key)
	cfg, err := c.Config()
	fmt.Printf("config: folders=%d devices=%d err=%v\n", len(cfg.Folders), len(cfg.Devices), err)
	for _, f := range cfg.Folders {
		fmt.Printf("  folder id=%q label=%q paused=%v\n", f.ID, f.Label, f.Paused)
	}
	st, err := c.SystemStatus()
	fmt.Printf("status: myID=%.7s err=%v\n", st.MyID, err)
	for _, d := range cfg.Devices {
		fmt.Printf("  device id=%.7s name=%q\n", d.DeviceID, d.Name)
	}
}

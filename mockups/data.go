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

// Package mockups contains static, fake-data renderings of candidate TUI
// layouts for human review (AGENTS.md stage 2). Each screen renders at a
// given width/height so it can be checked in a virtual terminal.
package mockups

type folder struct {
	Label, ID, Path, State string
	Pct                    int // sync percentage
	Global, Local          string
	Shared                 string
}

type device struct {
	Name, State, Address string
	Pct                  int
	Download, Upload     string
}

type alert struct {
	Kind, Title, Short, Time, Body string
	Buttons                        []string
}

var fakeFolders = []folder{
	{"Documents", "abcde-fghij", "~/Documents", "Up to Date", 100, "12.3 GiB", "12.3 GiB", "nas, laptop"},
	{"Photos", "kl3mn-op4qr", "~/Pictures", "Syncing", 72, "48.1 GiB", "34.6 GiB", "nas"},
	{"Music", "st5uv-wx6yz", "~/Music", "Paused", 100, "21.0 GiB", "21.0 GiB", "nas, phone"},
	{"Work", "ab7cd-ef8gh", "~/work", "Out of Sync", 91, "3.2 GiB", "2.9 GiB", "laptop"},
}

var fakeDevices = []device{
	{"nas", "Up to Date", "10.0.0.2:22000", 100, "1.2 MiB/s", "340 KiB/s"},
	{"laptop", "Syncing (72%)", "10.0.0.5:22000", 72, "0 B/s", "2.1 MiB/s"},
	{"phone", "Disconnected", "", 100, "", ""},
}

var fakeAlerts = []alert{
	{"device", "New Device", "workpc", "2026-07-12 09:14",
		`Device "workpc" (MFZWI3D-BONSGYC-... at 10.0.0.9:22000) wants to connect. Add new device?`,
		[]string{"Add Device", "Ignore", "Dismiss"}},
	{"folder", "Share Folder", "Books", "2026-07-12 09:20",
		`laptop wants to share folder "Books" (ij9kl-mn0op). Share this folder?`,
		[]string{"Share", "Ignore", "Dismiss"}},
}

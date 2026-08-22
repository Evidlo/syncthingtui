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

// Package app is the stage-3 interactive mockup: all chosen screen designs
// wired into one navigable bubbletea app, backed by fake data. REST wiring
// replaces this data in stage 4.
package app

type Folder struct {
	Label, ID, Path, State string
	Pct                    int
	Global, Local, Shared  string
}

type Device struct {
	ID, Name, State, Address string
	Pct                      int
	Download, Upload         string
	Compression, LastSeen    string
}

type Alert struct {
	Kind, Title, Short, Time, Body string
	ID, DeviceID, Addr             string // pending device/folder details (live mode)
	Buttons                        []string
}

var Folders = []Folder{
	{"Documents", "abcde-fghij", "~/Documents", "Up to Date", 100, "12.3 GiB", "12.3 GiB", "nas, laptop"},
	{"Photos", "kl3mn-op4qr", "~/Pictures", "Syncing", 72, "48.1 GiB", "34.6 GiB", "nas"},
	{"Music", "st5uv-wx6yz", "~/Music", "Paused", 100, "21.0 GiB", "21.0 GiB", "nas, phone"},
	{"Work", "ab7cd-ef8gh", "~/work", "Out of Sync", 91, "3.2 GiB", "2.9 GiB", "laptop"},
}

var Devices = []Device{
	{"", "nas", "Up to Date", "10.0.0.2:22000", 100, "1.2 MiB/s", "340 KiB/s", "Metadata Only", "2026-07-12 10:41"},
	{"", "laptop", "Syncing (72%)", "10.0.0.5:22000", 72, "0 B/s", "2.1 MiB/s", "Metadata Only", "2026-07-12 10:39"},
	{"", "phone", "Disconnected", "", 100, "", "", "All Data", "2026-07-10 08:12"},
}

var Alerts = []Alert{
	{"device", "New Device", "workpc", "2026-07-12 09:14",
		`Device "workpc" (MFZWI3D-BONSGYC-... at 10.0.0.9:22000) wants to connect. Add new device?`,
		"", "", "", []string{"Add Device", "Ignore", "Dismiss"}},
	{"folder", "Share Folder", "Books", "2026-07-12 09:20",
		`laptop wants to share folder "Books" (ij9kl-mn0op). Share this folder?`,
		"", "", "", []string{"Share", "Ignore", "Dismiss"}},
	{"folder", "New Folder", "Papers", "2026-07-12 09:25",
		`laptop wants to share folder "Papers" (qr1st-uv2wx). Add new folder?`,
		"", "", "", []string{"Add", "Ignore", "Dismiss"}},
	{"notice", "Notice", "", "2026-07-12 10:02",
		`Error on folder "Default Folder" (default): insufficient space on disk for database (~/.config/syncthing/index-v0.14.0.db): 0.2 % < 1 %`,
		"", "", "", []string{"OK"}},
}

var ThisDeviceStats = [][2]string{
	{"Download Rate", "1.2 MiB/s (4.3 GiB total)"},
	{"Upload Rate", "2.4 MiB/s (9.1 GiB total)"},
	{"Local State (Total)", "8,412 files, 84.6 GiB"},
	{"Listeners", "3/3"},
	{"Discovery", "4/5"},
	{"Uptime", "3d 2h 41m"},
	{"Identification", "MFZWI3D (this-machine)"},
	{"Version", "v2.0.13, Linux (64-bit)"},
}

var ActionItems = []string{"Settings", "Advanced", "Show ID", "Logs", "Support Bundle", "About", "Log Out", "Restart", "Shut Down"}

var AboutPaths = [][2]string{
	{"User Home", "/home/evan"},
	{"Configuration Directory", "/home/evan/.config/syncthing"},
	{"Configuration File", "/home/evan/.config/syncthing/config.xml"},
	{"Device Certificate", "/home/evan/.config/syncthing/cert.pem"},
	{"GUI / API HTTPS Certificate", "/home/evan/.config/syncthing/https-cert.pem"},
	{"Database Location", "/home/evan/.local/share/syncthing"},
	{"Log File", "-"},
	{"GUI Override Directory", "/home/evan/.config/syncthing/gui"},
}

const DeviceID = "MFZWI3D-BONSGYC-YLTMRWG-C43ENR5-QXGZDMM-FZWI3DP-BONSGYC-YLTMRW4"

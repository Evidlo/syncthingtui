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
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evidlo/syncthingtui/client"
)

// Live data: when a client is set, the App polls syncthing every pollInterval
// and feeds dataMsg into MainModel, replacing the fake data.
// TODO stage 4: switch to /rest/events long-poll instead of fixed polling.

const pollInterval = 3 * time.Second

type dataMsg struct {
	folders           []Folder
	devices           []Device
	alerts            []Alert
	stats             [][2]string
	stVersion         string
	myID              string
	paths             [][2]string
	inTotal, outTotal int64
	err               error
}

type pollTickMsg struct{}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func humanRate(bytesPerSec float64) string {
	return humanBytes(int64(bytesPerSec)) + "/s"
}

// lastSeenValid reports whether a device has ever actually been seen.
// syncthing reports the 1970 epoch (not Go's zero time) for never-seen devices.
func lastSeenValid(t time.Time) bool {
	return !t.IsZero() && t.Year() > 1971
}

// folderState mirrors the web GUI's folderStatus(): a "busy" db.State maps
// straight to a label; an idle folder is refined (out of sync / failed items /
// local additions / unshared / up to date) in the GUI's precedence order.
func folderState(cfg client.FolderCfg, db client.DBStatus) (string, int) {
	pct := 100
	if db.GlobalBytes > 0 {
		pct = int(100 * (db.GlobalBytes - db.NeedBytes) / db.GlobalBytes)
	}
	if cfg.Paused {
		return "Paused", pct
	}
	if db.State == "" {
		return "Unknown", pct
	}
	switch db.State {
	case "error":
		return "Stopped", pct
	case "scanning":
		return "Scanning", pct
	case "syncing":
		return "Syncing", pct
	case "sync-preparing":
		return "Preparing to Sync", pct
	case "cleaning":
		return "Cleaning Versions", pct
	case "sync-waiting":
		return "Waiting to Sync", pct
	case "scan-waiting":
		return "Waiting to Scan", pct
	case "clean-waiting":
		return "Waiting to Clean", pct
	}
	// db.State == "idle"
	switch {
	case db.NeedTotalItems > 0:
		return "Out of Sync", pct
	case db.PullErrors > 0:
		return "Failed Items", pct
	case db.ReceiveOnlyTotalItems > 0:
		if cfg.Type == "receiveonly" {
			return "Local Additions", pct
		}
		return "Local Data Unencrypted", pct
	case len(cfg.Devices) <= 1:
		return "Unshared", pct
	}
	return "Up to Date", pct
}

// fetch gathers everything the main view needs in one shot.
func fetch(c *client.Client) dataMsg {
	cfg, err := c.Config()
	if err != nil {
		return dataMsg{err: err}
	}
	status, err := c.SystemStatus()
	if err != nil {
		return dataMsg{err: err}
	}
	conns, _ := c.Connections()
	version, _ := c.Version()
	devStats, _ := c.DeviceStats()

	names := map[string]string{}
	for _, d := range cfg.Devices {
		names[d.DeviceID] = d.Name
	}

	var folders []Folder
	var localBytes, localFiles int64
	for _, f := range cfg.Folders {
		db, _ := c.DBStatus(f.ID)
		state, pct := folderState(f, db)
		var shared []string
		for _, d := range f.Devices {
			if d.DeviceID != status.MyID {
				shared = append(shared, names[d.DeviceID])
			}
		}
		label := f.Label
		if label == "" {
			label = f.ID
		}
		folders = append(folders, Folder{
			Label: label, ID: f.ID, Path: f.Path, State: state, Pct: pct,
			Global: humanBytes(db.GlobalBytes), Local: humanBytes(db.LocalBytes),
			Shared: strings.Join(shared, ", "),
		})
		localBytes += db.LocalBytes
		localFiles += db.LocalFiles
	}

	// A device is "unused" when it shares no folders with us; the web GUI
	// labels such devices "(Unused)" instead of showing sync progress.
	shares := map[string]bool{}
	for _, f := range cfg.Folders {
		for _, fd := range f.Devices {
			shares[fd.DeviceID] = true
		}
	}

	var devices []Device
	for _, d := range cfg.Devices {
		if d.DeviceID == status.MyID {
			continue
		}
		conn, known := conns.Connections[d.DeviceID]
		ds, hasStats := devStats[d.DeviceID]
		unused := !shares[d.DeviceID]
		suffix := ""
		if unused {
			suffix = " (Unused)"
		}
		state, pct := "Disconnected"+suffix, 100
		switch {
		case !known:
			// no connection record yet (e.g. just after startup)
			state = "Unknown"
		case d.Paused:
			state = "Paused" + suffix
		case conn.Connected && unused:
			state = "Connected (Unused)"
		case conn.Connected:
			comp, err := c.DeviceCompletion(d.DeviceID)
			if err == nil && comp.Completion >= 100 {
				state = "Up to Date"
			} else {
				state, pct = fmt.Sprintf("Syncing (%.0f%%)", comp.Completion), int(comp.Completion)
			}
		case !unused && hasStats && lastSeenValid(ds.LastSeen) &&
			time.Since(ds.LastSeen) >= 7*24*time.Hour:
			// shared device unseen for a week → the GUI's "inactive" state
			state = "Disconnected (Inactive)"
		}
		lastSeen := "-"
		if hasStats && lastSeenValid(ds.LastSeen) {
			lastSeen = ds.LastSeen.Local().Format("2006-01-02 15:04")
		}
		devices = append(devices, Device{
			ID: d.DeviceID, Name: d.Name, State: state, Address: conn.Address, Pct: pct,
			Compression: d.Compression, LastSeen: lastSeen,
		})
	}

	var alerts []Alert
	pendDev, _ := c.PendingDevices()
	for id, p := range pendDev {
		name := p.Name
		if name == "" {
			name = id[:7]
		}
		alerts = append(alerts, Alert{
			Kind: "device", Title: "New Device", Short: name, ID: id, Addr: p.Address,
			Time:    p.Time.Local().Format("2006-01-02 15:04:05"),
			Body:    fmt.Sprintf("Device %q (%s at %s) wants to connect. Add new device?", p.Name, id, p.Address),
			Buttons: []string{"Add Device", "Ignore", "Dismiss"},
		})
	}
	knownFolder := map[string]bool{}
	for _, f := range cfg.Folders {
		knownFolder[f.ID] = true
	}
	pendFold, _ := c.PendingFolders()
	for id, p := range pendFold {
		for devID, offer := range p.OfferedBy {
			label := offer.Label
			if label == "" {
				label = id
			}
			// like the web GUI: known folder → share it; unknown → add it
			title, question, accept := "New Folder", "Add new folder?", "Add"
			if knownFolder[id] {
				title, question, accept = "Share Folder", "Share this folder?", "Share"
			}
			alerts = append(alerts, Alert{
				Kind: "folder", Title: title, Short: label, ID: id, DeviceID: devID,
				Time:    offer.Time.Local().Format("2006-01-02 15:04:05"),
				Body:    fmt.Sprintf("%s wants to share folder %q (%s). %s", names[devID], label, id, question),
				Buttons: []string{accept, "Ignore", "Dismiss"},
			})
		}
	}
	// The REST API can only clear all system errors at once, so all notices
	// are combined into a single alert (OK clears them together).
	sysErrs, _ := c.Errors()
	if len(sysErrs) > 0 {
		var b strings.Builder
		for i, e := range sysErrs {
			if i > 0 {
				b.WriteString("\n\n")
			}
			fmt.Fprintf(&b, "%s: %s", e.When.Local().Format("2006-01-02 15:04:05"), e.Message)
		}
		alerts = append(alerts, Alert{
			Kind: "notice", Title: "Notice",
			Time:    sysErrs[len(sysErrs)-1].When.Local().Format("2006-01-02 15:04:05"),
			Body:    b.String(),
			Buttons: []string{"OK"},
		})
	}
	sort.Slice(alerts, func(i, j int) bool { return alerts[i].Time < alerts[j].Time })

	countOK := func(m map[string]client.ServiceStatus) string {
		ok := 0
		for _, v := range m {
			if v.Error == nil {
				ok++
			}
		}
		return fmt.Sprintf("%d/%d", ok, len(m))
	}
	up := time.Duration(status.Uptime) * time.Second
	stats := [][2]string{
		{"Download Rate", "-"}, // filled in by App from totals delta
		{"Upload Rate", "-"},
		{"Local State (Total)", fmt.Sprintf("%d files, %s", localFiles, humanBytes(localBytes))},
		{"Listeners", countOK(status.ConnectionServiceStatus)},
		{"Discovery", countOK(status.DiscoveryStatus)},
		{"Uptime", fmt.Sprintf("%dd %dh %dm", int(up.Hours())/24, int(up.Hours())%24, int(up.Minutes())%60)},
		{"Identification", status.MyID[:7]},
		{"Version", fmt.Sprintf("%s, %s (%s)", version.Version, version.OS, version.Arch)},
	}

	var paths [][2]string
	if pm, err := c.Paths(); err == nil {
		keys := make([]string, 0, len(pm))
		for k := range pm {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			paths = append(paths, [2]string{k, pm[k]})
		}
	}

	return dataMsg{folders: folders, devices: devices, alerts: alerts, stats: stats,
		stVersion: fmt.Sprintf("%s, %s (%s)", version.Version, version.OS, version.Arch),
		myID:      status.MyID,
		paths:     paths,
		inTotal:   conns.Total.InBytesTotal, outTotal: conns.Total.OutBytesTotal}
}

func fetchCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg { return fetch(c) }
}

func pollCmd() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg { return pollTickMsg{} })
}

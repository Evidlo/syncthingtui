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

func folderState(cfg client.FolderCfg, db client.DBStatus) (string, int) {
	pct := 100
	if db.GlobalBytes > 0 {
		pct = int(100 * (db.GlobalBytes - db.NeedBytes) / db.GlobalBytes)
	}
	switch {
	case cfg.Paused:
		return "Paused", pct
	case db.State == "syncing", db.State == "sync-preparing":
		return "Syncing", pct
	case db.State == "scanning":
		return "Scanning", pct
	case db.State == "error", db.State == "stopped":
		return "Error", pct
	case db.NeedBytes > 0:
		return "Out of Sync", pct
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
		conn := conns.Connections[d.DeviceID]
		unused := !shares[d.DeviceID]
		state, pct := "Disconnected", 100
		switch {
		case d.Paused:
			state = "Paused"
		case conn.Connected && unused:
			state = "Connected (Unused)"
		case conn.Connected:
			comp, err := c.DeviceCompletion(d.DeviceID)
			if err == nil && comp.Completion < 100 {
				state, pct = fmt.Sprintf("Syncing (%.0f%%)", comp.Completion), int(comp.Completion)
			} else {
				state = "Up to Date"
			}
		case unused:
			state = "Disconnected (Unused)"
		}
		lastSeen := "-"
		if ds, ok := devStats[d.DeviceID]; ok && !ds.LastSeen.IsZero() {
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

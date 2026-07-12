package app

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/evidlo/syncthingtui/client"
)

// FormView is the generic subview: infobar chrome, boxed sub-tabs, one Form
// per section. tab/⇧tab/numbers switch sections (when not editing); esc backs
// out (root App), or cancels the current field edit first.
//
// save/remove run against syncthing; when nil the buttons just close (fake
// mode / not-yet-wired views).
type FormView struct {
	title, sub    string
	sections      []string
	forms         []Form
	active        int
	width, height int
	save          func(vals map[string]any) error
	remove        func() error
}

func (v FormView) Init() tea.Cmd { return nil }

// allValues merges every section's savable fields.
func (v FormView) allValues() map[string]any {
	m := map[string]any{}
	for _, f := range v.forms {
		for k, val := range f.values() {
			m[k] = val
		}
	}
	return m
}

func (v FormView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width, v.height = msg.Width, msg.Height
		for i := range v.forms {
			v.forms[i], _ = v.forms[i].Update(msg)
		}
		return v, nil
	case tea.KeyMsg:
		if !v.forms[v.active].Editing() {
			key := msg.String()
			switch {
			case key == "tab":
				v.active = (v.active + 1) % len(v.sections)
				return v, nil
			case key == "shift+tab":
				v.active = (v.active + len(v.sections) - 1) % len(v.sections)
				return v, nil
			case len(key) == 1 && key[0] >= '1' && key[0] <= '0'+byte(len(v.sections)):
				v.active = int(key[0] - '1')
				return v, nil
			}
		}
	case buttonMsg:
		switch {
		case msg.label == "Save" && v.save != nil:
			vals, save, title := v.allValues(), v.save, v.title
			return v, func() tea.Msg {
				if err := save(vals); err != nil {
					return closeMsg{flash: title + " save failed: " + err.Error()}
				}
				return closeMsg{flash: title + " saved"}
			}
		case msg.label == "Remove" && v.remove != nil:
			remove, title := v.remove, v.title
			return v, func() tea.Msg {
				if err := remove(); err != nil {
					return closeMsg{flash: title + " remove failed: " + err.Error()}
				}
				return closeMsg{flash: title + " removed"}
			}
		}
		return v, func() tea.Msg { return closeMsg{flash: msg.label + ": " + v.title + " (not wired)"} }
	}
	var cmd tea.Cmd
	v.forms[v.active], cmd = v.forms[v.active].Update(msg)
	return v, cmd
}

// Editing reports whether the active section is in field-edit mode (esc then
// cancels the edit instead of leaving the subview).
func (v FormView) Editing() bool { return v.forms[v.active].Editing() }

func (v FormView) View() string {
	form := v.forms[v.active]
	body := boxedTabs(v.sections, v.active) + "\n" +
		lipgloss.NewStyle().PaddingLeft(2).Render(form.View())
	keys := keyHint("↑↓", "field", "enter", "edit/commit", "tab/⇧tab", "switch tabs", "esc", "cancel/back")
	if form.Editing() {
		keys = keyHint("enter", "commit", "esc", "cancel")
		if form.fields[form.cursor].kind == ftArea {
			keys = keyHint("ctrl+s", "commit", "esc", "cancel")
		}
	}
	return page(infobarTop(v.title, v.sub, v.width), body, keys, netRates, v.width, v.height)
}

// ── Edit Folder ──────────────────────────────────────────────────────────────

var subviewButtons = []string{"Save", "Close", "Remove"}

// randomFolderID mimics the web GUI's generated IDs, e.g. "abcde-fghij".
func randomFolderID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz234567"
	b := make([]byte, 11)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	b[5] = '-'
	return string(b)
}

// NewEditFolder is the fake-mode/mockup constructor.
func NewEditFolder(label string, isNew bool) FormView {
	f := client.FolderCfg{
		ID: "ab7cd-ef8gh", Label: label, Path: "~/" + label,
		Type: "sendreceive", Order: "random", RescanIntervalS: 3600,
		FSWatcherEnabled: true,
		Versioning:       client.VersioningCfg{Type: "simple", Params: map[string]string{"keep": "5"}, CleanupIntervalS: 3600},
	}
	devs := make([]client.DeviceCfg, len(Devices))
	for i, d := range Devices {
		devs[i] = client.DeviceCfg{DeviceID: d.Name, Name: d.Name}
	}
	f.Devices = []client.FolderDevice{{DeviceID: "nas"}}
	return newEditFolder(nil, f, devs, "", []string{"// One pattern per line", "*.tmp", "(?d).DS_Store"}, isNew)
}

// newEditFolder builds the folder editor; with a non-nil client Save/Remove
// hit the REST config API. Field keys follow the folder config schema
// (versioning.* and ignores are assembled specially in save).
func newEditFolder(c *client.Client, f client.FolderCfg, devs []client.DeviceCfg, myID string, ignores []string, isNew bool) FormView {
	title, sub := "Edit Folder", f.Label
	if isNew {
		title, sub = "Add Folder", "new"
		if f.ID == "" { // pending offers arrive with the remote's folder ID
			f.ID = randomFolderID()
		}
	}
	if f.Versioning.Params == nil {
		f.Versioning.Params = map[string]string{}
	}
	idField := RO("Folder ID", "Same on all cluster devices", f.ID)
	if isNew {
		idField = Text("Folder ID", "Same on all cluster devices", f.ID).K("id")
	}
	general := NewForm(subviewButtons,
		Text("Folder Label", "Optional label, can differ per device", f.Label).K("label"),
		idField,
		Text("Folder Path", "Created if it does not exist", f.Path).K("path"),
	)
	shared := map[string]bool{}
	for _, d := range f.Devices {
		shared[d.DeviceID] = true
	}
	var shareFields []FormField
	for _, d := range devs {
		if d.DeviceID == myID {
			continue
		}
		shareFields = append(shareFields,
			Bool(d.Name, "Share this folder with "+d.Name, shared[d.DeviceID]).K("device:"+d.DeviceID))
	}
	sharing := NewForm(subviewButtons, shareFields...)
	vers := NewForm(subviewButtons,
		Sel("File Versioning", "How to handle old versions",
			[]string{"No File Versioning", "Trash Can File Versioning", "Simple File Versioning", "Staggered File Versioning", "External File Versioning"},
			f.Versioning.Type).Vals("", "trashcan", "simple", "staggered", "external").K("versioning.type"),
		Text("Keep Versions", "Old versions to keep per file", f.Versioning.Params["keep"]).K("versioning.keep"),
		Text("Cleanup Interval", "Seconds between cleanup runs, zero disables",
			fmt.Sprint(f.Versioning.CleanupIntervalS)).Num().K("versioning.cleanupIntervalS"),
	)
	ignoresForm := NewForm(subviewButtons,
		// stored via /rest/db/ignores, not the config
		Area("Ignore Patterns", "One pattern per line", strings.Join(ignores, "\n")).K("ignores"),
	)
	advanced := NewForm(subviewButtons,
		Bool("Watch for Changes", "Use filesystem notifications", f.FSWatcherEnabled).K("fsWatcherEnabled"),
		Text("Full Rescan Interval", "Seconds between full rescans", fmt.Sprint(f.RescanIntervalS)).Num().K("rescanIntervalS"),
		Sel("Folder Type", "How this folder syncs",
			[]string{"Send & Receive", "Send Only", "Receive Only", "Receive Encrypted"},
			f.Type).Vals("sendreceive", "sendonly", "receiveonly", "receiveencrypted").K("type"),
		Sel("File Pull Order", "Order in which to download files",
			[]string{"Random", "Alphabetic", "Smallest First", "Largest First", "Oldest First", "Newest First"},
			f.Order).Vals("random", "alphabetic", "smallestFirst", "largestFirst", "oldestFirst", "newestFirst").K("order"),
		Bool("Ignore Permissions", "Disable syncing file permissions", f.IgnorePerms).K("ignorePerms"),
	)
	v := FormView{title: title, sub: sub,
		sections: []string{"General", "Sharing", "Versioning", "Ignore Patterns", "Advanced"},
		forms:    []Form{general, sharing, vers, ignoresForm, advanced}}
	if c == nil {
		return v
	}
	folderID := f.ID
	v.save = func(vals map[string]any) error {
		patch := map[string]any{}
		var deviceList []client.FolderDevice
		if myID != "" {
			deviceList = append(deviceList, client.FolderDevice{DeviceID: myID})
		}
		var ignoreLines []string
		for k, val := range vals {
			switch {
			case k == "ignores":
				ignoreLines = strings.Split(val.(string), "\n")
			case strings.HasPrefix(k, "device:"):
				if val.(bool) {
					deviceList = append(deviceList, client.FolderDevice{DeviceID: strings.TrimPrefix(k, "device:")})
				}
			case strings.HasPrefix(k, "versioning."):
				// assembled below
			default:
				patch[k] = val
			}
		}
		patch["devices"] = deviceList
		patch["versioning"] = map[string]any{
			"type":             vals["versioning.type"],
			"cleanupIntervalS": vals["versioning.cleanupIntervalS"],
			"params":           map[string]string{"keep": vals["versioning.keep"].(string)},
		}
		apply := func() error { return c.PatchFolder(folderID, patch) }
		if isNew {
			folderID = vals["id"].(string)
			patch["id"] = folderID
			apply = func() error { return c.PostFolder(patch) }
		}
		if err := apply(); err != nil {
			return err
		}
		return c.SetIgnores(folderID, ignoreLines)
	}
	v.remove = func() error { return c.DeleteFolder(folderID) }
	return v
}

// ── Edit Device ──────────────────────────────────────────────────────────────

// NewEditDevice is the fake-mode/mockup constructor.
func NewEditDevice(name string, isNew bool) FormView {
	d := client.DeviceCfg{
		DeviceID: "MFZWI3D-BONSGYC-...", Name: name,
		Addresses: []string{"dynamic"}, Compression: "metadata", AutoAcceptFolders: true,
	}
	return newEditDevice(nil, d, isNew)
}

func newEditDevice(c *client.Client, d client.DeviceCfg, isNew bool) FormView {
	title, sub := "Edit Device", d.Name
	if isNew {
		title, sub = "Add Device", "new"
	}
	idField := RO("Device ID", "Find it under Actions > Show ID on the other device", d.DeviceID)
	if isNew {
		idField = Text("Device ID", "Find it under Actions > Show ID on the other device", d.DeviceID).K("deviceID")
	}
	general := NewForm(subviewButtons,
		idField,
		Text("Device Name", "Shown instead of the ID; advertised to other devices", d.Name).K("name"),
	)
	sharing := NewForm(subviewButtons,
		Bool("Introducer", "Add devices from the introducer to our list", d.Introducer).K("introducer"),
		Bool("Auto Accept", "Auto create/share advertised folders", d.AutoAcceptFolders).K("autoAcceptFolders"),
	)
	advanced := NewForm(subviewButtons,
		Text("Addresses", `Comma separated or "dynamic"`, strings.Join(d.Addresses, ", ")).K("addresses"),
		Sel("Compression", "What data to compress",
			[]string{"All Data", "Metadata Only", "Off"},
			d.Compression).Vals("always", "metadata", "never").K("compression"),
		Text("Number of Connections", "Zero lets Syncthing decide", fmt.Sprint(d.NumConnections)).Num().K("numConnections"),
		Text("Incoming Rate Limit", "KiB/s, zero for no limit", fmt.Sprint(d.MaxRecvKbps)).Num().K("maxRecvKbps"),
		Text("Outgoing Rate Limit", "KiB/s, zero for no limit", fmt.Sprint(d.MaxSendKbps)).Num().K("maxSendKbps"),
		Bool("Untrusted", "Require password-protected folders", d.Untrusted).K("untrusted"),
	)
	v := FormView{title: title, sub: sub,
		sections: []string{"General", "Sharing", "Advanced"},
		forms:    []Form{general, sharing, advanced}}
	if c == nil {
		return v
	}
	deviceID := d.DeviceID
	v.save = func(vals map[string]any) error {
		patch := map[string]any{}
		for k, val := range vals {
			if k == "addresses" {
				var addrs []string
				for _, a := range strings.Split(val.(string), ",") {
					if a = strings.TrimSpace(a); a != "" {
						addrs = append(addrs, a)
					}
				}
				patch[k] = addrs
				continue
			}
			patch[k] = val
		}
		if isNew {
			id, _ := patch["deviceID"].(string)
			if id == "" {
				return fmt.Errorf("device ID is required")
			}
			return c.PostDevice(patch)
		}
		delete(patch, "deviceID")
		return c.PatchDevice(deviceID, patch)
	}
	v.remove = func() error { return c.DeleteDevice(deviceID) }
	return v
}

// ── Settings ─────────────────────────────────────────────────────────────────

func splitList(s string) []string {
	var out []string
	for _, e := range strings.Split(s, ",") {
		if e = strings.TrimSpace(e); e != "" {
			out = append(out, e)
		}
	}
	return out
}

// NewSettings is the fake-mode/mockup constructor.
func NewSettings() FormView {
	opts := client.OptionsCfg{
		ListenAddresses: []string{"default"}, GlobalAnnounceServers: []string{"default"},
		GlobalAnnounceEnabled: true, LocalAnnounceEnabled: true, RelaysEnabled: true,
		NATEnabled: true, URAccepted: -1, AutoUpgradeIntervalH: 12,
		MinHomeDiskFree: client.SizeCfg{Value: 1, Unit: "%"},
	}
	gui := client.GUICfg{Address: "127.0.0.1:8384", User: "evan", Password: "hunter2",
		UseTLS: true, Theme: "dark", APIKey: "abcDEF123..."}
	return newSettings(nil, opts, gui, "this-machine", "")
}

func newSettings(c *client.Client, opts client.OptionsCfg, gui client.GUICfg, selfName, myID string) FormView {
	settingsButtons := []string{"Save", "Close"}
	upgrades := "stable"
	switch {
	case opts.AutoUpgradeIntervalH == 0:
		upgrades = "none"
	case opts.UpgradeToPreReleases:
		upgrades = "candidate"
	}
	general := NewForm(settingsButtons,
		Text("Device Name", "Shown to other devices", selfName).K("self.name"),
		Text("Minimum Free Disk Space", `On the home (database) disk, e.g. "1 %" or "10 GB"`,
			fmt.Sprintf("%g %s", opts.MinHomeDiskFree.Value, opts.MinHomeDiskFree.Unit)).K("options.minHomeDiskFree"),
		RO("API Key", "Key for API access", gui.APIKey),
		Sel("Anonymous Usage Reporting", "Send anonymous usage statistics",
			[]string{"Version 3", "Version 2", "Undecided", "Disabled"},
			fmt.Sprint(opts.URAccepted)).Vals("3", "2", "0", "-1").K("options.urAccepted"),
		Sel("Automatic Upgrades", "Upgrade policy",
			[]string{"No Upgrades", "Stable Releases Only", "Stable Releases and Release Candidates"},
			upgrades).Vals("none", "stable", "candidate").K("upgrades"),
	)
	guiForm := NewForm(settingsButtons,
		Text("GUI Listen Address", "Non-privileged port 1024-65535", gui.Address).K("gui.address"),
		Text("GUI Authentication User", "Username for GUI access", gui.User).K("gui.user"),
		Pass("GUI Authentication Password", "Password for GUI access", gui.Password).K("gui.password"),
		Bool("Use HTTPS for GUI", "Enable HTTPS", gui.UseTLS).K("gui.useTLS"),
		Bool("Start Browser", "Open browser when Syncthing starts", opts.StartBrowser).K("options.startBrowser"),
		Sel("GUI Theme", "Web GUI theme",
			[]string{"Default", "Light", "Dark", "Black"},
			gui.Theme).Vals("default", "light", "dark", "black").K("gui.theme"),
	)
	connections := NewForm(settingsButtons,
		Text("Sync Protocol Listen Addresses", "Comma separated",
			strings.Join(opts.ListenAddresses, ", ")).K("options.listenAddresses"),
		Text("Incoming Rate Limit", "KiB/s, zero for no limit", fmt.Sprint(opts.MaxRecvKbps)).Num().K("options.maxRecvKbps"),
		Text("Outgoing Rate Limit", "KiB/s, zero for no limit", fmt.Sprint(opts.MaxSendKbps)).Num().K("options.maxSendKbps"),
		Bool("Limit Bandwidth in LAN", "Rate limit LAN connections too", opts.LimitBandwidthInLan).K("options.limitBandwidthInLan"),
		Bool("Enable NAT Traversal", "", opts.NATEnabled).K("options.natEnabled"),
		Bool("Local Discovery", "", opts.LocalAnnounceEnabled).K("options.localAnnounceEnabled"),
		Bool("Global Discovery", "", opts.GlobalAnnounceEnabled).K("options.globalAnnounceEnabled"),
		Bool("Enable Relaying", "Relay when direct connection fails", opts.RelaysEnabled).K("options.relaysEnabled"),
		Text("Global Discovery Servers", "Comma separated",
			strings.Join(opts.GlobalAnnounceServers, ", ")).K("options.globalAnnounceServers"),
	)
	v := FormView{title: "Settings", sub: selfName,
		sections: []string{"General", "GUI", "Connections"},
		forms:    []Form{general, guiForm, connections}}
	if c == nil {
		return v
	}
	oldPassword := gui.Password
	v.save = func(vals map[string]any) error {
		optsPatch, guiPatch := map[string]any{}, map[string]any{}
		for k, val := range vals {
			switch {
			case k == "self.name" || k == "upgrades":
				// handled below
			case k == "options.minHomeDiskFree":
				num, unit := 1.0, "%"
				fmt.Sscanf(val.(string), "%f %s", &num, &unit)
				optsPatch["minHomeDiskFree"] = map[string]any{"value": num, "unit": unit}
			case k == "options.urAccepted":
				n, _ := strconv.Atoi(val.(string))
				optsPatch["urAccepted"] = n
			case k == "options.listenAddresses" || k == "options.globalAnnounceServers":
				optsPatch[strings.TrimPrefix(k, "options.")] = splitList(val.(string))
			case strings.HasPrefix(k, "options."):
				optsPatch[strings.TrimPrefix(k, "options.")] = val
			case strings.HasPrefix(k, "gui."):
				guiPatch[strings.TrimPrefix(k, "gui.")] = val
			}
		}
		switch vals["upgrades"] {
		case "none":
			optsPatch["autoUpgradeIntervalH"] = 0
		case "stable":
			optsPatch["autoUpgradeIntervalH"], optsPatch["upgradeToPreReleases"] = 12, false
		case "candidate":
			optsPatch["autoUpgradeIntervalH"], optsPatch["upgradeToPreReleases"] = 12, true
		}
		if guiPatch["password"] == oldPassword {
			delete(guiPatch, "password") // unchanged bcrypt hash: don't re-hash it
		}
		if err := c.PatchOptions(optsPatch); err != nil {
			return err
		}
		if err := c.PatchGUI(guiPatch); err != nil {
			return err
		}
		return c.PatchDevice(myID, map[string]any{"name": vals["self.name"]})
	}
	return v
}

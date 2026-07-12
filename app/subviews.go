package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormView is the generic subview: infobar chrome, boxed sub-tabs, one Form
// per section. tab/⇧tab switch sections (when not editing a field); esc backs
// out (root App), or cancels the current field edit first.
type FormView struct {
	title, sub    string
	sections      []string
	forms         []Form
	active        int
	width, height int
}

func (v FormView) Init() tea.Cmd { return nil }

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
		// Save/Close/Remove all just close in the mockup (nothing persists).
		return v, func() tea.Msg { return closeMsg{flash: msg.label + ": " + v.title + " (mockup)"} }
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

// ── Concrete subviews ────────────────────────────────────────────────────────

var subviewButtons = []string{"Save", "Close", "Remove"}

func NewEditFolder(label string, isNew bool) FormView {
	title, sub, id := "Edit Folder", label, "ab7cd-ef8gh"
	if isNew {
		title, sub, id = "Add Folder", "new", "(generated)"
	}
	// Trailing comments name the syncthing config key each field maps to
	// (folder object in /rest/config/folders) — the upstream-parity anchor.
	general := NewForm(subviewButtons,
		Text("Folder Label", "Optional label, can differ per device", label), // label
		RO("Folder ID", "Same on all cluster devices", id),                   // id
		Text("Folder Path", "Created if it does not exist", "~/"+label),      // path
	)
	vers := NewForm(subviewButtons,
		Sel("File Versioning", "How to handle old versions",
			[]string{"No File Versioning", "Trash Can File Versioning", "Simple File Versioning", "Staggered File Versioning", "External File Versioning"},
			"Simple File Versioning"), // versioning.type
		Text("Keep Versions", "Old versions to keep per file", "5"),                     // versioning.params.keep
		Text("Cleanup Interval", "Seconds between cleanup runs, zero disables", "3600"), // versioning.cleanupIntervalS
	)
	shareFields := make([]FormField, len(Devices))
	for i, d := range Devices {
		shareFields[i] = Bool(d.Name, "Share this folder with "+d.Name, d.Name == "nas") // devices[]
	}
	sharing := NewForm(subviewButtons, shareFields...)
	ignores := NewForm(subviewButtons,
		// not in config: GET/POST /rest/db/ignores?folder=ID
		Area("Ignore Patterns", "One pattern per line", "// One pattern per line\n*.tmp\n(?d).DS_Store"),
	)
	advanced := NewForm(subviewButtons,
		Bool("Watch for Changes", "Use filesystem notifications", true),      // fsWatcherEnabled
		Text("Full Rescan Interval", "Seconds between full rescans", "3600"), // rescanIntervalS
		Sel("Folder Type", "How this folder syncs",
			[]string{"Send & Receive", "Send Only", "Receive Only", "Receive Encrypted"}, "Send & Receive"), // type
		Sel("File Pull Order", "Order in which to download files",
			[]string{"Random", "Alphabetic", "Smallest First", "Largest First", "Oldest First", "Newest First"}, "Random"), // order
		Bool("Ignore Permissions", "Disable syncing file permissions", false), // ignorePerms
	)
	return FormView{title: title, sub: sub,
		sections: []string{"General", "Sharing", "Versioning", "Ignore Patterns", "Advanced"},
		forms:    []Form{general, sharing, vers, ignores, advanced}}
}

func NewEditDevice(name string, isNew bool) FormView {
	title, sub := "Edit Device", name
	if isNew {
		title, sub = "Add Device", "new"
	}
	var idField FormField
	if isNew {
		idField = Text("Device ID", "Find it under Actions > Show ID on the other device", "")
	} else {
		idField = RO("Device ID", "Find it under Actions > Show ID on the other device", "MFZWI3D-BONSGYC-...")
	}
	// config keys: device object in /rest/config/devices
	general := NewForm(subviewButtons,
		idField, // deviceID
		Text("Device Name", "Shown instead of the ID; advertised to other devices", name), // name
	)
	sharing := NewForm(subviewButtons,
		Bool("Introducer", "Add devices from the introducer to our list", false), // introducer
		Bool("Auto Accept", "Auto create/share advertised folders", true),        // autoAcceptFolders
	)
	advanced := NewForm(subviewButtons,
		Text("Addresses", `Comma separated or "dynamic"`, "dynamic"), // addresses
		Sel("Compression", "What data to compress",
			[]string{"All Data", "Metadata Only", "Off"}, "Metadata Only"), // compression
		Text("Number of Connections", "Zero lets Syncthing decide", "0"), // numConnections
		Text("Incoming Rate Limit", "KiB/s, zero for no limit", "0"),     // maxRecvKbps
		Text("Outgoing Rate Limit", "KiB/s, zero for no limit", "0"),     // maxSendKbps
		Bool("Untrusted", "Require password-protected folders", false),   // untrusted
	)
	return FormView{title: title, sub: sub,
		sections: []string{"General", "Sharing", "Advanced"},
		forms:    []Form{general, sharing, advanced}}
}

func NewSettings() FormView {
	settingsButtons := []string{"Save", "Close"}
	// config keys: options/gui objects in /rest/config (except self device name)
	general := NewForm(settingsButtons,
		Text("Device Name", "Shown to other devices", "this-machine"),         // devices[self].name
		Text("Minimum Free Disk Space", "On the home (database) disk", "1 %"), // options.minHomeDiskFree
		RO("API Key", "Key for API access", "abcDEF123..."),                   // gui.apiKey
		Sel("Anonymous Usage Reporting", "Send anonymous usage statistics",
			[]string{"Version 3", "Version 2", "Undecided", "Disabled"}, "Disabled"), // options.urAccepted
		Sel("Automatic Upgrades", "Upgrade policy",
			[]string{"No Upgrades", "Stable Releases Only", "Stable Releases and Release Candidates"}, "Stable Releases Only"), // options.autoUpgradeIntervalH + upgradeToPreReleases
	)
	gui := NewForm(settingsButtons,
		Text("GUI Listen Address", "Non-privileged port 1024-65535", "127.0.0.1:8384"), // gui.address
		Text("GUI Authentication User", "Username for GUI access", "evan"),             // gui.user
		Pass("GUI Authentication Password", "Password for GUI access", "hunter2"),      // gui.password
		Bool("Use HTTPS for GUI", "Enable HTTPS", true),                                // gui.useTLS
		Bool("Start Browser", "Open browser when Syncthing starts", false),             // options.startBrowser
		Sel("GUI Theme", "Web GUI theme",
			[]string{"Default", "Light", "Dark", "Black"}, "Dark"), // gui.theme
	)
	connections := NewForm(settingsButtons,
		Text("Sync Protocol Listen Addresses", "Comma separated", "default"),    // options.listenAddresses
		Text("Incoming Rate Limit", "KiB/s, zero for no limit", "0"),            // options.maxRecvKbps
		Text("Outgoing Rate Limit", "KiB/s, zero for no limit", "0"),            // options.maxSendKbps
		Bool("Limit Bandwidth in LAN", "Rate limit LAN connections too", false), // options.limitBandwidthInLan
		Bool("Enable NAT Traversal", "", true),                                  // options.natEnabled
		Bool("Local Discovery", "", true),                                       // options.localAnnounceEnabled
		Bool("Global Discovery", "", true),                                      // options.globalAnnounceEnabled
		Bool("Enable Relaying", "Relay when direct connection fails", true),     // options.relaysEnabled
		Text("Global Discovery Servers", "Comma separated", "default"),          // options.globalAnnounceServers
	)
	return FormView{title: "Settings", sub: "this-machine",
		sections: []string{"General", "GUI", "Connections"},
		forms:    []Form{general, gui, connections}}
}

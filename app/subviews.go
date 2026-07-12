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
	general := NewForm(subviewButtons,
		Text("Folder Label", "Optional label, can differ per device", label),
		RO("Folder ID", "Same on all cluster devices", id),
		Text("Folder Path", "Created if it does not exist", "~/"+label),
	)
	vers := NewForm(subviewButtons,
		Sel("File Versioning", "How to handle old versions",
			[]string{"No File Versioning", "Trash Can File Versioning", "Simple File Versioning", "Staggered File Versioning", "External File Versioning"},
			"Simple File Versioning"),
		Text("Keep Versions", "Old versions to keep per file", "5"),
		Text("Cleanup Interval", "Seconds between cleanup runs, zero disables", "3600"),
	)
	shareFields := make([]FormField, len(Devices))
	for i, d := range Devices {
		shareFields[i] = Bool(d.Name, "Share this folder with "+d.Name, d.Name == "nas")
	}
	sharing := NewForm(subviewButtons, shareFields...)
	ignores := NewForm(subviewButtons,
		Area("Ignore Patterns", "One pattern per line", "// One pattern per line\n*.tmp\n(?d).DS_Store"),
	)
	advanced := NewForm(subviewButtons,
		Bool("Watch for Changes", "Use filesystem notifications", true),
		Text("Full Rescan Interval", "Seconds between full rescans", "3600"),
		Sel("Folder Type", "How this folder syncs",
			[]string{"Send & Receive", "Send Only", "Receive Only", "Receive Encrypted"}, "Send & Receive"),
		Sel("File Pull Order", "Order in which to download files",
			[]string{"Random", "Alphabetic", "Smallest First", "Largest First", "Oldest First", "Newest First"}, "Random"),
		Bool("Ignore Permissions", "Disable syncing file permissions", false),
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
	general := NewForm(subviewButtons,
		idField,
		Text("Device Name", "Shown instead of the ID; advertised to other devices", name),
	)
	sharing := NewForm(subviewButtons,
		Bool("Introducer", "Add devices from the introducer to our list", false),
		Bool("Auto Accept", "Auto create/share advertised folders", true),
	)
	advanced := NewForm(subviewButtons,
		Text("Addresses", `Comma separated or "dynamic"`, "dynamic"),
		Sel("Compression", "What data to compress",
			[]string{"All Data", "Metadata Only", "Off"}, "Metadata Only"),
		Text("Number of Connections", "Zero lets Syncthing decide", "0"),
		Text("Incoming Rate Limit", "KiB/s, zero for no limit", "0"),
		Text("Outgoing Rate Limit", "KiB/s, zero for no limit", "0"),
		Bool("Untrusted", "Require password-protected folders", false),
	)
	return FormView{title: title, sub: sub,
		sections: []string{"General", "Sharing", "Advanced"},
		forms:    []Form{general, sharing, advanced}}
}

func NewSettings() FormView {
	settingsButtons := []string{"Save", "Close"}
	general := NewForm(settingsButtons,
		Text("Device Name", "Shown to other devices", "this-machine"),
		Text("Minimum Free Disk Space", "On the home (database) disk", "1 %"),
		RO("API Key", "Key for API access", "abcDEF123..."),
		Sel("Anonymous Usage Reporting", "Send anonymous usage statistics",
			[]string{"Version 3", "Version 2", "Undecided", "Disabled"}, "Disabled"),
		Sel("Automatic Upgrades", "Upgrade policy",
			[]string{"No Upgrades", "Stable Releases Only", "Stable Releases and Release Candidates"}, "Stable Releases Only"),
	)
	gui := NewForm(settingsButtons,
		Text("GUI Listen Address", "Non-privileged port 1024-65535", "127.0.0.1:8384"),
		Text("GUI Authentication User", "Username for GUI access", "evan"),
		Pass("GUI Authentication Password", "Password for GUI access", "hunter2"),
		Bool("Use HTTPS for GUI", "Enable HTTPS", true),
		Bool("Start Browser", "Open browser when Syncthing starts", false),
		Sel("GUI Theme", "Web GUI theme",
			[]string{"Default", "Light", "Dark", "Black"}, "Dark"),
	)
	connections := NewForm(settingsButtons,
		Text("Sync Protocol Listen Addresses", "Comma separated", "default"),
		Text("Incoming Rate Limit", "KiB/s, zero for no limit", "0"),
		Text("Outgoing Rate Limit", "KiB/s, zero for no limit", "0"),
		Bool("Limit Bandwidth in LAN", "Rate limit LAN connections too", false),
		Bool("Enable NAT Traversal", "", true),
		Bool("Local Discovery", "", true),
		Bool("Global Discovery", "", true),
		Bool("Enable Relaying", "Relay when direct connection fails", true),
		Text("Global Discovery Servers", "Comma separated", "default"),
	)
	return FormView{title: "Settings", sub: "this-machine",
		sections: []string{"General", "GUI", "Connections"},
		forms:    []Form{general, gui, connections}}
}

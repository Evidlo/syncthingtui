package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/evidlo/syncthingtui/client"
)

// openMsg asks the root App to push a subview.
type openMsg struct{ view tea.Model }

func open(v tea.Model) tea.Cmd { return func() tea.Msg { return openMsg{v} } }

const (
	tabFolders = iota
	tabDevices
	tabAlerts
	tabThisDevice
	tabActions
)

// MainModel is the tabbed main view.
type MainModel struct {
	active        int
	width, height int
	folderIdx     int // 0..len(folders)+1; last two are Add Folder / Rescan All
	folderBtn     int
	deviceIdx     int // 0..len(devices)+1; last two are Add Device / Recent Changes
	deviceBtn     int
	alertIdx      int
	alertBtn      int
	actionIdx     int
	folders       []Folder
	devices       []Device
	alerts        []Alert
	stats         [][2]string
	stVersion     string
	myID          string
	paths         [][2]string
	actionConfirm string // pending Restart/Shut Down confirmation
	rates         string
	flash         string
	client        *client.Client // nil = fake mode
}

func NewMain(c *client.Client) MainModel {
	m := MainModel{client: c, rates: netRates, stats: ThisDeviceStats}
	if c == nil {
		m.folders = append([]Folder{}, Folders...)
		m.devices = append([]Device{}, Devices...)
		m.alerts = append([]Alert{}, Alerts...)
	} else {
		m.rates = ""
		m.stats = [][2]string{{"Status", "connecting to syncthing..."}}
	}
	if c == nil {
		m.stVersion = "v2.0.13, Linux (64-bit)"
		m.paths = AboutPaths
		m.myID = DeviceID
	}
	return m
}

// setData applies a live snapshot, keeping cursors in range.
func (m *MainModel) setData(d dataMsg) {
	m.folders, m.devices, m.alerts, m.stats = d.folders, d.devices, d.alerts, d.stats
	m.stVersion, m.paths, m.myID = d.stVersion, d.paths, d.myID
	m.folderIdx = clamp(m.folderIdx, 0, len(m.folders)+1)
	m.deviceIdx = clamp(m.deviceIdx, 0, len(m.devices)+1)
	m.alertIdx = clamp(m.alertIdx, 0, max(0, len(m.alerts)-1))
}

// openFolderEditor opens the folder editor, loading fresh config + ignores
// from syncthing in live mode.
func (m *MainModel) openFolderEditor(id, label string, isNew bool) tea.Cmd {
	if m.client == nil {
		return open(NewEditFolder(label, isNew))
	}
	c := m.client
	return func() tea.Msg {
		cfg, err := c.Config()
		if err != nil {
			return dataMsg{err: err}
		}
		status, err := c.SystemStatus()
		if err != nil {
			return dataMsg{err: err}
		}
		f := client.FolderCfg{Path: "~/", Type: "sendreceive", Order: "random",
			RescanIntervalS: 3600, FSWatcherEnabled: true}
		var ignores []string
		if !isNew {
			for _, fc := range cfg.Folders {
				if fc.ID == id {
					f = fc
				}
			}
			ignores, _ = c.Ignores(id)
		}
		return openMsg{view: newEditFolder(c, f, cfg.Devices, status.MyID, ignores, isNew)}
	}
}

// openDeviceEditor opens the device editor; prefill is used for new devices
// (e.g. accepting a pending device with known ID/name).
func (m *MainModel) openDeviceEditor(id string, prefill client.DeviceCfg, isNew bool) tea.Cmd {
	if m.client == nil {
		return open(NewEditDevice(prefill.Name, isNew))
	}
	c := m.client
	return func() tea.Msg {
		d := prefill
		if d.Addresses == nil {
			d.Addresses = []string{"dynamic"}
		}
		if d.Compression == "" {
			d.Compression = "metadata"
		}
		if !isNew {
			var err error
			if d, err = c.DeviceByID(id); err != nil {
				return dataMsg{err: err}
			}
		}
		return openMsg{view: newEditDevice(c, d, isNew)}
	}
}

// action runs fn against syncthing and refreshes; in fake mode it runs
// fallback locally instead.
func (m *MainModel) action(fn func(c *client.Client) error, fallback func()) tea.Cmd {
	if m.client == nil {
		fallback()
		return nil
	}
	c := m.client
	return func() tea.Msg {
		if err := fn(c); err != nil {
			return dataMsg{err: err}
		}
		return fetch(c)
	}
}

func (m MainModel) Init() tea.Cmd { return nil }

func (m MainModel) tabNames() []string {
	alerts := "Alerts"
	if n := len(m.alerts); n > 0 {
		alerts += dimStyle.Render(fmt.Sprintf(" (%d)", n))
	}
	return []string{"Folders", "Remote Devices", alerts, "This Device", "Actions"}
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		m.flash = ""
		key := msg.String()
		switch key {
		case "q", "esc", "backspace", "ctrl+c":
			return m, tea.Quit
		case "1", "2", "3", "4", "5":
			m.active = int(key[0] - '1')
			return m, nil
		case "tab":
			m.active = (m.active + 1) % 5
			return m, nil
		case "shift+tab":
			m.active = (m.active + 4) % 5
			return m, nil
		}
		switch m.active {
		case tabFolders:
			return m.updateFolders(key)
		case tabDevices:
			return m.updateDevices(key)
		case tabAlerts:
			return m.updateAlerts(key)
		case tabActions:
			return m.updateActions(key)
		}
	}
	return m, nil
}

func clamp(v, lo, hi int) int { return min(hi, max(lo, v)) }

func (m *MainModel) folderButtons() []string {
	pause := "Pause"
	if m.folderIdx < len(m.folders) && m.folders[m.folderIdx].State == "Paused" {
		pause = "Resume"
	}
	return []string{"Edit", pause, "Rescan"}
}

func (m MainModel) updateFolders(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.folderIdx, m.folderBtn = clamp(m.folderIdx-1, 0, len(m.folders)+1), 0
	case "down", "j":
		m.folderIdx, m.folderBtn = clamp(m.folderIdx+1, 0, len(m.folders)+1), 0
	case "left", "h":
		m.folderBtn = clamp(m.folderBtn-1, 0, len(m.folderButtons())-1)
	case "right", "l":
		m.folderBtn = clamp(m.folderBtn+1, 0, len(m.folderButtons())-1)
	case "enter":
		switch m.folderIdx {
		case len(m.folders):
			return m, m.openFolderEditor("", "", true)
		case len(m.folders) + 1:
			m.flash = "rescanning all folders"
			return m, m.action(
				func(c *client.Client) error { return c.Rescan("") },
				func() {})
		default:
			f := &m.folders[m.folderIdx]
			id := f.ID
			switch m.folderButtons()[m.folderBtn] {
			case "Pause":
				return m, m.action(
					func(c *client.Client) error { return c.SetFolderPaused(id, true) },
					func() { f.State = "Paused" })
			case "Resume":
				return m, m.action(
					func(c *client.Client) error { return c.SetFolderPaused(id, false) },
					func() { f.State = "Up to Date" })
			case "Rescan":
				m.flash = "rescanning " + f.Label
				return m, m.action(
					func(c *client.Client) error { return c.Rescan(id) },
					func() {})
			case "Edit":
				return m, m.openFolderEditor(f.ID, f.Label, false)
			}
		}
	}
	return m, nil
}

func (m *MainModel) deviceButtons() []string {
	pause := "Pause"
	// device paused states carry a "(Unused)" suffix, so match by prefix
	if m.deviceIdx < len(m.devices) && strings.HasPrefix(m.devices[m.deviceIdx].State, "Paused") {
		pause = "Resume"
	}
	return []string{"Edit", pause}
}

func (m MainModel) updateDevices(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.deviceIdx, m.deviceBtn = clamp(m.deviceIdx-1, 0, len(m.devices)+1), 0
	case "down", "j":
		m.deviceIdx, m.deviceBtn = clamp(m.deviceIdx+1, 0, len(m.devices)+1), 0
	case "left", "h":
		m.deviceBtn = clamp(m.deviceBtn-1, 0, len(m.deviceButtons())-1)
	case "right", "l":
		m.deviceBtn = clamp(m.deviceBtn+1, 0, len(m.deviceButtons())-1)
	case "enter":
		switch m.deviceIdx {
		case len(m.devices):
			return m, m.openDeviceEditor("", client.DeviceCfg{}, true)
		case len(m.devices) + 1:
			m.flash = "recent changes: not in mockup scope"
		default:
			d := &m.devices[m.deviceIdx]
			id := d.ID
			switch m.deviceButtons()[m.deviceBtn] {
			case "Pause":
				return m, m.action(
					func(c *client.Client) error { return c.SetDevicePaused(id, true) },
					func() { d.State = "Paused" })
			case "Resume":
				return m, m.action(
					func(c *client.Client) error { return c.SetDevicePaused(id, false) },
					func() { d.State = "Up to Date" })
			case "Edit":
				return m, m.openDeviceEditor(id, client.DeviceCfg{}, false)
			}
		}
	}
	return m, nil
}

func (m MainModel) updateAlerts(key string) (tea.Model, tea.Cmd) {
	if len(m.alerts) == 0 {
		return m, nil
	}
	sel := m.alerts[m.alertIdx]
	switch key {
	case "up", "k":
		m.alertIdx, m.alertBtn = clamp(m.alertIdx-1, 0, len(m.alerts)-1), 0
	case "down", "j":
		m.alertIdx, m.alertBtn = clamp(m.alertIdx+1, 0, len(m.alerts)-1), 0
	case "left", "h":
		m.alertBtn = clamp(m.alertBtn-1, 0, len(sel.Buttons)-1)
	case "right", "l":
		m.alertBtn = clamp(m.alertBtn+1, 0, len(sel.Buttons)-1)
	case "enter":
		button := sel.Buttons[m.alertBtn]
		m.flash = fmt.Sprintf("%s: %s", sel.Title, button)
		if m.client == nil { // fake mode: remove locally; live mode refetches truth
			m.alerts = append(m.alerts[:m.alertIdx], m.alerts[m.alertIdx+1:]...)
			m.alertIdx, m.alertBtn = clamp(m.alertIdx, 0, max(0, len(m.alerts)-1)), 0
		}
		switch button {
		case "OK": // clears the whole system error list, like the web GUI
			return m, m.action(
				func(c *client.Client) error { return c.ClearErrors() },
				func() {})
		case "Dismiss":
			return m, m.action(
				func(c *client.Client) error {
					if sel.Kind == "device" {
						return c.DismissPendingDevice(sel.ID)
					}
					return c.DismissPendingFolder(sel.ID)
				},
				func() {})
		case "Ignore":
			return m, m.action(
				func(c *client.Client) error {
					if sel.Kind == "device" {
						return c.IgnorePendingDevice(sel.ID, sel.Short, sel.Addr)
					}
					return c.IgnorePendingFolder(sel.DeviceID, sel.ID, sel.Short)
				},
				func() {})
		case "Share": // accept offer for a folder we already have
			return m, m.action(
				func(c *client.Client) error { return c.ShareFolderWithDevice(sel.ID, sel.DeviceID) },
				func() {})
		case "Add": // offered folder is new here: open Add Folder prefilled
			return m, m.openFolderOffer(sel.ID, sel.Short, sel.DeviceID)
		case "Add Device":
			return m, m.openDeviceEditor("", client.DeviceCfg{DeviceID: sel.ID, Name: sel.Short}, true)
		}
	}
	return m, nil
}

// openFolderOffer opens Add Folder prefilled from a pending offer (folder ID,
// label, and the offering device pre-shared).
func (m *MainModel) openFolderOffer(id, label, deviceID string) tea.Cmd {
	if m.client == nil {
		return open(NewEditFolder(label, true))
	}
	c := m.client
	return func() tea.Msg {
		cfg, err := c.Config()
		if err != nil {
			return dataMsg{err: err}
		}
		status, err := c.SystemStatus()
		if err != nil {
			return dataMsg{err: err}
		}
		f := client.FolderCfg{ID: id, Label: label, Path: "~/" + label,
			Type: "sendreceive", Order: "random", RescanIntervalS: 3600, FSWatcherEnabled: true,
			Devices: []client.FolderDevice{{DeviceID: deviceID}}}
		return openMsg{view: newEditFolder(c, f, cfg.Devices, status.MyID, nil, true)}
	}
}

func (m MainModel) updateActions(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.actionIdx, m.actionConfirm = clamp(m.actionIdx-1, 0, len(ActionItems)-1), ""
	case "down", "j":
		m.actionIdx, m.actionConfirm = clamp(m.actionIdx+1, 0, len(ActionItems)-1), ""
	case "enter":
		item := ActionItems[m.actionIdx]
		switch item {
		case "Settings":
			if m.client == nil {
				return m, open(NewSettings())
			}
			c := m.client
			return m, func() tea.Msg {
				opts, err := c.Options()
				if err != nil {
					return dataMsg{err: err}
				}
				gui, err := c.GUIConfig()
				if err != nil {
					return dataMsg{err: err}
				}
				status, err := c.SystemStatus()
				if err != nil {
					return dataMsg{err: err}
				}
				self, _ := c.DeviceByID(status.MyID)
				return openMsg{view: newSettings(c, opts, gui, self.Name, status.MyID)}
			}
		case "Show ID":
			return m, open(NewShowID(m.myID))
		case "About":
			return m, open(NewAbout(m.stVersion, m.paths))
		case "Restart", "Shut Down":
			if m.actionConfirm != item { // destructive: require a second enter
				m.actionConfirm = item
				m.flash = "press enter again to " + strings.ToLower(item) + " syncthing"
				return m, nil
			}
			m.actionConfirm = ""
			m.flash = item + " sent"
			return m, m.action(
				func(c *client.Client) error {
					if item == "Restart" {
						return c.Restart()
					}
					return c.Shutdown()
				},
				func() {})
		default:
			m.flash = item + ": not implemented"
		}
	}
	return m, nil
}

func (m MainModel) View() string {
	var body, keys string
	switch m.active {
	case tabFolders:
		body = m.viewFolders()
		keys = keyHint("↑↓", "select", "←→", "button", "enter", "confirm", "tab/⇧tab", "switch tabs", "esc", "quit")
	case tabThisDevice:
		body = m.viewThisDevice()
		keys = keyHint("tab/⇧tab", "switch tabs", "esc", "quit")
	case tabDevices:
		body = m.viewDevices()
		keys = keyHint("↑↓", "select", "←→", "button", "enter", "confirm", "tab/⇧tab", "switch tabs", "esc", "quit")
	case tabAlerts:
		body = m.viewAlerts()
		keys = keyHint("↑↓", "select", "←→", "button", "enter", "confirm", "tab/⇧tab", "switch tabs", "esc", "quit")
	case tabActions:
		body = m.viewActions()
		keys = keyHint("↑↓", "select", "enter", "open", "tab/⇧tab", "switch tabs", "esc", "quit")
	}
	right := m.rates
	if m.flash != "" {
		right = warnSt.Render(m.flash)
	}
	return page(boxedTabs(m.tabNames(), m.active), body, keys, right, m.width, m.height)
}

const leftW = 26

// splitPane renders the shared left-list / right-detail layout.
func (m MainModel) splitPane(list, detail string) string {
	innerH := m.height - 5
	left := lipgloss.NewStyle().Width(leftW).Height(max(1, innerH)).
		Border(lipgloss.NormalBorder(), false, true, false, false).BorderForeground(dim).
		Render(list)
	right := lipgloss.NewStyle().Padding(0, 1).Width(m.width - leftW - 1).Render(detail)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func listRow(selected bool, icon, label string) string {
	cursor, style := "  ", lipgloss.NewStyle()
	if selected {
		cursor, style = keyStyle.Render("▸ "), boldSt
	}
	return cursor + icon + " " + style.Render(label) + "\n"
}

func (m MainModel) viewFolders() string {
	var list strings.Builder
	for i, f := range m.folders {
		list.WriteString(listRow(i == m.folderIdx, stateStyle(f.State).Render("●"), f.Label))
	}
	list.WriteString("\n")
	list.WriteString(listRow(m.folderIdx == len(m.folders), keyStyle.Render("+"), "Add Folder"))
	list.WriteString(listRow(m.folderIdx == len(m.folders)+1, keyStyle.Render("⟳"), "Rescan All"))
	curLine := m.folderIdx
	if m.folderIdx >= len(m.folders) { // past the blank separator row
		curLine++
	}

	var detail string
	switch m.folderIdx {
	case len(m.folders):
		detail = dimStyle.Render("Create a new synced folder.")
	case len(m.folders) + 1:
		detail = dimStyle.Render("Rescan all folders now.")
	default:
		f := m.folders[m.folderIdx]
		detail = boldSt.Render(f.Label) + dimStyle.Render(" ("+f.ID+")") + "\n\n" +
			fmt.Sprintf("Folder Path      %s\n", f.Path) +
			fmt.Sprintf("State            %s\n", stateStyle(f.State).Render(f.State)) +
			fmt.Sprintf("Global State     %s\n", f.Global) +
			fmt.Sprintf("Local State      %s\n", f.Local) +
			fmt.Sprintf("Shared With      %s\n", f.Shared) + "\n" +
			progressBar(f.Pct, m.width-leftW-8) + fmt.Sprintf(" %d%%\n\n", f.Pct) +
			buttonRow(m.folderButtons(), m.folderBtn)
	}
	return m.splitPane(scrollToCursor(list.String(), curLine, m.height-5), detail)
}

func (m MainModel) viewDevices() string {
	var list strings.Builder
	for i, d := range m.devices {
		list.WriteString(listRow(i == m.deviceIdx, stateStyle(d.State).Render("●"), d.Name))
	}
	list.WriteString("\n")
	list.WriteString(listRow(m.deviceIdx == len(m.devices), keyStyle.Render("+"), "Add Device"))
	list.WriteString(listRow(m.deviceIdx == len(m.devices)+1, keyStyle.Render("≡"), "Recent Changes"))
	curLine := m.deviceIdx
	if m.deviceIdx >= len(m.devices) { // past the blank separator row
		curLine++
	}

	var detail string
	switch m.deviceIdx {
	case len(m.devices):
		detail = dimStyle.Render("Add a new remote device.")
	case len(m.devices) + 1:
		detail = dimStyle.Render("Show recently changed files.")
	default:
		d := m.devices[m.deviceIdx]
		detail = boldSt.Render(d.Name) + "\n\n" +
			fmt.Sprintf("State            %s\n", stateStyle(d.State).Render(d.State)) +
			fmt.Sprintf("Address          %s\n", d.Address) +
			fmt.Sprintf("Compression      %s\n", d.Compression) +
			fmt.Sprintf("Last Seen        %s\n", d.LastSeen) + "\n" +
			progressBar(d.Pct, m.width-leftW-8) + fmt.Sprintf(" %d%%\n\n", d.Pct) +
			buttonRow(m.deviceButtons(), m.deviceBtn)
	}
	return m.splitPane(scrollToCursor(list.String(), curLine, m.height-5), detail)
}

func (m MainModel) viewThisDevice() string {
	var b strings.Builder
	b.WriteString("\n")
	for _, kv := range m.stats {
		b.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render(fmt.Sprintf("%-21s", kv[0])), kv[1]))
	}
	return b.String()
}

func (m MainModel) viewAlerts() string {
	if len(m.alerts) == 0 {
		return "\n  " + dimStyle.Render("No pending alerts.")
	}
	var list strings.Builder
	kinds := map[string]string{"device": "Device", "folder": "Folder", "notice": "Notice"}
	for i, a := range m.alerts {
		label := kinds[a.Kind]
		if a.Short != "" {
			label += " (" + a.Short + ")"
		}
		list.WriteString(listRow(i == m.alertIdx, warnSt.Render("⚠"), label))
	}
	sel := m.alerts[m.alertIdx]
	detail := warnSt.Bold(true).Render("⚠ "+sel.Title) + "\n" +
		dimStyle.Render(sel.Time) + "\n\n" +
		lipgloss.NewStyle().Width(m.width-leftW-4).Render(sel.Body) + "\n\n" +
		buttonRow(sel.Buttons, m.alertBtn)
	return m.splitPane(scrollToCursor(list.String(), m.alertIdx, m.height-5), detail)
}

func (m MainModel) viewActions() string {
	var b strings.Builder
	b.WriteString("\n")
	for i, a := range ActionItems {
		b.WriteString("  " + listRow(i == m.actionIdx, "", a))
	}
	return b.String()
}

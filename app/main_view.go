package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// openMsg asks the root App to push a subview.
type openMsg struct{ view tea.Model }

func open(v tea.Model) tea.Cmd { return func() tea.Msg { return openMsg{v} } }

const (
	tabFolders = iota
	tabThisDevice
	tabDevices
	tabAlerts
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
	flash         string
}

func NewMain() MainModel {
	return MainModel{
		folders: append([]Folder{}, Folders...),
		devices: append([]Device{}, Devices...),
		alerts:  append([]Alert{}, Alerts...),
	}
}

func (m MainModel) Init() tea.Cmd { return nil }

func (m MainModel) tabNames() []string {
	alerts := "Alerts"
	if n := len(m.alerts); n > 0 {
		alerts += dimStyle.Render(fmt.Sprintf(" (%d)", n))
	}
	return []string{"Folders", "This Device", "Remote Devices", alerts, "Actions"}
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
			return m, open(NewEditFolder("", true))
		case len(m.folders) + 1:
			m.flash = "rescanned all folders (mockup)"
		default:
			f := &m.folders[m.folderIdx]
			switch m.folderButtons()[m.folderBtn] {
			case "Pause":
				f.State = "Paused"
			case "Resume":
				f.State = "Up to Date"
			case "Rescan":
				m.flash = "rescanned " + f.Label + " (mockup)"
			case "Edit":
				return m, open(NewEditFolder(f.Label, false))
			}
		}
	}
	return m, nil
}

func (m *MainModel) deviceButtons() []string {
	pause := "Pause"
	if m.deviceIdx < len(m.devices) && m.devices[m.deviceIdx].State == "Paused" {
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
			return m, open(NewEditDevice("", true))
		case len(m.devices) + 1:
			m.flash = "recent changes: not in mockup scope"
		default:
			d := &m.devices[m.deviceIdx]
			switch m.deviceButtons()[m.deviceBtn] {
			case "Pause":
				d.State = "Paused"
			case "Resume":
				d.State = "Up to Date"
			case "Edit":
				return m, open(NewEditDevice(d.Name, false))
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
		m.flash = fmt.Sprintf("%s: %s (mockup)", sel.Short, sel.Buttons[m.alertBtn])
		m.alerts = append(m.alerts[:m.alertIdx], m.alerts[m.alertIdx+1:]...)
		m.alertIdx, m.alertBtn = clamp(m.alertIdx, 0, max(0, len(m.alerts)-1)), 0
		if sel.Kind == "device" && strings.HasPrefix(m.flash, sel.Short+": Add") {
			return m, open(NewEditDevice(sel.Short, true))
		}
	}
	return m, nil
}

func (m MainModel) updateActions(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.actionIdx = clamp(m.actionIdx-1, 0, len(ActionItems)-1)
	case "down", "j":
		m.actionIdx = clamp(m.actionIdx+1, 0, len(ActionItems)-1)
	case "enter":
		switch ActionItems[m.actionIdx] {
		case "Settings":
			return m, open(NewSettings())
		case "Show ID":
			return m, open(NewShowID())
		case "Advanced": // stand-in: About/Paths lives here in the mockup
			return m, open(NewAbout())
		default:
			m.flash = ActionItems[m.actionIdx] + ": not in mockup scope"
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
	right := netRates
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
	return m.splitPane(list.String(), detail)
}

func (m MainModel) viewDevices() string {
	var list strings.Builder
	for i, d := range m.devices {
		list.WriteString(listRow(i == m.deviceIdx, stateStyle(d.State).Render("●"), d.Name))
	}
	list.WriteString("\n")
	list.WriteString(listRow(m.deviceIdx == len(m.devices), keyStyle.Render("+"), "Add Device"))
	list.WriteString(listRow(m.deviceIdx == len(m.devices)+1, keyStyle.Render("≡"), "Recent Changes"))

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
			fmt.Sprintf("Download Rate    %s\n", d.Download) +
			fmt.Sprintf("Upload Rate      %s\n", d.Upload) +
			"Compression      Metadata Only\n" +
			"Last Seen        2026-07-12 10:41\n" +
			"Folders          Documents, Photos, Music\n\n" +
			progressBar(d.Pct, m.width-leftW-8) + fmt.Sprintf(" %d%%\n\n", d.Pct) +
			buttonRow(m.deviceButtons(), m.deviceBtn)
	}
	return m.splitPane(list.String(), detail)
}

func (m MainModel) viewThisDevice() string {
	var b strings.Builder
	b.WriteString("\n")
	for _, kv := range ThisDeviceStats {
		b.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render(fmt.Sprintf("%-21s", kv[0])), kv[1]))
	}
	return b.String()
}

func (m MainModel) viewAlerts() string {
	if len(m.alerts) == 0 {
		return "\n  " + dimStyle.Render("No pending alerts.")
	}
	var list strings.Builder
	kinds := map[string]string{"device": "Device", "folder": "Folder"}
	for i, a := range m.alerts {
		list.WriteString(listRow(i == m.alertIdx, warnSt.Render("⚠"), kinds[a.Kind]+" ("+a.Short+")"))
	}
	sel := m.alerts[m.alertIdx]
	detail := warnSt.Bold(true).Render("⚠ "+sel.Title) + "\n" +
		dimStyle.Render(sel.Time) + "\n\n" +
		lipgloss.NewStyle().Width(m.width-leftW-4).Render(sel.Body) + "\n\n" +
		buttonRow(sel.Buttons, m.alertBtn)
	return m.splitPane(list.String(), detail)
}

func (m MainModel) viewActions() string {
	var b strings.Builder
	b.WriteString("\n")
	for i, a := range ActionItems {
		b.WriteString("  " + listRow(i == m.actionIdx, "", a))
	}
	return b.String()
}

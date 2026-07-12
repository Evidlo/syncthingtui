package mockups

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Renderer func(width, height int) string

var Screens = map[string]Renderer{
	"folders-v1-panels":     foldersV1Panels,
	"folders-v2-table":      foldersV2Table,
	"folders-v3-split":      foldersV3Split,
	"alerts-v1-list":        alertsV1List,
	"alerts-v2-focus":       alertsV2Focus,
	"alerts-v3-split":       alertsV3Split,
	"devices-split":         devicesSplit,
	"thisdevice-v1-kv":      thisDeviceV1KV,
	"thisdevice-v2-grid":    thisDeviceV2Grid,
	"actions-v1-list":       actionsV1List,
	"actions-v2-centered":   actionsV2Centered,
	"editfolder-v1-infobar": editFolderV1Infobar,
	"editfolder-v2-modal":   editFolderV2Modal,
	"editdevice-infobar":    editDeviceInfobar,
	"editdevice-sharing":    editDeviceSharing,
	"editdevice-advanced":   editDeviceAdvanced,
	"settings-infobar":      settingsInfobar,
	"settings-gui":          settingsGUI,
	"settings-connections":  settingsConnections,
	"editfolder-versioning": editFolderVersioning,
	"editfolder-ignores":    editFolderIgnores,
	"editfolder-advanced":   editFolderAdvanced,
	"showid":                showID,
	"about-paths":           aboutPaths,
}

func Names() []string {
	names := make([]string, 0, len(Screens))
	for n := range Screens {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

var (
	accent   = lipgloss.Color("4")
	dim      = lipgloss.Color("8")
	good     = lipgloss.Color("2")
	warn     = lipgloss.Color("3")
	bad      = lipgloss.Color("1")
	dimStyle = lipgloss.NewStyle().Foreground(dim)
	keyStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
)

func stateStyle(state string) lipgloss.Style {
	switch state {
	case "Up to Date":
		return lipgloss.NewStyle().Foreground(good)
	case "Syncing":
		return lipgloss.NewStyle().Foreground(accent)
	case "Out of Sync":
		return lipgloss.NewStyle().Foreground(bad)
	}
	return lipgloss.NewStyle().Foreground(dim)
}

// tabBar renders the main tab row (boxed active tab, superscript number
// hotkeys — style chosen from the old project) with the named tab active.
// The alerts count is dimmed, e.g. "Alerts (2)".
func tabBar(active string, width int) string {
	names := []string{"Folders", "This Device", "Remote Devices", "Alerts", "Actions"}
	sups := []string{"¹", "²", "³", "⁴", "⁵"}
	inactive := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.HiddenBorder())
	activeSt := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.NormalBorder()).BorderForeground(accent)
	tabs := make([]string, len(names))
	for i, n := range names {
		label := n
		if n == "Alerts" {
			label += dimStyle.Render(" (2)")
		}
		style := inactive
		if n == active {
			style = activeSt
		}
		tabs[i] = style.Render(label + dimStyle.Render(sups[i]))
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
}

func statusBar(keys, right string, width int) string {
	gap := width - lipgloss.Width(keys) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return dimStyle.Render(strings.Repeat("─", width)) + "\n" +
		keys + strings.Repeat(" ", gap) + right
}

func keyHint(pairs ...string) string {
	parts := make([]string, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, keyStyle.Render(pairs[i])+dimStyle.Render(":"+pairs[i+1]))
	}
	return strings.Join(parts, "  ")
}

// frame assembles tab bar + body + status bar, padding the body to fill height.
func frame(active, body, keys string, width, height int) string {
	top := tabBar(active, width)
	bottom := statusBar(keys, "↓1.2MiB/s ↑2.4MiB/s", width)
	used := lipgloss.Height(top) + lipgloss.Height(bottom)
	body = lipgloss.NewStyle().Width(width).Height(height - used).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, top, body, bottom)
}

func progressBar(pct, width int) string {
	filled := pct * width / 100
	return lipgloss.NewStyle().Foreground(accent).Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", width-filled))
}

// ── Folders V1: web-GUI-faithful expandable panels ──────────────────────────

func foldersV1Panels(width, height int) string {
	var panels []string
	for i, f := range fakeFolders {
		title := fmt.Sprintf(" %s %s", f.Label, dimStyle.Render("("+f.ID+")"))
		state := stateStyle(f.State).Render(f.State)
		head := title + strings.Repeat(" ", max(1, width-4-lipgloss.Width(title)-lipgloss.Width(state))) + state
		if i != 1 { // collapsed panels: single header line
			panels = append(panels, lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).BorderForeground(dim).
				Width(width-2).Render(head))
			continue
		}
		// expanded panel (selected)
		body := head + "\n" +
			dimStyle.Render(strings.Repeat("─", width-4)) + "\n" +
			fmt.Sprintf("  Folder Path      %s\n", f.Path) +
			fmt.Sprintf("  Global State     %s\n", f.Global) +
			fmt.Sprintf("  Local State      %s\n", f.Local) +
			fmt.Sprintf("  Shared With      %s\n", f.Shared) +
			fmt.Sprintf("  Sync Progress    %s %d%%\n", progressBar(f.Pct, 30), f.Pct) +
			"\n  " + buttonRow([]string{"Pause", "Versions", "Rescan", "Edit"}, 0)
		panels = append(panels, lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).BorderForeground(accent).
			Width(width-2).Render(body))
	}
	for _, action := range []string{"+ Add Folder", "⟳ Rescan All"} {
		panels = append(panels, lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).BorderForeground(dim).
			Width(width-2).Render(" "+action))
	}
	body := strings.Join(panels[:len(panels)-2], "\n") + "\n\n" +
		strings.Join(panels[len(panels)-2:], "\n")
	keys := keyHint("↑↓", "select", "enter", "expand", "p", "pause", "q", "quit")
	return frame("Folders", body, keys, width, height)
}

// ── Folders V2: dense table + detail pane for selection ─────────────────────

func foldersV2Table(width, height int) string {
	var b strings.Builder
	b.WriteString(dimStyle.Render(fmt.Sprintf("    %-12s %-13s %-16s %-12s %s", "LABEL", "ID", "PATH", "STATE", "SYNC")) + "\n")
	for i, f := range fakeFolders {
		cursor := "  "
		line := fmt.Sprintf("%-12s %-13s %-16s %-12s %3d%%", f.Label, f.ID, f.Path, f.State, f.Pct)
		if i == 1 {
			cursor = keyStyle.Render("▸ ")
			line = lipgloss.NewStyle().Bold(true).Render(line)
		}
		b.WriteString(cursor + stateStyle(f.State).Render("●") + " " + line + "\n")
	}
	b.WriteString("\n  " + keyStyle.Render("+") + " Add Folder\n  " + keyStyle.Render("⟳") + " Rescan All\n")
	sel := fakeFolders[1]
	detail := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(dim).Width(width).Render(
		lipgloss.NewStyle().Bold(true).Render(" "+sel.Label) + dimStyle.Render(" ("+sel.ID+")") + "\n" +
			fmt.Sprintf("  Path %-18s Global %-10s Local %s\n", sel.Path, sel.Global, sel.Local) +
			fmt.Sprintf("  Shared With: %s\n", sel.Shared) +
			fmt.Sprintf("  %s %d%%\n", progressBar(sel.Pct, width-10), sel.Pct) +
			"  " + buttonRow([]string{"Pause", "Versions", "Rescan", "Edit"}, 0))
	body := b.String() + "\n" + detail
	keys := keyHint("↑↓", "select", "p", "pause", "e", "edit", "q", "quit")
	return frame("Folders", body, keys, width, height)
}

// ── Folders V3: split pane, list left / detail right ────────────────────────

func foldersV3Split(width, height int) string {
	leftW := 26
	var list strings.Builder
	for i, f := range fakeFolders {
		cursor, style := "  ", lipgloss.NewStyle()
		if i == 1 {
			cursor, style = keyStyle.Render("▸ "), lipgloss.NewStyle().Bold(true)
		}
		list.WriteString(cursor + stateStyle(f.State).Render("●") + " " + style.Render(f.Label) + "\n")
	}
	list.WriteString("\n  " + keyStyle.Render("+") + " Add Folder\n  " + keyStyle.Render("⟳") + " Rescan All\n")
	sel := fakeFolders[1]
	detail := lipgloss.NewStyle().Bold(true).Render(sel.Label) + dimStyle.Render(" ("+sel.ID+")") + "\n\n" +
		fmt.Sprintf("Folder Path      %s\n", sel.Path) +
		fmt.Sprintf("State            %s\n", stateStyle(sel.State).Render(sel.State)) +
		fmt.Sprintf("Global State     %s\n", sel.Global) +
		fmt.Sprintf("Local State      %s\n", sel.Local) +
		fmt.Sprintf("Shared With      %s\n", sel.Shared) + "\n" +
		progressBar(sel.Pct, width-leftW-8) + fmt.Sprintf(" %d%%\n\n", sel.Pct) +
		buttonRow([]string{"Pause", "Versions", "Rescan", "Edit"}, 0)
	innerH := height - 5
	left := lipgloss.NewStyle().Width(leftW).Height(innerH).
		Border(lipgloss.NormalBorder(), false, true, false, false).BorderForeground(dim).
		Render(list.String())
	right := lipgloss.NewStyle().Padding(0, 1).Width(width - leftW - 1).Render(detail)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	keys := keyHint("↑↓", "select", "tab", "focus detail", "q", "quit")
	return frame("Folders", body, keys, width, height)
}

// ── Alerts V1: scrollable list of alert cards ────────────────────────────────

func alertsV1List(width, height int) string {
	var cards []string
	for i, a := range fakeAlerts {
		border, style := lipgloss.RoundedBorder(), lipgloss.NewStyle().Foreground(warn)
		head := style.Bold(true).Render(" ⚠ "+a.Title) +
			strings.Repeat(" ", max(1, width-6-lipgloss.Width(" ⚠ "+a.Title)-len(a.Time))) +
			dimStyle.Render(a.Time)
		body := head + "\n" +
			lipgloss.NewStyle().Width(width-6).Padding(0, 1).Render(a.Body) + "\n" +
			" " + buttonRow(a.Buttons, map[bool]int{true: 0, false: -1}[i == 0])
		bs := dim
		if i == 0 {
			bs = warn
		}
		cards = append(cards, lipgloss.NewStyle().Border(border).BorderForeground(bs).
			Width(width-2).Render(body))
	}
	body := strings.Join(cards, "\n")
	keys := keyHint("↑↓", "select alert", "←→", "select action", "enter", "confirm", "q", "quit")
	return frame("Alerts", body, keys, width, height)
}

// ── Alerts V2: one alert at a time, full focus ───────────────────────────────

func alertsV2Focus(width, height int) string {
	a := fakeAlerts[0]
	card := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(warn).
		Width(width-10).Padding(1, 2).Render(
		lipgloss.NewStyle().Bold(true).Foreground(warn).Render("⚠ "+a.Title) +
			"   " + dimStyle.Render(a.Time) + "\n\n" +
			lipgloss.NewStyle().Width(width-16).Render(a.Body) + "\n\n" +
			buttonRow(a.Buttons, 0))
	counter := dimStyle.Render("alert 1 of 2  (n: next, N: prev)")
	body := lipgloss.Place(width, height-6, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, card, "", counter))
	keys := keyHint("←→", "select action", "enter", "confirm", "n/N", "next/prev alert", "q", "quit")
	return frame("Alerts", body, keys, width, height)
}

// ── Alerts V3: split pane like folders-v3 ────────────────────────────────────

func alertsV3Split(width, height int) string {
	leftW := 26
	warnStyle := lipgloss.NewStyle().Foreground(warn)
	var list strings.Builder
	for i, a := range fakeAlerts {
		cursor, style := "  ", lipgloss.NewStyle()
		if i == 0 {
			cursor, style = keyStyle.Render("▸ "), lipgloss.NewStyle().Bold(true)
		}
		kind := map[string]string{"device": "Device", "folder": "Folder"}[a.Kind]
		list.WriteString(cursor + warnStyle.Render("⚠") + " " + style.Render(kind+" ("+a.Short+")") + "\n")
	}
	sel := fakeAlerts[0]
	detail := warnStyle.Bold(true).Render("⚠ "+sel.Title) + "\n" +
		dimStyle.Render(sel.Time) + "\n\n" +
		lipgloss.NewStyle().Width(width-leftW-4).Render(sel.Body) + "\n\n" +
		buttonRow(sel.Buttons, 0)
	innerH := height - 5
	left := lipgloss.NewStyle().Width(leftW).Height(innerH).
		Border(lipgloss.NormalBorder(), false, true, false, false).BorderForeground(dim).
		Render(list.String())
	right := lipgloss.NewStyle().Padding(0, 1).Width(width - leftW - 1).Render(detail)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	keys := keyHint("↑↓", "select alert", "←→", "select action", "enter", "confirm", "q", "quit")
	return frame("Alerts", body, keys, width, height)
}

// ── Remote Devices: split pane (baseline, mirrors folders-v3) ────────────────

func devicesSplit(width, height int) string {
	leftW := 26
	var list strings.Builder
	for i, d := range fakeDevices {
		cursor, style := "  ", lipgloss.NewStyle()
		if i == 0 {
			cursor, style = keyStyle.Render("▸ "), lipgloss.NewStyle().Bold(true)
		}
		dot := stateStyle(d.State).Render("●")
		list.WriteString(cursor + dot + " " + style.Render(d.Name) + "\n")
	}
	list.WriteString("\n  " + keyStyle.Render("+") + " Add Device\n  " + keyStyle.Render("≡") + " Recent Changes\n")
	sel := fakeDevices[0]
	detail := lipgloss.NewStyle().Bold(true).Render(sel.Name) + "\n\n" +
		fmt.Sprintf("State            %s\n", stateStyle(sel.State).Render(sel.State)) +
		fmt.Sprintf("Address          %s\n", sel.Address) +
		fmt.Sprintf("Download Rate    %s\n", sel.Download) +
		fmt.Sprintf("Upload Rate      %s\n", sel.Upload) +
		"Compression      Metadata Only\n" +
		"Introducer       No\n" +
		"Last Seen        2026-07-12 10:41\n" +
		"Version          v2.0.13\n" +
		"Folders          Documents, Photos, Music\n\n" +
		progressBar(sel.Pct, width-leftW-8) + fmt.Sprintf(" %d%%\n\n", sel.Pct) +
		buttonRow([]string{"Pause", "Edit"}, 0)
	innerH := height - 5
	left := lipgloss.NewStyle().Width(leftW).Height(innerH).
		Border(lipgloss.NormalBorder(), false, true, false, false).BorderForeground(dim).
		Render(list.String())
	right := lipgloss.NewStyle().Padding(0, 1).Width(width - leftW - 1).Render(detail)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	keys := keyHint("↑↓", "select", "tab", "focus detail", "q", "quit")
	return frame("Remote Devices", body, keys, width, height)
}

// ── This Device V1: aligned key/value status list ────────────────────────────

var thisDeviceStats = [][2]string{
	{"Download Rate", "1.2 MiB/s (4.3 GiB total)"},
	{"Upload Rate", "2.4 MiB/s (9.1 GiB total)"},
	{"Local State (Total)", "8,412 files, 84.6 GiB"},
	{"Listeners", "3/3"},
	{"Discovery", "4/5"},
	{"Uptime", "3d 2h 41m"},
	{"Identification", "MFZWI3D (this-machine)"},
	{"Version", "v2.0.13, Linux (64-bit)"},
}

func thisDeviceV1KV(width, height int) string {
	var b strings.Builder
	b.WriteString("\n")
	for _, kv := range thisDeviceStats {
		b.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render(fmt.Sprintf("%-21s", kv[0])), kv[1]))
	}
	keys := keyHint("1-5", "switch tab", "q", "quit")
	return frame("This Device", b.String(), keys, width, height)
}

// ── This Device V2: two-column grid of boxed stats ───────────────────────────

func thisDeviceV2Grid(width, height int) string {
	boxW := (width - 4) / 2
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).
		Width(boxW).Padding(0, 1)
	var rows []string
	for i := 0; i < len(thisDeviceStats); i += 2 {
		l := box.Render(dimStyle.Render(thisDeviceStats[i][0]) + "\n" + thisDeviceStats[i][1])
		r := box.Render(dimStyle.Render(thisDeviceStats[i+1][0]) + "\n" + thisDeviceStats[i+1][1])
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, l, r))
	}
	keys := keyHint("1-5", "switch tab", "q", "quit")
	return frame("This Device", strings.Join(rows, "\n"), keys, width, height)
}

// ── Actions V1: plain menu list ──────────────────────────────────────────────

var actionItems = []string{"Settings", "Advanced", "Show ID", "Logs", "Support Bundle", "Log Out", "Restart", "Shut Down"}

func actionsV1List(width, height int) string {
	var b strings.Builder
	b.WriteString("\n")
	for i, a := range actionItems {
		cursor, style := "  ", lipgloss.NewStyle()
		if i == 0 {
			cursor, style = keyStyle.Render("▸ "), lipgloss.NewStyle().Bold(true)
		}
		b.WriteString("  " + cursor + style.Render(a) + "\n")
	}
	keys := keyHint("↑↓", "select", "enter", "open", "q", "quit")
	return frame("Actions", b.String(), keys, width, height)
}

// ── Actions V2: centered menu box ────────────────────────────────────────────

func actionsV2Centered(width, height int) string {
	var b strings.Builder
	for i, a := range actionItems {
		cursor, style := "  ", lipgloss.NewStyle()
		if i == 0 {
			cursor, style = keyStyle.Render("▸ "), lipgloss.NewStyle().Bold(true)
		}
		b.WriteString(cursor + style.Render(fmt.Sprintf("%-16s", a)) + "\n")
	}
	menu := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).
		Padding(1, 3).Render(strings.TrimRight(b.String(), "\n"))
	body := lipgloss.Place(width, height-6, lipgloss.Center, lipgloss.Center, menu)
	keys := keyHint("↑↓", "select", "enter", "open", "q", "quit")
	return frame("Actions", body, keys, width, height)
}

// ── Subview building blocks ──────────────────────────────────────────────────

// subTabRow renders sub-tabs in the same boxed style as the main tab bar,
// with superscript number hotkeys.
func subTabRow(names []string, active int) string {
	sups := []string{"¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹"}
	inactive := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.HiddenBorder())
	activeSt := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.NormalBorder()).BorderForeground(accent)
	tabs := make([]string, len(names))
	for i, n := range names {
		style := inactive
		if i == active {
			style = activeSt
		}
		tabs[i] = style.Render(n + dimStyle.Render(sups[i]))
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
}

// formField renders a labeled input with dim description below. Readonly
// fields are greyed out entirely (label, value, no input affordance), matching
// the web GUI's editable/non-editable distinction.
// NOTE: placeholder styling — user wants the ▕value▏ look revisited later.
func formField(label, value, desc string, focused, readonly bool) string {
	if readonly {
		return dimStyle.Render(fmt.Sprintf("  %-14s   %-40s", label, value)) + "\n" +
			"                 " + dimStyle.Render(desc) + "\n\n"
	}
	box := dimStyle
	if focused {
		box = lipgloss.NewStyle().Foreground(accent)
	}
	return fmt.Sprintf("  %-14s %s\n", label, box.Render("▕ "+fmt.Sprintf("%-40s", value)+"▏")) +
		"                 " + dimStyle.Render(desc) + "\n\n"
}

// checkField renders a checkbox field with dim description below.
func checkField(label string, checked bool, desc string, focused bool) string {
	mark := "[ ]"
	if checked {
		mark = "[x]"
	}
	style := lipgloss.NewStyle()
	if focused {
		style = lipgloss.NewStyle().Foreground(accent)
	}
	return fmt.Sprintf("  %-14s %s\n", label, style.Render(mark)) +
		"                 " + dimStyle.Render(desc) + "\n\n"
}

// infobarFrame is the chosen subview chrome: same-height info bar with the
// title on the left and a boxed Back button (esc) on the right.
func infobarFrame(title, sub, body, keys string, width, height int) string {
	titleSt := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.HiddenBorder()).Bold(true)
	backSt := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.NormalBorder()).BorderForeground(accent)
	left := titleSt.Render(title + " " + dimStyle.Render("("+sub+")"))
	back := backSt.Render("Back" + dimStyle.Render("ᵉˢᶜ"))
	gap := width - lipgloss.Width(left) - lipgloss.Width(back)
	top := lipgloss.JoinHorizontal(lipgloss.Bottom, left, strings.Repeat(" ", max(0, gap)), back)
	bottom := statusBar(keys, "↓1.2MiB/s ↑2.4MiB/s", width)
	used := lipgloss.Height(top) + lipgloss.Height(bottom)
	body = lipgloss.NewStyle().Width(width).Height(height - used).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, top, body, bottom)
}

var subviewKeys = keyHint("↑↓", "field", "tab/⇧tab", "next/prev section", "enter", "edit", "esc", "back")

// selectField renders a select input (dropdown affordance ▾).
func selectField(label, value, desc string, focused bool) string {
	return formField(label, value+" ▾", desc, focused, false)
}

// ── Edit Device subview ──────────────────────────────────────────────────────

func editDeviceInfobar(width, height int) string {
	body := subTabRow([]string{"General", "Sharing", "Advanced"}, 0) + "\n" +
		formField("Device ID", "MFZWI3D-BONSGYC-...", "Find it under Actions > Show ID on the other device", false, true) +
		formField("Device Name", "workpc", "Shown instead of the ID; advertised to other devices", true, false) +
		"\n  " + buttonRow([]string{"Save", "Close", "Remove"}, -1)
	return infobarFrame("Edit Device", "workpc", body, subviewKeys, width, height)
}

// ── Settings subview ─────────────────────────────────────────────────────────

func settingsInfobar(width, height int) string {
	body := subTabRow([]string{"General", "GUI", "Connections"}, 0) + "\n" +
		formField("Device Name", "this-machine", "The name of this device shown to other devices", true, false) +
		formField("Min Free Space", "1 %", "Free space required on the home disk", false, false) +
		formField("API Key", "abcDEF123...", "Key for API access — enter to regenerate", false, true) +
		formField("Usage Report", "Disabled", "Send anonymous usage statistics", false, false) +
		formField("Upgrades", "Stable Releases Only", "Automatic upgrade policy", false, false) +
		"\n  " + buttonRow([]string{"Edit Folder Defaults", "Edit Device Defaults"}, -1) +
		"\n\n  " + buttonRow([]string{"Save", "Close"}, -1)
	return infobarFrame("Settings", "General", body, subviewKeys, width, height)
}

// ── Remaining sub-tab contents and small screens ─────────────────────────────

var editDeviceTabs = []string{"General", "Sharing", "Advanced"}
var editFolderTabs = []string{"General", "Versioning", "Ignore Patterns", "Advanced"}
var settingsTabs = []string{"General", "GUI", "Connections"}

func editDeviceSharing(width, height int) string {
	body := subTabRow(editDeviceTabs, 1) + "\n" +
		checkField("Introducer", false, "Add devices from the introducer to our device list", true) +
		checkField("Auto Accept", true, "Automatically create or share advertised folders at the default path", false) +
		"\n  " + buttonRow([]string{"Save", "Close", "Remove"}, -1)
	return infobarFrame("Edit Device", "workpc", body, subviewKeys, width, height)
}

func editDeviceAdvanced(width, height int) string {
	body := subTabRow(editDeviceTabs, 2) + "\n" +
		formField("Addresses", "dynamic", `Comma separated addresses or "dynamic"`, true, false) +
		selectField("Compression", "Metadata Only", "What data to compress when syncing", false) +
		formField("Connections", "0", "Concurrent connections, zero lets Syncthing decide", false, false) +
		formField("In Rate Limit", "0 KiB/s", "Maximum incoming rate, zero for no limit", false, false) +
		formField("Out Rate Limit", "0 KiB/s", "Maximum outgoing rate, zero for no limit", false, false) +
		checkField("Untrusted", false, "Shared folders must be protected by a password", false) +
		"\n  " + buttonRow([]string{"Save", "Close", "Remove"}, -1)
	return infobarFrame("Edit Device", "workpc", body, subviewKeys, width, height)
}

func settingsGUI(width, height int) string {
	body := subTabRow(settingsTabs, 1) + "\n" +
		formField("Listen Address", "127.0.0.1:8384", "Address and port for the GUI (non-privileged 1024-65535)", true, false) +
		formField("Auth User", "evan", "Username for GUI access", false, false) +
		formField("Auth Password", "••••••••", "Password for GUI access", false, false) +
		checkField("Use HTTPS", true, "Enable HTTPS for secure GUI access", false) +
		checkField("Start Browser", false, "Open browser automatically when Syncthing starts", false) +
		selectField("GUI Theme", "Dark", "Visual theme for the web GUI", false) +
		"\n  " + buttonRow([]string{"Save", "Close"}, -1)
	return infobarFrame("Settings", "GUI", body, subviewKeys, width, height)
}

func settingsConnections(width, height int) string {
	body := subTabRow(settingsTabs, 2) + "\n" +
		formField("Listen Addresses", "default", "Sync protocol addresses, comma separated", true, false) +
		formField("In Rate Limit", "0 KiB/s", "Global maximum incoming rate, zero for no limit", false, false) +
		formField("Out Rate Limit", "0 KiB/s", "Global maximum outgoing rate, zero for no limit", false, false) +
		checkField("LAN Rate Limit", false, "Apply rate limits to LAN connections as well", false) +
		checkField("NAT Traversal", true, "Attempt NAT traversal for incoming connections", false) +
		checkField("Local Discovery", true, "Discover devices on the local network", false) +
		checkField("Global Discovery", true, "Use global discovery servers to find devices", false) +
		checkField("Enable Relaying", true, "Allow relay connections when direct fails", false) +
		formField("Discovery Servers", "default", "Comma separated global discovery servers", false, false) +
		"\n  " + buttonRow([]string{"Save", "Close"}, -1)
	return infobarFrame("Settings", "Connections", body, subviewKeys, width, height)
}

func editFolderVersioning(width, height int) string {
	body := subTabRow(editFolderTabs, 1) + "\n" +
		selectField("Versioning", "Simple File Versioning", "How to handle old versions of files", true) +
		formField("Keep Versions", "5", "Old versions to keep per file", false, false) +
		formField("Clean Out After", "0 days", "Days to keep trashed files, zero means forever", false, false) +
		formField("Versions Path", "", "Empty for default .stversions directory", false, false) +
		formField("Cleanup Interval", "3600 seconds", "Cleanup run interval, zero to disable", false, false) +
		"\n  " + buttonRow([]string{"Save", "Close", "Remove"}, -1)
	return infobarFrame("Edit Folder", "foobar", body, subviewKeys, width, height)
}

func editFolderIgnores(width, height int) string {
	patterns := "// One pattern per line\n!important.txt\n*.tmp\n(?d).DS_Store\n#include more-patterns.txt"
	area := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).
		Width(width-6).Height(8).Padding(0, 1).Render(patterns)
	body := subTabRow(editFolderTabs, 2) + "\n" +
		"  Patterns " + dimStyle.Render("(one per line)") + "\n" +
		lipgloss.NewStyle().PaddingLeft(2).Render(area) + "\n\n" +
		"  " + buttonRow([]string{"Save", "Close", "Remove"}, -1)
	return infobarFrame("Edit Folder", "foobar", body, subviewKeys, width, height)
}

func editFolderAdvanced(width, height int) string {
	body := subTabRow(editFolderTabs, 3) + "\n" +
		checkField("Watch Changes", true, "Use filesystem notifications to detect changed items", true) +
		formField("Rescan Interval", "3600 seconds", "How often to run a full rescan", false, false) +
		selectField("Folder Type", "Send & Receive", "How this folder syncs with other devices", false) +
		selectField("Pull Order", "Random", "Order in which to download files", false) +
		formField("Min Free Space", "1 %", "Required free space before syncing stops", false, false) +
		checkField("Ignore Perms", false, "Disable comparing and syncing file permissions", false) +
		"\n  " + buttonRow([]string{"Save", "Close", "Remove"}, -1)
	return infobarFrame("Edit Folder", "foobar", body, subviewKeys, width, height)
}

// fakeQR renders a placeholder QR (finder patterns + noise) with half-block
// characters, 2 modules per row. The real TUI will render an actual QR the
// same way (e.g. skip2/go-qrcode).
func fakeQR(n int) string {
	mod := func(x, y int) bool {
		finder := func(cx, cy int) bool {
			dx, dy := x-cx, y-cy
			if dx < 0 || dy < 0 || dx > 6 || dy > 6 {
				return false
			}
			edge := dx == 0 || dy == 0 || dx == 6 || dy == 6
			core := dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4
			return edge || core
		}
		if finder(0, 0) || finder(n-7, 0) || finder(0, n-7) {
			return true
		}
		return (x*7+y*13+x*y)%5 < 2 // deterministic noise
	}
	var b strings.Builder
	for y := 0; y < n; y += 2 {
		for x := 0; x < n; x++ {
			top, bot := mod(x, y), y+1 < n && mod(x, y+1)
			b.WriteString(map[[2]bool]string{
				{true, true}: "█", {true, false}: "▀",
				{false, true}: "▄", {false, false}: " ",
			}[[2]bool{top, bot}])
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func showID(width, height int) string {
	id := "MFZWI3D-BONSGYC-YLTMRWG-C43ENR5-QXGZDMM-FZWI3DP-BONSGYC-YLTMRW4"
	idBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).
		Padding(1, 3).Bold(true).Render(id[:31] + "\n" + id[32:])
	body := lipgloss.Place(width, height-8, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			"Share this ID with other devices to connect.", "", idBox, "",
			fakeQR(25), "",
			buttonRow([]string{"Copy", "Close"}, 0)))
	return infobarFrame("Show ID", "this-machine", body,
		keyHint("←→", "select", "enter", "confirm", "esc", "back"), width, height)
}

func aboutPaths(width, height int) string {
	paths := [][2]string{
		{"User Home", "/home/evan"},
		{"Configuration Directory", "/home/evan/.config/syncthing"},
		{"Configuration File", "/home/evan/.config/syncthing/config.xml"},
		{"Device Certificate", "/home/evan/.config/syncthing/cert.pem"},
		{"GUI / API HTTPS Certificate", "/home/evan/.config/syncthing/https-cert.pem"},
		{"Database Location", "/home/evan/.local/share/syncthing"},
		{"Log File", "-"},
		{"GUI Override Directory", "/home/evan/.config/syncthing/gui"},
	}
	var b strings.Builder
	b.WriteString("\n")
	for _, kv := range paths {
		b.WriteString("  " + dimStyle.Render(fmt.Sprintf("%-29s", kv[0])) + kv[1] + "\n")
	}
	return infobarFrame("About", "Paths", b.String(),
		keyHint("esc", "back"), width, height)
}

// ── Edit Folder subview ──────────────────────────────────────────────────────

// editFolderForm renders the shared form body: sub-tab row, General fields,
// bottom buttons. First field focused.
func editFolderForm(width int) string {
	return subTabRow([]string{"General", "Versioning", "Ignore Patterns", "Advanced"}, 0) + "\n" +
		formField("Folder Label", "foobar", "Optional label, can differ per device", true, false) +
		formField("Folder ID", "ab7cd-ef8gh", "Same on all cluster devices", false, true) +
		formField("Folder Path", "~/foobar", "Created if it does not exist", false, false) +
		"\n  " + buttonRow([]string{"Save", "Close", "Remove"}, -1)
}

// editFolderV1Infobar: chosen subview chrome (see infobarFrame).
func editFolderV1Infobar(width, height int) string {
	return infobarFrame("Edit Folder", "foobar", editFolderForm(width), subviewKeys, width, height)
}

// editFolderV2Modal: centered bordered modal (main view assumed behind).
func editFolderV2Modal(width, height int) string {
	inner := lipgloss.NewStyle().Bold(true).Render("Edit Folder ") + dimStyle.Render("(foobar)") + "\n\n" +
		editFolderForm(width-12)
	modal := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).
		Padding(1, 2).Width(width - 8).Render(inner)
	body := lipgloss.Place(width, height-2, lipgloss.Center, lipgloss.Center, modal)
	bottom := statusBar(keyHint("↑↓", "field", "tab", "next section", "enter", "edit", "esc", "close"), "↓1.2MiB/s ↑2.4MiB/s", width)
	return lipgloss.JoinVertical(lipgloss.Left, body, bottom)
}

// buttonRow renders buttons; index selected is highlighted (-1 for none).
func buttonRow(labels []string, selected int) string {
	parts := make([]string, len(labels))
	for i, l := range labels {
		style := lipgloss.NewStyle().Foreground(dim)
		if i == selected {
			style = lipgloss.NewStyle().Reverse(true).Bold(true)
		}
		parts[i] = style.Render("[ " + l + " ]")
	}
	return strings.Join(parts, " ")
}

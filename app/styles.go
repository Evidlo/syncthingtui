package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	accent   = lipgloss.Color("4") // bootstrap "primary"
	dim      = lipgloss.Color("8") // bootstrap "default" (muted)
	good     = lipgloss.Color("2") // bootstrap "success"
	warn     = lipgloss.Color("3") // bootstrap "warning"
	bad      = lipgloss.Color("1") // bootstrap "danger"
	info     = lipgloss.Color("6") // bootstrap "info"
	dimStyle = lipgloss.NewStyle().Foreground(dim)
	keyStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	boldSt   = lipgloss.NewStyle().Bold(true)
	warnSt   = lipgloss.NewStyle().Foreground(warn)
)

// stateStyle colors a status dot to match the web GUI's folderClass/deviceClass
// (bootstrap color → our palette). See folderState (live.go) for the labels.
func stateStyle(state string) lipgloss.Style {
	switch {
	// success (green): folder idle / local additions; device up-to-date or
	// connected-but-unused (a live connection is still green)
	case state == "Up to Date", state == "Local Additions",
		strings.HasPrefix(state, "Connected"):
		return lipgloss.NewStyle().Foreground(good)
	// primary (blue): busy
	case strings.HasPrefix(state, "Syncing"), state == "Preparing to Sync",
		state == "Scanning", state == "Cleaning Versions":
		return lipgloss.NewStyle().Foreground(accent)
	// danger (red)
	case state == "Stopped", state == "Error", state == "Out of Sync",
		state == "Failed Items", state == "Local Data Unencrypted":
		return lipgloss.NewStyle().Foreground(bad)
	// warning (yellow): unshared + the waiting states
	case state == "Unshared", state == "Waiting to Sync",
		state == "Waiting to Scan", state == "Waiting to Clean":
		return warnSt
	// default (muted): paused
	case strings.HasPrefix(state, "Paused"):
		return dimStyle
	// info (cyan): unknown + every disconnected variant
	case state == "Unknown", strings.HasPrefix(state, "Disconnected"):
		return lipgloss.NewStyle().Foreground(info)
	}
	return lipgloss.NewStyle().Foreground(info)
}

var superscripts = []string{"¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹"}

// boxedTabs renders a tab row: boxed active tab, hidden borders otherwise,
// dim superscript hotkeys.
func boxedTabs(labels []string, active int) string {
	inactive := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.HiddenBorder())
	activeSt := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.NormalBorder()).BorderForeground(accent)
	tabs := make([]string, len(labels))
	for i, l := range labels {
		style := inactive
		if i == active {
			style = activeSt
		}
		tabs[i] = style.Render(l + dimStyle.Render(superscripts[i]))
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
}

// statusBar right-aligns `right` after the key hints; when both don't fit,
// the hints win and `right` is dropped.
func statusBar(keys, right string, width int) string {
	gap := width - lipgloss.Width(keys) - lipgloss.Width(right)
	if gap < 2 {
		right, gap = "", 0
	}
	return dimStyle.Render(strings.Repeat("─", width)) + "\n" +
		lipgloss.NewStyle().MaxWidth(width).Render(keys+strings.Repeat(" ", gap)+right)
}

func keyHint(pairs ...string) string {
	parts := make([]string, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, keyStyle.Render(pairs[i])+dimStyle.Render(":"+pairs[i+1]))
	}
	return strings.Join(parts, "  ")
}

func progressBar(pct, width int) string {
	if width < 1 {
		return ""
	}
	filled := pct * width / 100
	return lipgloss.NewStyle().Foreground(accent).Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", width-filled))
}

// infobarTop renders the subview header: bold title left, boxed Back button
// (esc) right — same height as the main tab bar.
func infobarTop(title, sub string, width int) string {
	titleSt := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.HiddenBorder()).Bold(true)
	backSt := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.NormalBorder()).BorderForeground(accent)
	left := titleSt.Render(title + " " + dimStyle.Render("("+sub+")"))
	back := backSt.Render("Back" + dimStyle.Render("ᵉˢᶜ"))
	gap := max(0, width-lipgloss.Width(left)-lipgloss.Width(back))
	return lipgloss.JoinHorizontal(lipgloss.Bottom, left, strings.Repeat(" ", gap), back)
}

// page pads/clips body between a top row and status bar to fill the window.
func page(top, body, keys, netRates string, width, height int) string {
	bottom := statusBar(keys, netRates, width)
	used := lipgloss.Height(top) + lipgloss.Height(bottom)
	body = lipgloss.NewStyle().Width(width).Height(max(0, height-used)).
		MaxWidth(width).MaxHeight(max(1, height-used)).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, top, body, bottom)
}

const netRates = "↓1.2MiB/s ↑2.4MiB/s"

func buttonRow(labels []string, selected int) string {
	parts := make([]string, len(labels))
	for i, l := range labels {
		style := dimStyle
		if i == selected {
			style = lipgloss.NewStyle().Reverse(true).Bold(true)
		}
		parts[i] = style.Render("[ " + l + " ]")
	}
	return strings.Join(parts, " ")
}

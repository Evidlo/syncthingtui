package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	accent   = lipgloss.Color("4")
	dim      = lipgloss.Color("8")
	good     = lipgloss.Color("2")
	warn     = lipgloss.Color("3")
	bad      = lipgloss.Color("1")
	dimStyle = lipgloss.NewStyle().Foreground(dim)
	keyStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	boldSt   = lipgloss.NewStyle().Bold(true)
	warnSt   = lipgloss.NewStyle().Foreground(warn)
)

func stateStyle(state string) lipgloss.Style {
	switch {
	case state == "Up to Date":
		return lipgloss.NewStyle().Foreground(good)
	case strings.HasPrefix(state, "Connected"):
		// connected but idle/unused is still a live connection → green
		return lipgloss.NewStyle().Foreground(good)
	case strings.HasPrefix(state, "Syncing"):
		return lipgloss.NewStyle().Foreground(accent)
	case state == "Out of Sync":
		return lipgloss.NewStyle().Foreground(bad)
	}
	return dimStyle
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

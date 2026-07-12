package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Show ID ──────────────────────────────────────────────────────────────────

type ShowIDView struct {
	btn           int
	width, height int
	flash         string
}

func NewShowID() ShowIDView { return ShowIDView{} }

func (v ShowIDView) Init() tea.Cmd { return nil }

func (v ShowIDView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width, v.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			v.btn = 0
		case "right", "l":
			v.btn = 1
		case "enter":
			if v.btn == 0 {
				v.flash = "copied to clipboard (mockup)"
			} else {
				return v, func() tea.Msg { return closeMsg{} }
			}
		}
	}
	return v, nil
}

// fakeQR renders a placeholder QR with half-block characters; the real app
// will render an actual QR (e.g. skip2/go-qrcode) the same way.
func fakeQR(n int) string {
	mod := func(x, y int) bool {
		finder := func(cx, cy int) bool {
			dx, dy := x-cx, y-cy
			if dx < 0 || dy < 0 || dx > 6 || dy > 6 {
				return false
			}
			return dx == 0 || dy == 0 || dx == 6 || dy == 6 ||
				(dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4)
		}
		if finder(0, 0) || finder(n-7, 0) || finder(0, n-7) {
			return true
		}
		return (x*7+y*13+x*y)%5 < 2
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

func (v ShowIDView) View() string {
	idBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).
		Padding(1, 3).Bold(true).Render(DeviceID[:31] + "\n" + DeviceID[32:])
	body := lipgloss.Place(v.width, max(1, v.height-5), lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			"Share this ID with other devices to connect.", "", idBox, "",
			fakeQR(25), "",
			buttonRow([]string{"Copy", "Close"}, v.btn)))
	right := netRates
	if v.flash != "" {
		right = warnSt.Render(v.flash)
	}
	return page(infobarTop("Show ID", "this-machine", v.width), body,
		keyHint("←→", "select", "enter", "confirm", "esc", "back"), right, v.width, v.height)
}

// ── About ────────────────────────────────────────────────────────────────────

// TUIVersion is this program's version, shown at the top of About.
const TUIVersion = "0.1.0-dev"

type AboutView struct {
	stVersion     string
	paths         [][2]string
	width, height int
}

// NewAbout shows version info (syncthingtui + syncthing build) and paths.
// Authors/Included Software from the web GUI are intentionally left out.
func NewAbout(stVersion string, paths [][2]string) AboutView {
	return AboutView{stVersion: stVersion, paths: paths}
}

func (v AboutView) Init() tea.Cmd { return nil }

func (v AboutView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		v.width, v.height = msg.Width, msg.Height
	}
	return v, nil
}

func (v AboutView) View() string {
	kw := 0
	for _, kv := range v.paths {
		kw = max(kw, len(kv[0]))
	}
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  " + boldSt.Render("syncthingtui") + "  " + TUIVersion + "\n")
	b.WriteString("  " + boldSt.Render("syncthing") + "     " + v.stVersion + "\n\n")
	b.WriteString("  " + dimStyle.Render("Paths") + "\n")
	if len(v.paths) == 0 {
		b.WriteString("  " + dimStyle.Render("(not available on this syncthing version)") + "\n")
	}
	for _, kv := range v.paths {
		b.WriteString("  " + dimStyle.Render(fmt.Sprintf("%-*s ", kw, kv[0])) + kv[1] + "\n")
	}
	return page(infobarTop("About", "syncthingtui", v.width), b.String(),
		keyHint("esc", "back"), netRates, v.width, v.height)
}

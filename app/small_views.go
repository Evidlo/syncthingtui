package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	qrcode "github.com/skip2/go-qrcode"
)

// ── Show ID ──────────────────────────────────────────────────────────────────

type ShowIDView struct {
	id            string
	btn           int
	width, height int
	flash         string
}

func NewShowID(id string) ShowIDView { return ShowIDView{id: id} }

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

// qrBlock renders text as a QR code using half-block characters (2 modules
// per terminal row).
func qrBlock(text string) string {
	q, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return ""
	}
	grid := q.Bitmap() // includes a 4-module quiet-zone border
	// keep 2 of the 4 border modules: 1 half-block row of margin
	const trim = 2
	grid = grid[trim : len(grid)-trim]
	var b strings.Builder
	for y := 0; y < len(grid); y += 2 {
		for x := trim; x < len(grid[y])-trim; x++ {
			top, bot := grid[y][x], y+1 < len(grid) && grid[y+1][x]
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
	id := v.id
	if len(id) < 63 {
		id = DeviceID // defensive: myID not fetched yet
	}
	idBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).
		Padding(0, 2).Bold(true).Render(id)
	body := lipgloss.Place(v.width, max(1, v.height-5), lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			"Share this ID with other devices to connect.", "", idBox,
			qrBlock(id), // its own trimmed quiet zone = 1 line of margin
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
